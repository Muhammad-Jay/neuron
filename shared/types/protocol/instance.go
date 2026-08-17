package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

const (
	HealthPath    = "/health"
	InstancesPath = "/v1/instances"
	InstanceByIDPath = "/v1/instances/%s"
	ExecutePath   = "/v1/instances/%s/executions"
)

type InstanceKey struct {
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

func (k InstanceKey) String() string {
	return fmt.Sprintf("%s@%s#%s:%s", k.SystemID, k.Version, k.Hash, k.Env)
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

func HashBlueprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
