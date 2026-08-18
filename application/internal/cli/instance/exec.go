package instance

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Instance Executions command",
		RunE:  instanceExecCmdHandler,
	}

	cmd.AddCommand(newExecListCmd())

	return cmd
}

func instanceExecCmdHandler(cmd *cobra.Command, args []string) error {
	fmt.Println("instanceExec called")
	return nil
}

func newExecListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all executions in an instance",
		RunE:  instanceExecListCmdHandler,
	}
}

func instanceExecListCmdHandler(cmd *cobra.Command, args []string) error {
	fmt.Println("Exec List called")
	return nil
}