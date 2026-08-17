package instances

import (
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/utils"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (h *Handler) GetInstanceByID(w http.ResponseWriter, r *http.Request) {
	id := utils.PathID(r.PathValue("id"))
	i, ok := h.instances.GetByID(id)
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: "instance",
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
