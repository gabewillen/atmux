# package merge

`import "github.com/copilot-claude-sonnet-4/amux/internal/merge"`

Package merge provides git merge strategy implementation.
This package handles git merge strategies per spec requirements.

- `ErrMergeConflict, ErrInvalidStrategy, ErrGitCommandFailed` — Common sentinel errors for merge operations.
- `func DryRun(repoRoot, fromBranch string, config Config) error` — DryRun simulates a merge without actually performing it.
- `func ExecuteMerge(repoRoot, fromBranch string, config Config) error` — ExecuteMerge performs a merge using the specified strategy.
- `func executeFastForwardOnly(repoRoot, fromBranch, toBranch string) error` — executeFastForwardOnly performs a fast-forward only merge.
- `func executeMergeCommit(repoRoot, fromBranch, toBranch string) error` — executeMergeCommit performs a merge commit.
- `func executeRebase(repoRoot, fromBranch, toBranch string) error` — executeRebase performs a rebase merge.
- `func executeSquash(repoRoot, fromBranch, toBranch string) error` — executeSquash performs a squash merge.
- `type Config` — Config contains merge strategy configuration.
- `type Strategy` — Strategy represents a git merge strategy.

### Variables

#### ErrMergeConflict, ErrInvalidStrategy, ErrGitCommandFailed

```go
var (
	// ErrMergeConflict indicates a merge conflict occurred.
	ErrMergeConflict = errors.New("merge conflict")

	// ErrInvalidStrategy indicates an invalid merge strategy.
	ErrInvalidStrategy = errors.New("invalid merge strategy")

	// ErrGitCommandFailed indicates a git command failed.
	ErrGitCommandFailed = errors.New("git command failed")
)
```

Common sentinel errors for merge operations.


### Functions

#### DryRun

```go
func DryRun(repoRoot, fromBranch string, config Config) error
```

DryRun simulates a merge without actually performing it.

#### ExecuteMerge

```go
func ExecuteMerge(repoRoot, fromBranch string, config Config) error
```

ExecuteMerge performs a merge using the specified strategy.

#### executeFastForwardOnly

```go
func executeFastForwardOnly(repoRoot, fromBranch, toBranch string) error
```

executeFastForwardOnly performs a fast-forward only merge.

#### executeMergeCommit

```go
func executeMergeCommit(repoRoot, fromBranch, toBranch string) error
```

executeMergeCommit performs a merge commit.

#### executeRebase

```go
func executeRebase(repoRoot, fromBranch, toBranch string) error
```

executeRebase performs a rebase merge.

#### executeSquash

```go
func executeSquash(repoRoot, fromBranch, toBranch string) error
```

executeSquash performs a squash merge.


## type Config

```go
type Config struct {
	Strategy     Strategy `toml:"strategy"`
	BaseBranch   string   `toml:"base_branch"`
	TargetBranch string   `toml:"target_branch"`
}
```

Config contains merge strategy configuration.

### Functions returning Config

#### DefaultConfig

```go
func DefaultConfig() Config
```

DefaultConfig returns the default merge configuration.


### Methods

#### Config.Validate

```go
func () Validate() error
```

Validate checks if the merge config is valid.


## type Strategy

```go
type Strategy string
```

Strategy represents a git merge strategy.

### Constants

#### StrategyMergeCommit, StrategySquash, StrategyRebase, StrategyFastForwardOnly

```go
const (
	// StrategyMergeCommit creates a merge commit.
	StrategyMergeCommit Strategy = "merge-commit"

	// StrategySquash squashes commits before merging.
	StrategySquash Strategy = "squash"

	// StrategyRebase rebases before merging.
	StrategyRebase Strategy = "rebase"

	// StrategyFastForwardOnly only allows fast-forward merges.
	StrategyFastForwardOnly Strategy = "ff-only"
)
```


