package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/connection"
	"github.com/Muhammad-Jay/neuron/application/daemon"
	"github.com/Muhammad-Jay/neuron/application/runtime"
	"github.com/spf13/viper"
)

// SetupClient initializes the connection, ensures the daemon is running
// (if local), and returns the initialized client alongside a cleanup function.
func SetupClient(ctx context.Context) (*client.Client, func(), error) {
	remoteURL := viper.GetString("remote")
	isRemote := remoteURL != ""
	isVerbose := viper.GetBool("verbose")

	var conn connection.Connection

	if isRemote {
		// Use HTTP for remote execution (No daemon needed)
		conn = connection.New(connection.NewHTTPTransport(nil, remoteURL))
	} else {
		// Use Unix Socket for local execution
		conn = connection.New(connection.NewLocal(client.DefaultSocketPath()))
	}

	cfg := daemon.DefaultConfig()
	binaryPath, err := NoreBinaryPath()
	if err != nil {
		return nil, nil, err
	}

	// MAGIC: AttachOutput controls whether the daemon prints to the console!
	// If false, the daemon runs silently in the background.
	cfg.AttachOutput = isVerbose
	cfg.BinaryPath = binaryPath

	rtManager := runtime.NewManager(cfg, conn)

	// Ensure N.O.R.E is running & accessible
	// If isRemote == true, this simply pings the remote server and skips the daemon!
	if err := rtManager.Ensure(ctx, isRemote); err != nil {
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

func NoreBinaryPath() (string, error) {
	// 1. Allow override via Viper/Env (e.g., NEURON_NORE_PATH=/usr/bin/nore)
	if customPath := viper.GetString("nore_path"); customPath != "" {
		return customPath, nil
	}

	// 2. Check if "nore" is installed globally in the systems's PATH
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
