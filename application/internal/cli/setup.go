package cli

import (
	"context"
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/connection"
	"github.com/Muhammad-Jay/neuron/application/daemon"
	"github.com/Muhammad-Jay/neuron/application/runtime"
	"github.com/spf13/viper"
)

// setupNoreClient initializes the connection, ensures the daemon is running
// (if local), and returns the initialized client alongside a cleanup function.
func setupNoreClient(ctx context.Context) (*client.Client, func(), error) {
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

	// MAGIC: AttachOutput controls whether the daemon prints to the console!
	// If false, the daemon runs silently in the background.
	cfg.AttachOutput = isVerbose

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