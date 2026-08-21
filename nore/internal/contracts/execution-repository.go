package contracts

import (
	"context"

	executionmodel "github.com/Muhammad-Jay/neuron/nore/internal/execution"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type ExecutionRepository interface {
	Add(execution *executionmodel.Execution) error
	Get(executionID core.ID) (*executionmodel.Execution, bool)
	Save(ctx context.Context, execution *executionmodel.Execution) error
	Delete(executionID core.ID)
	List() []*executionmodel.Execution
}
