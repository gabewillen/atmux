# package worktree

`import "github.com/agentflare-ai/amux/internal/worktree"`

- `func AddWorktree(repoPath, worktreePath, branch string) error` — AddWorktree adds a worktree.
- `func Commit(repoPath, message string) error` — Commit creates a commit with the given message.
- `func EnsureBranch(repoPath, branch, startPoint string) error` — EnsureBranch ensures a branch exists.
- `func GetHeadBranch(repoPath string) (string, error)`
- `func IsDirty(path string) (bool, error)` — IsDirty checks if the working directory has uncommitted changes.
- `func IsRepo(path string) bool` — IsRepo checks if the given path is a valid git repository.
- `func Merge(repoPath, branch, strategy string) error` — Merge merges the given branch into the current HEAD using the specified strategy.
- `func RemoveWorktree(repoPath, worktreePath string) error` — RemoveWorktree removes a worktree.
- `func WorktreeList(repoPath string) ([]string, error)` — WorktreeList returns a map of path -> branch/commit details (simplified).
- `type Manager` — Manager handles worktree lifecycle for agents.

### Functions

#### AddWorktree

```go
func AddWorktree(repoPath, worktreePath, branch string) error
```

AddWorktree adds a worktree.
git worktree add <path> <branch>

#### Commit

```go
func Commit(repoPath, message string) error
```

Commit creates a commit with the given message.
Useful for completing a squash merge which leaves changes staged.

#### EnsureBranch

```go
func EnsureBranch(repoPath, branch, startPoint string) error
```

EnsureBranch ensures a branch exists.
If it doesn't exist, it creates it starting from startPoint (default HEAD).

#### GetHeadBranch

```go
func GetHeadBranch(repoPath string) (string, error)
```

#### IsDirty

```go
func IsDirty(path string) (bool, error)
```

IsDirty checks if the working directory has uncommitted changes.

#### IsRepo

```go
func IsRepo(path string) bool
```

IsRepo checks if the given path is a valid git repository.

#### Merge

```go
func Merge(repoPath, branch, strategy string) error
```

Merge merges the given branch into the current HEAD using the specified strategy.
Supported strategies: merge-commit, squash, rebase, ff-only.

#### RemoveWorktree

```go
func RemoveWorktree(repoPath, worktreePath string) error
```

RemoveWorktree removes a worktree.

#### WorktreeList

```go
func WorktreeList(repoPath string) ([]string, error)
```

WorktreeList returns a map of path -> branch/commit details (simplified).
We might just need to check existence.


## type Manager

```go
type Manager struct {
	resolver *paths.Resolver
}
```

Manager handles worktree lifecycle for agents.

### Functions returning Manager

#### NewManager

```go
func NewManager() (*Manager, error)
```

NewManager creates a new worktree manager.


### Methods

#### Manager.Ensure

```go
func () Ensure(agent api.Agent) (string, error)
```

Ensure creates or validates the worktree for the given agent.
It ensures the target branch `amux/<slug>` exists and the worktree at `.amux/worktrees/<slug>` is active.

#### Manager.MergeAgent

```go
func () MergeAgent(agent api.Agent, strategy string, allowDirty bool) error
```

MergeAgent merges the agent's worktree branch back into the base branch.
It respects the allow_dirty configuration.

#### Manager.Remove

```go
func () Remove(agent api.Agent) error
```

Remove removes the worktree for the given agent.


