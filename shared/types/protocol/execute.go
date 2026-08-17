package protocol

import "github.com/Muhammad-Jay/neuron/shared/types/core"

type ExecutionItem struct {
	ID            core.ID  `json:"id"`
	CorrelationID core.ID  `json:"correlation_id"`
	Status        string   `json:"status"`
	StartedAt     *int64   `json:"started_at,omitempty"`
	CompletedAt   *int64   `json:"completed_at,omitempty"`
	Error         string  `json:"error,omitempty"`
}

type EventItem struct {
	ID            core.ID  `json:"id"`
	Type          string   `json:"type"`
	ServiceID     core.ID  `json:"service_id"`
	Payload       any      `json:"payload"`
}
