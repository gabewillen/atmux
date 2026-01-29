// Package agent provides the agent actor model, presence state machines, and
// roster management helpers.
package agent

import (
	"context"
	"sort"
	"sync"

	"github.com/stateforward/hsm-go/muid"

	"github.com/stateforward/amux/internal/errors"
	"github.com/stateforward/amux/internal/event"
	"github.com/stateforward/amux/pkg/api"
)

// Event type constants for presence and roster updates.
const (
	EventTypePresenceChanged = "presence.changed"
	EventTypeRosterUpdated   = "roster.updated"
)

// PresenceChangedPayload is the payload for presence.changed events.
// Per spec §6.1 and §6.5, presence indicates whether an agent can accept tasks.
// IDs are kept as muid.MUID here; JSON encoding is handled at the event system
// or control-plane layer per spec §9.1.3.1.
type PresenceChangedPayload struct {
	AgentID  muid.MUID
	Presence string
}

// RosterStore maintains the current roster of agents and their presence.
// Per spec §6.2 and §6.3, the roster MUST be updated in real time as presence
// changes occur and MUST be broadcast via presence.changed events.
type RosterStore struct {
	mu         sync.RWMutex
	entries    map[muid.MUID]api.RosterEntry
	dispatcher event.Dispatcher
}

// NewRosterStore creates a new RosterStore backed by the provided dispatcher.
// If dispatcher is nil, a local in-process dispatcher is created.
func NewRosterStore(dispatcher event.Dispatcher) *RosterStore {
	if dispatcher == nil {
		dispatcher = event.NewLocalDispatcher()
	}

	return &RosterStore{
		entries:    make(map[muid.MUID]api.RosterEntry),
		dispatcher: dispatcher,
	}
}

// UpsertAgent inserts or updates an agent in the roster and emits presence and
// roster events. Presence defaults to StateOnline when empty.
func (r *RosterStore) UpsertAgent(ctx context.Context, ag *api.Agent, presence string) error {
	if ag == nil {
		return errors.Wrap(errors.ErrInvalidInput, "agent must not be nil")
	}

	if presence == "" {
		presence = StateOnline
	}

	r.mu.Lock()
	r.entries[ag.ID] = api.RosterEntry{
		AgentID:  ag.ID,
		Name:     ag.Name,
		Adapter:  ag.Adapter,
		Presence: presence,
		RepoRoot: ag.RepoRoot,
	}
	snapshot := r.snapshotLocked()
	r.mu.Unlock()

	// Emit presence.changed for this agent.
	if err := r.dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypePresenceChanged,
		Payload: PresenceChangedPayload{
			AgentID:  ag.ID,
			Presence: presence,
		},
	}); err != nil {
		return errors.Wrap(err, "dispatch presence.changed")
	}

	// Emit roster.updated with the full snapshot.
	if err := r.dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypeRosterUpdated,
		Payload:   snapshot,
	}); err != nil {
		return errors.Wrap(err, "dispatch roster.updated")
	}

	return nil
}

// RemoveAgent removes an agent from the roster and emits a roster.updated event.
func (r *RosterStore) RemoveAgent(ctx context.Context, id muid.MUID) error {
	r.mu.Lock()
	delete(r.entries, id)
	snapshot := r.snapshotLocked()
	r.mu.Unlock()

	if err := r.dispatcher.Dispatch(ctx, event.BasicEvent{
		EventType: EventTypeRosterUpdated,
		Payload:   snapshot,
	}); err != nil {
		return errors.Wrap(err, "dispatch roster.updated")
	}

	return nil
}

// List returns a snapshot of the current roster entries, ordered deterministically
// by name (then by AgentID) for stable listings.
func (r *RosterStore) List() []api.RosterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.snapshotLocked()
}

// Dispatcher returns the underlying event dispatcher used by the store.
// This is primarily exposed for tests and integration wiring.
func (r *RosterStore) Dispatcher() event.Dispatcher {
	return r.dispatcher
}

// snapshotLocked builds a sorted slice of roster entries. Caller must hold
// either a read or write lock.
func (r *RosterStore) snapshotLocked() []api.RosterEntry {
	out := make([]api.RosterEntry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].AgentID < out[j].AgentID
		}
		return out[i].Name < out[j].Name
	})

	return out
}
