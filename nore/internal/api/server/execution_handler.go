package server

import (
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleInstanceExecution(w http.ResponseWriter, r *http.Request) {
	instID := pathID(r.PathValue("id"))
	execID := pathID(r.PathValue("execID"))

	i, ok := s.instances.GetByID(instID)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", instID))
		return
	}

	exec, ok := i.GetExecution(core.ID(execID))
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("execution %s not found", execID))
		return
	}

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

	writeJSON(w, http.StatusOK, protocol.Response{
		Message: "execution",
		Status:  http.StatusOK,
		Data:    item,
	})
}

type eventItem struct {
	ID      core.ID `json:"id"`
	Type    string  `json:"type"`
	Service string  `json:"service,omitempty"`
	Payload any     `json:"payload,omitempty"`
}

func (s *Server) handleExecutionEvents(w http.ResponseWriter, r *http.Request) {
	instID := pathID(r.PathValue("id"))
	execID := pathID(r.PathValue("execID"))

	i, ok := s.instances.GetByID(instID)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", instID))
		return
	}

	events, err := i.ListExecutionEvents(r.Context(), core.ID(execID))
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	items := make([]eventItem, 0, len(events))
	for _, evt := range events {
		items = append(items, eventItem{
			ID:      evt.Metadata.EventID,
			Type:    fmt.Sprintf("%d", evt.Type),
			Service: string(evt.Metadata.ServiceID),
			Payload: evt.Payload,
		})
	}

	writeJSON(w, http.StatusOK, protocol.Response{
		Message: "events",
		Status:  http.StatusOK,
		Data:    items,
	})
}