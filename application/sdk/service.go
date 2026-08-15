package sdk

import "github.com/Muhammad-Jay/neuron/shared/types/core"

// Service creates a developer-facing Service declaration.
//
// This does not replace core.Service. It is simply an ergonomic constructor
// that keeps application code readable.
func Service(
	id string,
	serviceType core.ServiceType,
) core.Service {
	return core.Service{
		Metadata: core.Metadata{
			ID:   core.ID(id),
			Name: id,
		},
		Type: serviceType,
		Inputs: make([]core.Port, 0),
		Outputs: make([]core.Port, 0),
		ServiceConfigurations: make(
			core.ServiceConfigurations,
		),
	}
}

// NamedService creates a Service with a human-readable name.
func NamedService(
	id string,
	name string,
	serviceType core.ServiceType,
) core.Service {
	service := Service(id, serviceType)
	service.Metadata.Name = name
	return service
}

// Input creates a service input declaration.
func Input(
	name string,
	valueType core.ValueType,
	required bool,
) core.Port {
	return core.Port{
		Name:     name,
		Type:     valueType,
		Required: required,
	}
}

// Output creates a service output declaration.
func Output(
	name string,
	valueType core.ValueType,
) core.Port {
	return core.Port{
		Name: name,
		Type: valueType,
	}
}