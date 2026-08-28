package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
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
	System *core.System `json:"systems"`
}

type InstanceResponse struct {
	ID                string        `json:"id"`
	BlueprintMetadata core.Metadata `json:"blueprint_metadata"`
	Status            string        `json:"status"`
	SystemID          string        `json:"system_id"`
	Version           string        `json:"version,omitempty"`
	Hash              string        `json:"hash,omitempty"`
	Env               string        `json:"env,omitempty"`
}

type ExecuteRequest struct {
	// Input is the free-form data passed to the execution's entry services.
	Input map[string]any `json:"input,omitempty"`
	// Mode selects how the server responds. "detach" returns immediately with
	// the execution accepted (HTTP 202); anything else (empty or "wait") waits
	// for the execution to finish and returns its final result (HTTP 200).
	Mode string `json:"mode,omitempty"`
}

type ExecuteResponse struct {
	ExecutionID core.ID    `json:"execution_id"`
	InstanceID  string     `json:"instance_id"`
	Status      string     `json:"status"`
	Time        time.Time  `json:"time"`
}

// ExecutionResult is returned by a wait-mode Execute request once the
// execution has reached a terminal state. Outputs aggregates every service
// output keyed by service ID.
type ExecutionResult struct {
	ExecutionID core.ID                   `json:"execution_id"`
	InstanceID  string                    `json:"instance_id"`
	Status      string                    `json:"status"`
	Error       string                    `json:"error,omitempty"`
	Outputs     map[string]map[string]any `json:"outputs,omitempty"`
}

func HashBlueprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
