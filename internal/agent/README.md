# package agent

`import "github.com/stateforward/amux/internal/agent"`

Package agent provides the agent actor model and HSM-driven lifecycle and presence state machines.

- `EventStart, EventReady, EventStop, EventError` — Lifecycle event constants.
- `EventTaskAssigned, EventTaskCompleted, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivityDetected` — Presence event constants per spec §6.5.
- `LifecycleModel` — LifecycleModel defines the agent lifecycle state machine per spec §5.4.
- `PresenceModel` — PresenceModel defines the agent presence state machine per spec §6.5.
- `StateOnline, StateBusy, StateOffline, StateAway` — Presence state constants matching spec §6.1.
- `StatePending, StateStarting, StateRunning, StateTerminated, StateErrored` — Lifecycle state constants matching spec §5.4.
- `type AgentActor` — AgentActor wraps an Agent with HSM-driven lifecycle and presence state machines.
- `type PresenceActor` — PresenceActor wraps an Agent with a presence state machine.

### Constants

#### StatePending, StateStarting, StateRunning, StateTerminated, StateErrored

```go
const (
	StatePending    = "pending"
	StateStarting   = "starting"
	StateRunning    = "running"
	StateTerminated = "terminated"
	StateErrored    = "errored"
)
```

Lifecycle state constants matching spec §5.4.

#### EventStart, EventReady, EventStop, EventError

```go
const (
	EventStart = "start"
	EventReady = "ready"
	EventStop  = "stop"
	EventError = "error"
)
```

Lifecycle event constants.

#### StateOnline, StateBusy, StateOffline, StateAway

```go
const (
	StateOnline  = "online"
	StateBusy    = "busy"
	StateOffline = "offline"
	StateAway    = "away"
)
```

Presence state constants matching spec §6.1.

#### EventTaskAssigned, EventTaskCompleted, EventPromptDetected, EventRateLimit, EventRateCleared, EventStuckDetected, EventActivityDetected

```go
const (
	EventTaskAssigned     = "task.assigned"
	EventTaskCompleted    = "task.completed"
	EventPromptDetected   = "prompt.detected"
	EventRateLimit        = "rate.limit"
	EventRateCleared      = "rate.cleared"
	EventStuckDetected    = "stuck.detected"
	EventActivityDetected = "activity.detected"
)
```

Presence event constants per spec §6.5.


### Variables

#### LifecycleModel

```go
var LifecycleModel = hsm.Define("agent.lifecycle",
	hsm.State(StatePending),

	hsm.State(StateStarting,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
	),

	hsm.State(StateRunning,
		hsm.Entry(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
		hsm.Exit(func(ctx context.Context, a *AgentActor, e hsm.Event) {

		}),
	),

	hsm.State(StateTerminated, hsm.Final(StateTerminated)),
	hsm.State(StateErrored, hsm.Final(StateErrored)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStart}), hsm.Source(StatePending), hsm.Target(StateStarting)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventReady}), hsm.Source(StateStarting), hsm.Target(StateRunning)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStop}), hsm.Source(StateRunning), hsm.Target(StateTerminated)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StatePending), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateStarting), hsm.Target(StateErrored)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventError}), hsm.Source(StateRunning), hsm.Target(StateErrored)),

	hsm.Initial(hsm.Target(StatePending)),
)
```

LifecycleModel defines the agent lifecycle state machine per spec §5.4.

State diagram:

	┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
	│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
	└─────────┘    └─────────┘    └─────────┘    └────────────┘
	                                   │
	                                   ▼
	                              ┌─────────┐
	                              │ Errored │
	                              └─────────┘

#### PresenceModel

```go
var PresenceModel = hsm.Define("agent.presence",
	hsm.State(StateOnline),
	hsm.State(StateBusy),
	hsm.State(StateOffline),
	hsm.State(StateAway),

	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskAssigned}), hsm.Source(StateOnline), hsm.Target(StateBusy)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventTaskCompleted}), hsm.Source(StateBusy), hsm.Target(StateOnline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventPromptDetected}), hsm.Source(StateBusy), hsm.Target(StateOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateBusy), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateLimit}), hsm.Source(StateOnline), hsm.Target(StateOffline)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventRateCleared}), hsm.Source(StateOffline), hsm.Target(StateOnline)),

	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOnline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateBusy), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventStuckDetected}), hsm.Source(StateOffline), hsm.Target(StateAway)),
	hsm.Transition(hsm.On(hsm.Event{Name: EventActivityDetected}), hsm.Source(StateAway), hsm.Target(StateOnline)),

	hsm.Initial(hsm.Target(StateOnline)),
)
```

