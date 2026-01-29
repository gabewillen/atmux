// Package agent implements agent orchestration (lifecycle, presence, messaging)
//
// The agent package provides core functionality for managing agents including:
//   - Lifecycle management (Pending → Starting → Running → Terminated/Errored)
//   - Presence management (Online ↔ Busy ↔ Offline ↔ Away)
//   - Roster maintenance for tracking all agents and their states
//   - Inter-agent messaging capabilities
package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/stateforward/hsm-go/muid"
	"github.com/stateforward/amux/pkg/api"
)

// Roster maintains a list of all agents and their current state
type Roster struct {
	agents map[muid.MUID]*RosterEntry
	mutex  sync.RWMutex
}

// RosterEntry represents an entry in the roster
type RosterEntry struct {
	ID       muid.MUID      `json:"id"`
	Name     string         `json:"name"`
	Adapter  string         `json:"adapter"`
	Presence PresenceState  `json:"presence"`
	RepoRoot string         `json:"repo_root"`
	HostID   muid.MUID      `json:"host_id,omitempty"`
	Task     string         `json:"task,omitempty"` // Current task if agent is busy
}

// NewRoster creates a new roster instance
func NewRoster() *Roster {
	return &Roster{
		agents: make(map[muid.MUID]*RosterEntry),
	}
}

// AddAgent adds an agent to the roster
func (r *Roster) AddAgent(agent *api.Agent, presence PresenceState) {
	var entryToNotify *RosterEntry

	r.mutex.Lock()
	entry := &RosterEntry{
		ID:       agent.ID,
		Name:     agent.Name,
		Adapter:  agent.Adapter,
		Presence: presence,
		RepoRoot: agent.RepoRoot,
		HostID:   agent.HostID,
	}

	r.agents[agent.ID] = entry
	// Make a copy of the entry to notify outside the lock
	copiedEntry := *entry
	entryToNotify = &copiedEntry
	r.mutex.Unlock()

	// Notify outside the lock to avoid potential deadlocks
	if entryToNotify != nil {
		r.presenceSubs.NotifyAll(entryToNotify)
	}
}

// RemoveAgent removes an agent from the roster
func (r *Roster) RemoveAgent(agentID muid.MUID) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.agents, agentID)
}

// UpdatePresence updates the presence state of an agent in the roster
func (r *Roster) UpdatePresence(agentID muid.MUID, presence PresenceState) {
	var entryToNotify *RosterEntry

	r.mutex.Lock()
	if entry, exists := r.agents[agentID]; exists {
		entry.Presence = presence
		// Make a copy of the entry to notify outside the lock
		copiedEntry := *entry
		entryToNotify = &copiedEntry
	}
	r.mutex.Unlock()

	// Notify outside the lock to avoid potential deadlocks
	if entryToNotify != nil {
		r.presenceSubs.NotifyAll(entryToNotify)
	}
}

// UpdateTask updates the current task of an agent in the roster
func (r *Roster) UpdateTask(agentID muid.MUID, task string) {
	var entryToNotify *RosterEntry

	r.mutex.Lock()
	if entry, exists := r.agents[agentID]; exists {
		entry.Task = task
		// Make a copy of the entry to notify outside the lock
		copiedEntry := *entry
		entryToNotify = &copiedEntry
	}
	r.mutex.Unlock()

	// Notify outside the lock to avoid potential deadlocks
	if entryToNotify != nil {
		r.presenceSubs.NotifyAll(entryToNotify)
	}
}

// GetAgent returns the roster entry for an agent
func (r *Roster) GetAgent(agentID muid.MUID) (*RosterEntry, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entry, exists := r.agents[agentID]
	return entry, exists
}

// GetAllAgents returns all agents in the roster
func (r *Roster) GetAllAgents() []*RosterEntry {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agents := make([]*RosterEntry, 0, len(r.agents))
	for _, entry := range r.agents {
		// Make a copy to prevent external modification of the stored entry
		agentCopy := *entry
		agents = append(agents, &agentCopy)
	}
	return agents
}

