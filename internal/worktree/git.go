package worktree

import (
	"os/exec"
	"strings"

	"github.com/agentflare-ai/amux/internal/errors"
)

// IsRepo checks if the given path is a valid git repository.
func IsRepo(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// EnsureBranch ensures a branch exists.
// If it doesn't exist, it creates it starting from startPoint (default HEAD).
func EnsureBranch(repoPath, branch, startPoint string) error {
	// Check if branch exists
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err == nil {
		return nil // Branch exists
	}

	// Create branch
	args := []string{"branch", branch}
	if startPoint != "" {
		args = append(args, startPoint)
	}
	cmd = exec.Command("git", args...)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "failed to create branch %q: %s", branch, string(out))
	}
	return nil
}

// AddWorktree adds a worktree.
// git worktree add <path> <branch>
func AddWorktree(repoPath, worktreePath, branch string) error {
	// Check if already exists?
	// git worktree add works, but if path exists and is not empty it might fail.
	// We assume caller handles directory flags/cleanup if needed, or we rely on git's error.

	cmd := exec.Command("git", "worktree", "add", worktreePath, branch)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		// If it's "already registered", that's fine?
		// "fatal: '...' is already a worktree"
		if strings.Contains(string(out), "already a worktree") {
			return nil
		}
		return errors.Wrapf(err, "failed to add worktree: %s", string(out))
	}
	return nil
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(repoPath, worktreePath string) error {
	// git worktree remove <path>
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath) // Force to ignore uncommitted changes?
	// Spec says "cleanup on remove as specified".
	// Phase 2 plan says "Remove(agent *api.Agent) error: Cleans up the worktree."
	// We should probably allow dirty? Or force?
	// Spec §5.3.4: "Pruning worktrees ... MUST use `git worktree remove --force`"
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		// If path doesn't exist etc.
		if strings.Contains(string(out), "is not a working tree") {
			return nil
		}
		return errors.Wrapf(err, "failed to remove worktree: %s", string(out))
	}
	return nil
}

// WorktreeList returns a map of path -> branch/commit details (simplified).
// We might just need to check existence.
func WorktreeList(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list worktrees")
	}

	var paths []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			paths = append(paths, strings.TrimPrefix(line, "worktree "))
		}
	}
	return paths, nil
}

func GetHeadBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current branch")
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDefaultBranch attempts to determine the default branch of the repository.
// It tries origin's HEAD first, then falls back to main/master.
func GetDefaultBranch(repoPath string) (string, error) {
	// Try to get the remote HEAD
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err == nil {
		// Output looks like "refs/remotes/origin/main"
		ref := strings.TrimSpace(string(out))
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main exists
	if err := EnsureBranch(repoPath, "main", ""); err == nil {
		return "main", nil
	}
	// Check if master exists
	if err := EnsureBranch(repoPath, "master", ""); err == nil {
		return "master", nil
	}

	return "", errors.New("could not determine default branch")
}

// IsDirty checks if the working directory has uncommitted changes.
func IsDirty(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return false, errors.Wrap(err, "failed to check dirty status")
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// Merge merges the given branch into the current HEAD using the specified strategy.
// Supported strategies: merge-commit, squash, rebase, ff-only.
func Merge(repoPath, branch, strategy string) error {
	var args []string
	switch strategy {
	case "squash":
		args = []string{"merge", "--squash", branch}
	case "rebase":
		args = []string{"rebase", branch}
	case "ff-only":
		args = []string{"merge", "--ff-only", branch}
	case "merge-commit":
		args = []string{"merge", "--no-ff", branch}
	default:
		// Default to standard merge if unknown, or error?
		// Plan says "Supported strategies: merge-commit, squash, rebase, ff-only"
		// Better to restrict.
		return errors.Errorf("unsupported merge strategy: %q", strategy)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "git merge failed (%s): %s", strategy, string(out))
	}
	return nil
}

// Checkout checks out the specified branch in the repository.
func Checkout(repoPath, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "git checkout failed: %s", string(out))
	}
	return nil
}

// Commit creates a commit with the given message.
// Useful for completing a squash merge which leaves changes staged.
func Commit(repoPath, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "git commit failed: %s", string(out))
	}
	return nil
}
