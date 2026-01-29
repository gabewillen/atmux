# package merge

`import "github.com/copilot-claude-sonnet-4/amux/internal/merge"`

Package merge provides git merge strategy implementation.
This package handles git merge strategies per spec requirements.

- `ErrMergeConflict, ErrInvalidStrategy, ErrGitCommandFailed` — Common sentinel errors for merge operations.
- `func DryRun(repoRoot, fromBranch string, config Config) error` — DryRun simulates a merge without actually performing it.
- `func ExecuteMerge(repoRoot, fromBranch string, config Config) error` — ExecuteMerge performs a merge using the specified strategy.
- `type Config` — Config contains merge strategy configuration.
- `type Strategy` — Strategy represents a git merge strategy.