// GetAgentsByPresence returns agents filtered by presence state
func (r *Roster) GetAgentsByPresence(presence PresenceState) []*RosterEntry {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var agents []*RosterEntry
	for _, entry := range r.agents {
		if entry.Presence == presence {
			// Make a copy to prevent external modification
			agentCopy := *entry
			agents = append(agents, &agentCopy)
		}
	}
	return agents
}

// GetOnlineAgents returns all agents that are currently online
func (r *Roster) GetOnlineAgents() []*RosterEntry {
	return r.GetAgentsByPresence(PresenceOnline)
}

// GetBusyAgents returns all agents that are currently busy
func (r *Roster) GetBusyAgents() []*RosterEntry {
	return r.GetAgentsByPresence(PresenceBusy)
}

// GetOfflineAgents returns all agents that are currently offline
func (r *Roster) GetOfflineAgents() []*RosterEntry {
	return r.GetAgentsByPresence(PresenceOffline)
}

// GetAwayAgents returns all agents that are currently away
func (r *Roster) GetAwayAgents() []*RosterEntry {
	return r.GetAgentsByPresence(PresenceAway)
}

// Size returns the number of agents in the roster
func (r *Roster) Size() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.agents)
}

// PresenceChangeCallback is a function that gets called when an agent's presence changes
type PresenceChangeCallback func(*RosterEntry)

// Subscription represents a subscription to presence changes
type Subscription struct {
	id       string
	callback PresenceChangeCallback
}

// presenceSubscriptions holds all presence change subscriptions
type presenceSubscriptions struct {
	subs map[string]PresenceChangeCallback
	nextID int
	mutex sync.RWMutex
}

// newPresenceSubscriptions creates a new presenceSubscriptions instance
func newPresenceSubscriptions() *presenceSubscriptions {
	return &presenceSubscriptions{
		subs: make(map[string]PresenceChangeCallback),
		nextID: 1,
	}
}

// Subscribe adds a new subscription to presence changes
func (ps *presenceSubscriptions) Subscribe(callback PresenceChangeCallback) string {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	id := fmt.Sprintf("sub_%d", ps.nextID)
	ps.nextID++

	ps.subs[id] = callback
	return id
}

// Unsubscribe removes a subscription
func (ps *presenceSubscriptions) Unsubscribe(id string) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	delete(ps.subs, id)
}

// NotifyAll notifies all subscribers about a presence change
func (ps *presenceSubscriptions) NotifyAll(entry *RosterEntry) {
	ps.mutex.RLock()
	defer ps.mutex.RUnlock()

	for _, callback := range ps.subs {
		// Call each subscriber's callback in a goroutine to avoid blocking
		go callback(entry)
	}
}

// Add a presence subscriptions field to the Roster
type Roster struct {
	agents map[muid.MUID]*RosterEntry
	mutex  sync.RWMutex
	presenceSubs *presenceSubscriptions
}

// NewRoster creates a new roster instance
func NewRoster() *Roster {
	return &Roster{
		agents: make(map[muid.MUID]*RosterEntry),
		presenceSubs: newPresenceSubscriptions(),
	}
}

// SubscribeToPresenceChanges allows components to subscribe to roster/presence changes
func (r *Roster) SubscribeToPresenceChanges(ctx context.Context, handler func(*RosterEntry)) string {
	return r.presenceSubs.Subscribe(handler)
}

// UnsubscribeFromPresenceChanges removes a subscription to presence changes
func (r *Roster) UnsubscribeFromPresenceChanges(subID string) {
	r.presenceSubs.Unsubscribe(subID)
}

// notifyPresenceChange notifies all subscribers about a presence change for an agent
func (r *Roster) notifyPresenceChange(agentID muid.MUID) {
	if entry, exists := r.GetAgent(agentID); exists {
		r.presenceSubs.NotifyAll(entry)
	}
}