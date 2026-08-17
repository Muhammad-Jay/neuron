package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Muhammad-Jay/neuron/application/daemon"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the local N.O.R.E. daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background N.O.R.E. daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		cfg := daemon.DefaultConfig()
		manager := daemon.NewManager(cfg, nil)

		fmt.Println("Stopping N.O.R.E. daemon...")

		if err := manager.Stop(ctx); err != nil {
			if errors.Is(err, daemon.ErrNotRunning) {
				fmt.Println("Daemon is not currently running.")
				return nil
			}
			return fmt.Errorf("failed to stop daemon: %w", err)
		}

		fmt.Println("Daemon stopped successfully.")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStopCmd)
}