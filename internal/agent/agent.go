// Package agent provides agent orchestration for amux.
//
// This package implements agent lifecycle management, presence tracking,
// and messaging. All operations are agent-agnostic; agent-specific behavior
// is delegated to adapters.
//
// See spec §5 for agent management requirements.
package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Manager manages agents.
type Manager struct {
	mu         sync.RWMutex
	agents     map[muid.MUID]*Agent
	slugs      map[string]muid.MUID // slug -> agent ID for collision detection
	dispatcher event.Dispatcher
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
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}

	return &Manager{
		agents:     make(map[muid.MUID]*Agent),
		slugs:      make(map[string]muid.MUID),
		dispatcher: dispatcher,
	}
}

// Add adds an agent. The agent's Slug will be computed from Name if not already set.
// Returns an error if the agent fails validation.
func (m *Manager) Add(ctx context.Context, cfg api.Agent) (*Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate ID if not set
	if cfg.ID == 0 {
		cfg.ID = ids.NewID()
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
			return nil, fmt.Errorf("agent slug collision: %q already in use", cfg.Slug)
		}
	}

	// Validate the agent
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}

	agent := &Agent{
		Agent:     cfg,
		lifecycle: api.LifecyclePending,
		presence:  api.PresenceOffline,
	}

	m.agents[cfg.ID] = agent
	m.slugs[cfg.Slug] = cfg.ID

	// Emit event
	_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeAgentAdded, cfg.ID, cfg))

	return agent, nil
}

// Remove removes an agent.
func (m *Manager) Remove(ctx context.Context, id muid.MUID) error {
	m.mu.Lock()
	agent, ok := m.agents[id]
	if ok {
		delete(m.agents, id)
		delete(m.slugs, agent.Slug)
	}
	m.mu.Unlock()

	if ok {
		_ = m.dispatcher.Dispatch(ctx, event.NewEvent(event.TypeAgentStopped, id, agent.Agent))
	}

	return nil
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
