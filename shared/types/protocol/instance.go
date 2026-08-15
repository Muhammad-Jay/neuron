package protocol

import (
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

const (
	HealthPath    = "/health"
	InstancesPath = "/v1/instances"
	ExecutePath   = "/v1/instances/%s/executions"
)

type InstanceKey struct {
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

type CreateInstanceRequest struct {
	Key    InstanceKey  `json:"key"`
	System *core.System `json:"system"`
}

type InstanceResponse struct {
	ID       string `json:"id"`
	BlueprintMetadata     core.Metadata `json:"blueprint_metadata"`
	Status   string `json:"status"`
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

type ExecuteRequest struct {
	Input map[string]any `json:"input,omitempty"`
}

type ExecuteResponse struct {
	ExecutionID core.ID `json:"execution_id"`
	InstanceID  string `json:"instance_id"`
	Status      string `json:"status"`
}