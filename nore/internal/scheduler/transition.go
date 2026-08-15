package scheduler

import (
	"context"
	"fmt"

	runtimemodel "github.com/Muhammad-Jay/neuron/nore/internal/runtime"
	"github.com/Muhammad-Jay/neuron/nore/internal/types"

	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
)

func buildTransitionEnvironment(execution *runtimemodel.Execution, sourceNode types.ExecutionNode, output map[string]any) resolver.Environment {
	service := sourceNode.Service
	return resolver.Environment{
		Source: map[string]any{
			"id": string(service.Metadata.ID), "name": service.Metadata.Name, "type": string(service.Type),
			"input": execution.Input(service.Metadata.ID), "output": cloneMap(output),
			"metadata": map[string]any{
				"id": string(service.Metadata.ID), "name": service.Metadata.Name,
				"description": service.Metadata.Description, "version": service.Metadata.Version,
			},
		},
		Execution: map[string]any{
			"id": string(execution.ID), "correlation_id": string(execution.CorrelationID),
			"input": execution.InitialInput(),
			"blueprint": map[string]any{
				"id": string(execution.Blueprint.Metadata.ID), "name": execution.Blueprint.Metadata.Name,
				"version": execution.Blueprint.Metadata.Version,
			},
		},
	}
}

func validateTransition(ctx context.Context, environment resolver.Environment, transition types.ExecutionTransition) error {
	for index, rule := range transition.Validations {
		if rule.Program == nil {
			return fmt.Errorf("connector %s contains an uncompiled validation %q", transition.ConnectorID, rule.Expression)
		}
		value, err := rule.Program.Evaluate(ctx, environment)
		if err != nil {
			return fmt.Errorf("connector %s validation %d failed to evaluate: %w", transition.ConnectorID, index, err)
		}
		valid, ok := value.(bool)
		if !ok {
			return fmt.Errorf("connector %s validation %q returned %T; expected bool", transition.ConnectorID, rule.Expression, value)
		}
		if !valid {
			return fmt.Errorf("connector %s: %s", transition.ConnectorID, rule.Message)
		}
	}
	return nil
}

func applyTransition(ctx context.Context, environment resolver.Environment, transition types.ExecutionTransition) (map[string]any, error) {
	// Empty mappings mean control flow only. No data is transferred.
	if len(transition.Mappings) == 0 {
		return map[string]any{}, nil
	}
	input := make(map[string]any)
	for _, mapping := range transition.Mappings {
		if mapping.Program == nil {
			return nil, fmt.Errorf("connector %s contains an uncompiled expression %q", transition.ConnectorID, mapping.Expression)
		}
		value, err := mapping.Program.Evaluate(ctx, environment)
		if err != nil {
			return nil, fmt.Errorf("connector %s expression %q failed: %w", transition.ConnectorID, mapping.Expression, err)
		}
		if err := setPath(input, mapping.TargetPath, value); err != nil {
			return nil, fmt.Errorf("connector %s could not assign target %q: %w", transition.ConnectorID, mapping.TargetPath, err)
		}
	}
	return input, nil
}
