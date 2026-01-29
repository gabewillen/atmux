// Package event provides event types and dispatch interfaces for amux.
//
// This package defines the core event system that enables event-driven
// architecture across all amux components. Events are dispatched through
// NATS subjects for both local and remote distribution.
//
// The package provides stable interfaces that can be implemented with
// either local/noop dispatch (during early development) or full
// network-aware dispatch (Phase 7 and beyond).
package event

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"
)

// Type represents an event type identifier.
type Type string

// Standard event types for the amux system.
const (
	// Agent lifecycle events
	TypeAgentAdded      Type = "agent.added"
	TypeAgentStarting   Type = "agent.starting"
	TypeAgentStarted    Type = "agent.started"
	TypeAgentStopping   Type = "agent.stopping"
	TypeAgentStopped    Type = "agent.stopped"
	TypeAgentTerminated Type = "agent.terminated"
	TypeAgentErrored    Type = "agent.errored"

	// Presence events
	TypePresenceChanged Type = "presence.changed"

	// PTY events
	TypePTYOutput   Type = "pty.output"
	TypePTYActivity Type = "pty.activity"
	TypePTYIdle     Type = "pty.idle"

	// Process events
	TypeProcessSpawned   Type = "process.spawned"
	TypeProcessCompleted Type = "process.completed"
	TypeProcessIO        Type = "process.io"

	// Config events
	TypeConfigFileChanged Type = "config.file_changed"
	TypeConfigReloaded    Type = "config.reloaded"
	TypeConfigUpdated     Type = "config.updated"
	TypeConfigReloadFailed Type = "config.reload_failed"

	// Connection events (remote)
	TypeConnectionEstablished Type = "connection.established"
	TypeConnectionLost        Type = "connection.lost"
	TypeConnectionRecovered   Type = "connection.recovered"

	// Shutdown events
	TypeShutdownInitiated Type = "shutdown.initiated"
	TypeShutdownForce     Type = "shutdown.force"
	TypeDrainComplete     Type = "drain.complete"
	TypeDrainTimeout      Type = "drain.timeout"
	TypeTerminateComplete Type = "terminate.complete"

	// Git merge events
	TypeGitMergeRequested Type = "git.merge.requested"
	TypeGitMergeCompleted Type = "git.merge.completed"
	TypeGitMergeConflict  Type = "git.merge.conflict"
	TypeGitMergeFailed    Type = "git.merge.failed"

	// Worktree events
	TypeWorktreeCreated Type = "worktree.created"
	TypeWorktreeRemoved Type = "worktree.removed"

	// Task events
	TypeTaskCancel Type = "task.cancel"

	// Adapter events
	TypeAdapterLoaded   Type = "adapter.loaded"
	TypeAdapterUnloaded Type = "adapter.unloaded"
)

// Event represents an immutable event record.
type Event struct {
	// ID is the unique event identifier.
	ID muid.MUID

	// Type is the event type.
	Type Type

	// Source is the ID of the entity that produced the event.
	Source muid.MUID

	// Target is the optional ID of the event target (for directed events).
	// Zero value means broadcast.
	Target muid.MUID

	// Timestamp is when the event was created (RFC 3339 UTC).
	Timestamp time.Time

	// Data is the event payload (type depends on event Type).
	Data any

	// TraceID is the optional trace context for observability.
	TraceID string
}

