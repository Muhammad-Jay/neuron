package executors

import (
	"context"
	"fmt"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

type LogExecutor struct{}

func (LogExecutor) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	fmt.Printf("[service=%s] input=%v config=%v\n", execution.Service.Metadata.Name, execution.Input, execution.ServiceConfigurations)
	return execution.Input, nil
}