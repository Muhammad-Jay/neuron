package sdk

import (
	"fmt"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type System struct {
	system        *core.System
	dslConnectors []*Connector // Tracks pointers so mappings are preserved
}

func NewSystem(metadata core.Metadata) *System {
	return &System{
		system: &core.System{
			Metadata: metadata,
			Specification: core.SystemSpec{
				Services:   make([]core.Service, 0),
				Triggers:   make([]core.Trigger, 0),
				Connectors: make([]core.Connector, 0),
			},
		},
		dslConnectors: make([]*Connector, 0),
	}
}

// New creates a new systems.
//
// Example:
//
//	var sys = mvp.New("Customer Platform", "1.0")
func New(name, version string) *System {
	return NewSystem(core.Metadata{
		ID:      core.NewID("system_"),
		Name:    name,
		Version: version,
	})
}

// AddServices adds a group of services.
func (s *System) AddServices(services ...core.Service) *System {
	s.system.Specification.Services = append(
		s.system.Specification.Services,
		services...,
	)
	return s
}

// AddService adds one service.
func (s *System) AddService(service core.Service) *System {
	return s.AddServices(service)
}

// Services creates a grouped service declaration.
func (s *System) Services(services ...core.Service) *System {
	return s.AddServices(services...)
}

// Connector creates a transition between two services.
func (s *System) Connector(
	source core.Service,
	target core.Service,
) *Connector {
	connector := NewConnector(
		source.Metadata.ID,
		target.Metadata.ID,
	)

	// FIX: Track the pointer so subsequent AddMappings() calls are saved.
	s.dslConnectors = append(s.dslConnectors, connector)

	return connector
}

// AddConnectors adds already-created types.Connectors.
func (s *System) AddConnectors(
	connectors ...core.Connector,
) *System {
	s.system.Specification.Connectors = append(
		s.system.Specification.Connectors,
		connectors...,
	)
	return s
}

// AddConnector adds already-created single core.Connector.
func (s *System) AddConnector(
	connector core.Connector,
) *System {
	s.system.Specification.Connectors = append(
		s.system.Specification.Connectors,
		connector,
	)
	return s
}

// Trigger registers a Service as a System entry point.
func (s *System) Trigger(service core.Service) *System {
	s.system.Specification.Triggers = append(
		s.system.Specification.Triggers,
		core.Trigger{Service: service},
	)
	return s
}

// Build returns the immutable specification consumed by N.O.R.E.
func (s *System) Build() *core.System {
	if s == nil || s.system == nil {
		panic("mvp: nil systems")
	}

	// Compile all the tracked DSL connectors right before building.
	// This ensures all AddMappings() and AddValidations() are captured.
	var finalConnectors []core.Connector
	finalConnectors = append(finalConnectors, s.system.Specification.Connectors...)
	for _, dslConn := range s.dslConnectors {
		finalConnectors = append(finalConnectors, dslConn.Core())
	}
	s.system.Specification.Connectors = finalConnectors

	// Clear dslConnectors to make Build() safe to call multiple times
	s.dslConnectors = nil

	if s.system.Metadata.Name == "" {
		panic("mvp: systems name is required")
	}
	if s.system.Metadata.Version == "" {
		panic("mvp: systems version is required")
	}
	if len(s.system.Specification.Services) == 0 {
		panic("mvp: systems must contain at least one service")
	}

	return s.system
}

func (s *System) MustBuild() *core.System {
	return s.Build()
}

// Metadata allows changing systems metadata without exposing core.System.
func (s *System) Metadata(name, version string) *System {
	s.system.Metadata.Name = name
	s.system.Metadata.Version = version
	return s
}

// Description sets the System description.
func (s *System) Description(description string) *System {
	s.system.Metadata.Description = description
	return s
}

// Label adds a System label.
func (s *System) Label(key, value string) *System {
	if s.system.Metadata.Labels == nil {
		s.system.Metadata.Labels = make(map[string]string)
	}
	s.system.Metadata.Labels[key] = value
	return s
}

// Validate performs only DSL-level structural checks.
func (s *System) Validate() error {
	if s == nil || s.system == nil {
		return fmt.Errorf("systems is nil")
	}
	if s.system.Metadata.Name == "" {
		return fmt.Errorf("systems name is required")
	}
	if s.system.Metadata.Version == "" {
		return fmt.Errorf("systems version is required")
	}
	if len(s.system.Specification.Services) == 0 {
		return fmt.Errorf("systems must contain at least one service")
	}
	return nil
}