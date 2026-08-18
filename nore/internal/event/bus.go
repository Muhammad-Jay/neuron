package event

import (
	"context"
	"errors"
	"sync"
)

var ErrBusClosed = errors.New("event bus is closed")

type Subscription interface {
	Events() <-chan Event
	Close() error
}

type subscription struct {
	id        uint64
	eventType Type
	channel   chan Event
	bus       *Bus
	mu        sync.RWMutex
	closed    bool
}

func (s *subscription) Events() <-chan Event { return s.channel }
func (s *subscription) Close() error         { return s.bus.unsubscribe(s) }

func (s *subscription) deliver(ctx context.Context, event Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil
	}
	select {
	case s.channel <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *subscription) closeLocal() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.channel)
}

type Bus struct {
	mu                 sync.RWMutex
	nextSubscriptionID uint64
	subscriptions      map[Type]map[uint64]*subscription
	closed             bool
}

func NewBus() *Bus {
	return &Bus{subscriptions: make(map[Type]map[uint64]*subscription)}
}

func (b *Bus) Subscribe(eventType Type, buffer int) (Subscription, error) {
	if buffer <= 0 {
		buffer = 64
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, ErrBusClosed
	}
	b.nextSubscriptionID++
	sub := &subscription{id: b.nextSubscriptionID, eventType: eventType, channel: make(chan Event, buffer), bus: b}
	if b.subscriptions[eventType] == nil {
		b.subscriptions[eventType] = make(map[uint64]*subscription)
	}
	b.subscriptions[eventType][sub.id] = sub
	return sub, nil
}

func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBusClosed
	}
	current := b.subscriptions[event.Type]
	all := b.subscriptions[All]
	subscribers := make([]*subscription, 0, len(current)+len(all))
	for _, sub := range current {
		subscribers = append(subscribers, sub)
	}
	for _, sub := range all {
		subscribers = append(subscribers, sub)
	}
	b.mu.RUnlock()
	for _, sub := range subscribers {
		if err := sub.deliver(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bus) unsubscribe(sub *subscription) error {
	b.mu.Lock()
	if !b.closed {
		current := b.subscriptions[sub.eventType]
		delete(current, sub.id)
		if len(current) == 0 {
			delete(b.subscriptions, sub.eventType)
		}
	}
	b.mu.Unlock()
	sub.closeLocal()
	return nil
}

func (b *Bus) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	var subscribers []*subscription
	for _, current := range b.subscriptions {
		for _, sub := range current {
			subscribers = append(subscribers, sub)
		}
	}
	b.subscriptions = nil
	b.mu.Unlock()
	for _, sub := range subscribers {
		sub.closeLocal()
	}
	return nil
}
