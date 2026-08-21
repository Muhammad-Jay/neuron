package executors

import (
	"context"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

// LogExecutor surfaces the service's input as a ServiceLog event and passes
// the input through unchanged, acting as a pass-through observability node.
type LogExecutor struct{}

// Execute emits a structured log of the service input/config and returns the
// input as output so downstream services keep receiving the same data.
func (LogExecutor) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	execution.Logger.Info(ctx,
		"[log] input received",
		contracts.F("input", execution.Input),
		contracts.F("config", execution.ServiceConfigurations),
	)
	return execution.Input, nil
}
