// Package protocol provides event dispatch and routing interfaces.
// Phase 0 introduces stable interfaces with noop/local implementations.
// Phase 7 will provide full network-aware routing via NATS.
package protocol

import (
	"context"
	"sync"
)

// Dispatcher provides event dispatch functionality.
// Phase 0: Local in-memory implementation
// Phase 7: Full network-aware routing
type Dispatcher interface {
	// Dispatch dispatches an event to all subscribers whose filter matches.
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
// Phase 0: Returns a local in-memory dispatcher that delivers to subscribers.
func NewDispatcher() Dispatcher {
	return &localDispatcher{
		subs: make(map[chan Event]EventFilter),
	}
}

// localDispatcher is a Phase 0 local in-memory event dispatcher (spec §6.2, §6.3).
type localDispatcher struct {
	mu   sync.RWMutex
	subs map[chan Event]EventFilter
}

// Dispatch delivers the event to all subscribers whose filter matches (spec §6.2 presence.changed, roster.updated).
func (d *localDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	snapshot := make(map[chan Event]EventFilter, len(d.subs))
	for ch, filter := range d.subs {
		snapshot[ch] = filter
	}
	d.mu.RUnlock()
	for ch, filter := range snapshot {
		if !matchFilter(filter, event) {
			continue
		}
		select {
		case ch <- event:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Non-blocking: skip if channel full to avoid blocking dispatch
		}
	}
	return nil
}

func matchFilter(filter EventFilter, event Event) bool {
	if len(filter.Types) == 0 {
		return true
	}
	for _, t := range filter.Types {
		if t == event.Type {
			return true
		}
	}
	return false
}

func (d *localDispatcher) Subscribe(filter EventFilter) (<-chan Event, func()) {
	ch := make(chan Event, 10)
	d.mu.Lock()
	d.subs[ch] = filter
	d.mu.Unlock()
	unsub := func() {
		d.mu.Lock()
		delete(d.subs, ch)
		d.mu.Unlock()
		close(ch)
	}
	return ch, unsub
}

// MessageRouter routes inter-agent messages (spec §6.4).
// Phase 4: Local delivery via Dispatch message.inbound.
// Phase 7: NATS P.comm.* subjects.
type MessageRouter interface {
	// Route delivers the message to the recipient(s); for local Phase 4 implementation
	// this dispatches a message.inbound event so subscribers can deliver to PTYs.
	Route(ctx context.Context, msg interface{}) error
}

// messageRouter is the Phase 4 local implementation: dispatches message.inbound (spec §6.4.2).
type messageRouter struct {
	Dispatcher Dispatcher
}

// NewMessageRouter creates a local message router that dispatches message.inbound (spec §6.4).
func NewMessageRouter(d Dispatcher) MessageRouter {
	return &messageRouter{Dispatcher: d}
}

func (m *messageRouter) Route(ctx context.Context, msg interface{}) error {
	return m.Dispatcher.Dispatch(ctx, Event{Type: "message.inbound", Data: msg})
}
