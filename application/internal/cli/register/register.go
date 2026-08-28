package register

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Register,
		Short: "Build, and register the current repo to N.O.R.E",
		RunE: registerCmdHandler,
	}


	return cmd
}

func registerCmdHandler(cmd *cobra.Command, args []string) error  {
	ctx := cmd.Context()

	c, cleanup, err := bootstrap.SetupClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	p, err := resolveAndParse(ctx)
	if err != nil {
		return err
	}

	sys, err := p.ParseSystem()
	if err != nil { return err }

	execSrc := p.GetExecutorSources()

	request := protocol.RegisterRequest{
		System: *sys,
		ExecutionConfigurations: execSrc,
	}

	result, err := c.Register(ctx, request)
	if err != nil { return err }

	fmt.Println(result.Message)

	return nil
}
