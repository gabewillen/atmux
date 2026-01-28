// Package event provides event dispatch interfaces for amux per spec §9.
//
// Phase 0: Provides stable interfaces with local/noop implementations.
// Phase 7 will add full hsmnet network-aware dispatch.
package event

import (
	"context"
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

// NewLocalDispatcher creates a new local-only dispatcher.
// This is a Phase 0 stub that will be enhanced in Phase 7.
func NewLocalDispatcher() Dispatcher {
	return &noopDispatcher{}
}

// noopDispatcher is a no-op implementation for Phase 0.
type noopDispatcher struct{}

func (d *noopDispatcher) Dispatch(ctx context.Context, event Event) error {
	// Phase 0: No-op
	return nil
}

func (d *noopDispatcher) Subscribe(ctx context.Context, filter EventFilter) (<-chan Event, error) {
	// Phase 0: Return closed channel
	ch := make(chan Event)
	close(ch)
	return ch, nil
}

func (d *noopDispatcher) Close() error {
	return nil
}
