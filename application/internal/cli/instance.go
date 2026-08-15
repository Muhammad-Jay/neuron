package cli

import (
	"github.com/spf13/cobra"
)

// instanceCmd represents the "instance" command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage N.O.R.E. system instances",
	Long:  `Create, list, and manage running system instances.`,
	// If the user runs 'neuron instance' without a subcommand, show the help menu.
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	// Attach the "instance" command to the "root" command
	rootCmd.AddCommand(instanceCmd)
}