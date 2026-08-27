package register

import (
	"fmt"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Build, and register the current repo to N.O.R.E",
		RunE: registerCmdHandler,
	}


	return cmd
}

func registerCmdHandler(cmd *cobra.Command, args []string) error  {
	fmt.Println("register called")

	return nil
}
