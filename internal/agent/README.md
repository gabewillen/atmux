# package agent

`import "github.com/agentflare-ai/amux/internal/agent"`

- `AgentModel` — AgentModel defines the combined agent state machine.
- `EventAgentMessage` — EventAgentMessage is the event name for inter-agent messages.
- `EventStart, EventStarted, EventStop, EventTerminated, EventError, EventActivityDetected, EventInactivity, EventSetPresence, EventRemoteDisconnect` — Event constants for agent state machines.
- `func SetPresenceAction(p api.Presence) func(context.Context, *AgentActor, hsm.Event)` — SetPresenceAction updates the agent's presence data.
- `func StartAction(ctx context.Context, a *AgentActor, event hsm.Event)` — StartAction handles the agent start sequence.
- `func StopAction(ctx context.Context, a *AgentActor, event hsm.Event)` — StopAction handles the agent stop sequence.
- `func on(name string) hsm.RedefinableElement` — Helpers for event definition
- `func presenceIs(p api.Presence) func(context.Context, *AgentActor, hsm.Event) bool` — Guards helpers
- `type AgentActor` — AgentActor wraps the public Agent struct and manages its state via HSM.
- `type MessagePayload` — MessagePayload represents the content of an inter-agent message.
- `type Roster` — Roster manages the collection of active agents.

### Constants

#### EventStart, EventStarted, EventStop, EventTerminated, EventError, EventActivityDetected, EventInactivity, EventSetPresence, EventRemoteDisconnect

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
	EventRemoteDisconnect = "remote.disconnect"  // Remote host disconnected
)
```

Event constants for agent state machines.

#### EventAgentMessage

```go
const EventAgentMessage = "agent.message"
```

EventAgentMessage is the event name for inter-agent messages.


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
			hsm.Entry(SetPresenceAction(api.PresenceOnline)),
			hsm.Transition(on(EventActivityDetected), hsm.Target("/agent/running/busy")),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/away")),
			hsm.Transition(on(EventRemoteDisconnect), hsm.Target("/agent/running/away")),
		),
		hsm.State("busy",
			hsm.Entry(SetPresenceAction(api.PresenceBusy)),
			hsm.Transition(on(EventInactivity), hsm.Target("/agent/running/online")),
			hsm.Transition(on(EventRemoteDisconnect), hsm.Target("/agent/running/away")),
		),
		hsm.State("away",
			hsm.Entry(SetPresenceAction(api.PresenceAway)),
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

#### SetPresenceAction

```go
func SetPresenceAction(p api.Presence) func(context.Context, *AgentActor, hsm.Event)
```

SetPresenceAction updates the agent's presence data.

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

#### AgentActor.RouteMessage

```go
func () RouteMessage(payload MessagePayload) bool
```

RouteMessage determines if a message is for this agent.

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


## type MessagePayload

```go
type MessagePayload struct {
	FromID  muid.MUID `json:"from_id"`
	ToID    muid.MUID `json:"to_id"`
	Content string    `json:"content"`
}
```

MessagePayload represents the content of an inter-agent message.

## type Roster

```go
type Roster struct {
	mu     sync.RWMutex
	agents map[muid.MUID]*AgentActor
}
```

Roster manages the collection of active agents.

### Functions returning Roster

#### NewRoster

```go
func NewRoster() *Roster
```

NewRoster creates a new empty Roster.


### Methods

#### Roster.Add

```go
func () Add(agent *AgentActor)
```

Add adds an agent to the roster.

#### Roster.Get

```go
func () Get(id muid.MUID) *AgentActor
```

Get retrieves an agent by ID.

#### Roster.List

```go
func () List() []api.RosterEntry
```

List returns a sorted list of roster entries.

#### Roster.Remove

```go
func () Remove(id muid.MUID)
```

Remove removes an agent from the roster by ID.


