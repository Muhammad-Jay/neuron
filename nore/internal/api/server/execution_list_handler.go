package server

import (
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleInstanceExecutions(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.PathValue("id"))
	i, ok := s.instances.GetByID(id)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	items := make([]executionItem, 0)
	for _, exec := range i.ListExecutions() {
		item := executionItem{
			ID:            exec.ID,
			CorrelationID: exec.CorrelationID,
			Status:        string(exec.Status()),
		}
		if started := exec.StartedAt(); started.IsZero() {
			ns := started.UnixNano()
			item.StartedAt = &ns
		}
		if completed := exec.CompletedAt(); completed.IsZero() {
			ns := completed.UnixNano()
			item.CompletedAt = &ns
		}
		if errMsg := exec.Error(); errMsg != "" {
			item.Error = errMsg
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, protocol.Response{
		Message: "executions",
		Status:  http.StatusOK,
		Data:    items,
	})
}