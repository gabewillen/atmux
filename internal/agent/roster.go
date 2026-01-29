// Package agent provides agent orchestration: lifecycle, presence, roster, and messaging.
// roster.go implements the roster store and listing per spec §6.2, §6.3.
package agent

import (
	"context"
	"sort"
	"sync"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

// RosterStore maintains the roster of all agents, host managers, and the director (spec §6.2).
// Mutations emit roster.updated via the configured Dispatcher so presence awareness is real-time (§6.3).
type RosterStore struct {
	Dispatcher protocol.Dispatcher
	mu         sync.RWMutex
	entries    map[api.ID]api.RosterEntry
}

// NewRosterStore creates a roster store that emits roster.updated on changes.
func NewRosterStore(d protocol.Dispatcher) *RosterStore {
	return &RosterStore{
		Dispatcher: d,
		entries:   make(map[api.ID]api.RosterEntry),
	}
}

// Add adds or replaces a roster entry and emits roster.updated (spec §6.2).
func (r *RosterStore) Add(ctx context.Context, e api.RosterEntry) {
	if !api.ValidRuntimeID(e.AgentID) {
		return
	}
	r.mu.Lock()
	r.entries[e.AgentID] = e
	r.mu.Unlock()
	r.emitRosterUpdated(ctx)
}

// Remove removes a participant by agent_id and emits roster.updated.
func (r *RosterStore) Remove(ctx context.Context, agentID api.ID) {
	r.mu.Lock()
	delete(r.entries, agentID)
	r.mu.Unlock()
	r.emitRosterUpdated(ctx)
}

// UpdatePresence updates the presence state for the given agent_id and emits roster.updated (spec §6.2).
func (r *RosterStore) UpdatePresence(ctx context.Context, agentID api.ID, presence string) {
	r.mu.Lock()
	if e, ok := r.entries[agentID]; ok {
		e.Presence = presence
		r.entries[agentID] = e
	}
	r.mu.Unlock()
	r.emitRosterUpdated(ctx)
}

// UpdateCurrentTask updates the current task for the given agent_id (optional, for busy agents §6.3).
func (r *RosterStore) UpdateCurrentTask(ctx context.Context, agentID api.ID, task string) {
	r.mu.Lock()
	if e, ok := r.entries[agentID]; ok {
		e.CurrentTask = task
		r.entries[agentID] = e
	}
	r.mu.Unlock()
	r.emitRosterUpdated(ctx)
}

// List returns a copy of the roster, ordered by agent_id (spec §6.2 ordering).
func (r *RosterStore) List() api.Roster {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.entries) == 0 {
		return nil
	}
	// Order by agent_id (deterministic ordering for listing)
	ids := make([]api.ID, 0, len(r.entries))
	for id := range r.entries {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make(api.Roster, len(ids))
	for i, id := range ids {
		out[i] = r.entries[id]
	}
	return out
}

func (r *RosterStore) emitRosterUpdated(ctx context.Context) {
	if r.Dispatcher == nil {
		return
	}
	list := r.List()
	_ = r.Dispatcher.Dispatch(ctx, protocol.Event{
		Type: "roster.updated",
		Data: list,
	})
}
