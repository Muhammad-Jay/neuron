package build

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the neuron systems",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Read values strictly from Viper, not from flag variables!
			logLevel := viper.GetString("log_level")
			isWatch := viper.GetBool("build.watch")

			fmt.Printf("Starting build... (Log Level: %s)\n", logLevel)

			if isWatch {
				fmt.Println("Watch mode is ON. Listening for file changes...")
			} else {
				fmt.Println("Watch mode is OFF. Running standard build...")
			}

			return nil
		},
	}

	// 2. Define the flag directly on the command
	cmd.Flags().BoolP("watch", "w", false, "watch for file changes")

	// 3. Bind the flag to a Viper key ("build.watch")
	// This makes it so viper.GetBool("build.watch") will pull from this flag.
	_ = viper.BindPFlag("build.watch", cmd.Flags().Lookup("watch"))

	return cmd
}
