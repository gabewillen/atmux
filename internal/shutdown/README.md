# package shutdown

`import "github.com/agentflare-ai/amux/internal/shutdown"`

Package shutdown provides HSM-based graceful shutdown for amux.

Shutdown is modeled as an event-driven process using HSM transitions
per spec §5.6. The shutdown sequence is:

	Running → Draining → Stopped    (graceful)
	Running → Terminating → Stopped (forced)
	Draining → Terminating → Stopped (drain timeout or second signal)

See spec §5.6.1-§5.6.4 for the shutdown HSM and signal handling.

- `type Controller` — Controller manages the shutdown process for the amux system.
- `type State` — State represents the shutdown state.

## type Controller

```go
type Controller struct {
	mu sync.Mutex

	state        State
	drainTimeout time.Duration
	sessions     *session.Manager
	dispatcher   event.Dispatcher

	// done is closed when the system reaches StateStopped.
	done chan struct{}

	// drainTimer fires when the drain timeout expires.
	drainTimer *time.Timer
}
```

Controller manages the shutdown process for the amux system.

### Functions returning Controller

#### NewController

```go
func NewController(sessions *session.Manager, dispatcher event.Dispatcher, drainTimeout time.Duration) *Controller
```

NewController creates a new shutdown controller.


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

This transitions directly to Terminating state, killing all sessions.

#### Controller.RequestShutdown

```go
func () RequestShutdown(ctx context.Context)
```

RequestShutdown initiates a graceful shutdown (SIGTERM/SIGINT handler).

This transitions the system from Running to Draining. All agents receive
a shutdown.initiated event and have drainTimeout to terminate gracefully.
If already draining, this escalates to forced termination.

See spec §5.6.2 for signal mapping.

#### Controller.State

```go
func () State() State
```

State returns the current shutdown state.

#### Controller.onDrainComplete

```go
func () onDrainComplete(ctx context.Context)
```

onDrainComplete is called when all sessions have stopped during drain.

#### Controller.onDrainTimeout

```go
func () onDrainTimeout(ctx context.Context)
```

onDrainTimeout is called when the drain timeout expires.

#### Controller.transitionToDraining

```go
func () transitionToDraining(ctx context.Context)
```

transitionToDraining moves to draining state. Caller must hold mu.

#### Controller.transitionToStopped

```go
func () transitionToStopped(ctx context.Context)
```

transitionToStopped moves to stopped state.

#### Controller.transitionToTerminating

```go
func () transitionToTerminating(ctx context.Context)
```

transitionToTerminating moves to terminating state. Caller must hold mu.


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


