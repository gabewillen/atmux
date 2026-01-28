# package paths

`import "github.com/stateforward/amux/internal/paths"`

Package paths implements a centralized path resolution system for the amux project

- `func EnsureDir(path string) error` — EnsureDir ensures that a directory exists, creating it if necessary
- `func EnsureParentDir(filePath string) error` — EnsureParentDir ensures that the parent directory of a file path exists
- `func ExpandHome(path string) string` — ExpandHome expands the ~ symbol to the user's home directory in the given path
- `type Config` — Config holds configuration for the path resolver
- `type Resolver` — Resolver provides centralized path resolution

### Functions

#### EnsureDir

```go
func EnsureDir(path string) error
```

EnsureDir ensures that a directory exists, creating it if necessary

#### EnsureParentDir

```go
func EnsureParentDir(filePath string) error
```

EnsureParentDir ensures that the parent directory of a file path exists

#### ExpandHome

```go
func ExpandHome(path string) string
```

ExpandHome expands the ~ symbol to the user's home directory in the given path


## type Config

```go
type Config struct {
	// BaseDir is the base directory for relative paths
	BaseDir string

	// HomeDir is the user's home directory (usually auto-detected)
	HomeDir string

	// RepoRoot is the root of the git repository
	RepoRoot string

	// CacheDir is the directory for cache files
	CacheDir string

	// ConfigDir is the directory for configuration files
	ConfigDir string
}
```

Config holds configuration for the path resolver

## type Resolver

```go
type Resolver struct {
	config Config
}
```

Resolver provides centralized path resolution

### Functions returning Resolver

#### New

```go
func New(config Config) *Resolver
```

New creates a new path resolver with the given configuration


### Methods

#### Resolver.AgentBranch

```go
func () AgentBranch(agentSlug string) string
```

AgentBranch returns the git branch name for an agent

#### Resolver.AmuxDir

```go
func () AmuxDir() string
```

AmuxDir returns the path to the .amux directory in the repo root

#### Resolver.CachePath

```go
func () CachePath(subpath ...string) string
```

CachePath returns a path within the cache directory

#### Resolver.ConfigPath

```go
func () ConfigPath(subpath ...string) string
```

ConfigPath returns a path within the config directory

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the resolved repository root

#### Resolver.Resolve

```go
func () Resolve(path string) string
```

Resolve expands a path relative to the resolver's configuration

#### Resolver.SocketPath

```go
func () SocketPath() string
```

SocketPath returns the path to the amux daemon socket

#### Resolver.WorktreeDir

```go
func () WorktreeDir(agentSlug string) string
```

WorktreeDir returns the path to a specific agent's worktree directory

#### Resolver.WorktreesDir

```go
func () WorktreesDir() string
```

WorktreesDir returns the path to the worktrees directory


