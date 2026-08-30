package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/config"
	"github.com/Muhammad-Jay/neuron/application/connection"
	"github.com/Muhammad-Jay/neuron/application/daemon"
	"github.com/Muhammad-Jay/neuron/application/runtime"
)

// Options carries everything SetupClient needs. It is filled in by the CLI
// layer so bootstrap never has to read configuration or flags itself.
type Options struct {
	Config config.Config

	// Verbose controls whether a spawned local daemon attaches its output.
	Verbose bool
}

// SetupClient initializes the connection, ensures the daemon is running
// (if local), and returns the initialized client alongside a cleanup function.
func SetupClient(ctx context.Context, opts Options) (*client.Client, func(), error) {
	cfg := opts.Config

	var conn connection.Connection

	if cfg.Daemon.Endpoint != "" {
		// Use HTTP for remote execution (No daemon needed)
		conn = connection.New(connection.NewHTTPTransport(nil, cfg.Daemon.Endpoint))
	} else {
		// Use Unix Socket for local execution
		conn = connection.New(connection.NewLocal(cfg.Daemon.Socket))
	}

	dCfg := daemon.ConfigFromEffective(cfg)

	binaryPath, err := NoreBinaryPath(cfg)
	if err != nil {
		return nil, nil, err
	}

	// MAGIC: AttachOutput controls whether the daemon prints to the console!
	// If false, the daemon runs silently in the background.
	dCfg.AttachOutput = opts.Verbose
	dCfg.BinaryPath = binaryPath

	rtManager := runtime.NewManager(dCfg, conn)

	// Ensure N.O.R.E is running & accessible
	// If daemon.Endpoint is set, we simply ping the remote server and skip the daemon!
	if err := rtManager.Ensure(ctx, cfg.Daemon.Endpoint != ""); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("failed to ensure N.O.R.E runtime: %w", err)
	}

	// Create the Client SDK to return to the CLI command
	c := client.New(conn)

	cleanup := func() {
		_ = c.Close()
	}

	return c, cleanup, nil
}

// NoreBinaryPath resolves the nore binary, honoring an explicit config path
// first, then the PATH, then a local development fallback.
func NoreBinaryPath(cfg config.Config) (string, error) {
	// 1. Allow override via config (e.g., NEURON_NORE_PATH=/usr/bin/nore)
	if customPath := cfg.Daemon.NorePath; customPath != "" {
		return customPath, nil
	}

	// 2. Check if "nore" is installed globally in the system's PATH
	if path, err := exec.LookPath("nore"); err == nil {
		return path, nil
	}

	// 3. Fallback for local development (assumes running from the neuron repo)
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "../../../nore/cmd/nore/nore"), nil
}
