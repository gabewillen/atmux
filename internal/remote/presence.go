package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func (m *HostManager) subscribePresence(ctx context.Context) error {
	if m == nil || m.dispatcher == nil {
		return fmt.Errorf("host manager: dispatcher unavailable")
	}
	_, err := m.dispatcher.Subscribe(ctx, protocol.Subject("events", "presence"), m.handlePresenceEvent)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	return nil
}

func (m *HostManager) handlePresenceEvent(event protocol.Event) {
	if event.Name != agent.EventPresenceChanged {
		return
	}
	var roster []api.RosterEntry
	if err := decodeEventPayload(event.Payload, &roster); err == nil {
		m.dispatchRosterToAdapters(context.Background(), roster)
		return
	}
	var payload agent.PresenceEvent
	if err := decodeEventPayload(event.Payload, &payload); err != nil {
		return
	}
	m.updateSessionPresence(payload.AgentID, payload.Presence)
}

func (m *HostManager) updateSessionPresence(id api.AgentID, presence string) {
	if m == nil {
		return
	}
	target := strings.ToLower(strings.TrimSpace(presence))
	if target == "" || id.IsZero() {
		return
	}
	m.mu.Lock()
	session := m.agentIndex[id]
	if session != nil {
		session.presence = target
	}
	m.mu.Unlock()
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
