package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// EnsureWorktree creates or reuses a worktree for the given agent.
// It returns the path to the worktree.
func EnsureWorktree(repoRoot api.RepoRoot, slug api.AgentSlug, baseBranch string) (string, error) {
	if err := slug.Validate(); err != nil {
		return "", fmt.Errorf("invalid agent slug: %w", err)
	}

	if baseBranch == "" {
		baseBranch = "HEAD"
	}

	worktreesDir := paths.DefaultWorktreesDir(string(repoRoot))
	agentWorktreePath := filepath.Join(worktreesDir, string(slug))

	// Check if exists
	info, err := os.Stat(agentWorktreePath)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("worktree path %s exists but is not a directory", agentWorktreePath)
		}
		// Idempotent reuse
		// In a real implementation we might check if it's actually a git worktree or valid state.
		return agentWorktreePath, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check worktree path: %w", err)
	}

	// Create directory structure
	// The spec says: "Worktrees shall be created in .amux/worktrees/{agent_slug}/ under the agent’s repo_root".
	// Ensure parent dir exists
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	// Create git worktree
	// branch name: amux/{slug}
	branchName := fmt.Sprintf("amux/%s", slug)

	// git worktree add -B amux/{slug} .amux/worktrees/{slug} {baseBranch}
	cmd := exec.Command("git", "worktree", "add", "-B", branchName, agentWorktreePath, baseBranch)
	cmd.Dir = string(repoRoot)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create git worktree: %s: %w", string(out), err)
	}

	return agentWorktreePath, nil
}

// RemoveWorktree removes the worktree for the given agent.
func RemoveWorktree(repoRoot api.RepoRoot, slug api.AgentSlug) error {
	worktreesDir := paths.DefaultWorktreesDir(string(repoRoot))
	agentWorktreePath := filepath.Join(worktreesDir, string(slug))

	// Check if it exists first
	if _, err := os.Stat(agentWorktreePath); os.IsNotExist(err) {
		return nil
	}

	// git worktree remove
	cmd := exec.Command("git", "worktree", "remove", "--force", agentWorktreePath)
	cmd.Dir = string(repoRoot)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove worktree: %s: %w", string(out), err)
	}

	// Prune to clean up metadata
	cmdPrune := exec.Command("git", "worktree", "prune")
	cmdPrune.Dir = string(repoRoot)
	_ = cmdPrune.Run() // Ignore error on prune

	return nil
}