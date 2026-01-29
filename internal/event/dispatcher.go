// Package event implements a basic event dispatcher that can be used by other packages
// This implementation will eventually be replaced with a full NATS-based implementation
package event

import (
	"context"
	"sync"
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

// BasicDispatcher is a simple in-memory event dispatcher
type BasicDispatcher struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewBasicDispatcher creates a new basic dispatcher
func NewBasicDispatcher() *BasicDispatcher {
	return &BasicDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

// Emit implements the Dispatcher interface
func (d *BasicDispatcher) Emit(ctx context.Context, event Event) error {
	d.mutex.RLock()
	handlers, exists := d.handlers[event.Type]
	d.mutex.RUnlock()

	if !exists {
		return nil // No handlers for this event type
	}

	// Call all handlers for this event type
	for _, handler := range handlers {
		// Run handlers concurrently but respect context cancellation
		go func(h EventHandler) {
			select {
			case <-ctx.Done():
				return
			default:
				// Ignore errors from individual handlers to prevent one bad handler from stopping others
				_ = h(ctx, event)
			}
		}(handler)
	}

	return nil
}

// Subscribe implements the Dispatcher interface
func (d *BasicDispatcher) Subscribe(eventType string, handler EventHandler) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.handlers[eventType] = append(d.handlers[eventType], handler)
	return nil
}

// Unsubscribe implements the Dispatcher interface
func (d *BasicDispatcher) Unsubscribe(eventType string, handler EventHandler) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	handlers, exists := d.handlers[eventType]
	if !exists {
		return nil // No handlers for this event type
	}

	// Find and remove the specific handler
	for i, h := range handlers {
		if &h == &handler { // Compare addresses to see if it's the same function
			// Remove the handler by slicing around it
			d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

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
var GlobalDispatcher Dispatcher = NewBasicDispatcher()

// EmitEvent is a convenience function to emit an event using the global dispatcher
func EmitEvent(ctx context.Context, eventType string, data map[string]interface{}) error {
	event := Event{
		Type:      eventType,
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