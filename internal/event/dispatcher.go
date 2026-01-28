// Package event implements a stable interface for event dispatch that can be used by other packages
// during Phase 0 before the full implementation is complete in Phase 7.
package event

import (
	"context"
)

// Event represents a system event
type Event struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Dispatcher defines the interface for dispatching events
type Dispatcher interface {
	// Emit emits an event to all interested listeners
	Emit(ctx context.Context, event Event) error

	// Subscribe registers a listener for events of a specific type
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe removes a listener for events of a specific type
	Unsubscribe(eventType string, handler EventHandler) error
}

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

// NoopDispatcher is a no-op implementation of the Dispatcher interface
type NoopDispatcher struct{}

// NewNoopDispatcher creates a new no-op dispatcher
func NewNoopDispatcher() *NoopDispatcher {
	return &NoopDispatcher{}
}

// Emit implements the Dispatcher interface
func (d *NoopDispatcher) Emit(ctx context.Context, event Event) error {
	// In the noop implementation, we just return nil (success) without doing anything
	return nil
}

// Subscribe implements the Dispatcher interface
func (d *NoopDispatcher) Subscribe(eventType string, handler EventHandler) error {
	// In the noop implementation, we just return nil (success) without doing anything
	return nil
}

// Unsubscribe implements the Dispatcher interface
func (d *NoopDispatcher) Unsubscribe(eventType string, handler EventHandler) error {
	// In the noop implementation, we just return nil (success) without doing anything
	return nil
}

// GlobalDispatcher is a global instance of the dispatcher that can be used by other packages
var GlobalDispatcher Dispatcher = NewNoopDispatcher()

// EmitEvent is a convenience function to emit an event using the global dispatcher
func EmitEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := Event{
		Type: eventType,
		// In a real implementation, we'd use the actual timestamp
		Timestamp: 0,
		Data:      data,
	}

	return GlobalDispatcher.Emit(ctx, event)
}

// SubscribeToEvent is a convenience function to subscribe to an event type
func SubscribeToEvent(eventType string, handler EventHandler) error {
	return GlobalDispatcher.Subscribe(eventType, handler)
}

// UnsubscribeFromEvent is a convenience function to unsubscribe from an event type
func UnsubscribeFromEvent(eventType string, handler EventHandler) error {
	return GlobalDispatcher.Unsubscribe(eventType, handler)
}