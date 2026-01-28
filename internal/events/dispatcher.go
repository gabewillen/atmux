// Package events provides event dispatch functionality.
// This package provides stable interfaces for event emission and subscription
// with noop implementations to unblock phased work.
package events

import (
	"context"
	"errors"
	"fmt"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
)

// Common sentinel errors for event operations.
var (
	// ErrDispatcherNotConnected indicates the dispatcher is not connected.
	ErrDispatcherNotConnected = errors.New("dispatcher not connected")

	// ErrSubscriptionFailed indicates subscription setup failed.
	ErrSubscriptionFailed = errors.New("subscription failed")

	// ErrEventEmitFailed indicates event emission failed.
	ErrEventEmitFailed = errors.New("event emit failed")
)

// Handler represents an event handler function.
type Handler func(ctx context.Context, event api.Event) error

// Dispatcher provides event dispatch functionality.
// Phase 0 provides a noop implementation to unblock later phases.
type Dispatcher struct {
	connected bool
	handlers  map[string][]Handler
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		connected: false,
		handlers:  make(map[string][]Handler),
	}
}

// Connect establishes the dispatcher connection.
// Phase 0: noop implementation.
func (d *Dispatcher) Connect(ctx context.Context) error {
	d.connected = true
	return nil
}

// Emit sends an event through the dispatch system.
// Phase 0: noop implementation that accepts but doesn't distribute events.
func (d *Dispatcher) Emit(ctx context.Context, event api.Event) error {
	if !d.connected {
		return fmt.Errorf("dispatcher not connected: %w", ErrDispatcherNotConnected)
	}

	// Phase 0: Accept events but don't actually dispatch them
	// Real implementation will be added in Phase 7
	return nil
}

// Subscribe registers a handler for events of the given type.
// Phase 0: noop implementation that accepts but doesn't invoke handlers.
func (d *Dispatcher) Subscribe(eventType string, handler Handler) error {
	if !d.connected {
		return fmt.Errorf("dispatcher not connected: %w", ErrDispatcherNotConnected)
	}

	if d.handlers[eventType] == nil {
		d.handlers[eventType] = make([]Handler, 0)
	}
	d.handlers[eventType] = append(d.handlers[eventType], handler)

	// Phase 0: Accept subscriptions but handlers won't be invoked yet
	// Real implementation will be added in Phase 7
	return nil
}

// Close shuts down the dispatcher.
func (d *Dispatcher) Close() error {
	d.connected = false
	d.handlers = make(map[string][]Handler)
	return nil
}