// Package worktree provides git worktree create/remove for agent isolation (spec §5.3, §5.3.1, §5.3.2).
// Worktrees are created under .amux/worktrees/{agent_slug}/ with branch amux/{agent_slug}.
package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// WorktreesDir is the relative path under repo root: .amux/worktrees
	WorktreesDir = ".amux/worktrees"
	// BranchPrefix is the prefix for agent worktree branches: amux/{agent_slug}
	BranchPrefix = "amux/"
)

// WorktreePath returns the absolute path to the worktree directory for agentSlug under repoRoot.
func WorktreePath(repoRoot, agentSlug string) string {
	return filepath.Join(repoRoot, WorktreesDir, agentSlug)
}

// BranchName returns the branch name for an agent worktree (spec §5.3.1).
func BranchName(agentSlug string) string {
	return BranchPrefix + agentSlug
}

// Create creates or reuses the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.1).
// Idempotent: if the worktree already exists and is valid, returns nil.
// Creates branch amux/{agentSlug} from current HEAD if needed, then "git worktree add".
func Create(repoRoot, agentSlug string) (string, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("absolute repo root: %w", err)
	}
	wtPath := WorktreePath(repoRoot, agentSlug)

	// Ensure .amux/worktrees exists
	wtDir := filepath.Dir(wtPath)
	if err := os.MkdirAll(wtDir, 0755); err != nil {
		return "", fmt.Errorf("create worktrees dir: %w", err)
	}

	// If path already exists, check if it's a valid worktree
	if _, err := os.Stat(wtPath); err == nil {
		// Verify it's a git worktree (has .git file or .git dir)
		gitMark := filepath.Join(wtPath, ".git")
		if _, err := os.Stat(gitMark); err == nil {
			return wtPath, nil
		}
		// Path exists but not a worktree; fail to avoid overwriting
		return "", fmt.Errorf("path exists and is not a worktree: %s", wtPath)
	}

	branch := BranchName(agentSlug)

	// Ensure branch exists (create from HEAD if not)
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		// Branch doesn't exist; create it from HEAD
		cmd = exec.Command("git", "branch", branch)
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("create branch %s: %w", branch, err)
		}
	}

	cmd = exec.Command("git", "worktree", "add", wtPath, branch)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree add: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return wtPath, nil
}

// Remove removes the worktree at .amux/worktrees/{agentSlug}/ (spec §5.3.2).
// Runs "git worktree remove .amux/worktrees/{agentSlug}".
func Remove(repoRoot, agentSlug string) error {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("absolute repo root: %w", err)
	}
	wtPath := WorktreePath(repoRoot, agentSlug)
	cmd := exec.Command("git", "worktree", "remove", wtPath)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Exists reports whether a worktree at .amux/worktrees/{agentSlug}/ exists and is a valid worktree.
func Exists(repoRoot, agentSlug string) bool {
	wtPath := WorktreePath(repoRoot, agentSlug)
	gitMark := filepath.Join(wtPath, ".git")
	_, err := os.Stat(gitMark)
	return err == nil
}
