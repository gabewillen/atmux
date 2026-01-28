// Package event provides stable interfaces for event system (to be fully implemented in Phase 7).
package event

import (
	"context"
)

// EventType represents the type of an event.
type EventType string

const (
	// Agent lifecycle events
	EventAgentSpawned    EventType = "agent.spawned"
	EventAgentStarted    EventType = "agent.started"
	EventAgentStopped    EventType = "agent.stopped"
	EventAgentTerminated EventType = "agent.terminated"
	EventAgentErrored    EventType = "agent.errored"

	// Connection events
	EventConnectionEstablished EventType = "connection.established"
	EventConnectionLost        EventType = "connection.lost"

	// Process events
	EventProcessSpawned EventType = "process.spawned"
	EventProcessExited  EventType = "process.exited"
	EventProcessIO      EventType = "process.io"

	// PTY events
	EventPTYActivity EventType = "pty.activity"
	EventPTYPattern  EventType = "pty.pattern"
	EventPTYDecoded  EventType = "pty.decoded"

	// Notification events
	EventNotificationBatch EventType = "notification.batch"
	EventNotificationError EventType = "notification.error"
)

// Event represents a system event with typed payload.
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	Timestamp int64                  `json:"timestamp"` // Unix timestamp
	Data      map[string]interface{} `json:"data"`
}

// Dispatcher handles event routing and delivery.
type Dispatcher interface {
	// Emit publishes an event to the event system
	Emit(ctx context.Context, event *Event) error

	// Subscribe registers for events of specific type(s)
	Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error

	// Unsubscribe removes an event subscription
	Unsubscribe(handler Handler) error

	// Shutdown gracefully shuts down the dispatcher
	Shutdown(ctx context.Context) error
}

// Handler processes incoming events.
type Handler interface {
	// Handle processes an event
	Handle(ctx context.Context, event *Event) error

	// HandlerID returns unique identifier for this handler
	HandlerID() string
}

// HandlerFunc adapts a function to Handler interface.
type HandlerFunc struct {
	id      string
	handler func(context.Context, *Event) error
}

// NewHandlerFunc creates a handler from a function.
func NewHandlerFunc(id string, fn func(context.Context, *Event) error) *HandlerFunc {
	return &HandlerFunc{
		id:      id,
		handler: fn,
	}
}

// Handle implements Handler interface.
func (h *HandlerFunc) Handle(ctx context.Context, event *Event) error {
	return h.handler(ctx, event)
}

// HandlerID implements Handler interface.
func (h *HandlerFunc) HandlerID() string {
	return h.id
}

// NoopDispatcher provides a no-op implementation for Phase 0.
type NoopDispatcher struct {
	handlers map[string]Handler
}

// NewNoopDispatcher creates a new no-op dispatcher.
func NewNoopDispatcher() *NoopDispatcher {
	return &NoopDispatcher{
		handlers: make(map[string]Handler),
	}
}

// Emit implements Dispatcher interface (no-op).
func (d *NoopDispatcher) Emit(ctx context.Context, event *Event) error {
	// TODO: implement actual event emission in Phase 7
	return nil
}

// Subscribe implements Dispatcher interface.
func (d *NoopDispatcher) Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error {
	d.handlers[handler.HandlerID()] = handler
	// TODO: implement actual subscription in Phase 7
	return nil
}

// Unsubscribe implements Dispatcher interface.
func (d *NoopDispatcher) Unsubscribe(handler Handler) error {
	delete(d.handlers, handler.HandlerID())
	return nil
}

// Shutdown implements Dispatcher interface.
func (d *NoopDispatcher) Shutdown(ctx context.Context) error {
	d.handlers = make(map[string]Handler)
	return nil
}
