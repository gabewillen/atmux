// Package git provides git operations for worktree management.
// This package implements git worktree isolation per spec requirements.
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Common sentinel errors for git operations.
var (
	// ErrWorktreeExists indicates a worktree already exists at the given path.
	ErrWorktreeExists = errors.New("worktree already exists")

	// ErrBranchExists indicates a branch already exists.
	ErrBranchExists = errors.New("branch already exists")

	// ErrGitCommandFailed indicates a git command failed.
	ErrGitCommandFailed = errors.New("git command failed")
)

// CreateWorktree creates a git worktree for the given agent slug.
// This implements worktree isolation per spec: .amux/worktrees/{agent_slug}/
// with branches amux/{agent_slug}
func CreateWorktree(repoRoot, agentSlug, worktreePath string) error {
	if worktreePath == "" {
		return fmt.Errorf("worktree path required")
	}

	// Check if worktree directory already exists
	if _, err := os.Stat(worktreePath); err == nil {
		// Directory exists - check if it's already a git worktree
		if isGitWorktree(worktreePath) {
			return nil // Idempotent - already exists
		}
		return fmt.Errorf("directory exists but is not a git worktree: %s", worktreePath)
	}

	branchName := fmt.Sprintf("amux/%s", agentSlug)
	
	// Create branch if it doesn't exist
	if err := createBranch(repoRoot, branchName); err != nil && !errors.Is(err, ErrBranchExists) {
		return fmt.Errorf("failed to create branch %s: %w", branchName, err)
	}

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create worktree: %s: %w", string(output), ErrGitCommandFailed)
	}

	return nil
}

// isGitWorktree checks if the given path is a git worktree
func isGitWorktree(path string) bool {
	gitFile := filepath.Join(path, ".git")
	if info, err := os.Stat(gitFile); err == nil {
		// If .git is a file (not directory), it's likely a worktree
		return !info.IsDir()
	}
	return false
}

// createBranch creates a new git branch if it doesn't exist
func createBranch(repoRoot, branchName string) error {
	// Check if branch exists
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err == nil {
		// Branch exists
		return ErrBranchExists
	}

	// Create branch
	cmd = exec.Command("git", "checkout", "-b", branchName)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		// If we're not on main/master, try creating from current branch
		if strings.Contains(string(output), "already exists") {
			return ErrBranchExists
		}
		return fmt.Errorf("failed to create branch %s: %s: %w", branchName, string(output), ErrGitCommandFailed)
	}

	// Switch back to original branch
	cmd = exec.Command("git", "checkout", "-")
	cmd.Dir = repoRoot
	_ = cmd.Run() // Ignore error - not critical

	return nil
}

// RemoveWorktree removes a git worktree and cleans up the branch
func RemoveWorktree(repoRoot, worktreePath string) error {
	if worktreePath == "" {
		return fmt.Errorf("worktree path required")
	}

	// Remove worktree
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		// If worktree doesn't exist, that's fine (idempotent)
		if !strings.Contains(string(output), "is not a working tree") {
			return fmt.Errorf("failed to remove worktree: %s: %w", string(output), ErrGitCommandFailed)
		}
	}

	return nil
}

// ListWorktrees returns a list of all git worktrees
func ListWorktrees(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %s: %w", string(output), ErrGitCommandFailed)
	}

	var worktrees []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			worktrees = append(worktrees, path)
		}
	}

	return worktrees, nil
}