package instance

import (
	"context"
	"fmt"
	"sync"

	"github.com/Muhammad-Jay/neuron/nore/internal/engine"
	"github.com/Muhammad-Jay/neuron/nore/internal/event"
	"github.com/Muhammad-Jay/neuron/nore/internal/planner"
	"github.com/Muhammad-Jay/neuron/nore/internal/registry"
	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
	"github.com/Muhammad-Jay/neuron/nore/internal/runtime"
	"github.com/Muhammad-Jay/neuron/nore/internal/scheduler"
	"github.com/Muhammad-Jay/neuron/nore/internal/types"
	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
)

type Status string

const (
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusStopping Status = "stopping"
	StatusStopped  Status = "stopped"
	StatusFailed   Status = "failed"
)

type Instance struct {
	ID        string
	Key       Key
	Blueprint *types.ExecutionBlueprint

	mu     sync.RWMutex
	status Status

	ctx    context.Context
	cancel context.CancelFunc
	bus    *event.Bus
	store  *runtime.MemoryStore

	scheduler *scheduler.Scheduler
	engine    *engine.ExecutorEngine
}

func New(
	parent context.Context,
	id string,
	key Key,
	system *shared.System,
	workers int,
) (*Instance, error) {
	if system == nil {
		return nil, fmt.Errorf("blueprint is required")
	}
	if workers <= 0 {
		workers = 8
	}

	ctx, cancel := context.WithCancel(parent)
	bus := event.NewBus()
	store := runtime.NewMemoryStore()
	reg := registry.New()
	reg.RegisterCoreServiceExecutors()

	sched, err := scheduler.New(bus, store)
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	celCompiler, err := resolver.NewCELCompiler(resolver.DefaultCELConfig())
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("create cel compiler: %w", err)
	}

	systemCompiler, err := planner.NewCompiler(celCompiler)
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("create compiler: %w", err)
	}

	blueprint, err := systemCompiler.Compile(*system)
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("compile system: %w", err)
	}

	execEngine, err := engine.NewExecutorEngine(bus, reg, store, workers)
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("create executor engine: %w", err)
	}

	i := &Instance{
		ID:        id,
		Key:       key,
		Blueprint: blueprint,
		status:    StatusStarting,
		ctx:       ctx,
		cancel:    cancel,
		bus:       bus,
		store:     store,
		scheduler: sched,
		engine:    execEngine,
	}

	return i, nil
}

func (i *Instance) Start() error {
	i.mu.Lock()
	if i.status == StatusRunning {
		i.mu.Unlock()
		return nil
	}
	if i.status != StatusStarting && i.status != StatusStopped {
		status := i.status
		i.mu.Unlock()
		return fmt.Errorf("instance %s cannot start from %s", i.ID, status)
	}
	i.status = StatusRunning
	i.mu.Unlock()

	go func() {
		if err := i.scheduler.Run(i.ctx); err != nil {
			i.fail(err)
		}
	}()
	go func() {
		if err := i.engine.Run(i.ctx); err != nil {
			i.fail(err)
		}
	}()

	return nil
}

func (i *Instance) fail(err error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.status == StatusRunning {
		i.status = StatusFailed
	}
}

func (i *Instance) Stop() error {
	i.mu.Lock()
	if i.status == StatusStopped {
		i.mu.Unlock()
		return nil
	}
	i.status = StatusStopping
	i.mu.Unlock()

	i.cancel()
	i.bus.Close()

	i.mu.Lock()
	i.status = StatusStopped
	i.mu.Unlock()
	return nil
}

func (i *Instance) Status() Status {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

func (i *Instance) Store() *runtime.MemoryStore {
	return i.store
}

func (i *Instance) Bus() *event.Bus {
	return i.bus
}

// Execute creates the execution context/store entry automatically.
// The instance owns the execution infrastructure; callers never construct
// a MemoryStore or Execution manually.
func (i *Instance) Execute(ctx context.Context, input map[string]any) (*runtime.Execution, error) {
	if i.Status() != StatusRunning {
		return nil, fmt.Errorf("instance %s is not running", i.ID)
	}

	execution, err := runtime.NewExecution(i.Blueprint, shared.NewID("request_"))
	if err != nil {
		return nil, err
	}
	if err := i.store.Add(execution); err != nil {
		return nil, err
	}

	if err := i.bus.Publish(
		ctx,
		event.New(
			event.ExecutionStarted,
			execution.ID,
			execution.CorrelationID,
			"",
			event.ExecutionStartedPayload{Input: input},
		),
	); err != nil {
		return nil, fmt.Errorf("publish execution start: %w", err)
	}

	return execution, nil
}
