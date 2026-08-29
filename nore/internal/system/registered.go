// Package system holds the durable representation of a registered system:
// a first-class N.O.R.E. resource that exists independently of any live
// Instance. Instances are reconstructed from a RegisteredSystem on demand.
package system

import (
	"time"

	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

// RegisteredSystem is the durable artifact persisted by /v1/register.
//
// This is the source of truth for a system's definition. An Instance is only
// a transient runtime built from it, so the registered artifact is reloadable
// and re-instantiable after a restart.
type RegisteredSystem struct {
	Key protocol.InstanceKey `json:"key"`

	System shared.System `json:"system"`

	// ExecutionConfigurations is stored opaquely. For CLI-sourced systems it is
	// a []project.ExecutorSource, which the runtime module cannot import.
	ExecutionConfigurations any `json:"execution_configurations,omitempty"`

	// BlueprintMetadata is the system-level metadata produced when the system
	// is compiled, kept so list views can surface graph metadata lazily.
	BlueprintMetadata shared.Metadata `json:"blueprint_metadata,omitempty"`

	RegisteredAt time.Time `json:"registered_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}