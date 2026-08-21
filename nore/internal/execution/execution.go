package execution

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/types"
	shared "github.com/Muhammad-Jay/neuron/shared/types/core"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type ServiceStatus string

const (
	ServicePending   ServiceStatus = "pending"
	ServiceReady     ServiceStatus = "ready"
	ServiceRunning   ServiceStatus = "running"
	ServiceCompleted ServiceStatus = "completed"
	ServiceFailed    ServiceStatus = "failed"
)

type ServiceExecutionState struct {
	Status      ServiceStatus
	StartedAt   *time.Time
	CompletedAt *time.Time
	Error       string
}

type Execution struct {
	ID             shared.ID
	CorrelationID  shared.ID
	Blueprint      *types.ExecutionBlueprint
	mu             sync.RWMutex
	status         Status
	initialInput   map[string]any
	inputs         map[shared.ID]map[string]any
	outputs        map[shared.ID]map[string]any
	states         map[shared.ID]ServiceExecutionState
	inFlight       int
	startedAt      *time.Time
	completedAt    *time.Time
	executionError string

	// done is closed exactly once when the execution reaches a terminal
	// state; Wait and Done poll it. It is a runtime-only device and is not
	// serialized in snapshots.
	done chan struct{}
}

func NewExecution(blueprint *types.ExecutionBlueprint, correlationID shared.ID) (*Execution, error) {
	if blueprint == nil {
		return nil, errors.New("execution blueprint is required")
	}
	if correlationID == "" {
		correlationID = shared.NewID("corr_")
	}
	states := make(map[shared.ID]ServiceExecutionState, len(blueprint.Nodes))
	for serviceID := range blueprint.Nodes {
		states[serviceID] = ServiceExecutionState{Status: ServicePending}
	}
	return &Execution{
		ID: shared.NewID("exec_"), CorrelationID: correlationID, Blueprint: blueprint,
		status: StatusPending, initialInput: make(map[string]any),
		inputs: make(map[shared.ID]map[string]any), outputs: make(map[shared.ID]map[string]any), states: states,
		done: make(chan struct{}),
	}, nil
}

// signalTerminal closes the done channel. It must be called exactly once,
// and only by a mutator that has already transitioned the execution into a
// terminal status (the bool-returning terminators are the sole valid callers).
func (e *Execution) signalTerminal() {
	select {
	case <-e.done:
	default:
		close(e.done)
	}
}

func (e *Execution) Start(initialInput map[string]any, initialServiceCount int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status != StatusPending {
		return fmt.Errorf("execution %s cannot start from status %s", e.ID, e.status)
	}
	if initialServiceCount <= 0 {
		return errors.New("execution must start with at least one entry service")
	}
	now := time.Now().UTC()
	e.status = StatusRunning
	e.startedAt = &now
	e.inFlight = initialServiceCount
	e.initialInput = cloneMap(initialInput)
	return nil
}

func (e *Execution) MarkServiceReady(serviceID shared.ID, input map[string]any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status != StatusRunning {
		return fmt.Errorf("execution %s is not running", e.ID)
	}
	state, exists := e.states[serviceID]
	if !exists {
		return fmt.Errorf("service %s is not in the blueprint", serviceID)
	}
	if state.Status != ServicePending {
		return fmt.Errorf("service %s cannot become ready from %s", serviceID, state.Status)
	}
	state.Status = ServiceReady
	e.states[serviceID] = state
	e.inputs[serviceID] = cloneMap(input)
	return nil
}

func (e *Execution) MarkServiceRunning(serviceID shared.ID) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, exists := e.states[serviceID]
	if !exists {
		return fmt.Errorf("service %s is not in the blueprint", serviceID)
	}
	if state.Status != ServiceReady {
		return fmt.Errorf("service %s cannot run from %s", serviceID, state.Status)
	}
	now := time.Now().UTC()
	state.Status = ServiceRunning
	state.StartedAt = &now
	e.states[serviceID] = state
	return nil
}

func (e *Execution) MarkServiceCompleted(serviceID shared.ID, output map[string]any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, exists := e.states[serviceID]
	if !exists {
		return fmt.Errorf("service %s is not in the blueprint", serviceID)
	}
	if state.Status != ServiceRunning {
		return fmt.Errorf("service %s cannot complete from %s", serviceID, state.Status)
	}
	now := time.Now().UTC()
	state.Status = ServiceCompleted
	state.CompletedAt = &now
	e.states[serviceID] = state
	e.outputs[serviceID] = cloneMap(output)
	return nil
}

func (e *Execution) MarkServiceFailed(serviceID shared.ID, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	state, exists := e.states[serviceID]
	if !exists {
		return
	}
	now := time.Now().UTC()
	state.Status = ServiceFailed
	state.CompletedAt = &now
	if err != nil {
		state.Error = err.Error()
	}
	e.states[serviceID] = state
}

func (e *Execution) CompleteCurrentAndSchedule(nextServiceCount int) (int, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status != StatusRunning {
		return e.inFlight, fmt.Errorf("execution %s is not running", e.ID)
	}
	if e.inFlight <= 0 {
		return e.inFlight, errors.New("execution in-flight counter is invalid")
	}
	e.inFlight += nextServiceCount - 1
	return e.inFlight, nil
}

func (e *Execution) MarkCompleted() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status != StatusRunning {
		return false
	}
	now := time.Now().UTC()
	e.status = StatusCompleted
	e.completedAt = &now
	e.signalTerminal()
	return true
}

func (e *Execution) MarkFailed(err error) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.status == StatusCompleted || e.status == StatusFailed || e.status == StatusCancelled {
		return false
	}
	now := time.Now().UTC()
	e.status = StatusFailed
	e.completedAt = &now
	if err != nil {
		e.executionError = err.Error()
	}
	e.signalTerminal()
	return true
}

func (e *Execution) IsTerminal() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status == StatusCompleted || e.status == StatusFailed || e.status == StatusCancelled
}

func (e *Execution) Input(serviceID shared.ID) map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return cloneMap(e.inputs[serviceID])
}

func (e *Execution) Output(serviceID shared.ID) map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return cloneMap(e.outputs[serviceID])
}

func (e *Execution) InitialInput() map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return cloneMap(e.initialInput)
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return map[string]any{}
	}
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = cloneValue(value)
	}
	return result
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = cloneValue(item)
		}
		return result
	default:
		return typed
	}
}
