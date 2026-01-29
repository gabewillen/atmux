# package git

`import "github.com/agentflare-ai/amux/internal/git"`

Package git provides git repository queries and merge-strategy helpers for amux.
Used for repo validation, base_branch resolution, and merge strategy selection (spec §5.3.4, §5.7, §5.7.1).

- `ValidMergeStrategies` — ValidMergeStrategies are the supported git merge strategies (spec §5.7.2).
- `func BaseBranch(repoRoot string) (string, error)` — BaseBranch returns the current branch name in repoRoot (spec §5.7.1).
- `func IsRepo(dir string) (bool, error)` — IsRepo reports whether dir is the root of a git repository.
- `func ResolveTargetBranch(repoRoot, configuredTarget string) (string, error)` — ResolveTargetBranch returns the branch to merge into (spec §5.7.1).
- `func Root(dir string) (string, error)` — Root returns the repository root directory containing dir, or empty string if not in a repo.
- `func ValidStrategy(s string) bool` — ValidStrategy returns true if s is one of merge-commit, squash, rebase, ff-only.

### Variables

#### ValidMergeStrategies

```go
var ValidMergeStrategies = []string{"merge-commit", "squash", "rebase", "ff-only"}
```

ValidMergeStrategies are the supported git merge strategies (spec §5.7.2).


### Functions

#### BaseBranch

```go
func BaseBranch(repoRoot string) (string, error)
```

BaseBranch returns the current branch name in repoRoot (spec §5.7.1).
It runs "git symbolic-ref --quiet --short HEAD" and returns the trimmed output.
If the command fails (detached HEAD or unborn branch), returns empty string and a non-nil error.

#### IsRepo

```go
func IsRepo(dir string) (bool, error)
```

IsRepo reports whether dir is the root of a git repository.
It runs "git rev-parse --is-inside-work-tree" from dir.
Non-zero exit (e.g. not a repo, missing dir) is treated as not a repo.

#### ResolveTargetBranch

```go
func ResolveTargetBranch(repoRoot, configuredTarget string) (string, error)
```

ResolveTargetBranch returns the branch to merge into (spec §5.7.1).
If configuredTarget is non-empty, returns it; otherwise returns base_branch from repoRoot.
If base_branch cannot be determined (detached HEAD), returns an error instructing the user to set git.merge.target_branch.

#### Root

```go
func Root(dir string) (string, error)
```

Root returns the repository root directory containing dir, or empty string if not in a repo.

#### ValidStrategy

```go
func ValidStrategy(s string) bool
```

ValidStrategy returns true if s is one of merge-commit, squash, rebase, ff-only.


