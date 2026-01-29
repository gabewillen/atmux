package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func (m *Manager) startPresenceRouting(ctx context.Context) error {
	if m == nil || m.dispatcher == nil {
		return nil
	}
	presenceSub, err := m.dispatcher.Subscribe(ctx, protocol.Subject("events", "presence"), m.handlePresenceEvent)
	if err != nil {
		return fmt.Errorf("presence routing: %w", err)
	}
	subjectPrefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	remoteSub, err := m.dispatcher.SubscribeRaw(ctx, protocol.Subject(subjectPrefix, "events", ">"), m.handleRemoteEvent)
	if err != nil {
		_ = presenceSub.Unsubscribe()
		return fmt.Errorf("presence routing: %w", err)
	}
	m.subs = append(m.subs, presenceSub, remoteSub)
	return nil
}

func (m *Manager) handlePresenceEvent(event protocol.Event) {
	if event.Name != agent.EventPresenceChanged {
		return
	}
	var payload agent.PresenceEvent
	if err := decodeEventPayload(event.Payload, &payload); err != nil {
		return
	}
	m.setPresence(context.Background(), payload.AgentID, payload.Presence, false)
}

func (m *Manager) handleRemoteEvent(msg protocol.Message) {
	hostID, err := remote.ParseEventsSubject(m.cfg.Remote.NATS.SubjectPrefix, msg.Subject)
	if err != nil {
		return
	}
	var event remote.EventMessage
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}
	switch event.Event.Name {
	case "connection.lost":
		m.updateRemotePresence(context.Background(), hostID, agent.PresenceAway, false)
	case "connection.established", "connection.recovered":
		m.updateRemotePresence(context.Background(), hostID, agent.PresenceOnline, true)
	}
}

func (m *Manager) updateRemotePresence(ctx context.Context, hostID api.HostID, presence string, onlyIfAway bool) {
	if m == nil {
		return
	}
	targetPresence := strings.ToLower(strings.TrimSpace(presence))
	if targetPresence == "" {
		return
	}
	var updated []api.AgentID
	m.mu.Lock()
	for id, state := range m.agents {
		if state == nil || !state.remote || state.remoteHost != hostID {
			continue
		}
		current := statePresence(state)
		if onlyIfAway && current != agent.PresenceAway && current != "" {
			continue
		}
		if current == targetPresence {
			continue
		}
		state.presence = targetPresence
		updated = append(updated, id)
	}
	m.mu.Unlock()
	for _, id := range updated {
		m.emitPresenceChanged(ctx, id, targetPresence)
	}
	if len(updated) > 0 {
		m.emitRosterUpdated(ctx)
	}
}

func (m *Manager) setPresence(ctx context.Context, id api.AgentID, presence string, emit bool) bool {
	targetPresence := strings.ToLower(strings.TrimSpace(presence))
	if targetPresence == "" {
		return false
	}
	m.mu.Lock()
	state := m.agents[id]
	if state == nil {
		m.mu.Unlock()
		return false
	}
	current := statePresence(state)
	if current == targetPresence {
		m.mu.Unlock()
		return false
	}
	state.presence = targetPresence
	m.mu.Unlock()
	if emit {
		m.emitPresenceChanged(ctx, id, targetPresence)
	}
	m.emitRosterUpdated(ctx)
	return true
}

func (m *Manager) emitPresenceChanged(ctx context.Context, id api.AgentID, presence string) {
	if m == nil || m.dispatcher == nil {
		return
	}
	event := protocol.Event{
		Name:       agent.EventPresenceChanged,
		Payload:    agent.PresenceEvent{AgentID: id, Presence: presence},
		OccurredAt: time.Now().UTC(),
	}
	_ = m.dispatcher.Publish(ctx, protocol.Subject("events", "presence"), event)
}

