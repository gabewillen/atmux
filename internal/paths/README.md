# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

Package paths centralizes filesystem path resolution for amux.

It is the single source of truth for repo-scoped, user-scoped, and runtime
paths derived from config and environment variables.

- `ErrRepoRootNotFound` — ErrRepoRootNotFound is returned when a git repository root cannot be located.
- `func AmuxRootForRepo(repoRoot string) string` — AmuxRootForRepo returns the .amux directory for the provided repo root.
- `func CanonicalizeRepoRoot(path string, homeDir string) (string, error)` — CanonicalizeRepoRoot applies repo_root canonicalization rules.
- `func FindRepoRoot(start string) (string, error)` — FindRepoRoot searches upward from start for a git repository root.
- `func PTYDirForRepo(repoRoot string) string` — PTYDirForRepo returns the PTY socket directory for the provided repo root.
- `func SlugifyAgent(name string) string` — SlugifyAgent derives the agent slug per the spec rules.
- `func UniqueAgentSlug(name string, used map[string]struct{}) string` — UniqueAgentSlug ensures the agent slug is unique within the provided set.
- `func WorktreePathForRepo(repoRoot, agentSlug string) string` — WorktreePathForRepo returns the worktree path for the provided repo root and slug.
- `func WorktreesDirForRepo(repoRoot string) string` — WorktreesDirForRepo returns the worktrees directory for the provided repo root.
- `func expandHomePath(path string, homeOverride string) (string, error)`
- `type Resolver` — Resolver resolves filesystem paths based on repo root and user home.

### Variables

#### ErrRepoRootNotFound

```go
var ErrRepoRootNotFound = errors.New("repo root not found")
```

ErrRepoRootNotFound is returned when a git repository root cannot be located.


### Functions

#### AmuxRootForRepo

```go
func AmuxRootForRepo(repoRoot string) string
```

AmuxRootForRepo returns the .amux directory for the provided repo root.

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(path string, homeDir string) (string, error)
```

CanonicalizeRepoRoot applies repo_root canonicalization rules.

#### FindRepoRoot

```go
func FindRepoRoot(start string) (string, error)
```

FindRepoRoot searches upward from start for a git repository root.

#### PTYDirForRepo

```go
func PTYDirForRepo(repoRoot string) string
```

PTYDirForRepo returns the PTY socket directory for the provided repo root.

#### SlugifyAgent

```go
func SlugifyAgent(name string) string
```

SlugifyAgent derives the agent slug per the spec rules.

#### UniqueAgentSlug

```go
func UniqueAgentSlug(name string, used map[string]struct{}) string
```

UniqueAgentSlug ensures the agent slug is unique within the provided set.

#### WorktreePathForRepo

```go
func WorktreePathForRepo(repoRoot, agentSlug string) string
```

WorktreePathForRepo returns the worktree path for the provided repo root and slug.

#### WorktreesDirForRepo

```go
func WorktreesDirForRepo(repoRoot string) string
```

WorktreesDirForRepo returns the worktrees directory for the provided repo root.

#### expandHomePath

```go
func expandHomePath(path string, homeOverride string) (string, error)
```


## type Resolver

```go
type Resolver struct {
	repoRoot string
	homeDir  string
}
```

Resolver resolves filesystem paths based on repo root and user home.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver(start string) (*Resolver, error)
```

NewResolver creates a resolver rooted at the discovered repo and user home.

#### NewResolverOptionalRepo

```go
func NewResolverOptionalRepo(start string) (*Resolver, error)
```

NewResolverOptionalRepo creates a resolver when repo root is optional.


### Methods

#### Resolver.AmuxRoot

```go
func () AmuxRoot() string
```

AmuxRoot returns the repo-scoped .amux directory path.

#### Resolver.CanonicalizeRepoRoot

```go
func () CanonicalizeRepoRoot(path string) (string, error)
```

CanonicalizeRepoRoot normalizes a repo_root path using the resolver's home.

#### Resolver.ExpandHome

```go
func () ExpandHome(path string) string
```

ExpandHome expands a leading ~/ in the provided path.

#### Resolver.HomeDir

```go
func () HomeDir() string
```

HomeDir returns the resolved user home directory.

#### Resolver.PTYDir

```go
func () PTYDir() string
```

PTYDir returns the repo-scoped PTY socket directory.

#### Resolver.ProjectAdapterConfigPath

```go
func () ProjectAdapterConfigPath(adapter string) string
```

ProjectAdapterConfigPath returns the per-adapter repo-scoped config path.

#### Resolver.ProjectAdapterWasmPath

```go
func () ProjectAdapterWasmPath(adapter string) string
```

ProjectAdapterWasmPath returns the per-adapter repo-scoped WASM module path.

#### Resolver.ProjectAdaptersDir

```go
func () ProjectAdaptersDir() string
```

ProjectAdaptersDir returns the repo-scoped adapters directory.

#### Resolver.ProjectConfigPath

```go
func () ProjectConfigPath() string
```

ProjectConfigPath returns the repo-scoped config path.

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the resolved repository root.

#### Resolver.SocketPath

```go
func () SocketPath() string
```

SocketPath returns the default daemon socket path under ~/.amux.

#### Resolver.UserAdapterConfigPath

```go
func () UserAdapterConfigPath(adapter string) string
```

UserAdapterConfigPath returns the per-adapter user config path.

#### Resolver.UserAdapterWasmPath

```go
func () UserAdapterWasmPath(adapter string) string
```

UserAdapterWasmPath returns the per-adapter user WASM module path.

#### Resolver.UserAdaptersDir

```go
func () UserAdaptersDir() string
```

UserAdaptersDir returns the user-scoped adapters directory.

#### Resolver.UserConfigPath

```go
func () UserConfigPath() string
```

UserConfigPath returns the user config path (~/.config/amux/config.toml).

#### Resolver.WorktreePath

```go
func () WorktreePath(agentSlug string) string
```

WorktreePath returns the worktree path for the given agent slug.

#### Resolver.WorktreesDir

```go
func () WorktreesDir() string
```

WorktreesDir returns the repo-scoped worktrees directory.


