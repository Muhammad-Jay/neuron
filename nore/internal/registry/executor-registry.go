package registry

import (
	"fmt"
	"sync"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	"github.com/Muhammad-Jay/neuron/nore/internal/executors"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type Registry struct {
	mu        sync.RWMutex
	executors map[core.ServiceType]contracts.Executor
}

func New() *Registry {
	return &Registry{executors: make(map[core.ServiceType]contracts.Executor)}
}

func (r *Registry) Register(serviceType core.ServiceType, executor contracts.Executor) error {
	if serviceType == "" || executor == nil {
		return fmt.Errorf("service type and executor are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.executors[serviceType]; exists {
		return fmt.Errorf("executor for service type %q is already registered", serviceType)
	}
	r.executors[serviceType] = executor
	return nil
}

func (r *Registry) Resolve(serviceType core.ServiceType) (contracts.Executor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	executor, exists := r.executors[serviceType]
	if !exists {
		return nil, fmt.Errorf("executor for service type %q was not found", serviceType)
	}
	return executor, nil
}

func (r *Registry) RegisterCoreServiceExecutors()  {
	must(r.Register("set", executors.SetExecutor{}))
	must(r.Register("ai", executors.AIMockExecutor{}))
	must(r.Register("log", executors.LogExecutor{}))
	must(r.Register("http", executors.HttpExecutor{}))
	must(r.Register("delay", executors.DelayExecutor{}))
	must(r.Register("command", executors.CommandExecutor{}))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}