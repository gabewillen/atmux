// Package agent implements agent orchestration (lifecycle, presence, messaging)
package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/stateforward/hsm-go/muid"
	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/git"
	"github.com/stateforward/amux/internal/ids"
	"github.com/stateforward/amux/internal/paths"
	"github.com/stateforward/amux/pkg/api"
)

// AgentProcess represents a running agent process
type AgentProcess struct {
	ID      muid.MUID
	Cmd     *exec.Cmd
	PTY     *os.File
	WorkDir string
}

// AgentManager manages multiple agents and their lifecycles
type AgentManager struct {
	agents      map[muid.MUID]*AgentActor
	processes   map[muid.MUID]*AgentProcess
	agentsMutex sync.RWMutex
	config      *config.Config
	resolver    *paths.Resolver
}

// NewAgentManager creates a new agent manager
func NewAgentManager(cfg *config.Config) (*AgentManager, error) {
	resolver := paths.New(paths.Config{
		BaseDir:  cfg.Core.RepoRoot,
		RepoRoot: cfg.Core.RepoRoot,
	})

	manager := &AgentManager{
		agents:    make(map[muid.MUID]*AgentActor),
		processes: make(map[muid.MUID]*AgentProcess),
		config:    cfg,
		resolver:  resolver,
	}

	return manager, nil
}

// AddAgent adds a new agent to the manager
func (m *AgentManager) AddAgent(ctx context.Context, name, about, adapter string, location *api.Location) error {
	// Validate inputs
	if name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	if adapter == "" {
		return fmt.Errorf("agent adapter cannot be empty")
	}
	if location == nil {
		return fmt.Errorf("agent location cannot be nil")
	}

	// Check if we're in a git repository
	repoRoot, err := m.findGitRepo(location)
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Normalize the agent name to create a slug
	slug := ids.NormalizeAgentSlug(name)

	// Create the agent
	agentID := muid.Make()
	agent := &api.Agent{
		ID:       agentID,
		Name:     name,
		Adapter:  adapter,
		Location: api.Location{Type: "worktree", RepoPath: fmt.Sprintf("worktree://%s", slug)}, // Store location as worktree reference
		RepoRoot: repoRoot,
		HostID:   0, // Will be set when the agent is actually deployed
	}

	// Create the worktree for the agent
	worktreePath := m.resolver.WorktreeDir(slug)
	if err := m.createOrReuseWorktree(repoRoot, worktreePath, slug); err != nil {
		return fmt.Errorf("failed to create worktree for agent: %w", err)
	}

	// Create the agent actor
	actor, err := NewAgentActor(agent, func(event interface{}) {
		// Handle agent events (placeholder)
	})
	if err != nil {
		return fmt.Errorf("failed to create agent actor: %w", err)
	}

	// Add to the manager
	m.agentsMutex.Lock()
	m.agents[agentID] = actor
	m.agentsMutex.Unlock()

	// Emit agent.added event
	// TODO: Implement proper event emission

	return nil
}

// findGitRepo determines the git repository root based on the location
func (m *AgentManager) findGitRepo(location *api.Location) (string, error) {
	// For local agents, we need to find the git repository
	if location.Type == "local" {
		// If repo_path is provided, use it
		if location.RepoPath != "" {
			if !m.isGitRepo(location.RepoPath) {
				return "", fmt.Errorf("location.repo_path is not a git repository: %s", location.RepoPath)
			}
			return location.RepoPath, nil
		}

		// Otherwise, use the configured repo root
		if !m.isGitRepo(m.config.Core.RepoRoot) {
			return "", fmt.Errorf("configured repo_root is not a git repository: %s", m.config.Core.RepoRoot)
		}
		return m.config.Core.RepoRoot, nil
	}

	return "", fmt.Errorf("unsupported location type: %s", location.Type)
}