func (m *Manager) emitRosterUpdated(ctx context.Context) {
	if m == nil || m.dispatcher == nil {
		return
	}
	entries, err := m.ListAgents()
	if err != nil {
		return
	}
	event := protocol.Event{
		Name:       "roster.updated",
		Payload:    entries,
		OccurredAt: time.Now().UTC(),
	}
	_ = m.dispatcher.Publish(ctx, protocol.Subject("events", "presence"), event)
}

func (m *Manager) rosterEntry(id api.AgentID, state *agentState) api.RosterEntry {
	if state == nil {
		return api.RosterEntry{Kind: api.RosterAgent, RuntimeID: id.RuntimeID, AgentID: &id, Presence: agent.PresenceOffline}
	}
	entry := api.RosterEntry{
		Kind:      api.RosterAgent,
		RuntimeID: id.RuntimeID,
		AgentID:   &id,
		Name:      state.config.Name,
		About:     state.config.About,
		Adapter:   api.AdapterRef(state.config.Adapter),
		RepoRoot:  state.repoRoot,
		Worktree:  state.worktree,
		Slug:      state.slug,
		Presence:  statePresence(state),
		Task:      state.task,
		Location:  locationForState(state),
	}
	return entry
}

func (m *Manager) systemRosterLocked() []api.RosterEntry {
	if m == nil || m.remoteDirector == nil {
		return nil
	}
	entries := make([]api.RosterEntry, 0, 2)
	directorID := m.remoteDirector.PeerID()
	if !directorID.IsZero() {
		entries = append(entries, api.RosterEntry{
			Kind:      api.RosterDirector,
			RuntimeID: directorID.RuntimeID,
			Name:      "director",
			Slug:      "director",
			Presence:  agent.PresenceOnline,
		})
	}
	localHost := m.remoteDirector.HostID()
	managerID := directorID
	if !managerID.IsZero() && localHost != "" {
		entries = append(entries, api.RosterEntry{
			Kind:      api.RosterManager,
			RuntimeID: managerID.RuntimeID,
			Name:      fmt.Sprintf("manager@%s", localHost.String()),
			Slug:      "manager",
			Presence:  agent.PresenceOnline,
		})
	}
	for _, host := range m.remoteDirector.Hosts() {
		if host.HostID == "" || host.PeerID.IsZero() {
			continue
		}
		if host.HostID == localHost {
			continue
		}
		presence := agent.PresenceAway
		if host.Connected && host.Ready {
			presence = agent.PresenceOnline
		}
		entries = append(entries, api.RosterEntry{
			Kind:      api.RosterManager,
			RuntimeID: host.PeerID.RuntimeID,
			Name:      fmt.Sprintf("manager@%s", host.HostID.String()),
			Slug:      fmt.Sprintf("manager@%s", host.HostID.String()),
			Presence:  presence,
		})
	}
	return entries
}

func locationForState(state *agentState) *api.Location {
	if state == nil {
		return nil
	}
	if state.runtime != nil {
		location := state.runtime.Location
		return &location
	}
	if state.config.Location.Type == "" {
		return nil
	}
	locType, err := api.ParseLocationType(state.config.Location.Type)
	if err != nil {
		return nil
	}
	location := api.Location{Type: locType, Host: state.config.Location.Host, RepoPath: state.config.Location.RepoPath}
	return &location
}

func sortRoster(entries []api.RosterEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		leftOrder := rosterOrder(left.Kind)
		rightOrder := rosterOrder(right.Kind)
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		leftName := strings.ToLower(strings.TrimSpace(left.Name))
		rightName := strings.ToLower(strings.TrimSpace(right.Name))
		if leftName != rightName {
			return leftName < rightName
		}
		return left.RuntimeID.String() < right.RuntimeID.String()
	})
}

func rosterOrder(kind api.RosterKind) int {
	switch kind {
	case api.RosterDirector:
		return 0
	case api.RosterManager:
		return 1
	case api.RosterAgent:
		return 2
	default:
		return 3
	}
}

func decodeEventPayload(payload any, dest any) error {
	if payload == nil {
		return fmt.Errorf("decode payload: empty")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}
