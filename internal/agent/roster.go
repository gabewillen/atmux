package agent

import (
	"sort"
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go/muid"
)

// Roster manages the collection of active agents.
type Roster struct {
	mu     sync.RWMutex
	agents map[muid.MUID]*AgentActor
}

// NewRoster creates a new empty Roster.
func NewRoster() *Roster {
	return &Roster{
		agents: make(map[muid.MUID]*AgentActor),
	}
}

// Add adds an agent to the roster.
func (r *Roster) Add(agent *AgentActor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.ID()] = agent
}

// Remove removes an agent from the roster by ID.
func (r *Roster) Remove(id muid.MUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, id)
}

// Get retrieves an agent by ID.
func (r *Roster) Get(id muid.MUID) *AgentActor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[id]
}

// List returns a sorted list of roster entries.
func (r *Roster) List() []api.RosterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]api.RosterEntry, 0, len(r.agents))
	for _, a := range r.agents {
		data := a.Data()
		status := string(data.State)
		if data.State == api.StateRunning {
			status = string(data.Presence)
		}

		entries = append(entries, api.RosterEntry{
			ID:       data.ID.String(),
			Name:     data.Name,
			Status:   status,
			Location: data.RepoRoot, // Using RepoRoot as location for now
		})
	}

	// Sort by Name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}
