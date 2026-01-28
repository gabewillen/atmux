# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api provides public types for the Agent Multiplexer (amux).

This package contains the stable API types that may be imported by external
packages. All types in this package are agent-agnostic; the Agent.Adapter
field is a string reference to an adapter name, not a typed dependency.

- `SpecVersion` — SpecVersion is the version of the specification this implementation targets.
- `type Agent` — Agent represents an active agent instance with a name, description, assigned adapter, and dedicated worktree.
- `type InvalidLocationTypeError` — InvalidLocationTypeError is returned when parsing an invalid location type string.
- `type LifecycleState` — LifecycleState represents the state of an agent's lifecycle.
- `type LocationType` — LocationType indicates whether an agent runs locally or via SSH.
- `type Location` — Location specifies where an agent runs.
- `type PresenceState` — PresenceState represents the availability state of an agent.
- `type RosterEntry` — RosterEntry represents an agent in the roster with presence information.
- `type Session` — Session represents an amux session containing one or more agent PTYs.

### Constants

#### SpecVersion

```go
const SpecVersion = "v1.22"
```

SpecVersion is the version of the specification this implementation targets.


## type Agent

```go
type Agent struct {
	// ID is the globally unique identifier assigned to this agent at runtime.
	// Used for HSM identity, event routing, and the remote protocol field agent_id.
	ID muid.MUID

	// Name is the configured agent name.
	Name string

	// About is a description of the agent's purpose.
	About string

	// Adapter is a string reference to the adapter name (agent-agnostic).
	// Example values: "claude-code", "cursor", "windsurf"
	Adapter string

	// RepoRoot is the canonical repository root path for this agent.
	// See spec §3.23 for canonicalization rules.
	RepoRoot string

	// Worktree is the absolute path to the agent's working directory within RepoRoot.
	// Located at .amux/worktrees/{agent_slug}/.
	Worktree string

	// Location specifies where the agent runs (local or SSH).
	Location Location
}
```

Agent represents an active agent instance with a name, description,
assigned adapter, and dedicated worktree.

The Adapter field is a string reference, not a typed dependency.
The agent structure has no knowledge of specific adapter implementations.
The adapter is loaded dynamically by name through the WASM registry.

## type InvalidLocationTypeError

```go
type InvalidLocationTypeError struct {
	Value string
}
```

InvalidLocationTypeError is returned when parsing an invalid location type string.

### Methods

#### InvalidLocationTypeError.Error

```go
func () Error() string
```

Error implements the error interface.


## type LifecycleState

```go
type LifecycleState string
```

LifecycleState represents the state of an agent's lifecycle.

### Constants

#### LifecyclePending, LifecycleStarting, LifecycleRunning, LifecycleTerminated, LifecycleErrored

```go
const (
	// LifecyclePending indicates the agent is pending initialization.
	LifecyclePending LifecycleState = "pending"

	// LifecycleStarting indicates the agent is starting up.
	LifecycleStarting LifecycleState = "starting"

	// LifecycleRunning indicates the agent is running.
	LifecycleRunning LifecycleState = "running"

	// LifecycleTerminated indicates the agent has terminated normally.
	LifecycleTerminated LifecycleState = "terminated"

	// LifecycleErrored indicates the agent has terminated with an error.
	LifecycleErrored LifecycleState = "errored"
)
```


## type Location

```go
type Location struct {
	// Type is the location type (Local or SSH).
	Type LocationType

	// Host is the SSH host or alias from ~/.ssh/config.
	// Only used when Type is LocationSSH.
	Host string

	// User is the SSH user (optional if defined in ssh config).
	User string

	// Port is the SSH port (optional if defined in ssh config).
	Port int

	// RepoPath is the path to the git repository root on the target host.
	// Required for SSH agents; optional for local agents to select a non-default repo.
	RepoPath string
}
```

Location specifies where an agent runs.

## type LocationType

```go
type LocationType int
```

LocationType indicates whether an agent runs locally or via SSH.

### Constants

#### LocationLocal, LocationSSH

```go
const (
	// LocationLocal indicates the agent runs on the local machine.
	LocationLocal LocationType = iota

	// LocationSSH indicates the agent runs on a remote host via SSH.
	LocationSSH
)
```


### Functions returning LocationType

#### ParseLocationType

```go
func ParseLocationType(s string) (LocationType, error)
```

ParseLocationType parses a case-insensitive string into a LocationType.
Returns LocationLocal for "local" and LocationSSH for "ssh".
Returns an error for any other value.


### Methods

#### LocationType.String

```go
func () String() string
```

String returns the string representation of the location type.


## type PresenceState

```go
type PresenceState string
```

PresenceState represents the availability state of an agent.

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	// PresenceOnline indicates the agent is available to accept tasks.
	PresenceOnline PresenceState = "online"

	// PresenceBusy indicates the agent is working on a task.
	PresenceBusy PresenceState = "busy"

	// PresenceOffline indicates the agent is offline.
	PresenceOffline PresenceState = "offline"

	// PresenceAway indicates the agent is temporarily unavailable
	// (e.g., remote connection lost).
	PresenceAway PresenceState = "away"
)
```


## type RosterEntry

```go
type RosterEntry struct {
	// Agent is the agent information.
	Agent Agent

	// Lifecycle is the current lifecycle state.
	Lifecycle LifecycleState

	// Presence is the current presence state.
	Presence PresenceState
}
```

RosterEntry represents an agent in the roster with presence information.

## type Session

```go
type Session struct {
	// ID is the unique session identifier.
	ID muid.MUID

	// Agents is the list of agent IDs in this session.
	Agents []muid.MUID
}
```

Session represents an amux session containing one or more agent PTYs.

