package instance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Key struct {
	SystemID string `json:"system_id"`
	Version  string `json:"version,omitempty"`
	Hash     string `json:"hash,omitempty"`
	Env      string `json:"env,omitempty"`
}

func (k Key) String() string {
	return fmt.Sprintf("%s@%s#%s:%s", k.SystemID, k.Version, k.Hash, k.Env)
}

func HashBlueprint(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
