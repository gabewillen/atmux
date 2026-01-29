package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

type actionSendInput struct {
	DataB64 string `json:"data_b64"`
}

type actionUpdatePresence struct {
	Presence string `json:"presence"`
}

type actionEmitEvent struct {
	Event adapter.Event `json:"event"`
}

func (m *HostManager) dispatchRosterToAdapters(ctx context.Context, roster []api.RosterEntry) {
	if m == nil {
		return
	}
	payload, err := json.Marshal(roster)
	if err != nil {
		return
	}
	event := adapter.Event{Type: agent.EventPresenceChanged, Payload: payload}
	m.mu.Lock()
	sessions := make([]*remoteSession, 0, len(m.sessions))
	for _, sess := range m.sessions {
		if sess == nil {
			continue
		}
		sessions = append(sessions, sess)
	}
	m.mu.Unlock()
	for _, sess := range sessions {
		m.dispatchAdapterEvent(ctx, sess, event)
	}
}

func (m *HostManager) dispatchAdapterEvent(ctx context.Context, session *remoteSession, event adapter.Event) {
	if session == nil || session.adapterRef == nil {
		return
	}
	actions, err := session.adapterRef.OnEvent(ctx, event)
	if err != nil {
		if m.logger != nil {
			m.logger.Printf("adapter event failed: agent=%s error=%v", session.slug, err)
		}
		return
	}
	m.executeAdapterActions(ctx, session, actions)
}

func (m *HostManager) executeAdapterActions(ctx context.Context, session *remoteSession, actions []adapter.Action) {
	if session == nil || len(actions) == 0 {
		return
	}
	for _, action := range actions {
		actionType := strings.ToLower(strings.TrimSpace(action.Type))
		switch actionType {
		case "send.input":
			m.handleActionSendInput(session, action.Payload)
		case "update.presence":
			m.handleActionUpdatePresence(ctx, session, action.Payload)
		case "emit.event":
			m.handleActionEmitEvent(ctx, action.Payload)
		}
	}
}

func (m *HostManager) handleActionSendInput(session *remoteSession, payload json.RawMessage) {
	if session == nil || session.runtime == nil {
		return
	}
	var req actionSendInput
	if err := json.Unmarshal(payload, &req); err != nil {
		if m.logger != nil {
			m.logger.Printf("adapter action send.input decode failed: %v", err)
		}
		return
	}
	if req.DataB64 == "" {
		return
	}
	data, err := base64.StdEncoding.DecodeString(req.DataB64)
	if err != nil {
		if m.logger != nil {
			m.logger.Printf("adapter action send.input decode failed: %v", err)
		}
		return
	}
	_ = session.runtime.Send(data)
}

func (m *HostManager) handleActionUpdatePresence(ctx context.Context, session *remoteSession, payload json.RawMessage) {
	if session == nil || session.agentRuntime == nil {
		return
	}
	var req actionUpdatePresence
	if err := json.Unmarshal(payload, &req); err != nil {
		if m.logger != nil {
			m.logger.Printf("adapter action update.presence decode failed: %v", err)
		}
		return
	}
	target := strings.ToLower(strings.TrimSpace(req.Presence))
	current := session.presence
	if strings.TrimSpace(current) == "" {
		current = agent.PresenceOnline
	}
	for _, evt := range presenceTransitionEvents(current, target) {
		_ = session.agentRuntime.EmitPresence(ctx, evt, nil)
	}
}

func (m *HostManager) handleActionEmitEvent(ctx context.Context, payload json.RawMessage) {
	if m == nil || m.dispatcher == nil {
		return
	}
	var req actionEmitEvent
	if err := json.Unmarshal(payload, &req); err != nil {
		if m.logger != nil {
			m.logger.Printf("adapter action emit.event decode failed: %v", err)
		}
		return
	}
	if strings.TrimSpace(req.Event.Type) == "" {
		return
	}
	var eventPayload any
	if len(req.Event.Payload) > 0 {
		eventPayload = json.RawMessage(req.Event.Payload)
	}
	event := protocol.Event{
		Name:       req.Event.Type,
		Payload:    eventPayload,
		OccurredAt: time.Now().UTC(),
	}
	_ = m.dispatcher.Publish(ctx, protocol.Subject("events", "adapter"), event)
}

func presenceTransitionEvents(current string, target string) []string {
	current = strings.ToLower(strings.TrimSpace(current))
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" || current == target {
		return nil
	}
	switch target {
	case agent.PresenceOnline:
		switch current {
		case agent.PresenceBusy:
			return []string{agent.EventTaskCompleted}
		case agent.PresenceOffline:
			return []string{agent.EventRateCleared}
		case agent.PresenceAway:
			return []string{agent.EventActivity}
		}
	case agent.PresenceBusy:
		switch current {
		case agent.PresenceOnline:
			return []string{agent.EventTaskAssigned}
		case agent.PresenceOffline:
			return []string{agent.EventRateCleared, agent.EventTaskAssigned}
		case agent.PresenceAway:
			return []string{agent.EventActivity, agent.EventTaskAssigned}
		}
	case agent.PresenceOffline:
		switch current {
		case agent.PresenceOnline, agent.PresenceBusy:
			return []string{agent.EventRateLimit}
		case agent.PresenceAway:
			return []string{agent.EventActivity, agent.EventRateLimit}
		}
	case agent.PresenceAway:
		if current != agent.PresenceAway {
			return []string{agent.EventStuckDetected}
		}
	}
	return nil
}
