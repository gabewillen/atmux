# package ids

`import "github.com/copilot-claude-sonnet-4/amux/internal/ids"`

Package ids provides identifier and normalization utilities for the amux system.
It implements the specifications for agent_id, peer_id, host_id, agent_slug,
and repo_root canonicalization as defined in the amux spec.

- `func AgentSlugFromName(name string) string` — AgentSlugFromName normalizes a human-readable agent name into a valid agent_slug.
- `func CanonicalizeRepoRoot(path string, isRemote bool) (string, error)` — CanonicalizeRepoRoot canonicalizes a repository root path.
- `func IsValidIdentifierName(name string) bool` — IsValidIdentifierName checks if a name is suitable for use in identifiers.
- `func ValidateAgentSlug(slug string) error` — ValidateAgentSlug validates that a string is a valid agent_slug per spec.
- `type AgentID` — AgentID represents a unique agent identifier.
- `type HostID` — HostID represents a unique host identifier.
- `type PeerID` — PeerID represents a unique peer identifier.

### Functions

#### AgentSlugFromName

```go
func AgentSlugFromName(name string) string
```

AgentSlugFromName normalizes a human-readable agent name into a valid agent_slug.
Per the spec: lowercase, non-[a-z0-9-] → '-', collapse, trim, max 63 chars.

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(path string, isRemote bool) (string, error)
```

CanonicalizeRepoRoot canonicalizes a repository root path.
For local paths, it resolves to absolute path.
For remote contexts, it expands ~ to user home directory.

#### IsValidIdentifierName

```go
func IsValidIdentifierName(name string) bool
```

IsValidIdentifierName checks if a name is suitable for use in identifiers.
It ensures the name contains printable characters and isn't excessively long.

#### ValidateAgentSlug

```go
func ValidateAgentSlug(slug string) error
```

ValidateAgentSlug validates that a string is a valid agent_slug per spec.


## type AgentID

```go
type AgentID = muid.MUID
```

AgentID represents a unique agent identifier.

### Functions returning AgentID

#### NewAgentID

```go
func NewAgentID() AgentID
```

NewAgentID generates a new unique agent identifier.


## type HostID

```go
type HostID = muid.MUID
```

HostID represents a unique host identifier.

### Functions returning HostID

#### NewHostID

```go
func NewHostID() HostID
```

NewHostID generates a new unique host identifier.


## type PeerID

```go
type PeerID = muid.MUID
```

PeerID represents a unique peer identifier.

### Functions returning PeerID

#### NewPeerID

```go
func NewPeerID() PeerID
```

NewPeerID generates a new unique peer identifier.


