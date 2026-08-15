package contracts

import (
	runtimemodel "github.com/Muhammad-Jay/neuron/nore/internal/runtime"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type ExecutionRepository interface {
	Add(execution *runtimemodel.Execution) error
	Get(executionID core.ID) (*runtimemodel.Execution, bool)
	Delete(executionID core.ID)
}