package agent

import (
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
)

// EventType represents the type of event.
type EventType string

const (
	EventPresenceUpdate EventType = "presence.update"
	EventMessage        EventType = "message"
)

// BusEvent represents an event on the bus.
type BusEvent struct {
	Type    EventType
	Source  api.AgentID
	Payload interface{}
}

// Subscription is a channel for receiving events.
type Subscription struct {
	C      chan BusEvent
	cancel func()
}

// Close unsubscribes.
func (s *Subscription) Close() {
	s.cancel()
}

// EventBus manages subscriptions and event distribution.
type EventBus struct {
	mu   sync.RWMutex
	subs map[*Subscription]struct{}
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[*Subscription]struct{}),
	}
}

// Subscribe returns a subscription for all events (for now).
// In a real implementation, we'd filter by topic.
func (b *EventBus) Subscribe() *Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &Subscription{
		C: make(chan BusEvent, 100),
	}
	sub.cancel = func() {
		b.unsubscribe(sub)
	}
	b.subs[sub] = struct{}{}
	return sub
}

func (b *EventBus) unsubscribe(sub *Subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subs, sub)
	close(sub.C)
}

// Publish sends an event to all subscribers.
func (b *EventBus) Publish(event BusEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub := range b.subs {
		select {
		case sub.C <- event:
		default:
			// Drop if full to prevent blocking
		}
	}
}
