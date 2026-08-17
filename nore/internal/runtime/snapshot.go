package runtime

import (
	"encoding/json"
	"time"

	"github.com/Muhammad-Jay/neuron/shared/types/core"
)

type ExecutionSnapshot struct {
	ID            core.ID                            `json:"id"`
	CorrelationID core.ID                            `json:"correlation_id"`
	Status        Status                             `json:"status"`
	InitialInput  map[string]any                     `json:"initial_input,omitempty"`
	Inputs        map[core.ID]map[string]any         `json:"inputs,omitempty"`
	Outputs       map[core.ID]map[string]any         `json:"outputs,omitempty"`
	States        map[core.ID]ServiceExecutionState  `json:"states,omitempty"`
	InFlight      int                                `json:"in_flight"`
	StartedAt     *time.Time                         `json:"started_at,omitempty"`
	CompletedAt   *time.Time                         `json:"completed_at,omitempty"`
	Error         string                             `json:"error,omitempty"`
}

func (e *Execution) Snapshot() *ExecutionSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	states := make(map[core.ID]ServiceExecutionState, len(e.states))
	for id, state := range e.states {
		states[id] = state
	}

	return &ExecutionSnapshot{
		ID:            e.ID,
		CorrelationID: e.CorrelationID,
		Status:        e.status,
		InitialInput:  cloneMap(e.initialInput),
		Inputs:        cloneInputsMap(e.inputs),
		Outputs:       cloneInputsMap(e.outputs),
		States:        states,
		InFlight:      e.inFlight,
		StartedAt:     e.startedAt,
		CompletedAt:   e.completedAt,
		Error:         e.executionError,
	}
}

func (e *Execution) Restore(snapshot *ExecutionSnapshot) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.status = snapshot.Status
	e.initialInput = cloneMap(snapshot.InitialInput)
	e.inputs = cloneInputsMap(snapshot.Inputs)
	e.outputs = cloneInputsMap(snapshot.Outputs)
	e.inFlight = snapshot.InFlight
	e.startedAt = snapshot.StartedAt
	e.completedAt = snapshot.CompletedAt
	e.executionError = snapshot.Error

	for id, state := range snapshot.States {
		e.states[id] = state
	}
}

func MarshalExecution(e *Execution) ([]byte, error) {
	return json.Marshal(e.Snapshot())
}

func UnmarshalExecution(data []byte) (*Execution, error) {
	var snap ExecutionSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	e := &Execution{
		ID:             snap.ID,
		CorrelationID:  snap.CorrelationID,
		status:         snap.Status,
		initialInput:   snap.InitialInput,
		inputs:         snap.Inputs,
		outputs:        snap.Outputs,
		states:         snap.States,
		inFlight:       snap.InFlight,
		startedAt:      snap.StartedAt,
		completedAt:    snap.CompletedAt,
		executionError: snap.Error,
	}
	return e, nil
}

func cloneInputsMap(source map[core.ID]map[string]any) map[core.ID]map[string]any {
	if source == nil {
		return make(map[core.ID]map[string]any)
	}
	result := make(map[core.ID]map[string]any, len(source))
	for id, m := range source {
		result[id] = cloneMap(m)
	}
	return result
}