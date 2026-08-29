package register

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/utils"
	"github.com/Muhammad-Jay/neuron/nore/internal/system"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req protocol.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	key, err := resolveKey(req)
	if err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, err)
		return
	}

	if _, err := h.compiler.Compile(req.System); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, fmt.Errorf("invalid system: %w", err))
		return
	}

	now := time.Now().UTC()
	reg := system.RegisteredSystem{
		Key:                     key,
		System:                  req.System,
		ExecutionConfigurations: req.ExecutionConfigurations,
		RegisteredAt:            now,
		UpdatedAt:               now,
	}

	created, err := h.systems.Register(r.Context(), reg)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	status := protocol.RegisterStatusRegistered
	message := "system registered"
	if !created {
		status = protocol.RegisterStatusAlreadyRegistered
		message = "system already registered"
	}

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: message,
		Status:  http.StatusOK,
		Data: protocol.RegisterResponse{
			Key:     key,
			Status:  status,
			Message: message,
		},
	})
}

// resolveKey derives the durable identity for a registration: the client key
// is honored when provided (Hash is validated against the content), otherwise
// the key is derived from the system metadata and content hash.
func resolveKey(req protocol.RegisterRequest) (protocol.InstanceKey, error) {
	if req.Key.SystemID != "" {
		if req.Key.Hash != "" {
			computed, err := protocol.HashSystem(req.System)
			if err != nil {
				return protocol.InstanceKey{}, fmt.Errorf("hash system: %w", err)
			}
			if computed != req.Key.Hash {
				return protocol.InstanceKey{}, fmt.Errorf("key.hash does not match system content")
			}
		}
		if req.Key.Version == "" {
			req.Key.Version = "latest"
		}
		if req.Key.Env == "" {
			req.Key.Env = "development"
		}
		return req.Key, nil
	}

	key, err := protocol.SystemKey(req.System, req.Key.Env)
	if err != nil {
		return protocol.InstanceKey{}, err
	}
	return key, nil
}