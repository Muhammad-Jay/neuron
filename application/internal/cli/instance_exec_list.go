package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var instanceExecListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all executions in an instance",
	RunE: instanceExecListCmdHandler,
}

func instanceExecListCmdHandler(cmd *cobra.Command, args []string) error {
	fmt.Println("Exec List called")
	return nil
}

func init() {
	instanceExecCmd.AddCommand(instanceExecListCmd)
}