// NewEvent creates a new event with the given type and data.
func NewEvent(eventType Type, source muid.MUID, data any) Event {
	return Event{
		ID:        muid.Make(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// WithTarget returns a copy of the event with the target set.
func (e Event) WithTarget(target muid.MUID) Event {
	e.Target = target
	return e
}

// WithTraceID returns a copy of the event with the trace ID set.
func (e Event) WithTraceID(traceID string) Event {
	e.TraceID = traceID
	return e
}

// MarshalJSON implements json.Marshaler for wire encoding.
// IDs are encoded as base-10 strings per spec §4.2.3.
func (e Event) MarshalJSON() ([]byte, error) {
	type jsonEvent struct {
		ID        string    `json:"id"`
		Type      Type      `json:"type"`
		Source    string    `json:"source"`
		Target    string    `json:"target,omitempty"`
		Timestamp string    `json:"timestamp"`
		Data      any       `json:"data,omitempty"`
		TraceID   string    `json:"trace_id,omitempty"`
	}

	je := jsonEvent{
		ID:        e.ID.String(),
		Type:      e.Type,
		Source:    e.Source.String(),
		Timestamp: e.Timestamp.Format(time.RFC3339Nano),
		Data:      e.Data,
		TraceID:   e.TraceID,
	}

	if e.Target != 0 {
		je.Target = e.Target.String()
	}

	return json.Marshal(je)
}

// Handler is a function that handles an event.
type Handler func(ctx context.Context, event Event) error

// Subscription represents an event subscription.
type Subscription struct {
	// ID is the unique subscription identifier.
	ID muid.MUID

	// Types is the list of event types to receive (empty means all).
	Types []Type

	// Handler is called for matching events.
	Handler Handler
}

// Dispatcher is the interface for event dispatch.
// Implementations may use local dispatch or network-aware routing.
type Dispatcher interface {
	// Dispatch sends an event to all matching subscribers.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe registers a handler for events.
	// Returns a function to unsubscribe.
	Subscribe(sub Subscription) (unsubscribe func())

	// Close shuts down the dispatcher.
	Close() error
}

// LocalDispatcher is a simple in-process event dispatcher.
// This is used during early development and for testing.
type LocalDispatcher struct {
	mu           sync.RWMutex
	subscribers  map[muid.MUID]*Subscription
	closed       bool
}

// NewLocalDispatcher creates a new local event dispatcher.
func NewLocalDispatcher() *LocalDispatcher {
	return &LocalDispatcher{
		subscribers: make(map[muid.MUID]*Subscription),
	}
}

// Dispatch sends an event to all matching subscribers.
func (d *LocalDispatcher) Dispatch(ctx context.Context, event Event) error {
	d.mu.RLock()
	if d.closed {
		d.mu.RUnlock()
		return ErrDispatcherClosed
	}

	// Collect matching subscribers
	var handlers []Handler
	for _, sub := range d.subscribers {
		if d.matches(sub, event) {
			handlers = append(handlers, sub.Handler)
		}
	}
	d.mu.RUnlock()

	// Dispatch to handlers (outside lock)
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			// Log error but continue dispatching
			continue
		}
	}

	return nil
}

// matches checks if a subscription matches an event.
func (d *LocalDispatcher) matches(sub *Subscription, event Event) bool {
	// Check target (if set, must match)
	if event.Target != 0 && event.Target != sub.ID {
		return false
	}

	// Check type filter
	if len(sub.Types) == 0 {
		return true
	}

	for _, t := range sub.Types {
		if t == event.Type {
			return true
		}
	}

	return false
}

// Subscribe registers a handler for events.
func (d *LocalDispatcher) Subscribe(sub Subscription) (unsubscribe func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if sub.ID == 0 {
		sub.ID = muid.Make()
	}

	d.subscribers[sub.ID] = &sub

	return func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		delete(d.subscribers, sub.ID)
	}
}

// Close shuts down the dispatcher.
func (d *LocalDispatcher) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	d.subscribers = nil
	return nil
}

// ErrDispatcherClosed is returned when dispatching to a closed dispatcher.
var ErrDispatcherClosed = &DispatcherClosedError{}

// DispatcherClosedError indicates the dispatcher is closed.
type DispatcherClosedError struct{}

func (e *DispatcherClosedError) Error() string {
	return "event dispatcher is closed"
}

// NoopDispatcher is a no-op dispatcher for testing.
type NoopDispatcher struct{}

// NewNoopDispatcher creates a new no-op dispatcher.
func NewNoopDispatcher() *NoopDispatcher {
	return &NoopDispatcher{}
}

// Dispatch is a no-op.
func (d *NoopDispatcher) Dispatch(ctx context.Context, event Event) error {
	return nil
}

// Subscribe returns a no-op unsubscribe function.
func (d *NoopDispatcher) Subscribe(sub Subscription) func() {
	return func() {}
}

// Close is a no-op.
func (d *NoopDispatcher) Close() error {
	return nil
}

// DefaultDispatcher is the global event dispatcher.
var (
	defaultDispatcher Dispatcher = NewLocalDispatcher()
	dispatcherMu      sync.RWMutex
)

// SetDefaultDispatcher sets the global event dispatcher.
func SetDefaultDispatcher(d Dispatcher) {
	dispatcherMu.Lock()
	defer dispatcherMu.Unlock()
	defaultDispatcher = d
}

// DefaultDispatcher returns the global event dispatcher.
func GetDefaultDispatcher() Dispatcher {
	dispatcherMu.RLock()
	defer dispatcherMu.RUnlock()
	return defaultDispatcher
}

// Dispatch sends an event using the default dispatcher.
func Dispatch(ctx context.Context, event Event) error {
	return GetDefaultDispatcher().Dispatch(ctx, event)
}

// Subscribe registers a handler using the default dispatcher.
func Subscribe(sub Subscription) func() {
	return GetDefaultDispatcher().Subscribe(sub)
}
