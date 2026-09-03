package compiler

import (
	"github.com/Muhammad-Jay/neuron/application/compiler/manifest"
)

// ExecutionConfigurations is the consolidated project/runtime configuration
// payload sent to N.O.R.E. alongside a registered system. N.O.R.E. persists
// it opaquely today but will consume it for runtime assembly (executor
// discovery, storage, inspector, runtime defaults) in the future.
//
// The structure is intentionally decoupled from both the manifest source
// languages and N.O.R.E.'s internal model, so it can evolve independently.
type ExecutionConfigurations struct {
	// ExecutorRegistries lists the registries that can supply executor
	// implementations for the system's services.
	ExecutorRegistries []manifest.ExecutorRegistry `json:"executor_registries,omitempty"`

	// ExecutorRequirements is an indexed view of the executors each
	// service requires, keyed by (type, version, source).
	ExecutorRequirements []manifest.ExecutorRequirement `json:"executor_requirements,omitempty"`

	// Runtime describes runtime execution defaults.
	Runtime manifest.RuntimeConfig `json:"runtime,omitempty"`

	// Storage describes the durable storage provider.
	Storage manifest.StorageConfig `json:"storage,omitempty"`

	// Inspector describes the inspector endpoint.
	Inspector manifest.InspectorConfig `json:"inspector,omitempty"`
}

// BuildExecutionConfigurations assembles the project/runtime configuration
// payload from a manifest for registration with N.O.R.E.
func BuildExecutionConfigurations(m *manifest.System) ExecutionConfigurations {
	regs := m.Config.ExecutorRegistries

	// Index executor requirements from services.
	reqs := collectExecutorRequirements(m.Services)

	return ExecutionConfigurations{
		ExecutorRegistries:   regs,
		ExecutorRequirements: reqs,
		Runtime:              m.Config.Runtime,
		Storage:              m.Config.Storage,
		Inspector:            m.Config.Inspector,
	}
}

// collectExecutorRequirements groups services by their executor key so
// N.O.R.E. can see which executors are needed and which services use each.
func collectExecutorRequirements(services []manifest.Service) []manifest.ExecutorRequirement {
	type key struct {
		Name     string
		Version  string
		Registry string
	}

	index := make(map[key]*manifest.ExecutorRequirement)

	for _, svc := range services {
		exec := svc.Executor
		k := key{exec.Name, exec.Version, exec.Registry}

		req, ok := index[k]
		if !ok {
			req = &manifest.ExecutorRequirement{
				Name:     exec.Name,
				Version:  exec.Version,
				Registry: exec.Registry,
			}
			index[k] = req
		}
		req.Services = append(req.Services, svc.Name)
	}

	result := make([]manifest.ExecutorRequirement, 0, len(index))
	for _, req := range index {
		result = append(result, *req)
	}
	return result
}
