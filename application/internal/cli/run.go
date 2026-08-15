package cli

import (
	"context"
	customersystem "development/systems/customer-system"
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/connection"

	"github.com/spf13/cobra"
)

var input []byte

var verbose bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Neuron System.",
	Long: "Run a Neuron System, and the internal N.O.R.E runtime will execute the System.",

	RunE: cmdHandler,
}

func cmdHandler(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Running Neuron System.")

	if verbose {
		fmt.Println("Running in verbose mode.")
	}

	//// Define it directly as a byte slice
	//input := []byte(`{ "target_url": "https://jsonplaceholder.typicode.com/todos/1"}`)
	//

	c := client.NewDefaultLocal()
	defer c.Close()

	sys := customersystem.System.Build()

	key := connection.InstanceKey{
		SystemID: string(sys.Metadata.ID),
	}

	res, err := c.EnsureInstance(ctx, key, sys)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Successfully created instance.")
	fmt.Println(res)

	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
