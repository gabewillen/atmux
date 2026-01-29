# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

- `ReservedID`
- `func EncodeID(id muid.MUID) string` — EncodeID returns the base-10 string representation of an ID.
- `func ParseID(s string) (muid.MUID, error)` — ParseID parses a base-10 string into an ID.
- `nonAlphanumeric, multipleDashes`
- `type AgentID` — AgentID is a globally unique identifier for an agent instance.
- `type AgentSlug` — AgentSlug is a normalized, filesystem-safe string derived from an agent's name.
- `type AgentState` — AgentState represents the lifecycle state of an agent.
- `type Agent` — Agent represents a managed agent instance.
- `type HostID` — HostID is a globally unique identifier for a host running amux.
- `type Presence` — Presence represents the availability state of an agent.
- `type RosterEntry` — RosterEntry represents a simplified view of an agent for listing.
- `type SessionID` — SessionID is a globally unique identifier for an amux session.
- `type Session` — Session represents an active PTY session for an agent.

### Constants

#### ReservedID

```go
const (
	// ReservedID is the sentinel value 0, which MUST NOT be used as a runtime ID.
	ReservedID muid.MUID = 0
)
```


### Variables

#### nonAlphanumeric, multipleDashes

```go
var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]+`)
	multipleDashes  = regexp.MustCompile(`-+`)
)
```


### Functions

#### EncodeID

```go
func EncodeID(id muid.MUID) string
```

EncodeID returns the base-10 string representation of an ID.
This matches the spec requirement for JSON encoding.

#### ParseID

```go
func ParseID(s string) (muid.MUID, error)
```

ParseID parses a base-10 string into an ID.


## type Agent

```go
type Agent struct {
	ID       muid.MUID  `json:"id"`
	Slug     AgentSlug  `json:"slug"`
	Name     string     `json:"name"`
	Adapter  string     `json:"adapter"` // Name of the adapter (e.g., "claude-code")
	State    AgentState `json:"state"`
	Presence Presence   `json:"presence"`
	RepoRoot string     `json:"repo_root"` // Canonical absolute path
	Worktree string     `json:"worktree"`  // Absolute path to worktree
}
```

Agent represents a managed agent instance.
This struct exposes public state; internal logic resides in the actor.

## type AgentID

```go
type AgentID = muid.MUID
```

AgentID is a globally unique identifier for an agent instance.

## type AgentSlug

```go
type AgentSlug string
```

AgentSlug is a normalized, filesystem-safe string derived from an agent's name.

### Functions returning AgentSlug

#### NewAgentSlug

```go
func NewAgentSlug(name string) AgentSlug
```

NewAgentSlug creates a normalized AgentSlug from a raw name.
Rules: lowercase, replace non-alphanumeric with dash, collapse dashes, trim dashes, max 63 chars.


### Methods

#### AgentSlug.String

```go
func () String() string
```


## type AgentState

```go
type AgentState string
```

AgentState represents the lifecycle state of an agent.

### Constants

#### StatePending, StateStarting, StateRunning, StateTerminated, StateErrored

```go
const (
	StatePending    AgentState = "pending"
	StateStarting   AgentState = "starting"
	StateRunning    AgentState = "running"
	StateTerminated AgentState = "terminated"
	StateErrored    AgentState = "errored"
)
```


## type HostID

```go
type HostID = muid.MUID
```

HostID is a globally unique identifier for a host running amux.

## type Presence

```go
type Presence string
```

Presence represents the availability state of an agent.

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	PresenceOnline  Presence = "online"
	PresenceBusy    Presence = "busy"
	PresenceOffline Presence = "offline"
	PresenceAway    Presence = "away"
)
```


## type RosterEntry

```go
type RosterEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"` // Combined state/presence summary
	Location string `json:"location"`
}
```

RosterEntry represents a simplified view of an agent for listing.

## type Session

```go
type Session struct {
	ID      muid.MUID `json:"id"`
	AgentID muid.MUID `json:"agent_id"`
	PTY     *os.File  `json:"-"` // PTY file descriptor (not serialized)
}
```

Session represents an active PTY session for an agent.

## type SessionID

```go
type SessionID = muid.MUID
```

SessionID is a globally unique identifier for an amux session.

