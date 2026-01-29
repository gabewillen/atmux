// Package worktree provides git worktree isolation for amux agents.
//
// Each agent operates within its own git worktree to ensure isolated file
// system changes, independent branch operations, and conflict-free parallel
// work. Worktrees are created under .amux/worktrees/{agent_slug}/ within
// the agent's repo_root.
//
// See spec §5.3 for worktree isolation requirements.
package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// BranchPrefix is the prefix for agent worktree branches.
// Each worktree branch is named amux/{agent_slug} per spec §5.3.1.
const BranchPrefix = "amux/"

// worktreesDir is the relative path from repo_root to the worktrees directory.
const worktreesDir = ".amux/worktrees"

// Manager manages git worktrees for agents.
type Manager struct {
	// gitPath is the path to the git executable.
	gitPath string
}

// NewManager creates a new worktree manager.
func NewManager() *Manager {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		gitPath = "git"
	}
	return &Manager{gitPath: gitPath}
}

// Create creates a git worktree for an agent. The worktree is created at
// {repoRoot}/.amux/worktrees/{slug}/ on branch amux/{slug}.
//
// If the worktree already exists and points to the correct branch, it is
// reused (idempotent). Returns the absolute path to the worktree directory.
//
// See spec §5.3.1 for naming and layout rules.
func (m *Manager) Create(repoRoot, slug string) (string, error) {
	if repoRoot == "" {
		return "", fmt.Errorf("worktree create: %w", amuxerrors.ErrNotInRepository)
	}

	// Validate the repo_root is a git repository
	if !isGitRepo(repoRoot) {
		return "", fmt.Errorf("worktree create: %q is %w", repoRoot, amuxerrors.ErrNotInRepository)
	}

	wtDir := filepath.Join(repoRoot, worktreesDir, slug)
	branch := BranchPrefix + slug

	// Check if worktree already exists and is valid
	if isValidWorktree(wtDir) {
		return wtDir, nil
	}

	// Ensure the parent directory exists
	parentDir := filepath.Dir(wtDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return "", fmt.Errorf("worktree create: mkdir: %w", err)
	}

	// Check if the branch already exists
	branchExists := m.branchExists(repoRoot, branch)

	if branchExists {
		// Create worktree using existing branch
		cmd := exec.Command(m.gitPath, "worktree", "add", wtDir, branch)
		cmd.Dir = repoRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("worktree create: git worktree add: %s: %w",
				strings.TrimSpace(string(output)), amuxerrors.ErrWorktreeCreateFailed)
		}
	} else {
		// Create worktree with a new branch from HEAD
		cmd := exec.Command(m.gitPath, "worktree", "add", "-b", branch, wtDir)
		cmd.Dir = repoRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("worktree create: git worktree add -b: %s: %w",
				strings.TrimSpace(string(output)), amuxerrors.ErrWorktreeCreateFailed)
		}
	}

	return wtDir, nil
}

// Remove removes a git worktree for an agent. It terminates any running
// processes in the worktree first (caller responsibility), then removes
// the worktree and optionally deletes the branch.
//
// See spec §5.3.2 for cleanup rules.
func (m *Manager) Remove(repoRoot, slug string, deleteBranch bool) error {
	if repoRoot == "" {
		return fmt.Errorf("worktree remove: %w", amuxerrors.ErrNotInRepository)
	}

	wtDir := filepath.Join(repoRoot, worktreesDir, slug)
	branch := BranchPrefix + slug

	// Remove the worktree (--force to handle dirty trees)
	cmd := exec.Command(m.gitPath, "worktree", "remove", "--force", wtDir)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the directory doesn't exist, that's fine
		if !os.IsNotExist(err) && !strings.Contains(string(output), "is not a working tree") {
			return fmt.Errorf("worktree remove: %s: %w",
				strings.TrimSpace(string(output)), amuxerrors.ErrWorktreeRemoveFailed)
		}
	}

	// Clean up the directory if it still exists
	if _, err := os.Stat(wtDir); err == nil {
		if err := os.RemoveAll(wtDir); err != nil {
			return fmt.Errorf("worktree remove: cleanup dir: %w", err)
		}
	}

	// Prune stale worktree references
	pruneCmd := exec.Command(m.gitPath, "worktree", "prune")
	pruneCmd.Dir = repoRoot
	_ = pruneCmd.Run()

	// Optionally delete the branch (default: preserve per spec §5.3.2)
	if deleteBranch {
		if err := m.deleteBranch(repoRoot, branch); err != nil {
			return fmt.Errorf("worktree remove: delete branch: %w", err)
		}
	}

	return nil
}

// Exists returns true if a worktree exists for the given slug.
func (m *Manager) Exists(repoRoot, slug string) bool {
	wtDir := filepath.Join(repoRoot, worktreesDir, slug)
	return isValidWorktree(wtDir)
}

// Path returns the worktree directory path for an agent slug.
func (m *Manager) Path(repoRoot, slug string) string {
	return filepath.Join(repoRoot, worktreesDir, slug)
}

// BranchName returns the branch name for an agent slug.
func BranchName(slug string) string {
	return BranchPrefix + slug
}

// IsDirty returns true if the worktree has uncommitted changes.
func (m *Manager) IsDirty(repoRoot, slug string) (bool, error) {
	wtDir := filepath.Join(repoRoot, worktreesDir, slug)
	if !isValidWorktree(wtDir) {
		return false, fmt.Errorf("worktree dirty check: worktree does not exist")
	}

	cmd := exec.Command(m.gitPath, "status", "--porcelain")
	cmd.Dir = wtDir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("worktree dirty check: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// BaseBranch determines the base branch for a repository by running
// git symbolic-ref --quiet --short HEAD in the repo_root.
//
// Returns the current branch name, or an error if in detached HEAD or
// unborn branch state. Per spec §5.7.1.
func (m *Manager) BaseBranch(repoRoot string) (string, error) {
	cmd := exec.Command(m.gitPath, "symbolic-ref", "--quiet", "--short", "HEAD")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("base branch: %w", amuxerrors.ErrDetachedHead)
	}

	branch := strings.TrimSpace(string(output))
	if branch == "" {
		return "", fmt.Errorf("base branch: %w", amuxerrors.ErrDetachedHead)
	}

	return branch, nil
}

// branchExists checks if a git branch exists in the repository.
func (m *Manager) branchExists(repoRoot, branch string) bool {
	cmd := exec.Command(m.gitPath, "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}

// deleteBranch deletes a local git branch.
func (m *Manager) deleteBranch(repoRoot, branch string) error {
	cmd := exec.Command(m.gitPath, "branch", "-D", branch)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "not found") {
			return nil
		}
		return fmt.Errorf("delete branch %q: %s", branch, strings.TrimSpace(string(output)))
	}
	return nil
}

// isGitRepo returns true if the directory is a git repository root.
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// isValidWorktree returns true if the directory is a valid git worktree.
func isValidWorktree(dir string) bool {
	gitFile := filepath.Join(dir, ".git")
	info, err := os.Stat(gitFile)
	if err != nil {
		return false
	}
	// Worktrees have a .git file (not directory) pointing to the main repo
	return !info.IsDir()
}
