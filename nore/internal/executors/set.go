package executors

import (
	"context"
	"maps"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

type SetExecutor struct{}

func (SetExecutor) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	output := make(map[string]any, len(execution.Input)+len(execution.ServiceConfigurations))
	maps.Copy(output, execution.Input)
	maps.Copy(output, execution.ServiceConfigurations)
	return output, nil
}