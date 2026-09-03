package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Process struct {
	mu   sync.RWMutex
	cmd  *exec.Cmd
	done chan struct{} // Closed when the process exits
}

func isProcessAlive(pid int, proc *os.Process) bool {
	if runtime.GOOS == "windows" {
		// On Windows, FindProcess errors.ts if the process doesn't exist
		_, err := os.FindProcess(pid)
		return err == nil
	}
	// On Unix, FindProcess always succeeds, so we send Signal 0 to check if it's alive
	return proc.Signal(syscall.Signal(0)) == nil
}

func (p *Process) Start(cfg Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil {
		return ErrAlreadyRunning
	}

	if cfg.BinaryPath == "" {
		return ErrInvalidConfig
	}

	cmd := exec.Command(cfg.BinaryPath, cfg.Args...)

	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}
	if len(cfg.Env) > 0 {
		cmd.Env = append(os.Environ(), cfg.Env...)
	}
	if cfg.AttachOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if cfg.PIDFile != "" {
		pidStr := strconv.Itoa(cmd.Process.Pid)
		if err := os.WriteFile(cfg.PIDFile, []byte(pidStr), 0644); err != nil {
			// If we fail to write the PID file, kill the process to avoid unmanageable zombies
			_ = cmd.Process.Kill()
			return fmt.Errorf("failed to write PID file: %w", err)
		}
	}

	p.cmd = cmd
	p.done = make(chan struct{})

	// Start the explicit process watcher
	go p.watch(cmd, cfg.PIDFile)

	return nil
}

// watch waits for the process to exit and cleans up the state.
// It guarantees Wait() is only ever called once.
func (p *Process) watch(cmd *exec.Cmd, pidFile string) {
	_ = cmd.Wait() // Blocks until process exits

	p.mu.Lock()
	defer p.mu.Unlock()

	p.cmd = nil
	if p.done != nil {
		close(p.done)
	}

	// Process exited (either naturally or killed), clean up PID file
	if pidFile != "" {
		_ = os.Remove(pidFile)
	}
}

// StopCrossProcess attempts to stop the daemon even if it was started by a different CLI command
func (p *Process) StopCrossProcess(ctx context.Context, cfg Config) error {
	if cfg.PIDFile == "" {
		return fmt.Errorf("no PID file configured")
	}

	pidData, err := os.ReadFile(cfg.PIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotRunning
		}
		return err
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	osProc, err := os.FindProcess(pid)
	if err != nil {
		return ErrNotRunning
	}

	if !isProcessAlive(pid, osProc) {
		_ = os.Remove(cfg.PIDFile)
		return ErrNotRunning
	}

	// Request graceful shutdown
	if err := osProc.Signal(os.Interrupt); err != nil {
		_ = osProc.Kill()
	}

	// Poll to ensure it actually died by watching the PID file.
	// The watch() goroutine in the running daemon will delete it when the process exits.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Timeout reached. Force kill and cleanup manually.
			_ = osProc.Kill()
			_ = os.Remove(cfg.PIDFile)
			return ctx.Err()
		case <-ticker.C:
			// Check if the PID file still exists
			if !isProcessAlive(pid, osProc) {
				// Process is dead! We can safely delete the PID file and exit.
				_ = os.Remove(cfg.PIDFile)
				return nil
			}
		}
	}
}

func (p *Process) Stop(ctx context.Context) error {
	p.mu.RLock()
	cmd := p.cmd
	done := p.done
	p.mu.RUnlock()

	if cmd == nil || cmd.Process == nil {
		return ErrNotRunning
	}

	// Request graceful shutdown
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		// Fallback if signal fails (e.g., Windows sometimes struggles with Interrupt)
		_ = cmd.Process.Kill()
	}

	select {
	case <-done:
		// Process exited successfully via our watcher
		return nil

	case <-ctx.Done():
		// Timeout reached, force kill
		_ = cmd.Process.Kill()

		// Wait for the watcher to acknowledge the kill so we don't leave zombies
		<-done
		return ctx.Err()
	}
}

func (p *Process) Running() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cmd != nil
}