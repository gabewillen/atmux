// Package protocol provides event dispatch and routing interfaces.
// Phase 0 introduces stable interfaces with noop/local implementations.
// Phase 7 will provide full network-aware routing via NATS.
package protocol

import (
	"context"
)

// Dispatcher provides event dispatch functionality.
// Phase 0: Local/noop implementation
// Phase 7: Full network-aware routing
type Dispatcher interface {
	// Dispatch dispatches an event to all subscribers.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe subscribes to events matching the given filter.
	Subscribe(filter EventFilter) (<-chan Event, func())
}

// Event represents a generic event in the system.
type Event struct {
	Type string
	Data interface{}
}

// EventFilter defines criteria for event subscription.
type EventFilter struct {
	Types []string // Event types to subscribe to (empty = all)
}

// NewDispatcher creates a new event dispatcher.
// Phase 0: Returns a local in-memory dispatcher
func NewDispatcher() Dispatcher {
	return &localDispatcher{
		subs: make(map[chan Event]EventFilter),
	}
}

// localDispatcher is a Phase 0 local in-memory event dispatcher.
type localDispatcher struct {
	subs map[chan Event]EventFilter
}

func (d *localDispatcher) Dispatch(ctx context.Context, event Event) error {
	// TODO: Implement local dispatch to subscribers
	_ = ctx
	_ = event
	return nil
}

func (d *localDispatcher) Subscribe(filter EventFilter) (<-chan Event, func()) {
	ch := make(chan Event, 10)
	d.subs[ch] = filter
	unsub := func() {
		delete(d.subs, ch)
		close(ch)
	}
	return ch, unsub
}
