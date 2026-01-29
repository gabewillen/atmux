# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

Package agent provides agent orchestration for amux.

This package implements agent lifecycle management, presence tracking,
and messaging. All operations are agent-agnostic; agent-specific behavior
is delegated to adapters.

See spec §5 for agent management requirements.

Lifecycle provides the HSM-based agent lifecycle state machine.

The lifecycle HSM implements the state transitions defined in spec §5.4:

	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
	└─────────┘    └─────────┘    └─────────┘    └────────────┘
	                                  │
	                                  ▼
	                             ┌─────────┐
	                             │ Errored │
	                             └─────────┘

Transitions:
  - "start" event: Pending → Starting
  - "ready" event: Starting → Running
  - "stop" event: Running → Terminated
  - "error" event: Any → Errored

Presence provides the HSM-based agent presence state machine.

The presence HSM implements the state transitions defined in spec §6.1 and §6.5:

	                    ┌──────────────────┐
	                    ▼                  │
	┌────────┐    ┌─────────┐    ┌────────┐
	│ Online │◀──▶│  Busy   │───▶│ Offline│
	└────────┘    └─────────┘    └────────┘
	     ▲              │              │
	     │              ▼              │
	     │         ┌────────┐          │
	     └─────────│  Away  │◀─────────┘
	               └────────┘

Transitions:
  - task.assigned: Online → Busy
  - task.completed: Busy → Online
  - prompt.detected: Busy → Online
  - rate.limit: * → Offline
  - rate.cleared: Offline → Online
  - stuck.detected: * → Away
  - activity.detected: Away → Online

- `LifecycleEventStart, LifecycleEventReady, LifecycleEventStop, LifecycleEventError` — LifecycleEvent names for lifecycle state transitions.
- `LifecycleModel` — LifecycleModel defines the HSM model for agent lifecycle.
- `PresenceEventTaskAssigned, PresenceEventTaskCompleted, PresenceEventPromptDetected, PresenceEventRateLimit, PresenceEventRateCleared, PresenceEventStuckDetected, PresenceEventActivityDetected` — PresenceEvent names for presence state transitions.
- `PresenceModel` — PresenceModel defines the HSM model for agent presence.
- `func DispatchActivityDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchActivityDetected sends an "activity.detected" event to transition from Away to Online.
- `func DispatchError(ctx context.Context, instance hsm.Instance, err error) <-chan struct{}` — DispatchError sends an "error" event to transition to Errored state.
- `func DispatchPromptDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchPromptDetected sends a "prompt.detected" event to transition from Busy to Online.
- `func DispatchRateCleared(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchRateCleared sends a "rate.cleared" event to transition from Offline to Online.
- `func DispatchRateLimit(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchRateLimit sends a "rate.limit" event to transition to Offline state.
- `func DispatchReady(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchReady sends a "ready" event to transition from Starting to Running.
- `func DispatchStart(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStart sends a "start" event to transition from Pending to Starting.
- `func DispatchStop(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStop sends a "stop" event to transition from Running to Terminated.
- `func DispatchStuckDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchStuckDetected sends a "stuck.detected" event to transition to Away state.
- `func DispatchTaskAssigned(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchTaskAssigned sends a "task.assigned" event to transition from Online to Busy.
- `func DispatchTaskCompleted(ctx context.Context, instance hsm.Instance) <-chan struct{}` — DispatchTaskCompleted sends a "task.completed" event to transition from Busy to Online.
- `type Agent` — Agent represents a managed agent instance.
- `type LifecycleHSM` — LifecycleHSM wraps an agent with HSM-driven lifecycle management.
- `type Manager` — Manager manages agents.
- `type PresenceHSM` — PresenceHSM wraps an agent with HSM-driven presence management.

### Constants

#### LifecycleEventStart, LifecycleEventReady, LifecycleEventStop, LifecycleEventError

```go
const (
	LifecycleEventStart = "start" // Pending → Starting
	LifecycleEventReady = "ready" // Starting → Running
	LifecycleEventStop  = "stop"  // Running → Terminated
	LifecycleEventError = "error" // Any → Errored
)
```

LifecycleEvent names for lifecycle state transitions.

#### PresenceEventTaskAssigned, PresenceEventTaskCompleted, PresenceEventPromptDetected, PresenceEventRateLimit, PresenceEventRateCleared, PresenceEventStuckDetected, PresenceEventActivityDetected

```go
const (
	PresenceEventTaskAssigned     = "task.assigned"     // Online → Busy
	PresenceEventTaskCompleted    = "task.completed"    // Busy → Online
	PresenceEventPromptDetected   = "prompt.detected"   // Busy → Online
	PresenceEventRateLimit        = "rate.limit"        // * → Offline
	PresenceEventRateCleared      = "rate.cleared"      // Offline → Online
	PresenceEventStuckDetected    = "stuck.detected"    // * → Away
	PresenceEventActivityDetected = "activity.detected" // Away → Online
)
```

PresenceEvent names for presence state transitions.


### Variables

#### LifecycleModel

```go
var LifecycleModel = hsm.Define(
	"agent.lifecycle",

	hsm.State("pending"),
	hsm.State("starting",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterStarting(ctx)
		}),
	),
	hsm.State("running",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterRunning(ctx)
		}),
		hsm.Exit(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onExitRunning(ctx)
		}),
	),
	hsm.State("terminated",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterTerminated(ctx)
		}),
	),
	hsm.State("errored",
		hsm.Entry(func(ctx context.Context, l *LifecycleHSM, e hsm.Event) {
			l.onEnterErrored(ctx, e)
		}),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStart}),
		hsm.Source("pending"),
		hsm.Target("starting"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventReady}),
		hsm.Source("starting"),
		hsm.Target("running"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventStop}),
		hsm.Source("running"),
		hsm.Target("terminated"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("pending"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("starting"),
		hsm.Target("errored"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: LifecycleEventError}),
		hsm.Source("running"),
		hsm.Target("errored"),
	),

	hsm.Initial(
		hsm.Target("pending"),
	),
)
```

LifecycleModel defines the HSM model for agent lifecycle.
See spec §5.4.

#### PresenceModel

```go
var PresenceModel = hsm.Define(
	"agent.presence",

	hsm.State("online",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOnline(ctx, e)
		}),
	),
	hsm.State("busy",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterBusy(ctx, e)
		}),
	),
	hsm.State("offline",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterOffline(ctx, e)
		}),
	),
	hsm.State("away",
		hsm.Entry(func(ctx context.Context, p *PresenceHSM, e hsm.Event) {
			p.onEnterAway(ctx, e)
		}),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskAssigned}),
		hsm.Source("online"),
		hsm.Target("busy"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventTaskCompleted}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventPromptDetected}),
		hsm.Source("busy"),
		hsm.Target("online"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("online"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateLimit}),
		hsm.Source("busy"),
		hsm.Target("offline"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventRateCleared}),
		hsm.Source("offline"),
		hsm.Target("online"),
	),

	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("online"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("busy"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventStuckDetected}),
		hsm.Source("offline"),
		hsm.Target("away"),
	),
	hsm.Transition(
		hsm.On(hsm.Event{Name: PresenceEventActivityDetected}),
		hsm.Source("away"),
		hsm.Target("online"),
	),

	hsm.Initial(
		hsm.Target("online"),
	),
)
```

PresenceModel defines the HSM model for agent presence.
See spec §6.1 and §6.5.


### Functions

#### DispatchActivityDetected

```go
func DispatchActivityDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchActivityDetected sends an "activity.detected" event to transition from Away to Online.

