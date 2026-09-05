// Package runtime materializes Installed executor artifacts into runnable
// Executor instances. It is the boundary N.O.R.E. sees: an Executor executes a
// Request and returns a Response. It never knows about registries, downloaders,
// or GitHub.
//
// Runtime adapters are pluggable by kind. "process" is implemented today;
// "container", "wasm", and "remote" adapters register the same contract later.
package runtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/Muhammad-Jay/neuron/application/executor"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// Executor is a materialized, runnable executor.
type Executor interface {
	// Type is the logical executor name.
	Type() string

	// Version is the exact resolved version.
	Version() string

	// Execute runs one request through the executor.
	Execute(ctx context.Context, req *shadexec.Request) (*shadexec.Response, error)
}

// Factory materializes an Executor from an installed package for one runtime
// kind.
type Factory func(ctx context.Context, installed *executor.Installed) (Executor, error)

// Runtime dispatches materialization to the adapter registered for an
// installed executor's runtime type.
type Runtime struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRuntime returns a Runtime with only explicit registrations.
func NewRuntime() *Runtime {
	return &Runtime{
		factories: make(map[string]Factory),
	}
}

// DefaultRuntime returns a Runtime with the built-in process adapter
// registered.
func DefaultRuntime() *Runtime {
	rt := NewRuntime()
	if err := rt.Register(ProcessKind, NewProcessFactory(nil)); err != nil {
		panic(err) // builtin registration cannot fail
	}
	return rt
}

// Register installs a materialization factory for a runtime kind. Registering
// an existing kind replaces the factory.
func (rt *Runtime) Register(kind string, factory Factory) error {
	if kind == "" {
		return fmt.Errorf("runtime kind is required")
	}
	if factory == nil {
		return fmt.Errorf("runtime factory is nil")
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.factories[kind] = factory
	return nil
}

// Materialize turns an installed executor into a runnable Executor using the
// adapter registered for installed.Runtime.Type.
func (rt *Runtime) Materialize(ctx context.Context, installed *executor.Installed) (Executor, error) {
	if installed == nil {
		return nil, fmt.Errorf("installed executor is nil")
	}
	kind := installed.Runtime.Type
	if kind == "" {
		kind = ProcessKind
	}

	rt.mu.RLock()
	factory := rt.factories[kind]
	rt.mu.RUnlock()

	if factory == nil {
		return nil, fmt.Errorf(
			"no runtime adapter registered for kind %q (supported: process; register container/wasm/remote adapters)",
			kind,
		)
	}

	return factory(ctx, installed)
}

// RuntimeKind is the stable kind identifier for the process adapter.
const ProcessKind = "process"
