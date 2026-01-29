# package git

`import "github.com/stateforward/amux/internal/git"`

Package git implements git operations and merge strategies for the amux project

- `func GetBaseBranch(repoPath string, targetBranch string) (string, error)` — GetBaseBranch determines the base branch for a repository Following the spec: run `git symbolic-ref --quiet --short HEAD` and use the output If that fails, use targetBranch if provided, otherwise return an error
- `func GetCurrentBranch(repoPath string) (string, error)` — GetCurrentBranch returns the current git branch name
- `func PerformMerge(repoPath string, opts MergeOptions) error` — PerformMerge executes a git merge operation based on the specified strategy
- `func performFFOnly(repoPath string, opts MergeOptions) error` — performFFOnly performs a fast-forward only merge
- `func performMergeCommit(repoPath string, opts MergeOptions) error` — performMergeCommit performs a standard merge-commit operation
- `func performRebase(repoPath string, opts MergeOptions) error` — performRebase performs a rebase operation
- `func performSquash(repoPath string, opts MergeOptions) error` — performSquash performs a squash merge operation
- `func simulateMerge(repoPath string, opts MergeOptions) error` — simulateMerge shows what would happen with the merge without actually performing it
- `type MergeOptions` — MergeOptions holds options for git merge operations
- `type MergeStrategy` — MergeStrategy represents different git merge strategies

### Functions

#### GetBaseBranch

```go
func GetBaseBranch(repoPath string, targetBranch string) (string, error)
```

GetBaseBranch determines the base branch for a repository
Following the spec: run `git symbolic-ref --quiet --short HEAD` and use the output
If that fails, use targetBranch if provided, otherwise return an error

#### GetCurrentBranch

```go
func GetCurrentBranch(repoPath string) (string, error)
```

GetCurrentBranch returns the current git branch name

#### PerformMerge

```go
func PerformMerge(repoPath string, opts MergeOptions) error
```

PerformMerge executes a git merge operation based on the specified strategy

#### performFFOnly

```go
func performFFOnly(repoPath string, opts MergeOptions) error
```

performFFOnly performs a fast-forward only merge

#### performMergeCommit

```go
func performMergeCommit(repoPath string, opts MergeOptions) error
```

performMergeCommit performs a standard merge-commit operation

#### performRebase

```go
func performRebase(repoPath string, opts MergeOptions) error
```

performRebase performs a rebase operation

#### performSquash

```go
func performSquash(repoPath string, opts MergeOptions) error
```

performSquash performs a squash merge operation

#### simulateMerge

```go
func simulateMerge(repoPath string, opts MergeOptions) error
```

simulateMerge shows what would happen with the merge without actually performing it


## type MergeOptions

```go
type MergeOptions struct {
	Strategy     MergeStrategy
	BaseBranch   string // Source branch to merge from
	TargetBranch string // Target branch to merge into
	DryRun       bool   // If true, only show what would be done
}
```

MergeOptions holds options for git merge operations

## type MergeStrategy

```go
type MergeStrategy string
```

MergeStrategy represents different git merge strategies

### Constants

#### MergeCommit, Squash, Rebase, FFOnly

```go
const (
	MergeCommit MergeStrategy = "merge-commit"
	Squash      MergeStrategy = "squash"
	Rebase      MergeStrategy = "rebase"
	FFOnly      MergeStrategy = "ff-only"
)
```


