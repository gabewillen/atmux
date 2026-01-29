# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

- `func EnsureDir(path string) error` — EnsureDir ensures that the directory exists.
- `type Resolver` — Resolver handles path resolution for the application.

### Functions

#### EnsureDir

```go
func EnsureDir(path string) error
```

EnsureDir ensures that the directory exists.


## type Resolver

```go
type Resolver struct {
	homeDir  string
	repoRoot string
}
```

Resolver handles path resolution for the application.

### Functions returning Resolver

#### NewResolver

```go
func NewResolver() (*Resolver, error)
```

NewResolver creates a new path resolver.


### Methods

#### Resolver.CanonicalizeRepoRoot

```go
func () CanonicalizeRepoRoot(path string) (string, error)
```

CanonicalizeRepoRoot resolves a repository root path to its canonical absolute form.
Rules (§3.23):
1. Expand ~/ to home directory.
2. Convert to absolute path.
3. Clean . and .. segments (implied by Abs).
4. Resolve symbolic links via EvalSymlinks.

#### Resolver.ConfigDir

```go
func () ConfigDir() string
```

ConfigDir returns the default configuration directory.
Linux: ~/.config/amux
macOS: ~/Library/Application Support/amux (or ~/.config/amux if preferred, spec says ~/.config/amux)

#### Resolver.Expand

```go
func () Expand(path string) string
```

Expand expands a path starting with ~/ to the user's home directory.

#### Resolver.ProjectConfigDir

```go
func () ProjectConfigDir(root string) string
```

ProjectConfigDir returns the project-local configuration directory (.amux).

#### Resolver.Resolve

```go
func () Resolve(path string) (string, error)
```

Resolve returns the absolute path for the given path.
It expands ~ and resolves relative paths against CWD.

#### Resolver.WorktreesDir

```go
func () WorktreesDir(root string) string
```

WorktreesDir returns the worktrees directory within the project config (.amux/worktrees).


