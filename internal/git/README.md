# package git

`import "github.com/copilot-claude-sonnet-4/amux/internal/git"`

Package git provides git operations for worktree management.
This package implements git worktree isolation per spec requirements.

- `ErrWorktreeExists, ErrBranchExists, ErrGitCommandFailed` — Common sentinel errors for git operations.
- `func CreateWorktree(repoRoot, agentSlug, worktreePath string) error` — CreateWorktree creates a git worktree for the given agent slug.
- `func ListWorktrees(repoRoot string) ([]string, error)` — ListWorktrees returns a list of all git worktrees
- `func RemoveWorktree(repoRoot, worktreePath string) error` — RemoveWorktree removes a git worktree and cleans up the branch

