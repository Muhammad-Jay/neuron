package connection

import (
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

const (
	HealthPath    = "/health"
	InstancesPath = "/v1/instances"
	ExecutePath   = "/v1/instances/%s/executions"
)

type HealthResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}

type InstanceKey struct {
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

type CreateInstanceRequest struct {
	Key       InstanceKey              `json:"key"`
	System *core.System `json:"system"`
}

type InstanceResponse struct {
	ID     string      `json:"id"`
	Key    InstanceKey `json:"key"`
	Status string      `json:"status"`
}

type ExecuteRequest struct {
	Input map[string]any `json:"input,omitempty"`
}

type ExecuteResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}
