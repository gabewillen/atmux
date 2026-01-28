# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

Package paths provides centralized filesystem path resolution for amux.

This package is the single source of truth for all filesystem path resolution
in the amux codebase. All subsystems MUST use this package for path resolution
and MUST NOT hardcode paths.

Path resolution follows these rules:
- Paths starting with ~/ are expanded to the user's home directory
- Paths are converted to absolute paths
- Paths are cleaned (. and .. segments resolved)
- Symbolic links are resolved where possible

- `DefaultResolver` — DefaultResolver is the default path resolver instance.
- `func ExpandHome(path string) string` — ExpandHome expands ~/ using the default resolver.
- `func FindModuleRoot(startDir string) (string, error)` — FindModuleRoot finds the Go module root using the default resolver.
- `func FindRepoRoot(startDir string) (string, error)` — FindRepoRoot finds the repository root using the default resolver.
- `func NormalizeAgentSlug(name string) string` — NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug.
- `func Resolve(path string) (string, error)` — Resolve resolves a path using the default resolver.
- `func init()`
- `type Resolver` — Resolver handles filesystem path resolution.

### Variables

#### DefaultResolver

```go
var DefaultResolver = &Resolver{}
```

DefaultResolver is the default path resolver instance.


### Functions

#### ExpandHome

```go
func ExpandHome(path string) string
```

ExpandHome expands ~/ using the default resolver.

#### FindModuleRoot

```go
func FindModuleRoot(startDir string) (string, error)
```

FindModuleRoot finds the Go module root using the default resolver.

#### FindRepoRoot

```go
func FindRepoRoot(startDir string) (string, error)
```

FindRepoRoot finds the repository root using the default resolver.

#### NormalizeAgentSlug

```go
func NormalizeAgentSlug(name string) string
```

NormalizeAgentSlug normalizes an agent name to a filesystem-safe slug.
Rules per spec §5.3.1:
- Convert to lowercase
- Replace any character not in [a-z0-9-] with -
- Collapse consecutive - characters to a single -
- Trim leading and trailing -
- Truncate to at most 63 characters
- If the result is empty, use "agent"

#### Resolve

```go
func Resolve(path string) (string, error)
```

Resolve resolves a path using the default resolver.

#### init

```go
func init()
```


## type Resolver

```go
type Resolver struct {
	mu sync.RWMutex

	// homeDir is the user's home directory
	homeDir string

	// configDir is the user config directory (~/.config/amux)
	configDir string

	// dataDir is the user data directory (~/.amux)
	dataDir string

	// repoRoot is the current repository root (if any)
	repoRoot string
}
```

Resolver handles filesystem path resolution.

### Methods

#### Resolver.AdapterDir

```go
func () AdapterDir(name string) string
```

AdapterDir returns the adapter directory for a given adapter name.
User adapter config: ~/.config/amux/adapters/{name}/

#### Resolver.ConfigDir

```go
func () ConfigDir() string
```

ConfigDir returns the user config directory (~/.config/amux).

#### Resolver.DaemonSocketPath

```go
func () DaemonSocketPath() string
```

DaemonSocketPath returns the daemon socket path.

#### Resolver.DataDir

```go
func () DataDir() string
```

DataDir returns the user data directory (~/.amux).

#### Resolver.ExpandHome

```go
func () ExpandHome(path string) string
```

ExpandHome expands a path that starts with ~/ to the user's home directory.

#### Resolver.FindModuleRoot

```go
func () FindModuleRoot(startDir string) (string, error)
```

FindModuleRoot searches upward from the given directory to find a Go module root.
Returns the directory containing go.mod, or an error if not found.

#### Resolver.FindRepoRoot

```go
func () FindRepoRoot(startDir string) (string, error)
```

FindRepoRoot searches upward from the given directory to find a git repository root.
Returns an empty string if no repository is found.

#### Resolver.HomeDir

```go
func () HomeDir() string
```

HomeDir returns the user's home directory.

#### Resolver.NATSDataDir

```go
func () NATSDataDir() string
```

NATSDataDir returns the NATS/JetStream data directory.

#### Resolver.PluginDir

```go
func () PluginDir() string
```

PluginDir returns the plugin registry directory.

#### Resolver.ProjectAdapterDir

```go
func () ProjectAdapterDir(name string) string
```

ProjectAdapterDir returns the project adapter directory within the repo.
Project adapter config: .amux/adapters/{name}/

#### Resolver.ProjectConfigFile

```go
func () ProjectConfigFile() string
```

ProjectConfigFile returns the path to the project config file within the repo.

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the current repository root.

#### Resolver.Resolve

```go
func () Resolve(path string) (string, error)
```

Resolve resolves a path, expanding ~ and making it absolute.

#### Resolver.SetRepoRoot

```go
func () SetRepoRoot(root string) error
```

SetRepoRoot sets the repository root for the resolver.

#### Resolver.SnapshotsDir

```go
func () SnapshotsDir() string
```

SnapshotsDir returns the test snapshots directory.

#### Resolver.UserConfigFile

```go
func () UserConfigFile() string
```

UserConfigFile returns the path to the user config file.

#### Resolver.WorktreeDir

```go
func () WorktreeDir(agentSlug string) string
```

WorktreeDir returns the worktree directory for an agent.
Pattern: {repo_root}/.amux/worktrees/{agent_slug}/

#### Resolver.canonicalize

```go
func () canonicalize(path string) (string, error)
```

canonicalize converts a path to its canonical form:
1. Expands ~/ to home directory
2. Converts to absolute path
3. Cleans . and .. segments
4. Resolves symbolic links where possible


