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
	persist    *persister

	// baseBranches tracks the base_branch per repo_root, recorded at the time
	// the first agent for that repository is added (spec §5.7.1).
	baseBranches map[string]string

	// mergeTargetBranch is the configured git.merge.target_branch fallback.
	// Per spec §5.7.1, when git symbolic-ref fails (detached HEAD), base_branch
	// MUST be set to this value. If this is also empty, the add operation MUST fail.
	mergeTargetBranch string

	// monitorUnsub is the unsubscribe function for the monitor event subscription.
	monitorUnsub func()

	// connectionUnsub is the unsubscribe function for the connection event subscription.
	connectionUnsub func()
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

	m := &Manager{
		agents:       make(map[muid.MUID]*Agent),
		slugs:        make(map[string]muid.MUID),
		hsms:         make(map[muid.MUID]*agentHSMs),
		dispatcher:   dispatcher,
		resolver:     resolver,
		worktrees:    worktree.NewManager(),
		persist:      newPersister(resolver),
		baseBranches: make(map[string]string),
	}

	m.setupMonitorSubscription()
	m.setupConnectionSubscription()
	return m
}

// setupMonitorSubscription subscribes to PTY monitor events and dispatches
// the appropriate presence HSM transitions per spec §7.6:
//   - pty.activity  -> ActivityDetected -> Busy
//   - pty.idle      -> PromptDetected   -> Online
//   - pty.stuck     -> StuckDetected    -> Away
func (m *Manager) setupMonitorSubscription() {
	unsub := m.dispatcher.Subscribe(event.Subscription{
		Types: []event.Type{
			event.TypePTYActivity,
			event.TypePTYIdle,
			event.TypePTYStuck,
		},
		Handler: func(ctx context.Context, evt event.Event) error {
			m.handleMonitorEvent(ctx, evt)
			return nil
		},
	})
	m.monitorUnsub = unsub
}

// handleMonitorEvent maps PTY monitor events to presence HSM transitions.
// Per spec §7.6:
//   - TypePTYActivity  -> task.assigned   (Online -> Busy)
//   - TypePTYIdle      -> prompt.detected (Busy -> Online)
//   - TypePTYStuck     -> stuck.detected  (* -> Away)
func (m *Manager) handleMonitorEvent(ctx context.Context, evt event.Event) {
	agentID := evt.Source

	m.mu.RLock()
	hsms, ok := m.hsms[agentID]
	m.mu.RUnlock()

	if !ok || hsms == nil || hsms.presenceInstance == nil {
		return
	}

	switch evt.Type {
	case event.TypePTYActivity:
		hsm.Dispatch(ctx, hsms.presenceInstance, hsm.Event{Name: PresenceEventTaskAssigned})
	case event.TypePTYIdle:
		hsm.Dispatch(ctx, hsms.presenceInstance, hsm.Event{Name: PresenceEventPromptDetected})
	case event.TypePTYStuck:
		hsm.Dispatch(ctx, hsms.presenceInstance, hsm.Event{Name: PresenceEventStuckDetected})
	}
}

// setupConnectionSubscription subscribes to connection events and dispatches
// the appropriate presence HSM transitions per spec §5.5.8 and §6.5:
//   - connection.lost      -> stuck.detected  (* -> Away)
//   - connection.recovered -> activity.detected (Away -> Online)
//
// Remote agents transition to Away when hub connection is lost and return
// to Online when the connection is recovered and replay is complete.
func (m *Manager) setupConnectionSubscription() {
	unsub := m.dispatcher.Subscribe(event.Subscription{
		Types: []event.Type{
			event.TypeConnectionLost,
			event.TypeConnectionRecovered,
		},
		Handler: func(ctx context.Context, evt event.Event) error {
			m.handleConnectionEvent(ctx, evt)
			return nil
		},
	})
	m.connectionUnsub = unsub
}

// handleConnectionEvent maps connection events to presence HSM transitions
// for remote agents per spec §5.5.8 and §6.5.
func (m *Manager) handleConnectionEvent(ctx context.Context, evt event.Event) {
	// Connection events contain the affected session IDs in their data payload.
	// We need to find the agents associated with those sessions.
	data, ok := evt.Data.(map[string]any)
	if !ok {
		return
	}

	sessions, ok := data["sessions"].([]string)
	if !ok {
		return
	}

	// For each affected session, find the agent and transition its presence
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sessionID := range sessions {
		// Find the agent with this session
		// Note: This is a simplified lookup; in production we'd have a session->agent mapping
		for agentID, hsms := range m.hsms {
			if hsms == nil || hsms.presenceInstance == nil {
				continue
			}

			// Check if this agent is remote and affected by the connection event
			agent, ok := m.agents[agentID]
			if !ok || agent.Location.Type != api.LocationSSH {
				continue
			}

			// Apply presence transition based on event type
			switch evt.Type {
			case event.TypeConnectionLost:
				// Hub disconnection -> Away per spec §5.5.8
				hsm.Dispatch(ctx, hsms.presenceInstance, hsm.Event{Name: PresenceEventStuckDetected})
			case event.TypeConnectionRecovered:
				// Reconnection -> Online per spec §5.5.8
				hsm.Dispatch(ctx, hsms.presenceInstance, hsm.Event{Name: PresenceEventActivityDetected})
			}
		}
		_ = sessionID // Use sessionID for proper routing when session mapping is available
	}
}

// LoadPersisted loads agent definitions from disk and registers them.
// This should be called on daemon startup to restore agents that survived a restart.
// Agents are loaded in their persisted state (lifecycle=pending, presence=online).
func (m *Manager) LoadPersisted(ctx context.Context) error {
	agents, err := m.persist.load()
	if err != nil {
		return fmt.Errorf("load persisted agents: %w", err)
	}

	for _, cfg := range agents {
		if _, addErr := m.Add(ctx, cfg); addErr != nil {
			// Log the error but continue loading other agents
			_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeAgentErrored, cfg.ID,
				map[string]any{"error": addErr.Error(), "agent": cfg.Name}))
		}
	}
	return nil
}

// persistAgents saves the current agent definitions to disk.
// Must be called with m.mu held (at least RLock).
func (m *Manager) persistAgents() {
	agents := make([]api.Agent, 0, len(m.agents))
	for _, a := range m.agents {
		a.mu.RLock()
		agents = append(agents, a.Agent)
		a.mu.RUnlock()
	}

	if err := m.persist.save(agents); err != nil {
		// Best effort: log the error but don't fail the operation
		_ = m.dispatcher.Dispatch(context.Background(), event.NewEvent(
			event.TypeAgentErrored, ids.BroadcastID,
			map[string]any{"error": err.Error(), "operation": "persist_agents"},
		))
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

	// Persist agents to disk
	m.persistAgents()

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

	// Persist agents to disk after removal
	m.mu.RLock()
	m.persistAgents()
	m.mu.RUnlock()

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