PresenceModel defines the agent presence state machine per spec §6.5.

State diagram:

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


## type AgentActor

```go
type AgentActor struct {
	hsm.HSM
	*api.Agent
}
```

AgentActor wraps an Agent with HSM-driven lifecycle and presence state machines.
Per spec §5.4, the lifecycle is managed as an HSM.

### Functions returning AgentActor

#### NewAgentActor

```go
func NewAgentActor(ctx context.Context, agent *api.Agent) *AgentActor
```

NewAgentActor creates a new AgentActor with initialized HSMs.
Per spec §5.4, the lifecycle starts in the "pending" state.


### Methods

#### AgentActor.ErrorAgent

```go
func () ErrorAgent(ctx context.Context, err error)
```

ErrorAgent transitions the agent to the Errored state from any state.
Per spec §5.4, this can be triggered from any state.

#### AgentActor.GetSimpleState

```go
func () GetSimpleState() string
```

GetSimpleState returns the current lifecycle state without the qualified name prefix.
For example, returns "pending" instead of "/agent.lifecycle/pending".

#### AgentActor.GetState

```go
func () GetState() string
```

GetState returns the current lifecycle state.

#### AgentActor.Ready

```go
func () Ready(ctx context.Context)
```

Ready transitions the agent from Starting to Running.
Per spec §5.4, this is triggered by the "ready" event after bootstrap completes.

#### AgentActor.StartAgent

```go
func () StartAgent(ctx context.Context)
```

StartAgent transitions the agent from Pending to Starting.
Per spec §5.4, this is triggered by the "start" event.

#### AgentActor.StopAgent

```go
func () StopAgent(ctx context.Context)
```

StopAgent transitions the agent from Running to Terminated.
Per spec §5.4, this is triggered by the "stop" event.


## type PresenceActor

```go
type PresenceActor struct {
	hsm.HSM
	AgentID string // For logging/debugging
}
```

PresenceActor wraps an Agent with a presence state machine.
Per spec §6.1, presence indicates whether an agent can accept tasks.

### Functions returning PresenceActor

#### NewPresenceActor

```go
func NewPresenceActor(ctx context.Context, agentID string) *PresenceActor
```

NewPresenceActor creates a new PresenceActor with initialized presence HSM.
Per spec §6.1, presence starts in the "online" state.


### Methods

#### PresenceActor.ActivityDetected

```go
func () ActivityDetected(ctx context.Context)
```

ActivityDetected transitions from Away to Online.

#### PresenceActor.GetPresenceState

```go
func () GetPresenceState() string
```

GetPresenceState returns the current presence state.

#### PresenceActor.GetSimplePresenceState

```go
func () GetSimplePresenceState() string
```

GetSimplePresenceState returns the current presence state without the qualified name prefix.
For example, returns "online" instead of "/agent.presence/online".

#### PresenceActor.PromptDetected

```go
func () PromptDetected(ctx context.Context)
```

PromptDetected transitions from Busy to Online.

#### PresenceActor.RateCleared

```go
func () RateCleared(ctx context.Context)
```

RateCleared transitions from Offline to Online.

#### PresenceActor.RateLimit

```go
func () RateLimit(ctx context.Context)
```

RateLimit transitions to Offline.

#### PresenceActor.StuckDetected

```go
func () StuckDetected(ctx context.Context)
```

StuckDetected transitions to Away from any state.

#### PresenceActor.TaskAssigned

```go
func () TaskAssigned(ctx context.Context)
```

TaskAssigned transitions from Online to Busy.

#### PresenceActor.TaskCompleted

```go
func () TaskCompleted(ctx context.Context)
```

TaskCompleted transitions from Busy to Online.


