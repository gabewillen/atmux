# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `AgentModel` — AgentModel defines the combined agent state machine.
- `EventStart, EventStarted, EventStop, EventTerminated, EventError, EventActivityDetected, EventInactivity, EventSetPresence` — Event constants for agent state machines.
- `func on(name string) hsm.RedefinableElement` — Helpers for event definition
- `func presenceIs(p api.Presence) func(context.Context, *AgentActor, hsm.Event) bool` — Guards helpers
- `type AgentActor` — AgentActor wraps the public Agent struct and manages its state via HSM.

### Constants

#### EventStart, EventStarted, EventStop, EventTerminated, EventError, EventActivityDetected, EventInactivity, EventSetPresence

```go
const (
	EventStart            = "agent.lifecycle.start"
	EventStarted          = "agent.lifecycle.started" // Emitted when process actually runs
	EventStop             = "agent.lifecycle.stop"
	EventTerminated       = "agent.lifecycle.terminated"
	EventError            = "agent.lifecycle.error"
	EventActivityDetected = "agent.presence.activity"
	EventInactivity       = "agent.presence.inactivity"
	EventSetPresence      = "agent.presence.set" // Manual override
)
```

Event constants for agent state machines.


### Variables

#### AgentModel

```go
var AgentModel = hsm.Define("agent",
	hsm.State("pending",
		hsm.Transition(on(EventStart), hsm.Target("starting")),
	),
	hsm.State("starting",
		hsm.Transition(on(EventStarted), hsm.Target("running")),
		hsm.Transition(on(EventError), hsm.Target("errored")),
		hsm.Transition(on(EventStop), hsm.Target("terminated")),
	),
	hsm.State("running",

		hsm.State("online",
			hsm.Transition(on(EventActivityDetected), hsm.Target("busy")),
			hsm.Transition(on(EventInactivity), hsm.Target("away")),
		),
		hsm.State("busy",
			hsm.Transition(on(EventInactivity), hsm.Target("online")),
		),
		hsm.State("away",
			hsm.Transition(on(EventActivityDetected), hsm.Target("online")),
		),

		hsm.Transition(
			on(EventSetPresence), hsm.Target("online"),
			hsm.Guard(presenceIs(api.PresenceOnline)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("busy"),
			hsm.Guard(presenceIs(api.PresenceBusy)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("away"),
			hsm.Guard(presenceIs(api.PresenceAway)),
		),

		hsm.Initial(hsm.Target("online")),

		hsm.Transition(on(EventStop), hsm.Target("terminated")),
		hsm.Transition(on(EventError), hsm.Target("errored")),
		hsm.Transition(on(EventTerminated), hsm.Target("terminated")),
	),
	hsm.State("terminated"),
	hsm.State("errored"),

	hsm.Initial(hsm.Target("pending")),
)
```

AgentModel defines the combined agent state machine.
Pending -> Starting -> Running (containing Presence) -> Terminated / Errored


### Functions

#### on

```go
func on(name string) hsm.RedefinableElement
```

Helpers for event definition

#### presenceIs

```go
func presenceIs(p api.Presence) func(context.Context, *AgentActor, hsm.Event) bool
```

Guards helpers


## type AgentActor

```go
type AgentActor struct {
	hsm.HSM
	data api.Agent
}
```

AgentActor wraps the public Agent struct and manages its state via HSM.

### Functions returning AgentActor

#### NewAgent

```go
func NewAgent(name, adapter, repoRoot string) *AgentActor
```

NewAgent creates a new AgentActor.


### Methods

#### AgentActor.Data

```go
func () Data() api.Agent
```

Data returns a copy of the public agent data.

#### AgentActor.ID

```go
func () ID() api.AgentID
```

ID returns the agent's ID.

#### AgentActor.SendActivity

```go
func () SendActivity()
```

SendActivity signals activity to the agent presence HSM.

#### AgentActor.Start

```go
func () Start()
```

Start initiates the agent start sequence.

#### AgentActor.Stop

```go
func () Stop()
```

Stop initiates the agent stop sequence.


