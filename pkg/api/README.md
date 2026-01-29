# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api provides public types for the Agent Multiplexer (amux).

This package contains the stable API types that may be imported by external
packages. All types in this package are agent-agnostic; the Agent.Adapter
field is a string reference to an adapter name, not a typed dependency.

- `SpecVersion` — SpecVersion is the version of the specification this implementation targets.
- `func itoa(i int) string` — itoa converts an integer to a string without importing strconv.
- `type AgentValidationError` — AgentValidationError represents an error validating an Agent.
- `type Agent` — Agent represents an active agent instance with a name, description, assigned adapter, and dedicated worktree.
- `type InvalidLocationTypeError` — InvalidLocationTypeError is returned when parsing an invalid location type string.
- `type LifecycleState` — LifecycleState represents the state of an agent's lifecycle.
- `type LocationType` — LocationType indicates whether an agent runs locally or via SSH.
- `type Location` — Location specifies where an agent runs.
- `type PresenceState` — PresenceState represents the availability state of an agent.
- `type RosterEntry` — RosterEntry represents an agent in the roster with presence information.
- `type SessionValidationError` — SessionValidationError represents an error validating a Session.
- `type Session` — Session represents an amux session containing one or more agent PTYs.

### Constants

#### SpecVersion

```go
const SpecVersion = "v1.22"
```

SpecVersion is the version of the specification this implementation targets.


### Functions

#### itoa

```go
func itoa(i int) string
```

itoa converts an integer to a string without importing strconv.


## type Agent

```go
type Agent struct {
	// ID is the globally unique identifier assigned to this agent at runtime.
	// Used for HSM identity, event routing, and the remote protocol field agent_id.
	// Must be non-zero (0 is reserved per spec §3.22).
	ID muid.MUID

	// Name is the configured agent name.
	Name string

	// Slug is the filesystem-safe identifier derived from Name.
	// Used for worktree directory names and git branch names.
	// See spec §5.3.1 for normalization rules.
	Slug string

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

### Methods

#### Agent.Validate

```go
func () Validate() error
```

Validate checks that the Agent meets all invariants:
  - ID must be non-zero (spec §3.22)
  - Name must be non-empty
  - Slug must be non-empty
  - Adapter must be non-empty
  - RepoRoot must be non-empty

Returns nil if valid, or an AgentValidationError describing the first violation.


## type AgentValidationError

```go
type AgentValidationError struct {
	Field   string
	Message string
}
```

AgentValidationError represents an error validating an Agent.

### Methods

#### AgentValidationError.Error

```go
func () Error() string
```

Error implements the error interface.


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
See spec §5.4 for the lifecycle state machine.

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
	// This is a final state.
	LifecycleTerminated LifecycleState = "terminated"

	// LifecycleErrored indicates the agent has terminated with an error.
	// This is a final state.
	LifecycleErrored LifecycleState = "errored"
)
```


### Methods

#### LifecycleState.IsFinal

```go
func () IsFinal() bool
```

IsFinal returns true if this is a terminal lifecycle state.
Final states are Terminated and Errored.

#### LifecycleState.IsValid

```go
func () IsValid() bool
```

IsValid returns true if this is a recognized lifecycle state.


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
See spec §6.1 and §6.5 for presence states and transitions.

### Constants

#### PresenceOnline, PresenceBusy, PresenceOffline, PresenceAway

```go
const (
	// PresenceOnline indicates the agent is available to accept tasks.
	PresenceOnline PresenceState = "online"

	// PresenceBusy indicates the agent is working on a task.
	PresenceBusy PresenceState = "busy"

	// PresenceOffline indicates the agent is rate-limited or temporarily unavailable.
	PresenceOffline PresenceState = "offline"

	// PresenceAway indicates the agent is connected but not responsive
	// (e.g., stuck, or remote connection lost).
	PresenceAway PresenceState = "away"
)
```


### Methods

#### PresenceState.CanAcceptTasks

```go
func () CanAcceptTasks() bool
```

CanAcceptTasks returns true if the agent can accept new tasks.
Only Online agents can accept tasks.

#### PresenceState.IsValid

```go
func () IsValid() bool
```

IsValid returns true if this is a recognized presence state.


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
	// Must be non-zero (0 is reserved per spec §3.22).
	ID muid.MUID

	// Agents is the list of agent IDs in this session.
	// All agent IDs must be non-zero.
	Agents []muid.MUID
}
```

Session represents an amux session containing one or more agent PTYs.

### Methods

#### Session.AddAgent

```go
func () AddAgent(id muid.MUID) bool
```

AddAgent adds an agent ID to the session if not already present.
Returns true if the agent was added, false if already present.

#### Session.HasAgent

```go
func () HasAgent(id muid.MUID) bool
```

HasAgent returns true if the session contains the given agent ID.

#### Session.RemoveAgent

```go
func () RemoveAgent(id muid.MUID) bool
```

RemoveAgent removes an agent ID from the session.
Returns true if the agent was removed, false if not present.

#### Session.Validate

```go
func () Validate() error
```

Validate checks that the Session meets all invariants:
  - ID must be non-zero (spec §3.22)
  - All agent IDs must be non-zero

Returns nil if valid, or a SessionValidationError describing the first violation.


## type SessionValidationError

```go
type SessionValidationError struct {
	Field   string
	Message string
}
```

SessionValidationError represents an error validating a Session.

### Methods

#### SessionValidationError.Error

```go
func () Error() string
```

Error implements the error interface.


