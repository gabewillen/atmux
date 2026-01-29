// Package roster implements the roster data model and listing outputs per spec §6.2.
// The roster maintains all agents, host managers, and the director with real-time updates.
package roster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/copilot-claude-sonnet-4/amux/pkg/api"
)

// Entry represents a single entry in the roster per spec §6.2.
type Entry struct {
	// ID is the runtime ID of the participant.
	ID muid.MUID `json:"id"`

	// Type indicates the entry type: "agent", "manager", or "director".
	Type string `json:"type"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Slug is the normalized slug for addressing.
	Slug string `json:"slug"`

	// Presence is the current presence state.
	Presence api.PresenceState `json:"presence"`

	// State is the current lifecycle state (for agents).
	State api.AgentState `json:"state,omitempty"`

	// HostID is the host identifier (for agents and managers).
	HostID string `json:"host_id,omitempty"`

	// About is a description of the participant's role.
	About string `json:"about,omitempty"`

	// CurrentTask describes what the participant is currently working on (if busy).
	CurrentTask string `json:"current_task,omitempty"`

	// LastSeen is when this participant was last active.
	LastSeen time.Time `json:"last_seen"`

	// Metadata contains additional participant-specific data.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Store manages the roster of all participants per spec §6.2.
type Store struct {
	// entries maps runtime ID to roster entry.
	entries map[muid.MUID]*Entry

	// slugIndex maps slug to runtime ID for lookup.
	slugIndex map[string]muid.MUID

	// mu protects concurrent access to the roster.
	mu sync.RWMutex

	// ctx is the store context.
	ctx context.Context

	// cancel cancels the store context.
	cancel context.CancelFunc

	// subscribers holds channels for presence change notifications.
	subscribers []chan<- PresenceChangeEvent

	// subMu protects the subscribers slice.
	subMu sync.RWMutex
}

// PresenceChangeEvent is emitted when a participant's presence changes.
type PresenceChangeEvent struct {
	// ParticipantID is the runtime ID of the participant.
	ParticipantID muid.MUID `json:"participant_id"`

	// OldPresence is the previous presence state.
	OldPresence api.PresenceState `json:"old_presence"`

	// NewPresence is the new presence state.
	NewPresence api.PresenceState `json:"new_presence"`

	// Timestamp is when the change occurred.
	Timestamp time.Time `json:"timestamp"`
}

// NewStore creates a new roster store.
func NewStore() *Store {
	ctx, cancel := context.WithCancel(context.Background())
	return &Store{
		entries:     make(map[muid.MUID]*Entry),
		slugIndex:   make(map[string]muid.MUID),
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make([]chan<- PresenceChangeEvent, 0),
	}
}

// AddAgent adds or updates an agent in the roster per spec requirements.
func (s *Store) AddAgent(agent *api.Agent, hostID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &Entry{
		ID:       agent.ID,
		Type:     "agent",
		Name:     agent.Name,
		Slug:     agent.Slug,
		Presence: agent.Presence,
		State:    agent.State,
		HostID:   hostID,
		LastSeen: time.Now().UTC(),
		Metadata: map[string]interface{}{
			"adapter":   agent.Adapter,
			"repo_root": agent.RepoRoot,
			"created_at": agent.CreatedAt,
		},
	}

	// Check for presence change
	if existing, exists := s.entries[agent.ID]; exists {
		if existing.Presence != entry.Presence {
			// Emit presence change event
			event := PresenceChangeEvent{
				ParticipantID: agent.ID,
				OldPresence:   existing.Presence,
				NewPresence:   entry.Presence,
				Timestamp:     time.Now().UTC(),
			}
			go s.notifySubscribers(event)
		}
	}

	s.entries[agent.ID] = entry
	s.slugIndex[agent.Slug] = agent.ID
	return nil
}

// AddManager adds or updates a host manager in the roster.
func (s *Store) AddManager(id muid.MUID, hostID string, presence api.PresenceState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slug := fmt.Sprintf("manager@%s", hostID)
	entry := &Entry{
		ID:       id,
		Type:     "manager",
		Name:     fmt.Sprintf("Host Manager (%s)", hostID),
		Slug:     slug,
		Presence: presence,
		HostID:   hostID,
		About:    fmt.Sprintf("Host manager for %s", hostID),
		LastSeen: time.Now().UTC(),
	}

	// Check for presence change
	if existing, exists := s.entries[id]; exists {
		if existing.Presence != entry.Presence {
			// Emit presence change event
			event := PresenceChangeEvent{
				ParticipantID: id,
				OldPresence:   existing.Presence,
				NewPresence:   entry.Presence,
				Timestamp:     time.Now().UTC(),
			}
			go s.notifySubscribers(event)
		}
	}

	s.entries[id] = entry
	s.slugIndex[slug] = id
	s.slugIndex["manager"] = id // Local manager alias
	return nil
}

// AddDirector adds or updates the director in the roster.
func (s *Store) AddDirector(id muid.MUID, presence api.PresenceState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &Entry{
		ID:       id,
		Type:     "director",
		Name:     "Director",
		Slug:     "director",
		Presence: presence,
		About:    "amux director - orchestrates all agents and managers",
		LastSeen: time.Now().UTC(),
	}

	// Check for presence change
	if existing, exists := s.entries[id]; exists {
		if existing.Presence != entry.Presence {
			// Emit presence change event
			event := PresenceChangeEvent{
				ParticipantID: id,
				OldPresence:   existing.Presence,
				NewPresence:   entry.Presence,
				Timestamp:     time.Now().UTC(),
			}
			go s.notifySubscribers(event)
		}
	}

	s.entries[id] = entry
	s.slugIndex["director"] = id
	return nil
}

// UpdatePresence updates the presence state of a participant.
func (s *Store) UpdatePresence(id muid.MUID, presence api.PresenceState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return fmt.Errorf("participant %s not found in roster", id)
	}

	oldPresence := entry.Presence
	entry.Presence = presence
	entry.LastSeen = time.Now().UTC()

	// Emit presence change event
	if oldPresence != presence {
		event := PresenceChangeEvent{
			ParticipantID: id,
			OldPresence:   oldPresence,
			NewPresence:   presence,
			Timestamp:     time.Now().UTC(),
		}
		go s.notifySubscribers(event)
	}

	return nil
}

// UpdateTask updates the current task for a participant.
func (s *Store) UpdateTask(id muid.MUID, task string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return fmt.Errorf("participant %s not found in roster", id)
	}

	entry.CurrentTask = task
	entry.LastSeen = time.Now().UTC()
	return nil
}

// Remove removes a participant from the roster.
func (s *Store) Remove(id muid.MUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return nil // Already removed
	}

	// Remove from slug index
	delete(s.slugIndex, entry.Slug)
	if entry.Type == "manager" && entry.HostID != "" {
		delete(s.slugIndex, "manager") // Remove local alias
	}

	// Remove from entries
	delete(s.entries, id)

	return nil
}

// GetByID retrieves a roster entry by runtime ID.
func (s *Store) GetByID(id muid.MUID) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[id]
	if !exists {
		return nil, fmt.Errorf("participant %s not found", id)
	}

	// Return a copy
	return copyEntry(entry), nil
}

// GetBySlug retrieves a roster entry by slug.
func (s *Store) GetBySlug(slug string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.slugIndex[slug]
	if !exists {
		return nil, fmt.Errorf("participant with slug '%s' not found", slug)
	}

	entry := s.entries[id]
	return copyEntry(entry), nil
}

// ListAll returns all roster entries per spec §6.2 requirements.
func (s *Store) ListAll() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		entries = append(entries, copyEntry(entry))
	}

	return entries
}

// ListAgents returns all agent entries.
func (s *Store) ListAgents() []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var agents []*Entry
	for _, entry := range s.entries {
		if entry.Type == "agent" {
			agents = append(agents, copyEntry(entry))
		}
	}

	return agents
}

// ListByPresence returns all entries with the specified presence.
func (s *Store) ListByPresence(presence api.PresenceState) []*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Entry
	for _, entry := range s.entries {
		if entry.Presence == presence {
			filtered = append(filtered, copyEntry(entry))
		}
	}

	return filtered
}

// Subscribe adds a subscriber for presence change events per spec §6.3.
func (s *Store) Subscribe() <-chan PresenceChangeEvent {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	ch := make(chan PresenceChangeEvent, 100) // Buffered to prevent blocking
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// notifySubscribers sends presence change events to all subscribers.
func (s *Store) notifySubscribers(event PresenceChangeEvent) {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// Skip if channel is full to prevent blocking
		}
	}
}

// Close gracefully shuts down the roster store.
func (s *Store) Close() error {
	s.cancel()

	s.subMu.Lock()
	defer s.subMu.Unlock()

	// Close all subscriber channels
	for _, ch := range s.subscribers {
		close(ch)
	}
	s.subscribers = s.subscribers[:0]

	return nil
}

// copyEntry creates a deep copy of a roster entry.
func copyEntry(src *Entry) *Entry {
	if src == nil {
		return nil
	}

	dst := &Entry{
		ID:          src.ID,
		Type:        src.Type,
		Name:        src.Name,
		Slug:        src.Slug,
		Presence:    src.Presence,
		State:       src.State,
		HostID:      src.HostID,
		About:       src.About,
		CurrentTask: src.CurrentTask,
		LastSeen:    src.LastSeen,
	}

	// Copy metadata map
	if src.Metadata != nil {
		dst.Metadata = make(map[string]interface{})
		for k, v := range src.Metadata {
			dst.Metadata[k] = v
		}
	}

	return dst
}