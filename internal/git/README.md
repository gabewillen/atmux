# package git

`import "github.com/copilot-claude-sonnet-4/amux/internal/git"`

Package git provides git operations for worktree management.
This package implements git worktree isolation per spec requirements.

- `ErrWorktreeExists, ErrBranchExists, ErrGitCommandFailed` — Common sentinel errors for git operations.
- `func CreateWorktree(repoRoot, agentSlug, worktreePath string) error` — CreateWorktree creates a git worktree for the given agent slug.
- `func ListWorktrees(repoRoot string) ([]string, error)` — ListWorktrees returns a list of all git worktrees
- `func RemoveWorktree(repoRoot, worktreePath string) error` — RemoveWorktree removes a git worktree and cleans up the branch
- `func createBranch(repoRoot, branchName string) error` — createBranch creates a new git branch if it doesn't exist
- `func isGitWorktree(path string) bool` — isGitWorktree checks if the given path is a git worktree

### Variables

#### ErrWorktreeExists, ErrBranchExists, ErrGitCommandFailed

```go
var (
	// ErrWorktreeExists indicates a worktree already exists at the given path.
	ErrWorktreeExists = errors.New("worktree already exists")

	// ErrBranchExists indicates a branch already exists.
	ErrBranchExists = errors.New("branch already exists")

	// ErrGitCommandFailed indicates a git command failed.
	ErrGitCommandFailed = errors.New("git command failed")
)
```

Common sentinel errors for git operations.


### Functions

#### CreateWorktree

```go
func CreateWorktree(repoRoot, agentSlug, worktreePath string) error
```

CreateWorktree creates a git worktree for the given agent slug.
This implements worktree isolation per spec: .amux/worktrees/{agent_slug}/
with branches amux/{agent_slug}

#### ListWorktrees

```go
func ListWorktrees(repoRoot string) ([]string, error)
```

ListWorktrees returns a list of all git worktrees

#### RemoveWorktree

```go
func RemoveWorktree(repoRoot, worktreePath string) error
```

RemoveWorktree removes a git worktree and cleans up the branch

#### createBranch

```go
func createBranch(repoRoot, branchName string) error
```

createBranch creates a new git branch if it doesn't exist

#### isGitWorktree

```go
func isGitWorktree(path string) bool
```

isGitWorktree checks if the given path is a git worktree


