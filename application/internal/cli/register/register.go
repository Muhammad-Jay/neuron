package register

import (
	"context"
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/compiler"
	"github.com/Muhammad-Jay/neuron/application/compiler/manifest"
	"github.com/Muhammad-Jay/neuron/application/config"
	"github.com/Muhammad-Jay/neuron/application/executor"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/internal/executorctl"
	"github.com/Muhammad-Jay/neuron/application/project"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
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

	configs := compiler.BuildExecutionConfigurations(m)

	// Resolve the executor requirements declared by services and freeze the
	// exact dependency set into the register payload, so N.O.R.E. can launch
	// Instances without resolving or installing anything itself.
	resolved, err := resolveFrozenExecutors(ctx, cfg, configs.ExecutorRequirements)
	if err != nil {
		return err
	}
	configs.ResolvedExecutors = resolved

	request := protocol.RegisterRequest{
		Key:                     key,
		System:                  *sys,
		ExecutionConfigurations: configs,
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

// resolveFrozenExecutors resolves each executor requirement through the wired
// catalog and freezes the results into the wire format persisted in a
// Deployment.
func resolveFrozenExecutors(ctx context.Context, cfg config.Config, requirements []manifest.ExecutorRequirement) ([]shadexec.ResolvedExecutor, error) {
	if len(requirements) == 0 {
		return nil, nil
	}

	catalog, err := executorctl.BuildCatalog(executorctl.CatalogConfig{
		ExecutorsConfig: cfg.Executors,
	})
	if err != nil {
		return nil, fmt.Errorf("build executor catalog: %w", err)
	}

	executorReqs := make([]executor.Requirement, 0, len(requirements))
	for _, req := range requirements {
		executorReqs = append(executorReqs, catalog.Require(req.Name, req.Version, []string{req.Registry}))
	}

	env, err := catalog.Resolve(ctx, executorReqs)
	if err != nil {
		return nil, fmt.Errorf("resolve executors: %w", err)
	}

	requested := make(map[string]string, len(executorReqs))
	for _, req := range executorReqs {
		requested[req.Type] = req.Version
	}

	frozen := make([]shadexec.ResolvedExecutor, 0, len(env.Executors))
	for _, installed := range env.Executors {
		frozen = append(frozen, *installed.Frozen(requested[installed.Type]))
	}
	return frozen, nil
}
