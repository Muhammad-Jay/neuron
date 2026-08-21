package contracts

import (
	"context"

	"github.com/Muhammad-Jay/neuron/nore/internal/event"
)

type EventBus interface {
	Subscribe(eventType event.Type, buffer int) (event.Subscription, error)
	Publish(ctx context.Context, event event.Event) error
	Close() error
}
