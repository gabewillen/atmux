# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api provides public API types for amux.
All types in this package are agent-agnostic and contain no adapter-specific knowledge.

Package api provides public API types for amux.
ids.go implements identifiers and normalization rules per spec §3, §4.2.3, §5.3.1.

Package api provides public API types for amux.
types.go defines Agent, Session, Location, AgentMessage, and Roster types per spec §5.1, §5.5.9, §6.2–§6.4.

- `DefaultAgentSlug` — DefaultAgentSlug is the slug used when normalization yields an empty string (spec §5.3.1).
- `ErrNotFound, ErrInvalidConfig, ErrNotReady` — Sentinel errors for common failure modes.
- `MaxAgentSlugLen` — MaxAgentSlugLen is the maximum length of a normalized agent_slug (spec §5.3.1).
- `func EncodeID(id ID) string` — EncodeID returns the base-10 string encoding of id for JSON/wire (spec §4.2.3).
- `func NormalizeAgentSlug(name string) string` — NormalizeAgentSlug derives a stable, filesystem-safe agent_slug from the agent name (spec §5.3.1).
- `func UniquifyAgentSlug(baseSlug string, existing map[string]struct{}) string` — UniquifyAgentSlug returns a slug that does not collide with existing.
- `func ValidRuntimeID(id ID) bool` — ValidRuntimeID returns true if id is non-zero and may be used as a runtime ID.
- `type AgentMessage` — AgentMessage is the inter-agent message payload (spec §6.4).
- `type Agent` — Agent is the core data structure for an agent instance (spec §5.1).
- `type ID` — ID is the runtime identifier type for agents, sessions, peers, and hosts.
- `type LocationType` — LocationType is the agent location kind (spec §5.1).
- `type Location` — Location describes where an agent runs: local or SSH (spec §5.1).
- `type RosterEntry` — RosterEntry is one participant in the roster (spec §6.2, §6.3).
- `type RosterKind` — RosterKind is the participant type in the roster (spec §6.2).
- `type Roster` — Roster is the full list of participants (spec §6.2).
- `type Session` — Session holds session identity and reference to an agent (spec §5.5.9).

### Constants

#### DefaultAgentSlug

```go
const DefaultAgentSlug = "agent"
```

DefaultAgentSlug is the slug used when normalization yields an empty string (spec §5.3.1).

#### MaxAgentSlugLen

```go
const MaxAgentSlugLen = 63
```

MaxAgentSlugLen is the maximum length of a normalized agent_slug (spec §5.3.1).


### Variables

#### ErrNotFound, ErrInvalidConfig, ErrNotReady

```go
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidConfig indicates configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrNotReady indicates the system is not ready for the requested operation.
	ErrNotReady = errors.New("system not ready")
)
```

Sentinel errors for common failure modes.


### Functions

#### EncodeID

```go
func EncodeID(id ID) string
```

EncodeID returns the base-10 string encoding of id for JSON/wire (spec §4.2.3).

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) string
```

NormalizeAgentSlug derives a stable, filesystem-safe agent_slug from the agent name (spec §5.3.1).
Rules: lowercase; replace non-[a-z0-9-] with '-'; collapse consecutive '-'; trim; truncate to 63 chars.
If the result is empty, returns DefaultAgentSlug ("agent").

#### UniquifyAgentSlug

```go
func UniquifyAgentSlug(baseSlug string, existing map[string]struct{}) string
```

UniquifyAgentSlug returns a slug that does not collide with existing.
If baseSlug is not in existing, returns baseSlug. Otherwise returns baseSlug-2, baseSlug-3, etc.

#### ValidRuntimeID

```go
func ValidRuntimeID(id ID) bool
```

ValidRuntimeID returns true if id is non-zero and may be used as a runtime ID.
The value 0 is reserved and MUST NOT be emitted as an entity ID (spec §3.22).


## type Agent

```go
type Agent struct {
	ID       ID
	Name     string
	About    string
	Adapter  string // String reference to adapter name (agent-agnostic)
	RepoRoot string // Canonical repository root path (§3.23, §5.3.4)
	Worktree string // Absolute path to the agent's working directory within RepoRoot
	Location Location
}
```

Agent is the core data structure for an agent instance (spec §5.1).
Lifecycle and presence are managed by HSMs in internal/agent; query via agent actor.

## type AgentMessage

```go
type AgentMessage struct {
	ID        ID
	From      ID
	To        ID
	ToSlug    string
	Content   string
	Timestamp time.Time // RFC 3339 UTC per spec §9.1.3.1
}
```

AgentMessage is the inter-agent message payload (spec §6.4).
From/To are set by the publishing component; ToSlug is the captured recipient token (e.g. agent_slug).

## type ID

```go
type ID uint64
```

ID is the runtime identifier type for agents, sessions, peers, and hosts.
It is a 64-bit value (muid-compatible); encoded as base-10 string in JSON per spec §4.2.3.

### Constants

#### BroadcastID

```go
const BroadcastID ID = 0
```

BroadcastID is the reserved sentinel for broadcast addressing (spec §3.22, §6.4).
Implementations MUST NOT assign 0 as a runtime ID for any entity.


### Functions returning ID

#### DecodeID

```go
func DecodeID(s string) (ID, error)
```

DecodeID parses a base-10 ID string. Returns an error for invalid or empty input.

#### NextRuntimeID

```go
func NextRuntimeID() ID
```

NextRuntimeID returns a new runtime ID, retrying until non-zero.
Callers MUST use this (or equivalent) so that 0 is never assigned (spec §3.22).


### Methods

#### ID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes id as a base-10 string per spec §4.2.3.

#### ID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes id from a base-10 string.


## type Location

```go
type Location struct {
	Type     LocationType
	Host     string // SSH host or alias from ~/.ssh/config
	User     string // SSH user (optional if in ssh config)
	Port     int    // SSH port (optional if in ssh config)
	RepoPath string // Path to git repository root on target host
}
```

Location describes where an agent runs: local or SSH (spec §5.1).

## type LocationType

```go
type LocationType int
```

LocationType is the agent location kind (spec §5.1).

### Constants

#### LocationLocal, LocationSSH

```go
const (
	LocationLocal LocationType = iota
	LocationSSH
)
```


## type Roster

```go
type Roster []RosterEntry
```

Roster is the full list of participants (spec §6.2).

## type RosterEntry

```go
type RosterEntry struct {
	AgentID     ID
	Name        string
	About       string
	Adapter     string
	Presence    string // HSM state name e.g. /agent.presence/online
	RepoRoot    string
	Kind        RosterKind
	CurrentTask string // Optional; for busy agents
}
```

RosterEntry is one participant in the roster (spec §6.2, §6.3).
At minimum: agent_id, name, adapter, presence, repo_root (§12); About and CurrentTask for presence awareness.

## type RosterKind

```go
type RosterKind int
```

RosterKind is the participant type in the roster (spec §6.2).

### Constants

#### RosterKindAgent, RosterKindManager, RosterKindDirector

```go
const (
	RosterKindAgent RosterKind = iota
	RosterKindManager
	RosterKindDirector
)
```


## type Session

```go
type Session struct {
	ID      ID
	AgentID ID
}
```

Session holds session identity and reference to an agent (spec §5.5.9).
PTY and replay buffer are owned by internal/agent; this type is the public/wire shape.

