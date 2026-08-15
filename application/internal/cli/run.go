package cli

import (
	customersystem "development/systems/customer-system"
	"fmt"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

var verbose bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Neuron System",
	Long:  "Run a Neuron System, and the internal N.O.R.E runtime will execute the System.",
	RunE:  runCmdHandler,
}

func runCmdHandler(cmd *cobra.Command, args []string) error {
	// Use the context provided by Cobra
	ctx := cmd.Context()

	if verbose {
		fmt.Println("Running in verbose mode.")
	}

	// 1. Setup client and ensure Daemon is running
	c, cleanup, err := setupNoreClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	// 2. Build the system
	sys := customersystem.System.Build()
	key := protocol.InstanceKey{
		SystemID: string(sys.Metadata.ID),
	}

	// 3. Ensure the instance exists on the running daemon
	res, err := c.EnsureInstance(ctx, key, sys)
	if err != nil {
		return fmt.Errorf("failed to ensure instance: %w", err)
	}

	fmt.Println("Successfully created instance.")
	fmt.Printf("Instance ID: %s\n", res.ID)

	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}