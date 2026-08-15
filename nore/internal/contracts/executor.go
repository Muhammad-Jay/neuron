package contracts

import (
	"context"

	core2 "github.com/Muhammad-Jay/neuron/shared/types/core"
)

type ExecutionContext struct {
	ExecutionID   core2.ID
	CorrelationID core2.ID

	Service core2.Service
	Input   map[string]any

	// ServiceConfigurations contains the fully resolved configuration for this
	// single Service execution. It contains no {{ ... }} placeholders.
	ServiceConfigurations map[string]any
}

type Executor interface {
	Execute(ctx context.Context, execution ExecutionContext) (map[string]any, error)
}

type ExecutorRegistry interface {
	Register(serviceType core2.ServiceType, executor Executor) error
	Resolve(serviceType core2.ServiceType) (Executor, error)
}