package register

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Register,
		Short: "Build, and register the current repo to N.O.R.E",
		RunE:  registerCmdHandler,
	}

	return cmd
}

func registerCmdHandler(cmd *cobra.Command, args []string) error {
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
	if err != nil {
		return err
	}

	key := p.GetInstanceKey()

	request := protocol.RegisterRequest{
		Key:                     key,
		System:                  *sys,
		ExecutionConfigurations: p.GetExecutorSources(),
	}

	result, err := c.Register(ctx, request)
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := project.SaveRegistrationKey(root, result.Key); err != nil {
		return err
	}

	printRegistration(result)

	return nil
}

func printRegistration(result protocol.RegisterResponse) {
	line := fmt.Sprintf("%s@%s#%s:%s", result.Key.SystemID, result.Key.Version, result.Key.Hash, result.Key.Env)
	if result.Status != "" {
		line += fmt.Sprintf(" (%s)", result.Status)
	}
	fmt.Println(line)
}