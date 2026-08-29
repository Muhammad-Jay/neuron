package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type InstanceKey struct {
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

// VersionLatest is the pseudo-version that resolves to the most recently
// registered version of a system. It is used when a key omits the version or
// explicitly asks for "latest". Resolution is time-based: whatever version was
// registered last wins, regardless of semver ordering.
const VersionLatest = "latest"

func (k InstanceKey) String() string {
	return fmt.Sprintf("%s@%s#%s:%s", k.SystemID, k.Version, k.Hash, k.Env)
}

// String returns the colon-encoded form used on the wire and in URL paths:
// systemID:version:hash[:env]. Env is optional and defaults to development.
func (k InstanceKey) ColonString() string {
	version := k.Version
	if version == "" {
		version = "latest"
	}
	env := k.Env
	if env == "" {
		env = "development"
	}
	return fmt.Sprintf("%s:%s:%s:%s", k.SystemID, version, k.Hash, env)
}

// ParseKey parses a colon-encoded InstanceKey: systemID[:version[:hash[:env]]].
// Missing version and hash default to latest and ""; an omitted env is left
// empty so the server can resolve the registration regardless of environment.
func ParseKey(s string) (InstanceKey, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return InstanceKey{}, fmt.Errorf("instance key is empty")
	}
	parts := strings.Split(s, ":")
	if len(parts) > 4 {
		return InstanceKey{}, fmt.Errorf("instance key %q has too many segments", s)
	}

	key := InstanceKey{
		SystemID: parts[0],
		Version:  VersionLatest,
	}
	if key.SystemID == "" {
		return InstanceKey{}, fmt.Errorf("instance key %q has no system id", s)
	}
	if len(parts) > 1 {
		key.Version = parts[1]
	}
	if len(parts) > 2 {
		key.Hash = parts[2]
	}
	if len(parts) > 3 {
		key.Env = parts[3]
	}
	return key, nil
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
