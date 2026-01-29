# package event

`import "github.com/agentflare-ai/amux/internal/event"`

Package event provides event types and dispatch interfaces for amux.

This package defines the core event system that enables event-driven
architecture across all amux components. Events are dispatched through
NATS subjects for both local and remote distribution.

The package provides stable interfaces that can be implemented with
either local/noop dispatch (during early development) or full
network-aware dispatch (Phase 7 and beyond).

- `ErrDispatcherClosed` — ErrDispatcherClosed is returned when dispatching to a closed dispatcher.
- `func Dispatch(ctx context.Context, event Event) error` — Dispatch sends an event using the default dispatcher.
- `func SetDefaultDispatcher(d Dispatcher)` — SetDefaultDispatcher sets the global event dispatcher.
- `func Subscribe(sub Subscription) func()` — Subscribe registers a handler using the default dispatcher.
- `type DispatcherClosedError` — DispatcherClosedError indicates the dispatcher is closed.
- `type Dispatcher` — Dispatcher is the interface for event dispatch.
- `type Event` — Event represents an immutable event record.
- `type Handler` — Handler is a function that handles an event.
- `type LocalDispatcher` — LocalDispatcher is a simple in-process event dispatcher.
- `type NoopDispatcher` — NoopDispatcher is a no-op dispatcher for testing.
- `type Subscription` — Subscription represents an event subscription.
- `type Type` — Type represents an event type identifier.

### Variables

#### ErrDispatcherClosed

```go
var ErrDispatcherClosed = &DispatcherClosedError{}
```

ErrDispatcherClosed is returned when dispatching to a closed dispatcher.


### Functions

#### Dispatch

```go
func Dispatch(ctx context.Context, event Event) error
```

Dispatch sends an event using the default dispatcher.

#### SetDefaultDispatcher

```go
func SetDefaultDispatcher(d Dispatcher)
```

SetDefaultDispatcher sets the global event dispatcher.

#### Subscribe

```go
func Subscribe(sub Subscription) func()
```

Subscribe registers a handler using the default dispatcher.


## type Dispatcher

```go
type Dispatcher interface {
	// Dispatch sends an event to all matching subscribers.
	Dispatch(ctx context.Context, event Event) error

	// Subscribe registers a handler for events.
	// Returns a function to unsubscribe.
	Subscribe(sub Subscription) (unsubscribe func())

	// Close shuts down the dispatcher.
	Close() error
}
```

Dispatcher is the interface for event dispatch.
Implementations may use local dispatch or network-aware routing.

### Variables

#### defaultDispatcher, dispatcherMu

```go
var (
	defaultDispatcher Dispatcher = NewLocalDispatcher()
	dispatcherMu      sync.RWMutex
)
```

DefaultDispatcher is the global event dispatcher.


### Functions returning Dispatcher

#### GetDefaultDispatcher

```go
func GetDefaultDispatcher() Dispatcher
```

DefaultDispatcher returns the global event dispatcher.


## type DispatcherClosedError

```go
type DispatcherClosedError struct{}
```

DispatcherClosedError indicates the dispatcher is closed.

### Methods

#### DispatcherClosedError.Error

```go
func () Error() string
```


## type Event

```go
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
```

Event represents an immutable event record.

### Functions returning Event

#### NewEvent

```go
func NewEvent(eventType Type, source muid.MUID, data any) Event
```

NewEvent creates a new event with the given type and data.


### Methods

#### Event.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON implements json.Marshaler for wire encoding.
IDs are encoded as base-10 strings per spec §4.2.3.

#### Event.WithTarget

```go
func () WithTarget(target muid.MUID) Event
```

WithTarget returns a copy of the event with the target set.

#### Event.WithTraceID

```go
func () WithTraceID(traceID string) Event
```

WithTraceID returns a copy of the event with the trace ID set.


## type Handler

```go
type Handler func(ctx context.Context, event Event) error
```

Handler is a function that handles an event.

## type LocalDispatcher

```go
type LocalDispatcher struct {
	mu          sync.RWMutex
	subscribers map[muid.MUID]*Subscription
	closed      bool
}
```

LocalDispatcher is a simple in-process event dispatcher.
This is used during early development and for testing.

### Functions returning LocalDispatcher

#### NewLocalDispatcher

```go
func NewLocalDispatcher() *LocalDispatcher
```

NewLocalDispatcher creates a new local event dispatcher.


### Methods

#### LocalDispatcher.Close

```go
func () Close() error
```

Close shuts down the dispatcher.

#### LocalDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

Dispatch sends an event to all matching subscribers.

#### LocalDispatcher.Subscribe

```go
func () Subscribe(sub Subscription) (unsubscribe func())
```

Subscribe registers a handler for events.

#### LocalDispatcher.matches

```go
func () matches(sub *Subscription, event Event) bool
```

matches checks if a subscription matches an event.


## type NoopDispatcher

```go
type NoopDispatcher struct{}
```

NoopDispatcher is a no-op dispatcher for testing.

### Functions returning NoopDispatcher

#### NewNoopDispatcher

```go
func NewNoopDispatcher() *NoopDispatcher
```

NewNoopDispatcher creates a new no-op dispatcher.


### Methods

#### NoopDispatcher.Close

```go
func () Close() error
```

Close is a no-op.

#### NoopDispatcher.Dispatch

```go
func () Dispatch(ctx context.Context, event Event) error
```

Dispatch is a no-op.

#### NoopDispatcher.Subscribe

```go
func () Subscribe(sub Subscription) func()
```

Subscribe returns a no-op unsubscribe function.


## type Subscription

```go
type Subscription struct {
	// ID is the unique subscription identifier.
	ID muid.MUID

	// Types is the list of event types to receive (empty means all).
	Types []Type

	// Handler is called for matching events.
	Handler Handler
}
```

Subscription represents an event subscription.

## type Type

```go
type Type string
```

Type represents an event type identifier.

### Constants

#### TypeAgentAdded, TypeAgentStarting, TypeAgentStarted, TypeAgentStopping, TypeAgentStopped, TypeAgentTerminated, TypeAgentErrored, TypePresenceChanged, TypePTYOutput, TypePTYActivity, TypePTYIdle, TypeProcessSpawned, TypeProcessCompleted, TypeProcessIO, TypeConfigFileChanged, TypeConfigReloaded, TypeConfigUpdated, TypeConfigReloadFailed, TypeConnectionEstablished, TypeConnectionLost, TypeConnectionRecovered, TypeShutdownInitiated, TypeShutdownForce, TypeDrainComplete, TypeDrainTimeout, TypeTerminateComplete, TypeGitMergeRequested, TypeGitMergeCompleted, TypeGitMergeConflict, TypeGitMergeFailed, TypeWorktreeCreated, TypeWorktreeRemoved, TypeTaskCancel, TypeAdapterLoaded, TypeAdapterUnloaded

```go
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
	TypeConfigFileChanged  Type = "config.file_changed"
	TypeConfigReloaded     Type = "config.reloaded"
	TypeConfigUpdated      Type = "config.updated"
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
```

Standard event types for the amux system.


