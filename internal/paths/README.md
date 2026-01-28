# package paths

`import "github.com/agentflare-ai/amux/internal/paths"`

- `func CanonicalizeRepoRoot(path string) (string, error)` — CanonicalizeRepoRoot canonicalizes a repository root path.
- `func DefaultConfigDir() (string, error)` — DefaultConfigDir returns the default configuration directory.
- `func DefaultSocketPath() (string, error)` — DefaultSocketPath returns the default daemon socket path.
- `func DefaultWorktreesDir(repoRoot string) string` — DefaultWorktreesDir returns the default worktrees directory.
- `func ExpandHome(path string) string` — ExpandHome expands a path starting with ~/ to the user's home directory.

### Functions

#### CanonicalizeRepoRoot

```go
func CanonicalizeRepoRoot(path string) (string, error)
```

CanonicalizeRepoRoot canonicalizes a repository root path.
It expands ~/, converts to absolute path, cleans ./.., and resolves symlinks.

#### DefaultConfigDir

```go
func DefaultConfigDir() (string, error)
```

DefaultConfigDir returns the default configuration directory.
~/.config/amux

#### DefaultSocketPath

```go
func DefaultSocketPath() (string, error)
```

DefaultSocketPath returns the default daemon socket path.
~/.amux/amuxd.sock

#### DefaultWorktreesDir

```go
func DefaultWorktreesDir(repoRoot string) string
```

DefaultWorktreesDir returns the default worktrees directory.
.amux/worktrees/

#### ExpandHome

```go
func ExpandHome(path string) string
```

ExpandHome expands a path starting with ~/ to the user's home directory.


