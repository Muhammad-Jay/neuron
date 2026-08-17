package cli

import "github.com/spf13/cobra"

var executionCmd = &cobra.Command{
	Use:   "execution",
	Short: "List all execution instances",
	Long: "List all execution instances",
	RunE: executionCmdHandler,
}

func executionCmdHandler(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func init() {
	RootCmd.AddCommand(executionCmd)
}