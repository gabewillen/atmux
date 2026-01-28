# package paths

`import "github.com/stateforward/amux/internal/paths"`

Package paths provides centralized filesystem path resolution for amux.

All filesystem paths are resolved through this package per spec §4.2.6, §4.2.8.
This ensures consistent handling of:
- Home directory expansion (~/)
- Repo-scoped .amux/ paths
- Configuration file paths
- Adapter/plugin registry paths

- `func Canonicalize(path string) (string, error)` — Canonicalize canonicalizes a path per spec §3.23: - Expands ~/ to home directory - Converts to absolute path - Cleans .
- `func ExpandHome(path string) (string, error)` — ExpandHome expands ~ prefix to user's home directory.
- `func UserConfigDir() (string, error)` — UserConfigDir returns the user configuration directory.
- `func UserConfigFile() (string, error)` — UserConfigFile returns the user configuration file path.
- `type Resolver` — Resolver resolves filesystem paths with consistent expansion and canonicalization.

### Functions

#### Canonicalize

```go
func Canonicalize(path string) (string, error)
```

Canonicalize canonicalizes a path per spec §3.23:
- Expands ~/ to home directory
- Converts to absolute path
- Cleans . and .. segments
- Resolves symbolic links (where possible)

#### ExpandHome

```go
func ExpandHome(path string) (string, error)
```

ExpandHome expands ~ prefix to user's home directory.

#### UserConfigDir

```go
func UserConfigDir() (string, error)
```

UserConfigDir returns the user configuration directory.
Default: ~/.config/amux/

#### UserConfigFile

```go
func UserConfigFile() (string, error)
```

UserConfigFile returns the user configuration file path.


## type Resolver

```go
type Resolver struct {
	repoRoot string // Canonical repository root path
}
```

Resolver resolves filesystem paths with consistent expansion and canonicalization.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver(repoRoot string) (*Resolver, error)
```

NewResolver creates a new path resolver for the given repository root.
repoRoot must be an absolute path to a git repository.


### Methods

#### Resolver.AmuxDir

```go
func () AmuxDir() string
```

AmuxDir returns the .amux directory path.

#### Resolver.ProjectConfig

```go
func () ProjectConfig() string
```

ProjectConfig returns the project configuration file path.

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the canonical repository root.

#### Resolver.WorktreePath

```go
func () WorktreePath(agentSlug string) string
```

WorktreePath returns the worktree path for the given agent slug.
Format: <repo_root>/.amux/worktrees/<agent_slug>/


