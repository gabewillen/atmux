# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

Package paths provides centralized filesystem path resolution for amux.
All filesystem paths MUST be resolved through this package to ensure
consistent handling of config/env overrides and repository root canonicalization.

- `func CanonicalizeRepoRoot(homeDir, rawPath string) (string, error)` — CanonicalizeRepoRoot produces the canonical repo_root per spec §3.23.
- `func expandHome(path, homeDir string) string` — expandHome expands ~ to the home directory.
- `type Resolver` — Resolver provides path resolution functionality.

### Functions

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(homeDir, rawPath string) (string, error)
```

CanonicalizeRepoRoot produces the canonical repo_root per spec §3.23.
It expands ~/ to homeDir, converts to absolute, cleans . and .., and resolves
symlinks where the OS provides a mechanism (e.g. EvalSymlinks).
If symlink resolution fails (permissions or unsupported), (a)-(c) are still applied.

#### expandHome

```go
func expandHome(path, homeDir string) string
```

expandHome expands ~ to the home directory.


## type Resolver

```go
type Resolver struct {
	configDir string
	homeDir   string
	repoRoot  string
}
```

Resolver provides path resolution functionality.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver(configDir, homeDir, repoRoot string) (*Resolver, error)
```

NewResolver creates a new path resolver with the given configuration.


### Methods

#### Resolver.AmuxDir

```go
func () AmuxDir() (string, error)
```

AmuxDir returns the path to the .amux directory in the repository root.

#### Resolver.ConfigDir

```go
func () ConfigDir() string
```

ConfigDir returns the user configuration directory.

#### Resolver.HomeDir

```go
func () HomeDir() string
```

HomeDir returns the user home directory.

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the canonical repository root path, or empty string if not set.

#### Resolver.WorktreePath

```go
func () WorktreePath(agentSlug string) (string, error)
```

WorktreePath returns the path to an agent's worktree directory.
The path is relative to the repository root: .amux/worktrees/{agent_slug}/


