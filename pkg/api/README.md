# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api defines public types shared between amux clients and the daemon.

The API types are stable, JSON-serializable, and enforce the wire conventions
required by the amux specification.

- `BroadcastID` — BroadcastID is the reserved runtime ID for broadcast messages.
- `ErrEmptyID, ErrZeroID`
- `ErrInvalidLocationType, ErrInvalidLocation, ErrInvalidAgent, ErrInvalidSession`
- `defaultIDGenerator`
- `func validateRepoRoot(repoRoot string) error`
- `func validateWorktree(repoRoot, worktree string) error`
- `type AdapterRef` — AdapterRef is the string name of an adapter loaded from the WASM registry.
- `type AgentID` — AgentID is the runtime identifier for an agent.
- `type AgentMessage` — AgentMessage represents a participant communication payload.
- `type Agent` — Agent describes the core metadata for a managed agent.
- `type HostID` — HostID is the identifier for a host manager.
- `type LocationType` — LocationType describes where an agent runs.
- `type Location` — Location describes where an agent should run.
- `type PeerID` — PeerID is the runtime identifier for a peer.
- `type RuntimeID` — RuntimeID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.
- `type SessionID` — SessionID is the runtime identifier for a session.
- `type Session` — Session describes runtime session metadata for an agent.
- `type TargetID` — TargetID is a runtime ID that permits the broadcast sentinel (0).

### Constants

#### BroadcastID

```go
const BroadcastID muid.MUID = 0
```

BroadcastID is the reserved runtime ID for broadcast messages.


### Variables

#### ErrInvalidLocationType, ErrInvalidLocation, ErrInvalidAgent, ErrInvalidSession

```go
var (
	// ErrInvalidLocationType is returned for unknown location types.
	ErrInvalidLocationType = errors.New("invalid location type")
	// ErrInvalidLocation is returned when a location violates invariants.
	ErrInvalidLocation = errors.New("invalid location")
	// ErrInvalidAgent is returned when an agent violates invariants.
	ErrInvalidAgent = errors.New("invalid agent")
	// ErrInvalidSession is returned when a session violates invariants.
	ErrInvalidSession = errors.New("invalid session")
)
```

#### ErrEmptyID, ErrZeroID

```go
var (
	// ErrEmptyID is returned when parsing an empty ID string.
	ErrEmptyID = errors.New("empty id")
	// ErrZeroID is returned when an ID is zero or reserved.
	ErrZeroID = errors.New("zero id")
)
```

#### defaultIDGenerator

```go
var defaultIDGenerator = muid.NewGenerator(muid.DefaultConfig(), 0, 0)
```


### Functions

#### validateRepoRoot

```go
func validateRepoRoot(repoRoot string) error
```

#### validateWorktree

```go
func validateWorktree(repoRoot, worktree string) error
```


## type AdapterRef

```go
type AdapterRef string
```

AdapterRef is the string name of an adapter loaded from the WASM registry.

## type Agent

```go
type Agent struct {
	ID       AgentID    `json:"agent_id"`
	Name     string     `json:"name"`
	About    string     `json:"about"`
	Adapter  AdapterRef `json:"adapter"`
	RepoRoot string     `json:"repo_root"`
	Worktree string     `json:"worktree"`
	Location Location   `json:"location"`
}
```

Agent describes the core metadata for a managed agent.

### Functions returning Agent

#### NewAgent

```go
func NewAgent(name, about string, adapter AdapterRef, repoRoot, worktree string, location Location) (Agent, error)
```

NewAgent constructs a new agent with a fresh ID.

#### NewAgentWithID

```go
func NewAgentWithID(id AgentID, name, about string, adapter AdapterRef, repoRoot, worktree string, location Location) (Agent, error)
```

NewAgentWithID constructs a new agent with the provided ID.


### Methods

#### Agent.Validate

```go
func () Validate() error
```

Validate checks agent invariants.


## type AgentID

```go
type AgentID struct {
	RuntimeID
}
```

AgentID is the runtime identifier for an agent.

### Functions returning AgentID

#### MustParseAgentID

```go
func MustParseAgentID(raw string) AgentID
```

MustParseAgentID parses a base-10 encoded agent ID string and panics on failure.

#### NewAgentID

```go
func NewAgentID() AgentID
```

NewAgentID creates a new agent runtime ID.

#### ParseAgentID

```go
func ParseAgentID(raw string) (AgentID, error)
```

ParseAgentID parses a base-10 encoded agent ID string.


### Methods

#### AgentID.IsZero

```go
func () IsZero() bool
```

IsZero reports whether the ID is the reserved zero value.

#### AgentID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### AgentID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### AgentID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### AgentID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.


## type AgentMessage

```go
type AgentMessage struct {
	ID        RuntimeID `json:"id"`
	From      RuntimeID `json:"from"`
	To        TargetID  `json:"to"`
	ToSlug    string    `json:"to_slug"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
