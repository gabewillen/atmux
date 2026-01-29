package worktree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Manager handles worktree lifecycle for agents.
type Manager struct {
	resolver *paths.Resolver
}

// NewManager creates a new worktree manager.
func NewManager() (*Manager, error) {
	r, err := paths.NewResolver()
	if err != nil {
		return nil, err
	}
	return &Manager{resolver: r}, nil
}

// Ensure creates or validates the worktree for the given agent.
// It ensures the target branch `amux/<slug>` exists and the worktree at `.amux/worktrees/<slug>` is active.
func (m *Manager) Ensure(agent api.Agent) (string, error) {
	if agent.RepoRoot == "" {
		return "", fmt.Errorf("agent %s has no repo root", agent.Name)
	}

	// 1. Validate RepoRoot
	if !IsRepo(agent.RepoRoot) {
		return "", fmt.Errorf("path %s is not a git repository", agent.RepoRoot)
	}

	// 2. Resolve target path
	// .amux/worktrees/<slug>
	worktreeDir := m.resolver.WorktreesDir(agent.RepoRoot)
	targetPath := filepath.Join(worktreeDir, agent.Slug.String())

	// 3. Ensure branch `amux/<slug>`
	// We'll base it off HEAD for now as per simple start defaults.
	// Ideally we'd have configuration for base branch.
	branchName := fmt.Sprintf("amux/%s", agent.Slug)

	// Get current HEAD to use as start point if branch doesn't exist
	// We can pass "" to EnsureBranch to let git default to HEAD, but being explicit is safer?
	// Actually git branch <name> defaults to HEAD.
	if err := EnsureBranch(agent.RepoRoot, branchName, ""); err != nil {
		return "", errors.Wrap(err, "failed to ensure agent branch")
	}

	// 4. Create Worktree
	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil {
		// Verify it's actually a worktree?
		// For now assume yes if it exists in the expected location.
		// We could use `git worktree list` to be sure.
		// But returning success is fine for idempotency.
		return targetPath, nil
	}

	if err := AddWorktree(agent.RepoRoot, targetPath, branchName); err != nil {
		return "", errors.Wrap(err, "failed to create worktree")
	}

	return targetPath, nil
}

// Remove removes the worktree for the given agent.
func (m *Manager) Remove(agent api.Agent) error {
	if agent.RepoRoot == "" {
		return nil // Nothing to do
	}

	worktreeDir := m.resolver.WorktreesDir(agent.RepoRoot)
	targetPath := filepath.Join(worktreeDir, agent.Slug.String())

	return RemoveWorktree(agent.RepoRoot, targetPath)
}

// MergeAgent merges the agent's worktree branch back into the base branch.
// It respects the allow_dirty configuration.
func (m *Manager) MergeAgent(agent api.Agent, strategy string, allowDirty bool) error {
	if agent.RepoRoot == "" {
		return fmt.Errorf("agent %s has no repo root", agent.Name)
	}

	worktreeDir := m.resolver.WorktreesDir(agent.RepoRoot)
	targetPath := filepath.Join(worktreeDir, agent.Slug.String())

	// 1. Check dirty status
	dirty, err := IsDirty(targetPath)
	if err != nil {
		return errors.Wrap(err, "failed to check dirty status")
	}
	if dirty && !allowDirty {
		return errors.Errorf("worktree is dirty and allow_dirty is false")
	}

	// 2. Determine branches
	// Use explicit branch name: amux/<slug>
	agentBranch := fmt.Sprintf("amux/%s", agent.Slug)

	// 3. Resolve base branch (target for merge)
	// TODO: Support configuration override for base_branch
	baseBranch, err := GetDefaultBranch(agent.RepoRoot)
	if err != nil {
		return errors.Wrap(err, "failed to resolve base branch")
	}

	// 4. Checkout base branch in repo root to prepare for merge
	// Note: This modifies the state of the main repo check out.
	if err := Checkout(agent.RepoRoot, baseBranch); err != nil {
		return errors.Wrapf(err, "failed to checkout base branch %q", baseBranch)
	}

	// 5. Execute merge
	return Merge(agent.RepoRoot, agentBranch, strategy)
}
