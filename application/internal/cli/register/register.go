package register

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/compiler"
	"github.com/Muhammad-Jay/neuron/application/compiler/manifest"
	"github.com/Muhammad-Jay/neuron/application/config"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Register,
		Short: "Register the current project to N.O.R.E",
		Long:  "Load the built .neuron/manifest.json, compile it to a core.System, and register it with N.O.R.E. Requires `neuron build` to have been run first.",
		RunE:  registerCmdHandler,
	}

	return cmd
}

func registerCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, ok := config.FromContext(ctx)
	if !ok {
		return fmt.Errorf("configuration not loaded")
	}

	c, cleanup, err := bootstrap.SetupClient(ctx, bootstrap.Options{Config: cfg})
	if err != nil {
		return err
	}
	defer cleanup()

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Load the canonical manifest produced by `neuron build`.
	m, err := manifest.LoadFromProjectRoot(root)
	if err != nil {
		return fmt.Errorf("load manifest (run `neuron build` first): %w", err)
	}

	// Compile the manifest into the runtime core.System representation.
	comp := compiler.New()
	sys, err := comp.Compile(m)
	if err != nil {
		return fmt.Errorf("compile manifest: %w", err)
	}

	// Compute the instance key from the manifest + compiled system.
	key, err := comp.InstanceKey(m)
	if err != nil {
		return fmt.Errorf("compute instance key: %w", err)
	}

	request := protocol.RegisterRequest{
		Key:                     key,
		System:                  *sys,
		ExecutionConfigurations: compiler.BuildExecutionConfigurations(m),
	}

	result, err := c.Register(ctx, request)
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