// isGitRepo checks if a directory is a git repository
func (m *AgentManager) isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// createOrReuseWorktree creates or reuses a git worktree for the agent
func (m *AgentManager) createOrReuseWorktree(repoRoot, worktreePath, agentSlug string) error {
	// Ensure the worktrees directory exists
	worktreesDir := m.resolver.WorktreesDir()
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		// Worktree already exists, reuse it
		return nil
	}

	// Create the worktree
	branchName := m.resolver.AgentBranch(agentSlug)

	// First, create the branch if it doesn't exist
	cmd := exec.Command("git", "branch", "-f", branchName)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		// If the branch doesn't exist yet, create it from the current branch
		cmd = exec.Command("git", "symbolic-ref", "--short", "HEAD")
		cmd.Dir = repoRoot
		currentBranchBytes, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w", err)
		}
		currentBranch := strings.TrimSpace(string(currentBranchBytes))

		// Create the branch from the current branch
		cmd = exec.Command("git", "branch", "-f", branchName, currentBranch)
		cmd.Dir = repoRoot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branchName, err)
		}
	}

	// Create the worktree
	cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
	cmd.Dir = repoRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree at %s: %w", worktreePath, err)
	}

	return nil
}

// ListAgents returns a list of all agents
func (m *AgentManager) ListAgents() []*api.Agent {
	m.agentsMutex.RLock()
	defer m.agentsMutex.RUnlock()

	agents := make([]*api.Agent, 0, len(m.agents))
	for _, actor := range m.agents {
		agents = append(agents, actor.Agent)
	}
	return agents
}

// RemoveAgent removes an agent from the manager
func (m *AgentManager) RemoveAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Stop the agent if it's running
	if actor.CurrentLifecycleState() == LifecycleRunning {
		if err := m.StopAgent(ctx, agentID); err != nil {
			return fmt.Errorf("failed to stop agent before removal: %w", err)
		}
	}

	// Remove from the map
	delete(m.agents, agentID)
	delete(m.processes, agentID)

	// TODO: Cleanup worktree directory if needed

	return nil
}

// GetAgent returns an agent by ID
func (m *AgentManager) GetAgent(agentID muid.MUID) (*AgentActor, bool) {
	m.agentsMutex.RLock()
	defer m.agentsMutex.RUnlock()

	actor, exists := m.agents[agentID]
	return actor, exists
}

// SpawnAgent starts a new agent process in its worktree
func (m *AgentManager) SpawnAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Get the agent slug to determine the worktree path
	slug := ids.NormalizeAgentSlug(actor.Agent.Name)
	worktreePath := m.resolver.WorktreeDir(slug)

	// Create the command to run the agent
	// For now, we'll use a placeholder command - in reality, this would run the adapter
	cmd := exec.CommandContext(ctx, "sleep", "3600") // Placeholder command
	cmd.Dir = worktreePath

	// Start the command with a PTY
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start agent PTY: %w", err)
	}

	// Create the process record
	process := &AgentProcess{
		ID:      agentID,
		Cmd:     cmd,
		PTY:     ptmx,
		WorkDir: worktreePath,
	}

	// Store the process
	m.processes[agentID] = process

	// Transition the agent to Starting state
	if err := actor.Start(ctx); err != nil {
		// Clean up if we can't transition the state
		ptmx.Close()
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill agent process: %w", err)
		}
		delete(m.processes, agentID)
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}

// StartAgent transitions an agent to the Running state
func (m *AgentManager) StartAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Transition to Running state
	if err := actor.Ready(ctx); err != nil {
		return fmt.Errorf("failed to mark agent as ready: %w", err)
	}

	return nil
}

// AttachAgent connects to an existing agent process
func (m *AgentManager) AttachAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Check if the agent is already running
	if actor.CurrentLifecycleState() != LifecycleRunning {
		return fmt.Errorf("agent is not in running state: %s", actor.CurrentLifecycleState())
	}

	// For now, just return - attaching would involve connecting to the existing PTY
	// In a real implementation, this would establish a connection to the running process
	return nil
}

