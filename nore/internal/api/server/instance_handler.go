package server

import (
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleInstance(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.PathValue("id"))
	i, ok := s.instances.GetByID(id)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	writeJSON(w, http.StatusOK, protocol.Response{
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