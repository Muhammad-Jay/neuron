package execution

import "github.com/spf13/cobra"

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "execution",
		Short: "List all execution instances",
		Long:  "List all execution instances",
		RunE:  executionCmdHandler,
	}
}

func executionCmdHandler(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}