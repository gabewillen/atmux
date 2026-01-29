// Package agent provides agent orchestration for amux.
//
// This package implements agent lifecycle management, presence tracking,
// and messaging. All operations are agent-agnostic; agent-specific behavior
// is delegated to adapters.
//
// The Add flow validates that the agent's repository exists, computes a
// unique slug, creates an isolated git worktree, and registers the agent
// for lifecycle management.
//
// See spec §5 for agent management requirements.
package agent

import (
	"context"
	"fmt"
	"sync"

	hsm "github.com/stateforward/hsm-go"
	"github.com/stateforward/hsm-go/muid"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Manager manages agents, including worktree isolation and lifecycle tracking.
type Manager struct {
	mu         sync.RWMutex
	agents     map[muid.MUID]*Agent
	slugs      map[string]muid.MUID // slug -> agent ID for collision detection
	hsms       map[muid.MUID]*agentHSMs
	sessions   SessionSpawner
	dispatcher event.Dispatcher
	resolver   *paths.Resolver
	worktrees  *worktree.Manager

	// baseBranches tracks the base_branch per repo_root, recorded at the time
	// the first agent for that repository is added (spec §5.7.1).
	baseBranches map[string]string

	// mergeTargetBranch is the configured git.merge.target_branch fallback.
	// Per spec §5.7.1, when git symbolic-ref fails (detached HEAD), base_branch
	// MUST be set to this value. If this is also empty, the add operation MUST fail.
	mergeTargetBranch string
}

// Agent represents a managed agent instance.
type Agent struct {
	mu sync.RWMutex
	api.Agent

	// Lifecycle state
	lifecycle api.LifecycleState

	// Presence state
	presence api.PresenceState
}

// NewManager creates a new agent manager.
func NewManager(dispatcher event.Dispatcher) *Manager {
	return NewManagerWithResolver(dispatcher, nil)
}

// NewManagerWithResolver creates a new agent manager with a specific path resolver.
func NewManagerWithResolver(dispatcher event.Dispatcher, resolver *paths.Resolver) *Manager {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	if resolver == nil {
		resolver = paths.DefaultResolver
	}

	return &Manager{
		agents:       make(map[muid.MUID]*Agent),
		slugs:        make(map[string]muid.MUID),
		hsms:         make(map[muid.MUID]*agentHSMs),
		dispatcher:   dispatcher,
		resolver:     resolver,
		worktrees:    worktree.NewManager(),
		baseBranches: make(map[string]string),
	}
}

// Add adds a new agent. The agent's Slug is computed from Name, a worktree is
// created for isolation, and the agent is registered for lifecycle management.
//
// For local agents, if RepoRoot is empty, it is resolved from the Location.RepoPath
// or from the current working directory. The repo_root must be a valid git repository.
//
// See spec §5.2 for the agent add flow.
func (m *Manager) Add(ctx context.Context, cfg api.Agent) (*Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if cfg.ID == 0 {
		cfg.ID = ids.NewID()
	}

	// Resolve RepoRoot for local agents if not set
	if cfg.Location.Type == api.LocationLocal && cfg.RepoRoot == "" {
		repoRoot, err := m.resolveLocalRepoRoot(cfg.Location)
		if err != nil {
			return nil, fmt.Errorf("agent add: %w", err)
		}
		cfg.RepoRoot = repoRoot
	}

	// For SSH agents, RepoPath is required (validated elsewhere in remote flow)
	if cfg.Location.Type == api.LocationSSH && cfg.Location.RepoPath == "" {
		return nil, fmt.Errorf("agent add: SSH agent requires location.repo_path")
	}

	// Compute slug if not set
	if cfg.Slug == "" {
		cfg.Slug = ids.UniqueAgentSlug(cfg.Name, func(slug string) bool {
			_, exists := m.slugs[slug]
			return exists
		})
	} else {
		// Validate the provided slug doesn't collide
		if existingID, exists := m.slugs[cfg.Slug]; exists && existingID != cfg.ID {
			return nil, fmt.Errorf("agent add: %w: %q already in use",
				amuxerrors.ErrAgentSlugCollision, cfg.Slug)
		}
	}

	// Create worktree for local agents
	if cfg.Location.Type == api.LocationLocal && cfg.RepoRoot != "" {
		wtPath, err := m.worktrees.Create(cfg.RepoRoot, cfg.Slug)
		if err != nil {
			return nil, fmt.Errorf("agent add: %w", err)
		}
		cfg.Worktree = wtPath

		// Record base_branch for this repo if not yet recorded (spec §5.7.1).
		// Per spec: if git symbolic-ref fails (detached HEAD), base_branch MUST
		// be set to git.merge.target_branch if configured; otherwise the director
		// MUST fail the add operation.
		if _, recorded := m.baseBranches[cfg.RepoRoot]; !recorded {
			baseBranch, err := m.worktrees.BaseBranch(cfg.RepoRoot)
			if err == nil {
				m.baseBranches[cfg.RepoRoot] = baseBranch
			} else if m.mergeTargetBranch != "" {
				m.baseBranches[cfg.RepoRoot] = m.mergeTargetBranch
			} else {
				return nil, fmt.Errorf("agent add: cannot determine base branch (detached HEAD or unborn branch): set git.merge.target_branch in config")
			}
		}
	}

	// Validate the agent
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("agent add: validation failed: %w", err)
	}

	agent := &Agent{
		Agent:     cfg,
		lifecycle: api.LifecyclePending,
		presence:  api.PresenceOnline,
	}

	m.agents[cfg.ID] = agent
	m.slugs[cfg.Slug] = cfg.ID

	// Create and start lifecycle + presence HSMs
	lhsm := NewLifecycleHSM(agent, m.dispatcher)
	lInstance := lhsm.Start(ctx)

	phsm := NewPresenceHSM(agent, m.dispatcher)
	pInstance := phsm.Start(ctx)

	m.hsms[cfg.ID] = &agentHSMs{
		lifecycle:         lhsm,
		presence:          phsm,
		lifecycleInstance: lInstance,
		presenceInstance:  pInstance,
	}

	// Emit agent.added event
	_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeAgentAdded, cfg.ID, cfg))

	return agent, nil
}

