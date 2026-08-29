package protocol

import (
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

// RegisterRequest registers a system definition durably with N.O.R.E. It does
// not create or start an Instance; instances are created lazily on the first
// execution request for the system's key.
type RegisterRequest struct {
	System core.System `json:"system"`

	// Key is the client-computed identity (SystemID, Version, Hash, Env). When
	// provided the server validates Hash matches the system content; when empty
	// the server derives the key from the system metadata.
	Key InstanceKey `json:"key,omitempty"`

	// ExecutionConfigurations carries opaque executor-requirement metadata that
	// N.O.R.E. persists alongside the system for later runtime assembly.
	ExecutionConfigurations any `json:"execution_configurations,omitempty"`
}

// RegisterStatus describes the outcome of a registration against a durable key.
type RegisterStatus string

const (
	RegisterStatusRegistered        RegisterStatus = "registered"
	RegisterStatusAlreadyRegistered RegisterStatus = "already_registered"
	RegisterStatusReplaced          RegisterStatus = "replaced"
)

type RegisterResponse struct {
	Key     InstanceKey    `json:"key"`
	Status  RegisterStatus `json:"status"`
	Message string         `json:"message,omitempty"`
}