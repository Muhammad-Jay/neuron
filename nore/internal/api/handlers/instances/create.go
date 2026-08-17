package instances

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/utils"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	// 2. Use the protocol struct instead of an anonymous struct
	var req protocol.CreateInstanceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	if req.Key.SystemID == "" {
		utils.ErrorJSON(w, http.StatusBadRequest, fmt.Errorf("key.system_id is required"))
		return
	}
	if req.System == nil {
		utils.ErrorJSON(w, http.StatusBadRequest, fmt.Errorf("system is required"))
		return
	}

	// Map protocol DTO to internal domain model
	domainKey := protocol.InstanceKey{
		SystemID: req.Key.SystemID,
		Version:  req.Key.Version,
		Hash:     req.Key.Hash,
		Env:      req.Key.Env,
	}

	i, _, err := h.instances.GetOrCreate(domainKey, req.System)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: "instance ready",
		Status:  http.StatusOK,
		Data: protocol.InstanceResponse{
			ID:       i.ID,
			Status:   string(i.Status()),
			SystemID: i.Key.SystemID,
			Version:  i.Key.Version,
			Hash:     i.Key.Hash,
			Env:      i.Key.Env,
		},
	})
}