#### DispatchError

```go
func DispatchError(ctx context.Context, instance hsm.Instance, err error) <-chan struct{}
```

DispatchError sends an "error" event to transition to Errored state.
The error parameter is stored and can be retrieved via LastError().

#### DispatchPromptDetected

```go
func DispatchPromptDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchPromptDetected sends a "prompt.detected" event to transition from Busy to Online.

#### DispatchRateCleared

```go
func DispatchRateCleared(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchRateCleared sends a "rate.cleared" event to transition from Offline to Online.

#### DispatchRateLimit

```go
func DispatchRateLimit(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchRateLimit sends a "rate.limit" event to transition to Offline state.

#### DispatchReady

```go
func DispatchReady(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchReady sends a "ready" event to transition from Starting to Running.

#### DispatchStart

```go
func DispatchStart(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStart sends a "start" event to transition from Pending to Starting.

#### DispatchStop

```go
func DispatchStop(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStop sends a "stop" event to transition from Running to Terminated.

#### DispatchStuckDetected

```go
func DispatchStuckDetected(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchStuckDetected sends a "stuck.detected" event to transition to Away state.

#### DispatchTaskAssigned

```go
func DispatchTaskAssigned(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchTaskAssigned sends a "task.assigned" event to transition from Online to Busy.

#### DispatchTaskCompleted

```go
func DispatchTaskCompleted(ctx context.Context, instance hsm.Instance) <-chan struct{}
```

DispatchTaskCompleted sends a "task.completed" event to transition from Busy to Online.


## type Agent

```go
type Agent struct {
	mu sync.RWMutex
	api.Agent

	// Lifecycle state
	lifecycle api.LifecycleState

	// Presence state
	presence api.PresenceState
}
```

Agent represents a managed agent instance.

### Methods

#### Agent.Lifecycle

```go
func () Lifecycle() api.LifecycleState
```

Lifecycle returns the agent's lifecycle state.

#### Agent.Presence

```go
func () Presence() api.PresenceState
```

Presence returns the agent's presence state.

#### Agent.SetLifecycle

```go
func () SetLifecycle(state api.LifecycleState)
```

SetLifecycle sets the agent's lifecycle state.

#### Agent.SetPresence

```go
func () SetPresence(state api.PresenceState)
```

SetPresence sets the agent's presence state.


## type LifecycleHSM

```go
type LifecycleHSM struct {
	hsm.HSM

	mu             sync.RWMutex
	agent          *Agent
	lifecycleState api.LifecycleState
	dispatcher     event.Dispatcher
	lastError      error
}
```

LifecycleHSM wraps an agent with HSM-driven lifecycle management.

### Functions returning LifecycleHSM

#### NewLifecycleHSM

```go
func NewLifecycleHSM(agent *Agent, dispatcher event.Dispatcher) *LifecycleHSM
```

NewLifecycleHSM creates a new lifecycle HSM for an agent.


### Methods

#### LifecycleHSM.Agent

```go
func () Agent() *Agent
```

Agent returns the associated agent.

#### LifecycleHSM.LastError

```go
func () LastError() error
```

LastError returns the last error that caused the errored state.

#### LifecycleHSM.LifecycleState

```go
func () LifecycleState() api.LifecycleState
```

LifecycleState returns the current lifecycle state.

#### LifecycleHSM.Start

```go
func () Start(ctx context.Context) *LifecycleHSM
```

Start initializes and starts the lifecycle HSM.
Returns the started HSM instance.

#### LifecycleHSM.onEnterErrored

```go
func () onEnterErrored(ctx context.Context, e hsm.Event)
```

Entry action for Errored state

#### LifecycleHSM.onEnterRunning

```go
func () onEnterRunning(ctx context.Context)
```

Entry action for Running state

#### LifecycleHSM.onEnterStarting

```go
func () onEnterStarting(ctx context.Context)
```

Entry action for Starting state

#### LifecycleHSM.onEnterTerminated

```go
func () onEnterTerminated(ctx context.Context)
```

Entry action for Terminated state

#### LifecycleHSM.onExitRunning

```go
func () onExitRunning(ctx context.Context)
```

Exit action for Running state

#### LifecycleHSM.setLifecycleState

```go
func () setLifecycleState(state api.LifecycleState)
```

setLifecycleState updates the internal state and synchronizes with the agent.


## type Manager

```go
type Manager struct {
	mu         sync.RWMutex
	agents     map[muid.MUID]*Agent
	slugs      map[string]muid.MUID // slug -> agent ID for collision detection
	dispatcher event.Dispatcher
}
```

Manager manages agents.

### Functions returning Manager

#### NewManager

```go
func NewManager(dispatcher event.Dispatcher) *Manager
```

NewManager creates a new agent manager.


### Methods

#### Manager.Add

```go
func () Add(ctx context.Context, cfg api.Agent) (*Agent, error)
```

Add adds an agent. The agent's Slug will be computed from Name if not already set.
Returns an error if the agent fails validation.

#### Manager.Get

```go
func () Get(id muid.MUID) *Agent
```

Get returns an agent by ID.

#### Manager.GetBySlug

```go
func () GetBySlug(slug string) *Agent
```

GetBySlug returns an agent by its slug.

#### Manager.List

```go
func () List() []*Agent
```

List returns all agents.

#### Manager.Remove

```go
func () Remove(ctx context.Context, id muid.MUID) error
```

Remove removes an agent.

#### Manager.Roster

```go
func () Roster() []api.RosterEntry
```

Roster returns the roster entries for all agents.

#### Manager.SlugExists

```go
func () SlugExists(slug string) bool
```

SlugExists returns true if an agent with the given slug exists.


## type PresenceHSM

```go
type PresenceHSM struct {
	hsm.HSM

	mu            sync.RWMutex
	agent         *Agent
	presenceState api.PresenceState
	dispatcher    event.Dispatcher
}
```

PresenceHSM wraps an agent with HSM-driven presence management.

### Functions returning PresenceHSM

#### NewPresenceHSM

```go
func NewPresenceHSM(agent *Agent, dispatcher event.Dispatcher) *PresenceHSM
```

NewPresenceHSM creates a new presence HSM for an agent.


### Methods

#### PresenceHSM.Agent

```go
func () Agent() *Agent
```

Agent returns the associated agent.

#### PresenceHSM.PresenceState

```go
func () PresenceState() api.PresenceState
```

PresenceState returns the current presence state.

#### PresenceHSM.Start

```go
func () Start(ctx context.Context) *PresenceHSM
```

Start initializes and starts the presence HSM.
Returns the started HSM instance.

#### PresenceHSM.onEnterAway

```go
func () onEnterAway(ctx context.Context, e hsm.Event)
```

Entry action for Away state

#### PresenceHSM.onEnterBusy

```go
func () onEnterBusy(ctx context.Context, e hsm.Event)
```

Entry action for Busy state

#### PresenceHSM.onEnterOffline

```go
func () onEnterOffline(ctx context.Context, e hsm.Event)
```

Entry action for Offline state

#### PresenceHSM.onEnterOnline

```go
func () onEnterOnline(ctx context.Context, e hsm.Event)
```

Entry action for Online state

#### PresenceHSM.setPresenceState

```go
func () setPresenceState(state api.PresenceState)
```

setPresenceState updates the internal state and synchronizes with the agent.


