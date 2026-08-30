package daemon

import (
	"context"
	"errors"
	"fmt"
	"time"

	noredaemon "github.com/Muhammad-Jay/neuron/application/daemon"
	"github.com/Muhammad-Jay/neuron/application/config"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Daemon,
		Short: "Manage the local N.O.R.E. daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newStopCmd())

	return cmd
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the background N.O.R.E. daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			cfg, ok := config.FromContext(cmd.Context())
			if !ok {
				return fmt.Errorf("configuration not loaded")
			}

			manager := noredaemon.NewManager(noredaemon.ConfigFromEffective(cfg), nil)

			fmt.Println("Stopping N.O.R.E. daemon...")

			if err := manager.Stop(ctx); err != nil {
				if errors.Is(err, noredaemon.ErrNotRunning) {
					fmt.Println("Daemon is not currently running.")
					return nil
				}
				return fmt.Errorf("failed to stop daemon: %w", err)
			}

			fmt.Println("Daemon stopped successfully.")
			return nil
		},
	}
}
