package daemon

import (
	"context"
	"os"
	"os/exec"
	"sync"
)

type Process struct {
	mu   sync.RWMutex
	cmd  *exec.Cmd
	done chan struct{} // Closed when the process exits
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

	p.cmd = cmd
	p.done = make(chan struct{})

	// Start the explicit process watcher
	go p.watch(cmd)

	return nil
}

// watch waits for the process to exit and cleans up the state.
// It guarantees Wait() is only ever called once.
func (p *Process) watch(cmd *exec.Cmd) {
	_ = cmd.Wait() // Blocks until process exits

	p.mu.Lock()
	defer p.mu.Unlock()

	p.cmd = nil
	if p.done != nil {
		close(p.done)
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