# package paths

`import "github.com/copilot-claude-sonnet-4/amux/internal/paths"`

Package paths provides shared path resolution functionality.
This package resolves all filesystem paths from config/env and repo_root,
maintaining .amux/ directory structure invariants.

- `ErrRepoNotFound, ErrInvalidSlug, ErrPathResolveFailed` — Common sentinel errors for path operations.
- `func normalizeAgentSlug(slug string) (string, error)` — normalizeAgentSlug creates a valid agent slug per spec requirements: lowercase, non-[a-z0-9-] → -, collapse, trim, max 63 chars
- `type Resolver` — Resolver handles filesystem path resolution with .amux/ invariants.

### Variables

#### ErrRepoNotFound, ErrInvalidSlug, ErrPathResolveFailed

```go
var (
	// ErrRepoNotFound indicates no git repository was found.
	ErrRepoNotFound = errors.New("git repository not found")

	// ErrInvalidSlug indicates an invalid agent slug.
	ErrInvalidSlug = errors.New("invalid agent slug")

	// ErrPathResolveFailed indicates path resolution failed.
	ErrPathResolveFailed = errors.New("path resolve failed")
)
```

Common sentinel errors for path operations.


### Functions

#### normalizeAgentSlug

```go
func normalizeAgentSlug(slug string) (string, error)
```

normalizeAgentSlug creates a valid agent slug per spec requirements:
lowercase, non-[a-z0-9-] → -, collapse, trim, max 63 chars


## type Resolver

```go
type Resolver struct {
	repoRoot string
	amuxDir  string
	homeDir  string
}
```

Resolver handles filesystem path resolution with .amux/ invariants.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver(repoRoot string) (*Resolver, error)
```

NewResolver creates a new path resolver for the given repository.


### Methods

#### Resolver.AmuxDir

```go
func () AmuxDir() string
```

AmuxDir returns the .amux directory path.

#### Resolver.ConfigDir

```go
func () ConfigDir() string
```

ConfigDir returns the user configuration directory.

#### Resolver.SnapshotsDir

```go
func () SnapshotsDir() string
```

SnapshotsDir returns the snapshots directory in the repository.

#### Resolver.SocketPath

```go
func () SocketPath() string
```

SocketPath returns the daemon socket path.

#### Resolver.WorktreeDir

```go
func () WorktreeDir(agentSlug string) (string, error)
```

WorktreeDir returns the worktree directory for the given agent slug.


