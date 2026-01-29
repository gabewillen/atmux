# package git

`import "github.com/agentflare-ai/amux/internal/git"`

Package git provides worktree and merge helpers for local repositories.

- `ErrDirtyWorktree, ErrBranchMissing, ErrMergeConflict, ErrDetachedHead, ErrInvalidStrategy`
- `ErrRepoRequired`
- `ErrWorktreeConflict`
- `func isMissingRef(exitCode int, output []byte) bool`
- `func parseWorktrees(output string) map[string]Worktree`
- `type ExecFunc` — ExecFunc executes a git command in the provided repo root.
- `type ExecResult` — ExecResult captures command output and exit status.
- `type MergeOptions` — MergeOptions configures a merge operation.
- `type MergeResult` — MergeResult describes a merge attempt.
- `type MergeStrategy` — MergeStrategy identifies a supported merge strategy.
- `type Runner` — Runner executes git commands.
- `type Worktree` — Worktree describes a git worktree entry.

### Variables

#### ErrDirtyWorktree, ErrBranchMissing, ErrMergeConflict, ErrDetachedHead, ErrInvalidStrategy

```go
var (
	// ErrDirtyWorktree is returned when a worktree has uncommitted changes.
	ErrDirtyWorktree = errors.New("dirty worktree")
	// ErrBranchMissing is returned when a required branch is missing.
	ErrBranchMissing = errors.New("branch missing")
	// ErrMergeConflict is returned when a merge conflict is detected.
	ErrMergeConflict = errors.New("merge conflict")
	// ErrDetachedHead is returned when base branch detection fails.
	ErrDetachedHead = errors.New("detached head")
	// ErrInvalidStrategy is returned for unsupported strategies.
	ErrInvalidStrategy = errors.New("invalid merge strategy")
)
```

#### ErrRepoRequired

```go
var (
	// ErrRepoRequired is returned when a repo root is required.
	ErrRepoRequired = errors.New("repo root required")
)
```

#### ErrWorktreeConflict

```go
var (
	// ErrWorktreeConflict is returned when a worktree path is in use.
	ErrWorktreeConflict = errors.New("worktree path conflict")
)
```


### Functions

#### isMissingRef

```go
func isMissingRef(exitCode int, output []byte) bool
```

#### parseWorktrees

```go
func parseWorktrees(output string) map[string]Worktree
```


## type ExecFunc

```go
type ExecFunc func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error)
```

ExecFunc executes a git command in the provided repo root.

## type ExecResult

```go
type ExecResult struct {
	Output   []byte
	ExitCode int
}
```

ExecResult captures command output and exit status.

### Functions returning ExecResult

#### defaultExec

```go
func defaultExec(ctx context.Context, repoRoot string, args ...string) (ExecResult, error)
```


## type MergeOptions

```go
type MergeOptions struct {
	RepoRoot     string
	WorktreePath string
	AgentSlug    string
	Strategy     MergeStrategy
	TargetBranch string
	BaseBranch   string
	AllowDirty   bool
}
```

MergeOptions configures a merge operation.

## type MergeResult

```go
type MergeResult struct {
	TargetBranch string
	Strategy     MergeStrategy
}
```

MergeResult describes a merge attempt.

## type MergeStrategy

```go
type MergeStrategy string
```

MergeStrategy identifies a supported merge strategy.

### Constants

#### StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFFOnly

```go
const (
	// StrategyMergeCommit performs a merge commit.
	StrategyMergeCommit MergeStrategy = "merge-commit"
	// StrategySquash performs a squash merge.
	StrategySquash MergeStrategy = "squash"
	// StrategyRebase rebases and fast-forwards.
	StrategyRebase MergeStrategy = "rebase"
	// StrategyFFOnly performs a fast-forward only merge.
	StrategyFFOnly MergeStrategy = "ff-only"
)
```


## type Runner

```go
type Runner struct {
	Exec ExecFunc
}
```

Runner executes git commands.

### Functions returning Runner

#### NewRunner

```go
func NewRunner() *Runner
```

NewRunner constructs a Runner using the default executor.


### Methods

#### Runner.DetectBaseBranch

```go
func () DetectBaseBranch(ctx context.Context, repoRoot, fallback string) (string, error)
```

DetectBaseBranch determines the base branch for a repository.

#### Runner.EnsureWorktree

```go
func () EnsureWorktree(ctx context.Context, repoRoot, agentSlug string) (Worktree, error)
```

EnsureWorktree creates or reuses a worktree for the agent slug.

#### Runner.ListWorktrees

```go
func () ListWorktrees(ctx context.Context, repoRoot string) (map[string]Worktree, error)
```

ListWorktrees returns worktrees keyed by path.

#### Runner.Merge

```go
func () Merge(ctx context.Context, opts MergeOptions) (MergeResult, error)
```

Merge integrates the agent branch into the target branch.

#### Runner.RemoveWorktree

```go
func () RemoveWorktree(ctx context.Context, repoRoot, agentSlug string, deleteBranch bool) error
```

RemoveWorktree removes the worktree and optionally deletes the branch.

#### Runner.ensureBranch

```go
func () ensureBranch(ctx context.Context, repoRoot, branch string) (bool, error)
```

#### Runner.ensureBranchExists

```go
func () ensureBranchExists(ctx context.Context, repoRoot, branch string) error
```

#### Runner.ensureClean

```go
func () ensureClean(ctx context.Context, worktreePath string, allowDirty bool) error
```

#### Runner.hasConflicts

```go
func () hasConflicts(ctx context.Context, repoRoot string) error
```

#### Runner.mergeCommit

```go
func () mergeCommit(ctx context.Context, repoRoot, agentBranch, target string) error
```

#### Runner.mergeFFOnly

```go
func () mergeFFOnly(ctx context.Context, repoRoot, agentBranch, target string) error
```

#### Runner.mergeRebase

```go
func () mergeRebase(ctx context.Context, worktreePath, repoRoot, agentBranch, target string) error
```

#### Runner.mergeSquash

```go
func () mergeSquash(ctx context.Context, repoRoot, agentBranch, target string) error
```

#### Runner.run

```go
func () run(ctx context.Context, repoRoot string, args ...string) (ExecResult, error)
```


## type Worktree

```go
type Worktree struct {
	Path     string
	Branch   string
	Detached bool
	Existing bool
}
```

Worktree describes a git worktree entry.

