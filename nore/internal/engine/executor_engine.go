package engine

import (
	"context"
	"fmt"

	"sync"

	runtimemodel "github.com/Muhammad-Jay/neuron/nore/internal/runtime"
	"github.com/Muhammad-Jay/neuron/shared/types/core"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	"github.com/Muhammad-Jay/neuron/nore/internal/event"
	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
)

type ExecutorEngine struct {
	bus        contracts.EventBus
	registry   contracts.ExecutorRegistry
	executions contracts.ExecutionRepository
	ready      event.Subscription
	semaphore  chan struct{}
}

func NewExecutorEngine(bus contracts.EventBus, registry contracts.ExecutorRegistry, executions contracts.ExecutionRepository, maxConcurrency int) (*ExecutorEngine, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 8
	}
	ready, err := bus.Subscribe(event.ServiceReady, maxConcurrency*4)
	if err != nil {
		return nil, err
	}
	return &ExecutorEngine{bus: bus, registry: registry, executions: executions, ready: ready, semaphore: make(chan struct{}, maxConcurrency)}, nil
}

func (e *ExecutorEngine) Run(ctx context.Context) error {
	var workers sync.WaitGroup
	defer func() {
		workers.Wait()
		_ = e.ready.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case received, open := <-e.ready.Events():
			if !open {
				return nil
			}
			select {
			case e.semaphore <- struct{}{}:
			case <-ctx.Done():
				return nil
			}
			workers.Add(1)
			go func(received event.Event) {
				defer workers.Done()
				defer func() { <-e.semaphore }()
				e.executeService(ctx, received)
			}(received)
		}
	}
}

func (e *ExecutorEngine) executeService(ctx context.Context, received event.Event) {
	execution, exists := e.executions.Get(received.Metadata.ExecutionID)
	if !exists || execution.IsTerminal() {
		return
	}
	serviceID := received.Metadata.ServiceID
	node, exists := execution.Blueprint.Nodes[serviceID]
	if !exists {
		e.publishFailure(ctx, execution, serviceID, fmt.Errorf("service %s does not exist in the blueprint", serviceID))
		return
	}

	input := execution.Input(serviceID)
	resolvedConfig, err := node.Configurations.Resolve(ctx, resolver.ServiceEnvironment{
		Input: input,
		Execution: map[string]any{
			"id": string(execution.ID), "correlation_id": string(execution.CorrelationID),
			"input": execution.InitialInput(),
			"blueprint": map[string]any{
				"id": string(execution.Blueprint.Metadata.ID), "name": execution.Blueprint.Metadata.Name,
				"version": execution.Blueprint.Metadata.Version,
			},
		},
		Service: map[string]any{
			"id": string(node.Service.Metadata.ID), "name": node.Service.Metadata.Name,
			"type": string(node.Service.Type), "version": node.Service.Metadata.Version,
		},
	})
	if err != nil {
		e.publishFailure(ctx, execution, serviceID, fmt.Errorf("resolve configurations for service %s: %w", serviceID, err))
		return
	}

	if err := execution.MarkServiceRunning(serviceID); err != nil {
		e.publishFailure(ctx, execution, serviceID, err)
		return
	}
	if err := e.bus.Publish(ctx, event.New(event.ServiceStarted, execution.ID, execution.CorrelationID, serviceID, event.ServiceStartedPayload{})); err != nil {
		e.publishFailure(ctx, execution, serviceID, err)
		return
	}
	executor, err := e.registry.Resolve(node.Service.Type)
	if err != nil {
		e.publishFailure(ctx, execution, serviceID, err)
		return
	}
	output, err := executor.Execute(ctx, contracts.ExecutionContext{
		ExecutionID: execution.ID, CorrelationID: execution.CorrelationID,
		Service: node.Service, Input: input, ServiceConfigurations: resolvedConfig,
	})
	if err != nil {
		e.publishFailure(ctx, execution, serviceID, err)
		return
	}
	if output == nil {
		output = map[string]any{}
	}
	if err := execution.MarkServiceCompleted(serviceID, output); err != nil {
		e.publishFailure(ctx, execution, serviceID, err)
		return
	}
	_ = e.bus.Publish(ctx, event.New(event.ServiceCompleted, execution.ID, execution.CorrelationID, serviceID, event.ServiceCompletedPayload{Output: output}))
}

func (e *ExecutorEngine) publishFailure(ctx context.Context, execution *runtimemodel.Execution, serviceID core.ID, err error) {
	execution.MarkServiceFailed(serviceID, err)
	_ = e.bus.Publish(ctx, event.New(event.ServiceFailed, execution.ID, execution.CorrelationID, serviceID, event.ServiceFailedPayload{Message: err.Error()}))
}