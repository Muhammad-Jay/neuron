package instance

import (
	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

var instanceID string

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage N.O.R.E. system instances",
		Long:  `Create, list, and manage running system instances.`,
		// If the user runs 'neuron instance' without a subcommand, show the help menu.
		RunE: func(cmd *cobra.Command, args []string) error {
			//if len(args) == 0 {
			//	return cmd.Help()
			//}
			//
			//if len(args) == 1 {
			//	instanceID = args[0]
			//
			//	instance, err := getInstance(cmd, instanceID)
			//
			//	if instance == nil && err != nil {
			//		return err
			//	}
			//
			//	if instance == nil {
			//		fmt.Println("Instance not found")
			//		return nil
			//	}
			//
			//	printInstances([]protocol.InstanceResponse{*instance})
			//}

			return nil
		},
	}

	cmd.AddCommand(newListCmd())

	return cmd
}

func getInstance(cmd *cobra.Command, id string) (*protocol.InstanceResponse, error) {
	ctx := cmd.Context()

	c, cleanup, err := bootstrap.SetupClient(ctx)
	if err != nil {
		return &protocol.InstanceResponse{}, err
	}
	defer cleanup()

	instance, err := c.GetInstanceById(ctx, id)
	if err != nil {
		return &protocol.InstanceResponse{}, err
	}

	return &instance, nil
}
