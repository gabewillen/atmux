# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

- `type AgentID` — AgentID is a globally unique identifier for an agent.
- `type AgentSlug` — AgentSlug is a filesystem-safe identifier derived from an agent's name.
- `type HostID` — HostID is a stable identifier for a host.
- `type PeerID` — PeerID is a unique identifier for a peer in the hsmnet.
- `type ProcessID` — ProcessID is a unique identifier for a process.
- `type RepoKey` — repoKey represents a stable, session-scoped identifier for a repository.
- `type RepoRoot` — RepoRoot is a canonical absolute path to a git repository.
- `type SessionID` — SessionID is a unique identifier for a session.

## type AgentID

```go
type AgentID muid.MUID
```

AgentID is a globally unique identifier for an agent.
It is serialized as a base-10 string in JSON.

### Methods

#### AgentID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes AgentID as a base-10 string.

#### AgentID.String

```go
func () String() string
```

#### AgentID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes AgentID from a base-10 string.


## type AgentSlug

```go
type AgentSlug string
```

AgentSlug is a filesystem-safe identifier derived from an agent's name.

### Functions returning AgentSlug

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) AgentSlug
```

NormalizeAgentSlug derives a stable, filesystem-safe identifier from a name.
Rules: lowercase, non-[a-z0-9-] -> -, collapse -, trim -, max 63 chars.


### Methods

#### AgentSlug.String

```go
func () String() string
```

String returns the string representation.

#### AgentSlug.Validate

```go
func () Validate() error
```

Validate checks if the slug is valid.


## type HostID

```go
type HostID string
```

HostID is a stable identifier for a host.

## type PeerID

```go
type PeerID muid.MUID
```

PeerID is a unique identifier for a peer in the hsmnet.
It is serialized as a base-10 string in JSON.

### Methods

#### PeerID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes PeerID as a base-10 string.

#### PeerID.String

```go
func () String() string
```

#### PeerID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes PeerID from a base-10 string.


## type ProcessID

```go
type ProcessID muid.MUID
```

ProcessID is a unique identifier for a process.
It is serialized as a base-10 string in JSON.

### Methods

#### ProcessID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes ProcessID as a base-10 string.

#### ProcessID.String

```go
func () String() string
```

#### ProcessID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes ProcessID from a base-10 string.


## type RepoKey

```go
type RepoKey string
```

repoKey represents a stable, session-scoped identifier for a repository.
Derived from (location.type, location.host, repo_root).

## type RepoRoot

```go
type RepoRoot string
```

RepoRoot is a canonical absolute path to a git repository.

### Functions returning RepoRoot

#### ParseRepoRoot

```go
func ParseRepoRoot(path string) (RepoRoot, error)
```

ParseRepoRoot validates that the path looks like a repo root (simple check).
Full canonicalization requires filesystem access (internal/paths).


### Methods

#### RepoRoot.String

```go
func () String() string
```


## type SessionID

```go
type SessionID muid.MUID
```

SessionID is a unique identifier for a session.
It is serialized as a base-10 string in JSON.

### Methods

#### SessionID.MarshalJSON

```go
func () MarshalJSON() ([]byte, error)
```

MarshalJSON encodes SessionID as a base-10 string.

#### SessionID.String

```go
func () String() string
```

#### SessionID.UnmarshalJSON

```go
func () UnmarshalJSON(data []byte) error
```

UnmarshalJSON decodes SessionID from a base-10 string.


