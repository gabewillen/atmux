# package api

`import "github.com/copilot-claude-sonnet-4/amux/pkg/api"`

Package api provides public types for the amux system.
These types are agent-agnostic - Agent.Adapter is a string reference
to maintain the separation between core and adapter implementations.

- `type AgentMessage` — AgentMessage represents a message sent between agents per spec §6.4.
- `type AgentState` — AgentState represents agent lifecycle states per HSM pattern.
- `type Agent` — Agent represents an agent instance in the system.
- `type Event` — Event represents a system event in the amux architecture.
- `type PresenceState` — PresenceState represents agent presence states per HSM pattern.
- `type Session` — Session represents an active agent session with PTY and process tracking.

## type Agent

```go
type Agent struct {
	// ID is the unique identifier for this agent instance.
	ID muid.MUID `json:"id"`

	// Slug is the normalized agent slug (derived from Name).
	Slug string `json:"slug"`

	// Adapter is the string name of the WASM adapter (e.g., "test-adapter-1", "test-adapter-2").
	Adapter string `json:"adapter"`

	// Name is a human-readable name for this agent instance.
	Name string `json:"name"`

	// State represents the current lifecycle state.
	State AgentState `json:"state"`

	// Presence represents the current presence state.
	Presence PresenceState `json:"presence"`

	// RepoRoot is the canonical repository root path for this agent.
	RepoRoot string `json:"repo_root"`

	// CreatedAt is when this agent was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this agent was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// Config contains adapter-specific configuration (opaque to core).
	Config map[string]interface{} `json:"config,omitempty"`
}
```

Agent represents an agent instance in the system.
The Adapter field is a string reference to maintain agent-agnostic design.

## type AgentMessage

```go
type AgentMessage struct {
	// ID is the unique identifier for this message.
	ID muid.MUID `json:"id"`

	// From is the sender runtime ID (set by publishing component).
	From muid.MUID `json:"from"`

	// To is the recipient runtime ID (set by publishing component, or BroadcastID).
	To muid.MUID `json:"to"`

	// ToSlug is the recipient token captured from text (typically agent_slug); case-insensitive.
	ToSlug string `json:"to_slug"`

	// Content is the message content.
	Content string `json:"content"`

	// Timestamp is when this message was sent.
	Timestamp time.Time `json:"timestamp"`
}
```

AgentMessage represents a message sent between agents per spec §6.4.

## type AgentState

```go
type AgentState string
```

AgentState represents agent lifecycle states per HSM pattern.

### Constants

#### AgentStatePending, AgentStateStarting, AgentStateRunning, AgentStateTerminated, AgentStateErrored

```go
const (
	// AgentStatePending indicates the agent is pending startup.
	AgentStatePending AgentState = "pending"

	// AgentStateStarting indicates the agent is in the process of starting.
	AgentStateStarting AgentState = "starting"

	// AgentStateRunning indicates the agent is running and operational.
	AgentStateRunning AgentState = "running"

	// AgentStateTerminated indicates the agent has terminated normally.
	AgentStateTerminated AgentState = "terminated"

	// AgentStateErrored indicates the agent has encountered an error.
	AgentStateErrored AgentState = "errored"
)
```


## type Event

```go
type Event struct {
	// ID is the unique identifier for this event.
	ID muid.MUID `json:"id"`

	// Type is the event type (e.g., "agent.started", "process.spawned").
	Type string `json:"type"`

	// AgentID is the ID of the agent associated with this event (if any).
	AgentID *muid.MUID `json:"agent_id,omitempty"`

	// SessionID is the ID of the session associated with this event (if any).
	SessionID *muid.MUID `json:"session_id,omitempty"`

	// Timestamp is when this event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Data contains event-specific data.
	Data map[string]interface{} `json:"data,omitempty"`
}
```

Event represents a system event in the amux architecture.

## type PresenceState

```go
type PresenceState string
```

PresenceState represents agent presence states per HSM pattern.

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	// PresenceOnline indicates the agent is online and available.
	PresenceOnline PresenceState = "online"

	// PresenceBusy indicates the agent is busy with a task.
	PresenceBusy PresenceState = "busy"

	// PresenceOffline indicates the agent is offline.
	PresenceOffline PresenceState = "offline"

	// PresenceAway indicates the agent is away/idle.
	PresenceAway PresenceState = "away"
)
```


## type Session

```go
type Session struct {
	// ID is the unique identifier for this session.
	ID muid.MUID `json:"id"`

	// AgentID is the ID of the agent that owns this session.
	AgentID muid.MUID `json:"agent_id"`

	// PTYPath is the path to the PTY device for this session.
	PTYPath string `json:"pty_path,omitempty"`

	// ProcessPID is the process ID of the main process (if any).
	ProcessPID int `json:"process_pid,omitempty"`

	// WorkingDirectory is the current working directory for the session.
	WorkingDirectory string `json:"working_directory"`

	// Environment contains session-specific environment variables.
	Environment map[string]string `json:"environment,omitempty"`

	// CreatedAt is when this session was created.
	CreatedAt time.Time `json:"created_at"`

	// LastActivity is when this session last had activity.
	LastActivity time.Time `json:"last_activity"`
}
```

Session represents an active agent session with PTY and process tracking.

