package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	id := pathID(r.PathValue("id"))
	i, ok := s.instances.GetByID(id)
	if !ok {
		errorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	var body struct {
		Input map[string]any `json:"input,omitempty"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	input := MergeMaps(GetQueryParams(r), body.Input)

	execution, err := i.Execute(r.Context(), input)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusAccepted, protocol.Response{
		Message: "execution accepted",
		Status:  http.StatusAccepted,
		Data: protocol.ExecuteResponse{
			ExecutionID: execution.ID,
			InstanceID:  i.ID,
			Status: string(i.Status()),
		},
	})
}

type executionItem struct {
	ID            core.ID  `json:"id"`
	CorrelationID core.ID  `json:"correlation_id"`
	Status        string   `json:"status"`
	StartedAt     *int64   `json:"started_at,omitempty"`
	CompletedAt   *int64   `json:"completed_at,omitempty"`
	Error         string  `json:"error,omitempty"`
}