// StopAgent gracefully stops an agent process
func (m *AgentManager) StopAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Get the process
	process, exists := m.processes[agentID]
	if !exists {
		return fmt.Errorf("agent process not found: %s", agentID.String())
	}

	// Transition to Terminated state
	if err := actor.Terminate(ctx); err != nil {
		return fmt.Errorf("failed to terminate agent: %w", err)
	}

	// Close the PTY
	if process.PTY != nil {
		process.PTY.Close()
	}

	// Kill the process
	if process.Cmd != nil && process.Cmd.Process != nil {
		if err := process.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill agent process: %w", err)
		}
	}

	// Remove the process from the map
	delete(m.processes, agentID)

	return nil
}

// KillAgent forcefully kills an agent process
func (m *AgentManager) KillAgent(ctx context.Context, agentID muid.MUID) error {
	m.agentsMutex.Lock()
	defer m.agentsMutex.Unlock()

	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Get the process
	process, exists := m.processes[agentID]
	if !exists {
		return fmt.Errorf("agent process not found: %s", agentID.String())
	}

	// Transition to Errored state
	if err := actor.Error(ctx, fmt.Errorf("agent killed")); err != nil {
		return fmt.Errorf("failed to mark agent as errored: %w", err)
	}

	// Close the PTY
	if process.PTY != nil {
		process.PTY.Close()
	}

	// Kill the process
	if process.Cmd != nil && process.Cmd.Process != nil {
		if err := process.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill agent process: %w", err)
		}
	}

	// Remove the process from the map
	delete(m.processes, agentID)

	return nil
}

// RestartAgent stops and then starts an agent
func (m *AgentManager) RestartAgent(ctx context.Context, agentID muid.MUID) error {
	// First, stop the agent
	if err := m.StopAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to stop agent for restart: %w", err)
	}

	// Then, spawn the agent again
	if err := m.SpawnAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to spawn agent after stop: %w", err)
	}

	// Mark as ready
	if err := m.StartAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to start agent after spawn: %w", err)
	}

	return nil
}

// SetDefaultMergeStrategy sets the default merge strategy for the manager
func (m *AgentManager) SetDefaultMergeStrategy(strategy git.MergeStrategy) {
	// In a real implementation, this would store the strategy in the config
	// For now, we'll just acknowledge the setting
	_ = strategy
}

// GetDefaultMergeStrategy returns the default merge strategy
func (m *AgentManager) GetDefaultMergeStrategy() git.MergeStrategy {
	// In a real implementation, this would read from the config
	// For now, we'll return merge-commit as the default
	return git.MergeCommit
}

// MergeAgentChanges performs a git merge of the agent's worktree changes to the target branch
func (m *AgentManager) MergeAgentChanges(ctx context.Context, agentID muid.MUID, targetBranch string, strategy git.MergeStrategy) error {
	actor, exists := m.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID.String())
	}

	// Get the agent slug to determine the worktree path
	slug := ids.NormalizeAgentSlug(actor.Agent.Name)
	worktreePath := m.resolver.WorktreeDir(slug)

	// Determine the base branch (the agent's dedicated branch)
	baseBranch := m.resolver.AgentBranch(slug)

	// Perform the merge
	opts := git.MergeOptions{
		Strategy:    strategy,
		BaseBranch:  baseBranch,
		TargetBranch: targetBranch,
		DryRun:      false, // In a real implementation, this could be configurable
	}

	return git.PerformMerge(worktreePath, opts)
}

// GetBaseBranchForRepo determines the base branch for a repository
// Following the spec: run `git symbolic-ref --quiet --short HEAD` and use the output
// If that fails, use targetBranch if provided, otherwise return an error
func (m *AgentManager) GetBaseBranchForRepo(repoPath, targetBranch string) (string, error) {
	return git.GetBaseBranch(repoPath, targetBranch)
}