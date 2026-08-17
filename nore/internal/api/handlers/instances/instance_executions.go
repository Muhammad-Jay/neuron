package instances

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/utils"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	id := utils.PathID(r.PathValue("id"))
	i, ok := h.instances.GetByID(id)
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	items := make([]protocol.ExecutionItem, 0)
	for _, exec := range i.ListExecutions() {
		item := protocol.ExecutionItem{
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

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: "executions",
		Status:  http.StatusOK,
		Data:    items,
	})
}

func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
	id := utils.PathID(r.PathValue("id"))
	i, ok := h.instances.GetByID(id)
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", id))
		return
	}

	var body struct {
		Input map[string]any `json:"input,omitempty"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	input := utils.MergeMaps(utils.GetQueryParams(r), body.Input)

	execution, err := i.Execute(r.Context(), input)
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusAccepted, protocol.Response{
		Message: "execution accepted",
		Status:  http.StatusAccepted,
		Data: protocol.ExecuteResponse{
			ExecutionID: execution.ID,
			InstanceID:  i.ID,
			Status:      string(i.Status()),
		},
	})
}

func (h *Handler) GetExecutionState(w http.ResponseWriter, r *http.Request) {
	instID := utils.PathID(r.PathValue("id"))
	execID := utils.PathID(r.PathValue("execID"))

	i, ok := h.instances.GetByID(instID)
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", instID))
		return
	}

	exec, ok := i.GetExecution(core.ID(execID))
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("execution %s not found", execID))
		return
	}

	item := protocol.ExecutionItem{
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

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: "execution",
		Status:  http.StatusOK,
		Data:    item,
	})
}

func (h *Handler) GetExecutionEvents(w http.ResponseWriter, r *http.Request) {
	instID := utils.PathID(r.PathValue("id"))
	execID := utils.PathID(r.PathValue("execID"))

	i, ok := h.instances.GetByID(instID)
	if !ok {
		utils.ErrorJSON(w, http.StatusNotFound, fmt.Errorf("instance %s not found", instID))
		return
	}

	events, err := i.ListExecutionEvents(r.Context(), core.ID(execID))
	if err != nil {
		utils.ErrorJSON(w, http.StatusInternalServerError, err)
		return
	}

	items := make([]protocol.EventItem, 0, len(events))
	for _, evt := range events {
		items = append(items, protocol.EventItem{
			ID:        evt.Metadata.EventID,
			Type:      fmt.Sprintf("%d", evt.Type),
			ServiceID: evt.Metadata.ServiceID,
			Payload:   evt.Payload,
		})
	}

	utils.WriteJSON(w, http.StatusOK, protocol.Response{
		Message: "events",
		Status:  http.StatusOK,
		Data:    items,
	})
}
