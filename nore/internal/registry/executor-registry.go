package registry

import (
	"fmt"
	"sort"
	"strings"
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

func (r *Registry) RegisterCoreServiceExecutors() {
	builtins := map[string]contracts.Executor{
		"set":     executors.SetExecutor{},
		"ai":      executors.AIMockExecutor{},
		"log":     executors.LogExecutor{},
		"http":    executors.HttpExecutor{},
		"delay":   executors.DelayExecutor{},
		"command": executors.CommandExecutor{},
	}

	// Register each in-process executor under its canonical namespaced name
	// (neuron:core:set, ...) plus the legacy bare name so blueprints authored
	// before the namespace existed still resolve.
	for name, executor := range builtins {
		must(r.Register(core.CoreName(name), executor))
		must(r.Register(core.ServiceType(name), executor))
	}
}

// Close releases resources held by registered executors that implement
// contracts.ExecutorCloser (e.g. wasm runtimes). It is idempotent and safe to
// call from an instance shutdown path once no execution is in flight.
func (r *Registry) Close() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var failures []string
	types := make([]string, 0, len(r.executors))
	for serviceType := range r.executors {
		types = append(types, string(serviceType))
	}
	sort.Strings(types)

	for _, serviceType := range types {
		closer, ok := r.executors[core.ServiceType(serviceType)].(contracts.ExecutorCloser)
		if !ok {
			continue
		}
		if err := closer.Close(); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", serviceType, err))
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("close executors: %s", strings.Join(failures, "; "))
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