// Remove removes an agent, cleaning up its worktree if configured.
// The deleteBranch parameter controls whether the agent's git branch is deleted.
// If the agent has a running session, it is stopped first.
func (m *Manager) Remove(ctx context.Context, id muid.MUID, deleteBranch bool) error {
	m.mu.Lock()
	agent, ok := m.agents[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("agent remove: %w", amuxerrors.ErrAgentNotFound)
	}
	agentCfg := agent.Agent
	agentHSMs := m.hsms[id]
	spawner := m.sessions
	delete(m.agents, id)
	delete(m.slugs, agent.Slug)
	delete(m.hsms, id)
	m.mu.Unlock()

	// Stop running session if any
	if agentHSMs != nil && spawner != nil &&
		agentHSMs.lifecycle.LifecycleState() == api.LifecycleRunning {
		agentHSMs.setStopping(true)
		_ = spawner.StopAgent(ctx, id)
		spawner.RemoveSession(id)
	}

	// Stop HSMs
	if agentHSMs != nil {
		<-hsm.Stop(ctx, agentHSMs.lifecycleInstance)
		<-hsm.Stop(ctx, agentHSMs.presenceInstance)
	}

	// Clean up worktree for local agents
	if agentCfg.Location.Type == api.LocationLocal && agentCfg.RepoRoot != "" {
		if err := m.worktrees.Remove(agentCfg.RepoRoot, agentCfg.Slug, deleteBranch); err != nil {
			// Log but don't fail the remove
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeWorktreeRemoved, id,
				map[string]any{"error": err.Error(), "slug": agentCfg.Slug}))
		} else {
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeWorktreeRemoved, id,
				map[string]any{"slug": agentCfg.Slug}))
		}
	}

	_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeAgentStopped, id, agentCfg))
	return nil
}

// BaseBranch returns the recorded base branch for a repository.
func (m *Manager) BaseBranch(repoRoot string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	branch, ok := m.baseBranches[repoRoot]
	return branch, ok
}

// SetMergeTargetBranch sets the configured git.merge.target_branch fallback.
// Per spec §5.7.1, this value is used as base_branch when the repository is
// in detached HEAD state (git symbolic-ref fails).
func (m *Manager) SetMergeTargetBranch(branch string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mergeTargetBranch = branch
}

// resolveLocalRepoRoot resolves the repo_root for a local agent.
// If location.RepoPath is set, it is validated; otherwise the current
// working directory's repo root is used.
func (m *Manager) resolveLocalRepoRoot(loc api.Location) (string, error) {
	if loc.RepoPath != "" {
		// Validate the specified path is a git repository
		resolved, err := m.resolver.Resolve(loc.RepoPath)
		if err != nil {
			return "", fmt.Errorf("resolve repo path: %w", err)
		}
		repoRoot, err := m.resolver.FindRepoRoot(resolved)
		if err != nil || repoRoot == "" {
			return "", fmt.Errorf("repo path %q: %w", loc.RepoPath, amuxerrors.ErrNotInRepository)
		}
		return repoRoot, nil
	}

	// Use the resolver's repo root (from CWD)
	repoRoot := m.resolver.RepoRoot()
	if repoRoot != "" {
		return repoRoot, nil
	}

	// Try to find repo root from CWD
	wd, err := paths.FindRepoRoot(".")
	if err != nil || wd == "" {
		return "", fmt.Errorf("no git repository found: %w", amuxerrors.ErrNotInRepository)
	}
	return wd, nil
}

// GetBySlug returns an agent by its slug.
func (m *Manager) GetBySlug(slug string) *Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if id, ok := m.slugs[slug]; ok {
		return m.agents[id]
	}
	return nil
}

// SlugExists returns true if an agent with the given slug exists.
func (m *Manager) SlugExists(slug string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.slugs[slug]
	return ok
}

// Get returns an agent by ID.
func (m *Manager) Get(id muid.MUID) *Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agents[id]
}

// List returns all agents.
func (m *Manager) List() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a)
	}
	return agents
}

// Roster returns the roster entries for all agents.
func (m *Manager) Roster() []api.RosterEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]api.RosterEntry, 0, len(m.agents))
	for _, a := range m.agents {
		a.mu.RLock()
		entries = append(entries, api.RosterEntry{
			Agent:     a.Agent,
			Lifecycle: a.lifecycle,
			Presence:  a.presence,
		})
		a.mu.RUnlock()
	}
	return entries
}

// Lifecycle returns the agent's lifecycle state.
func (a *Agent) Lifecycle() api.LifecycleState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lifecycle
}

// SetLifecycle sets the agent's lifecycle state.
func (a *Agent) SetLifecycle(state api.LifecycleState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lifecycle = state
}

// Presence returns the agent's presence state.
func (a *Agent) Presence() api.PresenceState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.presence
}

// SetPresence sets the agent's presence state.
func (a *Agent) SetPresence(state api.PresenceState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.presence = state
}
