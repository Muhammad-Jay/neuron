package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Message: "N.O.R.E. is healthy",
		Status:  http.StatusOK,
		Data:    map[string]any{"service": "nore"},
	})
}

func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	items := make([]map[string]any, 0)
	for _, i := range s.instances.List() {
		items = append(items, map[string]any{
			"id": i.ID, "status": i.Status(),
			"system_id": i.Key.SystemID, "version": i.Key.Version,
			"hash": i.Key.Hash, "env": i.Key.Env,
		})
	}
	writeJSON(w, http.StatusOK, Response{
		Message: "instances",
		Status:  http.StatusOK,
		Data:    items,
	})
}

func (s *Server) handleCreateInstance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key       instance.Key `json:"key"`
		System *core.System    `json:"system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}
	if req.Key.SystemID == "" {
		errorJSON(w, http.StatusBadRequest, fmt.Errorf("key.system_id is required"))
		return
	}
	if req.System == nil {
		errorJSON(w, http.StatusBadRequest, fmt.Errorf("system is required"))
		return
	}

	i, _, err := s.instances.GetOrCreate(req.Key, req.System)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Message: "instance ready",
		Status:  http.StatusOK,
		Data: map[string]any{
			"id": i.ID, "status": i.Status(),
			"system_id": i.Key.SystemID, "version": i.Key.Version,
			"hash": i.Key.Hash, "env": i.Key.Env,
		},
	})
}

func (s *Server) handleInstance(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.PathValue("id"))
	i, ok := s.instances.GetByID(id)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}
	writeJSON(w, http.StatusOK, Response{
		Message: "instance",
		Status:  http.StatusOK,
		Data: map[string]any{
			"id": i.ID, "status": i.Status(),
			"system_id": i.Key.SystemID, "version": i.Key.Version,
			"hash": i.Key.Hash, "env": i.Key.Env,
		},
	})
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.PathValue("id"))
	i, ok := s.instances.GetByID(id)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	var req struct {
		Input map[string]any `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
		return
	}

	execution, err := i.Execute(r.Context(), req.Input)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	// The request returns once the execution has been accepted. A separate
	// status/events API can be added without changing the execution model.
	writeJSON(w, http.StatusAccepted, Response{
		Message: "execution accepted",
		Status:  http.StatusAccepted,
		Data: map[string]any{
			"execution_id": execution.ID,
			"instance_id":  i.ID,
			"status":       "running",
		},
	})
}
