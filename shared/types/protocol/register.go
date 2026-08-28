package protocol

import "github.com/Muhammad-Jay/neuron/shared/types/core"

type RegisterRequest struct {
	System core.System `json:"system"`
	ExecutionConfigurations any `json:"execution_configurations,omitempty"`
}
type RegisterResponse struct {
	Message string `json:"message"`
}
