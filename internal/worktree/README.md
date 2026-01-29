# package worktree

`import "github.com/agentflare-ai/amux/internal/worktree"`

Package worktree provides git worktree create/remove for agent isolation (spec §5.3, §5.3.1, §5.3.2).
Worktrees are created under .amux/worktrees/{agent_slug}/ with branch amux/{agent_slug}.

- `WorktreesDir, BranchPrefix`
- `func BranchName(agentSlug string) string` — BranchName returns the branch name for an agent worktree (spec §5.3.1).
- `func Create(repoRoot, agentSlug string) (string, error)` — Create creates or reuses the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.1).
- `func Exists(repoRoot, agentSlug string) bool` — Exists reports whether a worktree at .amux/worktrees/{agentSlug}/ exists and is a valid worktree.
- `func Remove(repoRoot, agentSlug string) error` — Remove removes the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.2).
- `func WorktreePath(repoRoot, agentSlug string) string` — WorktreePath returns the absolute path to the worktree directory for agentSlug under repoRoot.

### Constants

#### WorktreesDir, BranchPrefix

```go
const (
	// WorktreesDir is the relative path under repo root: .amux/worktrees
	WorktreesDir = ".amux/worktrees"
	// BranchPrefix is the prefix for agent worktree branches: amux/{agent_slug}
	BranchPrefix = "amux/"
)
```


### Functions

#### BranchName

```go
func BranchName(agentSlug string) string
```

BranchName returns the branch name for an agent worktree (spec §5.3.1).

#### Create

```go
func Create(repoRoot, agentSlug string) (string, error)
```

Create creates or reuses the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.1).
Idempotent: if the worktree already exists and is valid, returns nil.
Creates branch amux/{agentSlug} from current HEAD if needed, then "git worktree add".

#### Exists

```go
func Exists(repoRoot, agentSlug string) bool
```

Exists reports whether a worktree at .amux/worktrees/{agentSlug}/ exists and is a valid worktree.

#### Remove

```go
func Remove(repoRoot, agentSlug string) error
```

Remove removes the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.2).
Runs "git worktree remove .amux/worktrees/{agentSlug}".

#### WorktreePath

```go
func WorktreePath(repoRoot, agentSlug string) string
```

WorktreePath returns the absolute path to the worktree directory for agentSlug under repoRoot.


