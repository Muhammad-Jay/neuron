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

	c, cleanup, err := setupNoreClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	sys := customersystem.System.Build()

	input := map[string]any {
		"target_url": "https://jsonplaceholder.typicode.com/todos/1",
	}

	var key = protocol.InstanceKey{SystemID: string(sys.Metadata.ID)}

	res, err := c.Execute(ctx, key, sys, input)

	if err != nil {
		return err
	}

	fmt.Println(res)

	return nil
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}