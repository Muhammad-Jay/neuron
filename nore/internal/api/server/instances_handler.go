package server

import (
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items := make([]protocol.InstanceResponse, 0)

	all := parseBool(q, "all", false)
	status := q.Get("status")

	for _, i := range s.instances.List(protocol.ListOptions{ Status: status, All: all }) {
		items = append(items, protocol.InstanceResponse{
			ID:       i.ID,
			Status:   string(i.Status()),
			SystemID: i.Key.SystemID,
			BlueprintMetadata:     i.Blueprint.Metadata,
			Version:  i.Key.Version,
			Hash:     i.Key.Hash,
			Env:      i.Key.Env,
		})
	}

	writeJSON(w, http.StatusOK, protocol.Response{
		Message: "instances",
		Status:  http.StatusOK,
		Data:    items,
	})
}