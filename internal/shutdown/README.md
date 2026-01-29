# package shutdown

`import "github.com/agentflare-ai/amux/internal/shutdown"`

Package shutdown provides HSM-based graceful shutdown for amux.

Shutdown is modeled as an event-driven process using HSM transitions
per spec §5.6. The shutdown sequence is:

	Running → Draining → Stopped    (graceful)
	Running → Terminating → Stopped (forced)
	Draining → Terminating → Stopped (drain timeout or second signal)

See spec §5.6.1-§5.6.4 for the shutdown HSM and signal handling.

- `EventShutdownRequest, EventShutdownForce, EventDrainComplete, EventDrainTimeout, EventTerminateComplete` — HSM event names for shutdown transitions per spec §5.6.1-§5.6.2.
- `ShutdownModel` — ShutdownModel defines the HSM model for system shutdown per spec §5.6.1.
- `type Controller` — Controller manages the shutdown process for the amux system using an HSM-driven state machine per spec §5.6.1.
- `type State` — State represents the shutdown state.

### Constants

#### EventShutdownRequest, EventShutdownForce, EventDrainComplete, EventDrainTimeout, EventTerminateComplete

```go
const (
	EventShutdownRequest   = "shutdown.request"
	EventShutdownForce     = "shutdown.force"
	EventDrainComplete     = "drain.complete"
	EventDrainTimeout      = "drain.timeout"
	EventTerminateComplete = "terminate.complete"
)
```

HSM event names for shutdown transitions per spec §5.6.1-§5.6.2.


### Variables

#### ShutdownModel

```go
var ShutdownModel = hsm.Define(
	"system.shutdown",

	hsm.State("running"),
	hsm.State("draining",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterDraining(ctx)
		}),
	),
	hsm.State("terminating",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterTerminating(ctx)
		}),
	),
	hsm.State("stopped",
		hsm.Entry(func(ctx context.Context, c *Controller, e hsm.Event) {
			c.onEnterStopped(ctx)
		}),
	),

	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownRequest}),
		hsm.Source("running"), hsm.Target("draining")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source("running"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventShutdownForce}),
		hsm.Source("draining"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventDrainComplete}),
		hsm.Source("draining"), hsm.Target("stopped")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventDrainTimeout}),
		hsm.Source("draining"), hsm.Target("terminating")),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTerminateComplete}),
		hsm.Source("terminating"), hsm.Target("stopped")),

	hsm.Initial(hsm.Target("running")),
)
```

ShutdownModel defines the HSM model for system shutdown per spec §5.6.1.


## type Controller

```go
type Controller struct {
	hsm.HSM

	mu           sync.Mutex
	state        State
	drainTimeout time.Duration
	sessions     *session.Manager
	dispatcher   event.Dispatcher
	requested    bool

	// done is closed when the system reaches StateStopped.
	done chan struct{}
	// doneOnce prevents double-close of the done channel.
	doneOnce sync.Once

	// drainTimer fires when the drain timeout expires.
	drainTimer *time.Timer
}
```

Controller manages the shutdown process for the amux system using
an HSM-driven state machine per spec §5.6.1.

### Functions returning Controller

#### NewController

```go
func NewController(sessions *session.Manager, dispatcher event.Dispatcher, drainTimeout time.Duration) *Controller
```

NewController creates a new shutdown controller. The HSM is initialized
in the Running state per spec §5.6.1.


### Methods

#### Controller.Done

```go
func () Done() <-chan struct{}
```

Done returns a channel that is closed when the system reaches StateStopped.

#### Controller.ForceShutdown

```go
func () ForceShutdown(ctx context.Context)
```

ForceShutdown forces immediate termination.

This dispatches shutdown.force, transitioning to Terminating from
either Running or Draining state.

#### Controller.RequestShutdown

```go
func () RequestShutdown(ctx context.Context)
```

RequestShutdown initiates a graceful shutdown (SIGTERM/SIGINT handler).

First call dispatches shutdown.request (Running → Draining).
Second call dispatches shutdown.force (Draining → Terminating).

See spec §5.6.2 for signal mapping.

#### Controller.ShutdownState

```go
func () ShutdownState() State
```

ShutdownState returns the current shutdown state.
Named ShutdownState (not State) to avoid shadowing hsm.HSM.State()
which is required by the hsm.Instance interface.

#### Controller.onEnterDraining

```go
func () onEnterDraining(ctx context.Context)
```

onEnterDraining is the entry action for the draining state.
It dispatches shutdown.initiated, stops all sessions gracefully,
and starts the drain timeout timer per spec §5.6.3-§5.6.4.

#### Controller.onEnterStopped

```go
func () onEnterStopped(ctx context.Context)
```

onEnterStopped is the entry action for the stopped state.
It closes the done channel to signal shutdown completion.

#### Controller.onEnterTerminating

```go
func () onEnterTerminating(ctx context.Context)
```

onEnterTerminating is the entry action for the terminating state.
It cancels the drain timer, dispatches shutdown.force, and kills
all sessions per spec §5.6.4.


## type State

```go
type State string
```

State represents the shutdown state.

### Constants

#### StateRunning, StateDraining, StateTerminating, StateStopped

```go
const (
	// StateRunning is the normal operating state.
	StateRunning State = "running"

	// StateDraining indicates graceful shutdown is in progress.
	StateDraining State = "draining"

	// StateTerminating indicates forced termination is in progress.
	StateTerminating State = "terminating"

	// StateStopped indicates the system has fully stopped.
	StateStopped State = "stopped"
)
```


