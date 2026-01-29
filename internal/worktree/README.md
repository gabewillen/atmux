# package worktree

`import "github.com/agentflare-ai/amux/internal/worktree"`

Package worktree provides git worktree isolation for amux agents.

Each agent operates within its own git worktree to ensure isolated file
system changes, independent branch operations, and conflict-free parallel
work. Worktrees are created under .amux/worktrees/{agent_slug}/ within
the agent's repo_root.

See spec §5.3 for worktree isolation requirements.

- `BranchPrefix` — BranchPrefix is the prefix for agent worktree branches.
- `func BranchName(slug string) string` — BranchName returns the branch name for an agent slug.
- `func isGitRepo(dir string) bool` — isGitRepo returns true if the directory is a git repository root.
- `func isValidWorktree(dir string) bool` — isValidWorktree returns true if the directory is a valid git worktree.
- `type Manager` — Manager manages git worktrees for agents.
- `worktreesDir` — worktreesDir is the relative path from repo_root to the worktrees directory.

### Constants

#### BranchPrefix

```go
const BranchPrefix = "amux/"
```

BranchPrefix is the prefix for agent worktree branches.
Each worktree branch is named amux/{agent_slug} per spec §5.3.1.

#### worktreesDir

```go
const worktreesDir = ".amux/worktrees"
```

worktreesDir is the relative path from repo_root to the worktrees directory.


### Functions

#### BranchName

```go
func BranchName(slug string) string
```

BranchName returns the branch name for an agent slug.

#### isGitRepo

```go
func isGitRepo(dir string) bool
```

isGitRepo returns true if the directory is a git repository root.

#### isValidWorktree

```go
func isValidWorktree(dir string) bool
```

isValidWorktree returns true if the directory is a valid git worktree.


## type Manager

```go
type Manager struct {
	// gitPath is the path to the git executable.
	gitPath string
}
```

Manager manages git worktrees for agents.

### Functions returning Manager

#### NewManager

```go
func NewManager() *Manager
```

NewManager creates a new worktree manager.


### Methods

#### Manager.BaseBranch

```go
func () BaseBranch(repoRoot string) (string, error)
```

BaseBranch determines the base branch for a repository by running
git symbolic-ref --quiet --short HEAD in the repo_root.

Returns the current branch name, or an error if in detached HEAD or
unborn branch state. Per spec §5.7.1.

#### Manager.Create

```go
func () Create(repoRoot, slug string) (string, error)
```

Create creates a git worktree for an agent. The worktree is created at
{repoRoot}/.amux/worktrees/{slug}/ on branch amux/{slug}.

If the worktree already exists and points to the correct branch, it is
reused (idempotent). Returns the absolute path to the worktree directory.

See spec §5.3.1 for naming and layout rules.

#### Manager.Exists

```go
func () Exists(repoRoot, slug string) bool
```

Exists returns true if a worktree exists for the given slug.

#### Manager.IsDirty

```go
func () IsDirty(repoRoot, slug string) (bool, error)
```

IsDirty returns true if the worktree has uncommitted changes.

#### Manager.Path

```go
func () Path(repoRoot, slug string) string
```

Path returns the worktree directory path for an agent slug.

#### Manager.Remove

```go
func () Remove(repoRoot, slug string, deleteBranch bool) error
```

Remove removes a git worktree for an agent. It terminates any running
processes in the worktree first (caller responsibility), then removes
the worktree and optionally deletes the branch.

See spec §5.3.2 for cleanup rules.

#### Manager.branchExists

```go
func () branchExists(repoRoot, branch string) bool
```

branchExists checks if a git branch exists in the repository.

#### Manager.deleteBranch

```go
func () deleteBranch(repoRoot, branch string) error
```

deleteBranch deletes a local git branch.


