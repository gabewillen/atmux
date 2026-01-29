# package api

`import "github.com/stateforward/amux/pkg/api"`

Package api provides public types for the amux system.

- `BroadcastID` — BroadcastID is the reserved ID value (0) for broadcast messages to all participants.
- `ErrInvalidLocationType, ErrReservedID, ErrInvalidAgent`
- `func CanonicalizeRepoRoot(repoPath string) (string, error)` — CanonicalizeRepoRoot canonicalizes a repository root path.
- `func GenerateID() muid.MUID` — GenerateID generates a new muid.MUID and ensures it is not the reserved value 0.
- `func NormalizeAgentSlug(name string) string` — NormalizeAgentSlug derives a stable, filesystem-safe identifier from an agent name.
- `type AgentMessage` — AgentMessage represents a message between agents, host managers, or the director.
- `type Agent` — Agent represents a coding agent instance managed by amux.
- `type LocationType` — LocationType indicates whether an agent runs locally or remotely.
- `type Location` — Location specifies where an agent runs.
- `type Session` — Session represents a collection of agents managed together.

### Constants

#### BroadcastID

```go
const BroadcastID muid.MUID = 0
```

BroadcastID is the reserved ID value (0) for broadcast messages to all participants.
Per spec §3.22, implementations SHALL NOT assign 0 as a runtime ID for any agent,
process, session, peer, or message.


### Variables

#### ErrInvalidLocationType, ErrReservedID, ErrInvalidAgent

```go
var (
	// ErrInvalidLocationType is returned when parsing an invalid location type string.
	ErrInvalidLocationType = errors.New("invalid location type: must be 'local' or 'ssh'")

	// ErrReservedID is returned when attempting to use the reserved ID value 0.
	ErrReservedID = errors.New("cannot use reserved ID value 0")

	// ErrInvalidAgent is returned when an agent structure fails validation.
	ErrInvalidAgent = errors.New("invalid agent configuration")
)
```


### Functions

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(repoPath string) (string, error)
```

CanonicalizeRepoRoot canonicalizes a repository root path.
Per spec §3.23:
  - Expand ~/ to the target host's home directory
  - Convert to an absolute path
  - Clean ./.. segments
  - Resolve symbolic links to their target path where possible

If symbolic link resolution is not possible (insufficient permissions or missing OS support),
this function still applies (a)-(c) and treats the result as canonical.

#### GenerateID

```go
func GenerateID() muid.MUID
```

GenerateID generates a new muid.MUID and ensures it is not the reserved value 0.
Per spec §3.22: If an ID generator produces 0, the implementation SHALL generate a new ID.

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) string
```

NormalizeAgentSlug derives a stable, filesystem-safe identifier from an agent name.
Per spec §5.3.1:
  - Convert to lowercase
  - Replace any character not in [a-z0-9-] with -
  - Collapse consecutive - characters to a single -
  - Trim leading and trailing -
  - Truncate to at most 63 characters
  - If the result is empty, use "agent"


## type Agent

```go
type Agent struct {
	// ID is the runtime identifier for this agent instance.
	// Per spec §3.21, this is a globally unique identifier used for HSM identity,
	// event routing, and the remote protocol field `agent_id`.
	ID muid.MUID

	// Name is the human-readable name for this agent.
	Name string

	// About is a description of the agent's role or purpose.
	About string

	// Adapter is a string reference to the adapter name (e.g., "claude-code", "cursor").
	// Per spec §5.1, this is agent-agnostic - the adapter is loaded dynamically by name
	// through the WASM registry.
	Adapter string

	// RepoRoot is the canonical repository root path for this agent.
	// Per spec §5.1 and §3.23, this is the canonicalized absolute path.
	RepoRoot string

	// Worktree is the absolute path to the agent's working directory within RepoRoot.
	// Per spec §5.3.1, this is typically `.amux/worktrees/{agent_slug}/`.
	Worktree string

	// Location specifies where this agent runs (local or SSH).
	Location Location
}
```

Agent represents a coding agent instance managed by amux.
Per spec §5.1, an agent consists of the required properties.
The Adapter field is a string reference, not a typed dependency - the agent structure
has no knowledge of specific adapter implementations.

## type AgentMessage

```go
type AgentMessage struct {
	// ID is the unique identifier for this message.
	ID muid.MUID

	// From is the sender runtime ID (set by publishing component).
	From muid.MUID

	// To is the recipient runtime ID (set by publishing component, or BroadcastID).
	To muid.MUID

	// ToSlug is the recipient token captured from text (typically agent_slug); case-insensitive.
	ToSlug string

	// Content is the message content.
	Content string
}
```

AgentMessage represents a message between agents, host managers, or the director.
Per spec §6.4, agents can communicate with each other using these messages.

## type Location

```go
type Location struct {
	// Type indicates whether this is a local or SSH location.
	Type LocationType

	// Host is the SSH host or alias from ~/.ssh/config.
	// Only used when Type is LocationSSH.
	Host string

	// User is the SSH user (optional if configured in ssh config).
	// Only used when Type is LocationSSH.
	User string

	// Port is the SSH port (optional if configured in ssh config).
	// Only used when Type is LocationSSH.
	Port int

	// RepoPath is the path to the git repository root on the target host.
	// Per spec §5.1:
	// - Required for SSH agents
	// - Optional for local agents to select a non-default repo
	RepoPath string
}
```

Location specifies where an agent runs.
Per spec §5.1, agents can run locally or on a remote host via SSH.

## type LocationType

```go
type LocationType int
```

LocationType indicates whether an agent runs locally or remotely.

### Constants

#### LocationLocal, LocationSSH

```go
const (
	// LocationLocal indicates the agent runs on the same host as the director.
	LocationLocal LocationType = iota

	// LocationSSH indicates the agent runs on a remote host accessed via SSH.
	LocationSSH
)
```


### Functions returning LocationType

#### ParseLocationType

```go
func ParseLocationType(s string) (LocationType, error)
```

ParseLocationType parses a location type string (case-insensitive).
Per spec §5.1, valid values are "local" and "ssh".


### Methods

#### LocationType.String

```go
func () String() string
```

String returns the string representation of a LocationType.


## type Session

```go
type Session struct {
	// ID is the runtime identifier for this session.
	ID muid.MUID

	// Agents is the list of agents in this session.
	Agents []*Agent
}
```

Session represents a collection of agents managed together.
Per spec §3.5, a session contains one or more agent PTYs.

