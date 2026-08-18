package instance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	"github.com/Muhammad-Jay/neuron/nore/internal/engine"
	"github.com/Muhammad-Jay/neuron/nore/internal/event"
	"github.com/Muhammad-Jay/neuron/nore/internal/execution"
	"github.com/Muhammad-Jay/neuron/nore/internal/planner"
	"github.com/Muhammad-Jay/neuron/nore/internal/registry"
	"github.com/Muhammad-Jay/neuron/nore/internal/resolver"
	"github.com/Muhammad-Jay/neuron/nore/internal/runtime"
	"github.com/Muhammad-Jay/neuron/nore/internal/scheduler"
	"github.com/Muhammad-Jay/neuron/nore/internal/storage"
	"github.com/Muhammad-Jay/neuron/nore/internal/types"
	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
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
	Key       protocol.InstanceKey
	Blueprint *types.ExecutionBlueprint

	mu     sync.RWMutex
	wg     sync.WaitGroup
	status Status

	ctx       context.Context
	cancel    context.CancelFunc
	bus       *event.Bus
	createdAt time.Time

	store      contracts.ExecutionRepository
	eventStore *event.Store

	scheduler *scheduler.Scheduler
	engine    *engine.ExecutorEngine

	execPersister *executionPersister
}

func New(
	parent context.Context,
	id string,
	key protocol.InstanceKey,
	system *shared.System,
	workers int,
	persistentStore storage.Store,
) (*Instance, error) {
	if system == nil {
		return nil, fmt.Errorf("blueprint is required")
	}
	if workers <= 0 {
		workers = 8
	}

	ctx, cancel := context.WithCancel(parent)
	bus := event.NewBus()

	store := execution.NewExecutionStore(persistentStore)
	evtStore := event.NewStore(persistentStore)

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

	persister, err := newExecutionPersister(bus, store)
	if err != nil {
		cancel()
		bus.Close()
		return nil, fmt.Errorf("create execution persister: %w", err)
	}

	i := &Instance{
		ID:            id,
		Key:           key,
		Blueprint:     blueprint,
		status:        StatusStarting,
		ctx:           ctx,
		cancel:        cancel,
		bus:           bus,
		createdAt:     time.Now().UTC(),
		store:         store,
		eventStore:    evtStore,
		scheduler:     sched,
		engine:        execEngine,
		execPersister: persister,
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

	if i.scheduler == nil || i.engine == nil {
		return fmt.Errorf("instance %s has no runtime; it was restored from metadata", i.ID)
	}

	i.wg.Add(4)

	go func() {
		defer i.wg.Done()
		if err := i.scheduler.Run(i.ctx); err != nil {
			i.fail(err)
		}
	}()
	go func() {
		defer i.wg.Done()
		if err := i.engine.Run(i.ctx); err != nil {
			i.fail(err)
		}
	}()
	go func() {
		defer i.wg.Done()
		i.persistEvents(i.ctx)
	}()
	go func() {
		defer i.wg.Done()
		_ = i.execPersister.Run(i.ctx)
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
	if i.bus != nil {
		i.wg.Wait()
		i.bus.Close()
	}

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

func (i *Instance) Store() contracts.ExecutionRepository {
	return i.store
}

func (i *Instance) Bus() *event.Bus {
	return i.bus
}

func (i *Instance) EventStore() *event.Store {
	return i.eventStore
}

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

func (i *Instance) ListExecutions() []*runtime.Execution {
	return i.store.List()
}

func (i *Instance) GetExecution(id shared.ID) (*runtime.Execution, bool) {
	return i.store.Get(id)
}

func (i *Instance) ListExecutionEvents(ctx context.Context, executionID shared.ID) ([]event.Event, error) {
	return i.eventStore.List(ctx, executionID)
}

func (i *Instance) persistEvents(ctx context.Context) {
	sub, err := i.bus.Subscribe(event.All, 256)
	if err != nil {
		return
	}
	defer sub.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-sub.Events():
			if !ok {
				return
			}
			_ = i.eventStore.Save(ctx, evt)
		}
	}
}
