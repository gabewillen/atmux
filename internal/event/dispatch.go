// Package event provides event dispatch interfaces for amux per spec §9.
//
// Phase 0: Provides stable interfaces with local implementations.
// Phase 7 will add full hsmnet network-aware dispatch.
package event

import (
	"context"
	"strings"
	"sync"
)

// Dispatcher is the interface for event dispatch.
type Dispatcher interface {
	// Dispatch sends an event to the event bus.
	Dispatch(ctx context.Context, event Event) error
	
	// Subscribe subscribes to events matching the given filter.
	Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error)
	
	// Close closes the dispatcher.
	Close() error
}

// Event represents a generic event in the system.
type Event interface {
	// Type returns the event type (e.g., "agent.started", "process.spawned").
	Type() string
	
	// Data returns the event payload.
	Data() any
}

// EventFilter filters events by type or other criteria.
type EventFilter interface {
	// Matches returns true if the event matches this filter.
	Matches(event Event) bool
}

// BasicEvent is a simple Event implementation used for local dispatch.
type BasicEvent struct {
	EventType string
	Payload   any
}

// Type implements Event.Type.
func (e BasicEvent) Type() string {
	return e.EventType
}

// Data implements Event.Data.
func (e BasicEvent) Data() any {
	return e.Payload
}

// TypeFilter matches events by type prefix.
type TypeFilter struct {
	// Prefix is matched against the event type using strings.HasPrefix.
	// An empty prefix matches all events.
	Prefix string
}

// Matches returns true if the event type starts with the configured prefix.
func (f TypeFilter) Matches(event Event) bool {
	if f.Prefix == "" {
		return true
	}
	return strings.HasPrefix(event.Type(), f.Prefix)
}

// NewLocalDispatcher creates a new local-only dispatcher.
// This is a Phase 0 implementation that keeps all dispatch in-process.
// Phase 7 will replace this with a network-aware dispatcher behind the same interface.
func NewLocalDispatcher() Dispatcher {
	return &localDispatcher{}
}

// subscriber represents a single subscription on the local dispatcher.
type subscriber struct {
	filter EventFilter
	ch     chan Event
}

// localDispatcher is an in-memory dispatcher suitable for single-process tests.
// It is safe for concurrent use by multiple goroutines.
type localDispatcher struct {
	mu          sync.RWMutex
	subscribers []subscriber
}

// Dispatch sends an event to all matching subscribers.
func (d *localDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	subs := make([]subscriber, len(d.subscribers))
	copy(subs, d.subscribers)
	d.mu.RUnlock()

	for _, sub := range subs {
		if sub.filter == nil || sub.filter.Matches(event) {
			select {
			case sub.ch <- event:
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Drop event for slow subscribers to avoid blocking the caller.
			}
		}
	}

	return nil
}

// Subscribe registers a new subscriber with the given filter.
// The returned channel is buffered to reduce the risk of blocking producers.
func (d *localDispatcher) Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error) {
	ch := make(chan Event, 16)

	d.mu.Lock()
	d.subscribers = append(d.subscribers, subscriber{filter: filter, ch: ch})
	d.mu.Unlock()

	// For the Phase 0 local dispatcher we ignore ctx for simplicity; callers
	// are expected to stop reading when done. Phase 7 will handle cancellation
	// and networked unsubscription semantics.
	_ = ctx

	return ch, nil
}

// Close closes all subscriber channels and clears internal state.
func (d *localDispatcher) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, sub := range d.subscribers {
		close(sub.ch)
	}
	d.subscribers = nil

	return nil
}
