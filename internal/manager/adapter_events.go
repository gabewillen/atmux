package manager

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

func (m *Manager) dispatchRosterToAdapters(ctx context.Context, roster []api.RosterEntry) {
	if m == nil {
		return
	}
	payload, err := json.Marshal(roster)
	if err != nil {
		return
	}
	event := adapter.Event{Type: agent.EventPresenceChanged, Payload: payload}
	m.mu.Lock()
	states := make([]*agentState, 0, len(m.agents))
	for _, state := range m.agents {
		if state == nil || state.remote {
			continue
		}
		states = append(states, state)
	}
	m.mu.Unlock()
	for _, state := range states {
		m.dispatchAdapterEvent(ctx, state, event)
	}
}

func (m *Manager) dispatchAdapterEvent(ctx context.Context, state *agentState, event adapter.Event) {
	if state == nil || state.adapter == nil {
		return
	}
	actions, err := state.adapter.OnEvent(ctx, event)
	if err != nil {
		m.warnf("adapter event failed: agent=%s error=%v", state.slug, err)
		return
	}
	m.executeAdapterActions(ctx, state, actions)
}

func (m *Manager) executeAdapterActions(ctx context.Context, state *agentState, actions []adapter.Action) {
	if state == nil || len(actions) == 0 {
		return
	}
	for _, action := range actions {
		actionType := strings.ToLower(strings.TrimSpace(action.Type))
		switch actionType {
		case "send.input":
			m.handleActionSendInput(ctx, state, action.Payload)
		case "update.presence":
			m.handleActionUpdatePresence(ctx, state, action.Payload)
		case "emit.event":
			m.handleActionEmitEvent(ctx, action.Payload)
		}
	}
}

func (m *Manager) handleActionSendInput(ctx context.Context, state *agentState, payload json.RawMessage) {
	if state == nil || state.session == nil {
		return
	}
	var req actionSendInput
	if err := json.Unmarshal(payload, &req); err != nil {
		m.warnf("adapter action send.input decode failed: %v", err)
		return
	}
	if req.DataB64 == "" {
		return
	}
	data, err := base64.StdEncoding.DecodeString(req.DataB64)
	if err != nil {
		m.warnf("adapter action send.input decode failed: %v", err)
		return
	}
	if err := state.session.Send(data); err != nil {
		m.warnf("adapter action send.input failed: %v", err)
	}
}

func (m *Manager) handleActionUpdatePresence(ctx context.Context, state *agentState, payload json.RawMessage) {
	if state == nil || state.runtime == nil {
		return
	}
	var req actionUpdatePresence
	if err := json.Unmarshal(payload, &req); err != nil {
		m.warnf("adapter action update.presence decode failed: %v", err)
		return
	}
	target := strings.ToLower(strings.TrimSpace(req.Presence))
	for _, evt := range presenceTransitionEvents(statePresence(state), target) {
		_ = state.runtime.EmitPresence(ctx, evt, nil)
	}
}

func (m *Manager) handleActionEmitEvent(ctx context.Context, payload json.RawMessage) {
	if m == nil || m.dispatcher == nil {
		return
	}
	var req actionEmitEvent
	if err := json.Unmarshal(payload, &req); err != nil {
		m.warnf("adapter action emit.event decode failed: %v", err)
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
