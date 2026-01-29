// Roster provides the roster management for amux.
//
// The roster contains all participants in a session: agents, host managers,
// and the director. It is updated in real-time as presence changes occur
// and broadcast via roster.updated events.
//
// See spec §6.2 for roster requirements and §6.3 for presence awareness.
package agent

import (
	"context"
	"sync"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Roster manages all participants in a session.
// The roster includes agents, host managers (manager agents), and the director.
//
// See spec §6.2.
type Roster struct {
	mu           sync.RWMutex
	participants map[muid.MUID]*api.Participant
	slugIndex    map[string]muid.MUID // slug -> participant ID for lookups
	dispatcher   event.Dispatcher
	directorID   muid.MUID
}

// NewRoster creates a new roster with the given event dispatcher.
func NewRoster(dispatcher event.Dispatcher) *Roster {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	return &Roster{
		participants: make(map[muid.MUID]*api.Participant),
		slugIndex:    make(map[string]muid.MUID),
		dispatcher:   dispatcher,
	}
}

// SetDirector registers the director in the roster.
// There can only be one director; calling this replaces any existing director.
func (r *Roster) SetDirector(id muid.MUID, name, about string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove old director if exists
	if r.directorID != 0 {
		if old, ok := r.participants[r.directorID]; ok {
			delete(r.slugIndex, old.Slug)
		}
		delete(r.participants, r.directorID)
	}

	p := &api.Participant{
		ID:       id,
		Type:     api.ParticipantDirector,
		Name:     name,
		Slug:     api.DirectorSlug,
		About:    about,
		Presence: api.PresenceOnline,
	}
	r.participants[id] = p
	r.slugIndex[api.DirectorSlug] = id
	r.directorID = id

	r.emitRosterUpdated()
}

// DirectorID returns the director's runtime ID, or 0 if no director is registered.
func (r *Roster) DirectorID() muid.MUID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.directorID
}

// AddManager registers a host manager in the roster.
// The slug is "manager@<hostID>" for remote managers, or "manager" for local.
func (r *Roster) AddManager(id muid.MUID, name, hostID, about string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slug := api.ManagerSlug
	if hostID != "" {
		slug = "manager@" + hostID
	}

	p := &api.Participant{
		ID:       id,
		Type:     api.ParticipantManager,
		Name:     name,
		Slug:     slug,
		About:    about,
		HostID:   hostID,
		Presence: api.PresenceOnline,
	}
	r.participants[id] = p
	r.slugIndex[slug] = id

	r.emitRosterUpdated()
}

// AddAgent registers an agent in the roster.
func (r *Roster) AddAgent(agent *api.Agent, lifecycle api.LifecycleState, presence api.PresenceState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hostID := ""
	if agent.Location.Type == api.LocationSSH {
		hostID = agent.Location.Host
	}

	p := &api.Participant{
		ID:        agent.ID,
		Type:      api.ParticipantAgent,
		Name:      agent.Name,
		Slug:      agent.Slug,
		About:     agent.About,
		HostID:    hostID,
		Presence:  presence,
		Lifecycle: lifecycle,
	}
	r.participants[agent.ID] = p
	r.slugIndex[agent.Slug] = agent.ID

	r.emitRosterUpdated()
}

// RemoveParticipant removes a participant from the roster.
func (r *Roster) RemoveParticipant(id muid.MUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.participants[id]; ok {
		delete(r.slugIndex, p.Slug)
		delete(r.participants, id)
		if id == r.directorID {
			r.directorID = 0
		}
		r.emitRosterUpdated()
	}
}

// UpdatePresence updates the presence state of a participant.
func (r *Roster) UpdatePresence(id muid.MUID, presence api.PresenceState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.participants[id]; ok {
		if p.Presence != presence {
			p.Presence = presence
			r.emitRosterUpdated()
		}
	}
}

// UpdateLifecycle updates the lifecycle state of an agent participant.
func (r *Roster) UpdateLifecycle(id muid.MUID, lifecycle api.LifecycleState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.participants[id]; ok && p.Type == api.ParticipantAgent {
		if p.Lifecycle != lifecycle {
			p.Lifecycle = lifecycle
			r.emitRosterUpdated()
		}
	}
}

// UpdateCurrentTask updates the current task of a busy participant.
// Per spec §6.3, this enables other agents to know what busy agents are working on.
func (r *Roster) UpdateCurrentTask(id muid.MUID, task string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.participants[id]; ok {
		if p.CurrentTask != task {
			p.CurrentTask = task
			r.emitRosterUpdated()
		}
	}
}

// Get returns a participant by ID.
func (r *Roster) Get(id muid.MUID) *api.Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.participants[id]; ok {
		// Return a copy to avoid races
		cp := *p
		return &cp
	}
	return nil
}

// GetBySlug returns a participant by slug.
// Slug lookup is case-insensitive per spec §6.4.1.3.
func (r *Roster) GetBySlug(slug string) *api.Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact match first
	if id, ok := r.slugIndex[slug]; ok {
		if p, ok := r.participants[id]; ok {
			cp := *p
			return &cp
		}
	}

	// Try case-insensitive match
	for s, id := range r.slugIndex {
		if equalFoldASCII(s, slug) {
			if p, ok := r.participants[id]; ok {
				cp := *p
				return &cp
			}
		}
	}

	return nil
}

// List returns all participants in the roster.
func (r *Roster) List() []api.Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]api.Participant, 0, len(r.participants))
	for _, p := range r.participants {
		result = append(result, *p)
	}
	return result
}

// ListAgents returns all agent participants.
func (r *Roster) ListAgents() []api.Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []api.Participant
	for _, p := range r.participants {
		if p.Type == api.ParticipantAgent {
			result = append(result, *p)
		}
	}
	return result
}

// ListManagers returns all manager participants.
func (r *Roster) ListManagers() []api.Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []api.Participant
	for _, p := range r.participants {
		if p.Type == api.ParticipantManager {
			result = append(result, *p)
		}
	}
	return result
}

// emitRosterUpdated dispatches a roster.updated event.
// Must be called with r.mu held.
func (r *Roster) emitRosterUpdated() {
	// Build roster snapshot while holding the lock
	roster := make([]api.Participant, 0, len(r.participants))
	for _, p := range r.participants {
		roster = append(roster, *p)
	}

	// Capture directorID before releasing the lock (via defer in caller)
	directorID := r.directorID
	dispatcher := r.dispatcher

	go func() {
		_ = dispatcher.Dispatch(context.Background(), event.NewEvent(
			event.TypeRosterUpdated,
			directorID,
			map[string]any{
				"roster": roster,
			},
		))
	}()
}

// equalFoldASCII compares two strings case-insensitively (ASCII only).
// This is faster than strings.EqualFold for ASCII-only slugs.
func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
