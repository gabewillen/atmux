package agent

import (
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
)

// Registry manages the set of known agents.
type Registry struct {
	mu     sync.RWMutex
	agents map[api.AgentID]*Agent
}

// GlobalRegistry is the default instance (singleton pattern for simplicity in this phase).
var GlobalRegistry = NewRegistry()

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[api.AgentID]*Agent),
	}
}

// Register adds an agent to the registry.
func (r *Registry) Register(a *Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[a.ID] = a
}

// Unregister removes an agent from the registry.
func (r *Registry) Unregister(id api.AgentID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, id)
}

// Get retrieves an agent by ID.
func (r *Registry) Get(id api.AgentID) (*Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[id]
	return a, ok
}

// List returns all agents.
func (r *Registry) List() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Agent, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

// GetRoster returns the list of roster entries.
func (r *Registry) GetRoster() []RosterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	roster := make([]RosterEntry, 0, len(r.agents))
	for _, a := range r.agents {
		entry := RosterEntry{
			AgentID:  a.ID,
			Name:     a.Name,
			Adapter:  a.Config.Adapter,
			Presence: a.CurrentPresence,
			RepoRoot: a.RepoRoot,
		}
		roster = append(roster, entry)
	}
	return roster
}