```

AgentMessage represents a participant communication payload.

## type HostID

```go
type HostID string
```

HostID is the identifier for a host manager.

### Functions returning HostID

#### MustParseHostID

```go
func MustParseHostID(raw string) HostID
```

MustParseHostID parses a host ID and panics on failure.

#### ParseHostID

```go
func ParseHostID(raw string) (HostID, error)
```

ParseHostID validates a host ID.


### Methods

#### HostID.String

```go
func () String() string
```

String returns the host ID as a string.


## type Location

```go
type Location struct {
	Type     LocationType `json:"type"`
	Host     string       `json:"host,omitempty"`
	User     string       `json:"user,omitempty"`
	Port     int          `json:"port,omitempty"`
	RepoPath string       `json:"repo_path,omitempty"`
}
```

Location describes where an agent should run.

### Methods

#### Location.Validate

```go
func () Validate() error
```

Validate checks location invariants.


## type LocationType

```go
type LocationType int
```

LocationType describes where an agent runs.

### Constants

#### LocationLocal, LocationSSH

```go
const (
	// LocationLocal represents a local agent.
	LocationLocal LocationType = iota
	// LocationSSH represents an agent running on a remote SSH host.
	LocationSSH
)
```


### Functions returning LocationType

#### ParseLocationType

```go
func ParseLocationType(raw string) (LocationType, error)
```

ParseLocationType parses a string into a location type.


### Methods

#### LocationType.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the location type as a string.

#### LocationType.String

```go
func () String() string
```

String returns the string form of the location type.

#### LocationType.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a location type from a string.


## type PeerID

```go
type PeerID struct {
	RuntimeID
}
```

PeerID is the runtime identifier for a peer.

### Functions returning PeerID

#### MustParsePeerID

```go
func MustParsePeerID(raw string) PeerID
```

MustParsePeerID parses a base-10 encoded peer ID string and panics on failure.

#### NewPeerID

```go
func NewPeerID() PeerID
```

NewPeerID creates a new peer runtime ID.

#### ParsePeerID

```go
func ParsePeerID(raw string) (PeerID, error)
```

ParsePeerID parses a base-10 encoded peer ID string.


### Methods

#### PeerID.IsZero

```go
func () IsZero() bool
```

IsZero reports whether the ID is the reserved zero value.

#### PeerID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### PeerID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### PeerID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### PeerID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.


## type RuntimeID

```go
type RuntimeID struct {
	value muid.MUID
}
```

RuntimeID is a JSON-safe wrapper around muid.MUID that encodes as base-10 strings.

### Functions returning RuntimeID

#### MustParseRuntimeID

```go
func MustParseRuntimeID(raw string) RuntimeID
```

MustParseRuntimeID parses a base-10 encoded ID string and panics on failure.

#### NewRuntimeID

```go
func NewRuntimeID() RuntimeID
```

NewRuntimeID creates a new non-zero ID suitable for runtime use.

#### ParseRuntimeID

```go
func ParseRuntimeID(raw string) (RuntimeID, error)
```

ParseRuntimeID parses a base-10 encoded ID string.


### Methods

#### RuntimeID.IsZero

```go
func () IsZero() bool
```

IsZero reports whether the ID is the reserved zero value.

#### RuntimeID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### RuntimeID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### RuntimeID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### RuntimeID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.


## type Session

```go
type Session struct {
	ID       SessionID `json:"session_id"`
	AgentID  AgentID   `json:"agent_id"`
	RepoRoot string    `json:"repo_root"`
	Worktree string    `json:"worktree"`
	Location Location  `json:"location"`
}
```

Session describes runtime session metadata for an agent.

### Functions returning Session

#### NewSession

```go
func NewSession(agentID AgentID, repoRoot, worktree string, location Location) (Session, error)
```

NewSession constructs a new session with a fresh ID.

#### NewSessionWithID

```go
func NewSessionWithID(id SessionID, agentID AgentID, repoRoot, worktree string, location Location) (Session, error)
```

NewSessionWithID constructs a new session with the provided ID.


### Methods

#### Session.Validate

```go
func () Validate() error
```

Validate checks session invariants.


## type SessionID

```go
type SessionID struct {
	RuntimeID
}
```

SessionID is the runtime identifier for a session.

### Functions returning SessionID

#### MustParseSessionID

```go
func MustParseSessionID(raw string) SessionID
```

MustParseSessionID parses a base-10 encoded session ID string and panics on failure.

#### NewSessionID

```go
func NewSessionID() SessionID
```

NewSessionID creates a new session runtime ID.

#### ParseSessionID

```go
func ParseSessionID(raw string) (SessionID, error)
```

ParseSessionID parses a base-10 encoded session ID string.


### Methods

#### SessionID.IsZero

```go
func () IsZero() bool
```

IsZero reports whether the ID is the reserved zero value.

#### SessionID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### SessionID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### SessionID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### SessionID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.


## type TargetID

```go
type TargetID struct {
	value muid.MUID
}
```

TargetID is a runtime ID that permits the broadcast sentinel (0).

### Functions returning TargetID

#### ParseTargetID

```go
func ParseTargetID(raw string) (TargetID, error)
```

ParseTargetID parses a base-10 encoded ID string, allowing zero.

#### TargetIDFromRuntime

```go
func TargetIDFromRuntime(id RuntimeID) TargetID
```

TargetIDFromRuntime converts a runtime ID to a target ID.


### Methods

#### TargetID.IsBroadcast

```go
func () IsBroadcast() bool
```

IsBroadcast reports whether the ID is the broadcast sentinel.

#### TargetID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes the ID as a JSON string containing a base-10 integer.

#### TargetID.String

```go
func () String() string
```

String returns the base-10 string form of the ID.

#### TargetID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes a JSON string containing a base-10 integer ID.

#### TargetID.Value

```go
func () Value() muid.MUID
```

Value returns the underlying muid.MUID.


