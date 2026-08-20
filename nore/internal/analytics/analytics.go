package analytics

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	"github.com/Muhammad-Jay/neuron/nore/internal/event"
)

type Analytics struct {
	bus    contracts.EventBus
	logger *slog.Logger

	executionStarted event.Subscription
	executionFailed  event.Subscription
	serviceCompleted event.Subscription
	serviceFailed    event.Subscription
}

// New 2. Inject the logger via the constructor
func New(bus contracts.EventBus, logger *slog.Logger) (*Analytics, error) {
	if bus == nil {
		return nil, fmt.Errorf("event bus is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	started, err := bus.Subscribe(event.ExecutionStarted, 64)
	if err != nil {
		return nil, err
	}

	executionFailed, err :=bus.Subscribe(event.ExecutionFailed, 64)
	if err != nil {
		_ = started.Close()
		return nil, err
	}

	completed, err := bus.Subscribe(event.ServiceCompleted, 64)
	if err != nil {
		_ = started.Close()
		return nil, err
	}
	failed, err := bus.Subscribe(event.ServiceFailed, 64)
	if err != nil {
		_ = started.Close()
		_ = completed.Close()
		return nil, err
	}

	return &Analytics{
		bus:              bus,
		logger:           logger,
		executionStarted: started,
		executionFailed:  executionFailed,
		serviceCompleted: completed,
		serviceFailed:    failed,
	}, nil
}

func (a *Analytics) Serve(ctx context.Context) error {
	defer a.closeSubscriptions()

	for {
		select {
		case <-ctx.Done():
			return nil

		case received, open := <-a.executionStarted.Events():
			if !open {
				return nil
			}
			a.logger.Info("Execution started",
				slog.String("execution_id", string(received.Metadata.ExecutionID)),
				slog.String("correlation_id", string(received.Metadata.CorrelationID)),
				slog.Time("occurred_at", received.Metadata.OccurredAt.UTC()),
			)

		case received, open := <-a.executionFailed.Events():
			if !open {
				return nil
			}

			var errMsg string
			switch payload := received.Payload.(type) {
			case event.ExecutionFailedPayload:
				errMsg = payload.Message
			case string:
				errMsg = payload
			case fmt.Stringer:
				errMsg = payload.String()
			default:
				errMsg = fmt.Sprintf("%v", received.Payload)
			}

			a.logger.Info("Execution failed",
				slog.String("execution_id", string(received.Metadata.ExecutionID)),
				slog.String("correlation_id", string(received.Metadata.CorrelationID)),
				slog.String("error_message", errMsg),
				slog.Time("occurred_at", received.Metadata.OccurredAt.UTC()),
			)

		case received, open := <-a.serviceCompleted.Events():
			if !open {
				return nil
			}
			a.logger.Info("Service completed",
				slog.String("execution_id", string(received.Metadata.ExecutionID)),
				slog.String("service_id", string(received.Metadata.ServiceID)),
				slog.String("correlation_id", string(received.Metadata.CorrelationID)),
				slog.Time("occurred_at", received.Metadata.OccurredAt.UTC()),
			)

		case received, open := <-a.serviceFailed.Events():
			if !open {
				return nil
			}
			a.logger.Error("Service failed",
				slog.String("execution_id", string(received.Metadata.ExecutionID)),
				slog.String("service_id", string(received.Metadata.ServiceID)),
				slog.String("correlation_id", string(received.Metadata.CorrelationID)),
				slog.Time("occurred_at", received.Metadata.OccurredAt.UTC()),
			)
		}
	}
}

func (a *Analytics) closeSubscriptions() {
	_ = a.executionStarted.Close()
	_ = a.executionFailed.Close()
	_ = a.serviceCompleted.Close()
	_ = a.serviceFailed.Close()
}