# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

Package paths provides centralized filesystem path resolution for amux.

- `func findRepoRoot() (string, error)` — findRepoRoot walks up the directory tree looking for .git.
- `func normalizeAgentSlug(name string) string` — normalizeAgentSlug converts agent names to lowercase alphanumerics with hyphens.
- `type Config` — Config holds path configuration.
- `type Resolver` — Resolver provides centralized path resolution.

### Functions

#### findRepoRoot

```go
func findRepoRoot() (string, error)
```

findRepoRoot walks up the directory tree looking for .git.

#### normalizeAgentSlug

```go
func normalizeAgentSlug(name string) string
```

normalizeAgentSlug converts agent names to lowercase alphanumerics with hyphens.


## type Config

```go
type Config struct {
	HomeDir      string
	ConfigDir    string
	DataDir      string
	RuntimeDir   string
	SocketPath   string
	RegistryRoot string
	ModelsRoot   string
	HooksRoot    string
}
```

Config holds path configuration.

## type Resolver

```go
type Resolver struct {
	repoRoot string
	config   Config
}
```

Resolver provides centralized path resolution.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver(config Config) (*Resolver, error)
```

NewResolver creates a new path resolver.


### Methods

#### Resolver.AgentWorktree

```go
func () AgentWorktree(agentSlug string) string
```

AgentWorktree returns the worktree path for a specific agent slug.

#### Resolver.AmuxRoot

```go
func () AmuxRoot() string
```

AmuxRoot returns the .amux directory within the repo.

#### Resolver.ConfigPath

```go
func () ConfigPath(name string) string
```

ConfigPath returns the path to a config file.

#### Resolver.DataPath

```go
func () DataPath(parts ...string) string
```

DataPath returns a path in the data directory.

#### Resolver.HooksPath

```go
func () HooksPath(parts ...string) string
```

HooksPath returns a path in the hooks directory.

#### Resolver.ModelsPath

```go
func () ModelsPath(parts ...string) string
```

ModelsPath returns a path in the models directory.

#### Resolver.RegistryPath

```go
func () RegistryPath(parts ...string) string
```

RegistryPath returns a path in the registry directory.

#### Resolver.RepoRoot

```go
func () RepoRoot() string
```

RepoRoot returns the canonical repository root path.

#### Resolver.SocketPath

```go
func () SocketPath() string
```

SocketPath returns the Unix socket path for daemon communication.

#### Resolver.WorktreeRoot

```go
func () WorktreeRoot() string
```

WorktreeRoot returns the worktrees directory for agents.


