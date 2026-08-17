package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var instanceExecCmd = &cobra.Command{
	Use: "exec",
	Short: "Instance Executions command",
	RunE: instanceExecCmdHandler,
}

func instanceExecCmdHandler(cmd *cobra.Command, args []string) error {
	fmt.Println("instanceExec called")
	return nil
}

func init() {
	instanceCmd.AddCommand(instanceExecCmd)
}
