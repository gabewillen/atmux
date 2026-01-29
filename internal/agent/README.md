# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `AgentModel` — AgentModel defines the combined agent state machine.
- `EventStart, EventStarted, EventStop, EventTerminated, EventError, EventActivityDetected, EventInactivity, EventSetPresence` — Event constants for agent state machines.
- `func StartAction(ctx context.Context, a *AgentActor, event hsm.Event)` — StartAction handles the agent start sequence.
- `func StopAction(ctx context.Context, a *AgentActor, event hsm.Event)` — StopAction handles the agent stop sequence.
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
		hsm.Transition(on(EventStart), hsm.Target("/agent/starting")),
	),
	hsm.State("starting",
		hsm.Entry[*AgentActor](StartAction),
		hsm.Transition(on(EventStarted), hsm.Target("/agent/running")),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored")),
		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated")),
	),
	hsm.State("running",

		hsm.State("online",
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/busy")),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/away")),
		),
		hsm.State("busy",
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/online")),
		),
		hsm.State("away",
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/online")),
		),

		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/online"),
			hsm.Guard(presenceIs(api.PresenceOnline)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/busy"),
			hsm.Guard(presenceIs(api.PresenceBusy)),
		),
		hsm.Transition(
			on(EventSetPresence), hsm.Target("/agent/running/away"),
			hsm.Guard(presenceIs(api.PresenceAway)),
		),

		hsm.Initial(hsm.Target("/agent/running/online")),

		hsm.Transition(on(EventStop), hsm.Target("/agent/terminated"), hsm.Effect[*AgentActor](StopAction)),
		hsm.Transition(on(EventError), hsm.Target("/agent/errored"), hsm.Effect[*AgentActor](StopAction)),
		hsm.Transition(on(EventTerminated), hsm.Target("/agent/terminated"), hsm.Effect[*AgentActor](StopAction)),
	),
	hsm.State("terminated",
		hsm.Entry[*AgentActor](StopAction),
	),
	hsm.State("errored",
		hsm.Entry[*AgentActor](StopAction),
	),

	hsm.Initial(hsm.Target("pending")),
)
```

AgentModel defines the combined agent state machine.
Pending -> Starting -> Running (containing Presence) -> Terminated / Errored


### Functions

#### StartAction

```go
func StartAction(ctx context.Context, a *AgentActor, event hsm.Event)
```

StartAction handles the agent start sequence.

#### StopAction

```go
func StopAction(ctx context.Context, a *AgentActor, event hsm.Event)
```

StopAction handles the agent stop sequence.

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
	data     api.Agent
	worktree *worktree.Manager
	ptyFile  *os.File
}
```

AgentActor wraps the public Agent struct and manages its state via HSM.

### Functions returning AgentActor

#### NewAgent

```go
func NewAgent(name, adapter, repoRoot string, wtMgr *worktree.Manager) *AgentActor
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

#### AgentActor.PtyFile

```go
func () PtyFile() *os.File
```

PtyFile returns the underlying PTY file descriptor if started.

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


