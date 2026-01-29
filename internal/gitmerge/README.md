# package gitmerge

`import "github.com/agentflare-ai/amux/internal/gitmerge"`

Package gitmerge implements git merge strategy selection and execution for amux.

This package specifies how changes made in an agent worktree branch
(amux/{agent_slug}) are integrated into a target branch within the same
repository. Merge execution is performed by running local git commands
in the corresponding repo_root.

Supported strategies: merge-commit, squash, rebase, ff-only.

See spec §5.7 for merge strategy requirements.

- `func isConflictError(err error) bool` — isConflictError checks if a git error indicates merge conflicts.
- `func isDirtyWorktree(gitPath, dir string) (bool, error)` — isDirtyWorktree checks if a worktree has uncommitted changes.
- `type Executor` — Executor executes git merge operations.
- `type Request` — Request represents a merge integration request.
- `type Result` — Result represents the result of a merge operation.
- `type Strategy` — Strategy represents a supported merge strategy.

### Functions

#### isConflictError

```go
func isConflictError(err error) bool
```

isConflictError checks if a git error indicates merge conflicts.

#### isDirtyWorktree

```go
func isDirtyWorktree(gitPath, dir string) (bool, error)
```

isDirtyWorktree checks if a worktree has uncommitted changes.


## type Executor

```go
type Executor struct {
	gitPath    string
	dispatcher event.Dispatcher
}
```

Executor executes git merge operations.

### Functions returning Executor

#### NewExecutor

```go
func NewExecutor(dispatcher event.Dispatcher) *Executor
```

NewExecutor creates a new merge executor.


### Methods

#### Executor.Execute

```go
func () Execute(ctx context.Context, req Request) (*Result, error)
```

Execute performs a merge operation per the given request.

It validates preconditions, executes the merge using the selected strategy,
and emits the appropriate events.

See spec §5.7.2-§5.7.5 for strategy behavior, preconditions, and events.

#### Executor.branchExists

```go
func () branchExists(repoRoot, branch string) bool
```

branchExists checks if a branch exists in the repository.

#### Executor.doFFOnly

```go
func () doFFOnly(repoRoot, targetBranch, sourceBranch string) error
```

doFFOnly fast-forwards target to source, failing if not a direct descendant.

#### Executor.doMergeCommit

```go
func () doMergeCommit(repoRoot, targetBranch, sourceBranch string) error
```

doMergeCommit performs a non-fast-forward merge commit.

#### Executor.doRebase

```go
func () doRebase(repoRoot, targetBranch, sourceBranch string) error
```

doRebase rebases source onto target and fast-forwards target.

The source branch is checked out in its worktree, so the rebase runs inside
the worktree directory rather than checking out the branch in the main repo
(git prohibits checking out a branch that is already in a worktree).

#### Executor.doSquash

```go
func () doSquash(repoRoot, targetBranch, sourceBranch string) error
```

doSquash squashes all commits from source into a single commit on target.

#### Executor.executeStrategy

```go
func () executeStrategy(req Request, targetBranch, sourceBranch string) (*Result, error)
```

executeStrategy runs the appropriate git commands for the selected strategy.

#### Executor.getHeadSHA

```go
func () getHeadSHA(repoRoot string) string
```

getHeadSHA returns the current HEAD commit SHA.

#### Executor.gitCheckout

```go
func () gitCheckout(repoRoot, branch string) error
```

gitCheckout checks out a branch.

#### Executor.gitMergeAbort

```go
func () gitMergeAbort(repoRoot string) error
```

gitMergeAbort aborts a merge in progress.

#### Executor.isGitRepo

```go
func () isGitRepo(dir string) bool
```

isGitRepo returns true if the directory is a git repository root.

#### Executor.validatePreconditions

```go
func () validatePreconditions(req Request, targetBranch, sourceBranch string) error
```

validatePreconditions checks merge preconditions per spec §5.7.3.


## type Request

```go
type Request struct {
	// RepoRoot is the absolute path to the repository root.
	RepoRoot string

	// AgentSlug is the agent's slug (branch is amux/{agent_slug}).
	AgentSlug string

	// Strategy is the merge strategy to use.
	Strategy Strategy

	// TargetBranch is the branch to merge into. If empty, uses base_branch.
	TargetBranch string

	// BaseBranch is the branch recorded when the first agent was added.
	BaseBranch string

	// AllowDirty permits merging from a dirty worktree.
	AllowDirty bool

	// AgentID is the agent's runtime ID for event emission.
	AgentID muid.MUID
}
```

Request represents a merge integration request.

## type Result

```go
type Result struct {
	// Strategy is the strategy that was used.
	Strategy Strategy

	// TargetBranch is the branch that was merged into.
	TargetBranch string

	// SourceBranch is the agent branch that was merged from.
	SourceBranch string

	// CommitSHA is the resulting commit hash, if applicable.
	CommitSHA string

	// Conflict indicates merge conflicts were detected.
	Conflict bool
}
```

Result represents the result of a merge operation.

## type Strategy

```go
type Strategy string
```

Strategy represents a supported merge strategy.

### Constants

#### StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFFOnly

```go
const (
	// StrategyMergeCommit creates a non-fast-forward merge commit.
	StrategyMergeCommit Strategy = "merge-commit"

	// StrategySquash squashes all commits into a single commit on target_branch.
	StrategySquash Strategy = "squash"

	// StrategyRebase rebases the agent branch onto target_branch and fast-forwards.
	StrategyRebase Strategy = "rebase"

	// StrategyFFOnly fast-forwards target_branch only if direct descendant.
	StrategyFFOnly Strategy = "ff-only"
)
```


### Functions returning Strategy

#### ParseStrategy

```go
func ParseStrategy(s string) (Strategy, error)
```

ParseStrategy parses a strategy string. Returns an error for unsupported values.

#### ValidStrategies

```go
func ValidStrategies() []Strategy
```

ValidStrategies returns all supported merge strategies.


