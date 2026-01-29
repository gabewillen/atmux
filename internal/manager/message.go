package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func (m *Manager) startMessageRouting(ctx context.Context) error {
	if m == nil || m.dispatcher == nil {
		return nil
	}
	outboundSub, err := m.dispatcher.Subscribe(ctx, protocol.Subject("events", "message"), m.handleOutboundMessage)
	if err != nil {
		return fmt.Errorf("message routing: %w", err)
	}
	hostID := m.localHostID()
	if hostID == "" {
		_ = outboundSub.Unsubscribe()
		return fmt.Errorf("message routing: %w", ErrAgentInvalid)
	}
	prefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	managerSub, err := m.dispatcher.SubscribeRaw(ctx, remote.ManagerCommSubject(prefix, hostID), m.handleCommMessage)
	if err != nil {
		_ = outboundSub.Unsubscribe()
		return fmt.Errorf("message routing: %w", err)
	}
	agentSub, err := m.dispatcher.SubscribeRaw(ctx, protocol.Subject(prefix, "comm", "agent", hostID.String(), ">"), m.handleCommMessage)
	if err != nil {
		_ = outboundSub.Unsubscribe()
		_ = managerSub.Unsubscribe()
		return fmt.Errorf("message routing: %w", err)
	}
	broadcastSub, err := m.dispatcher.SubscribeRaw(ctx, remote.BroadcastCommSubject(prefix), m.handleCommMessage)
	if err != nil {
		_ = outboundSub.Unsubscribe()
		_ = managerSub.Unsubscribe()
		_ = agentSub.Unsubscribe()
		return fmt.Errorf("message routing: %w", err)
	}
	m.subs = append(m.subs, outboundSub, managerSub, agentSub, broadcastSub)
	return nil
}

func (m *Manager) handleOutboundMessage(event protocol.Event) {
	if event.Name != "message.outbound" {
		return
	}
	var payload api.OutboundMessage
	if err := decodeEventPayload(event.Payload, &payload); err != nil {
		return
	}
	m.routeOutboundMessage(context.Background(), payload)
}

func (m *Manager) routeOutboundMessage(ctx context.Context, payload api.OutboundMessage) {
	if strings.TrimSpace(payload.ToSlug) == "" || strings.TrimSpace(payload.Content) == "" {
		return
	}
	senderID, state, ok := m.resolveSender(payload)
	if !ok || state == nil || state.remote {
		return
	}
	msg, ok := m.buildAgentMessage(senderID, payload)
	if !ok {
		return
	}
	m.publishAgentMessage(senderID, msg)
}

func (m *Manager) resolveSender(payload api.OutboundMessage) (api.AgentID, *agentState, bool) {
	if payload.AgentID != nil && !payload.AgentID.IsZero() {
		m.mu.Lock()
		state := m.agents[*payload.AgentID]
		m.mu.Unlock()
		if state == nil {
			return api.AgentID{}, nil, false
		}
		return *payload.AgentID, state, true
	}
	if strings.TrimSpace(payload.From) == "" {
		return api.AgentID{}, nil, false
	}
	runtimeID, err := api.ParseRuntimeID(payload.From)
	if err != nil {
		return api.AgentID{}, nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, state := range m.agents {
		if state == nil {
			continue
		}
		if id.Value() == runtimeID.Value() {
			return id, state, true
		}
	}
	return api.AgentID{}, nil, false
}

func (m *Manager) buildAgentMessage(sender api.AgentID, payload api.OutboundMessage) (api.AgentMessage, bool) {
	msg := api.AgentMessage{
		ToSlug:  payload.ToSlug,
		Content: payload.Content,
	}
	if payload.ID != "" && payload.ID != "0" {
		id, err := api.ParseRuntimeID(payload.ID)
		if err != nil {
			return api.AgentMessage{}, false
		}
		msg.ID = id
	} else {
		msg.ID = api.NewRuntimeID()
	}
	if payload.From != "" {
		from, err := api.ParseRuntimeID(payload.From)
		if err != nil {
			return api.AgentMessage{}, false
		}
		msg.From = from
	} else {
		msg.From = sender.RuntimeID
	}
	if payload.Timestamp != "" {
		if ts, err := time.Parse(time.RFC3339Nano, payload.Timestamp); err == nil {
			msg.Timestamp = ts.UTC()
		}
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	if payload.To != "" {
		target, err := api.ParseTargetID(payload.To)
		if err == nil {
			msg.To = target
			return msg, true
		}
	}
	target, ok := m.resolveToID(payload.ToSlug)
	if !ok {
		return api.AgentMessage{}, false
	}
	msg.To = target
	return msg, true
}

func (m *Manager) resolveToID(slug string) (api.TargetID, bool) {
	target := strings.ToLower(strings.TrimSpace(slug))
	switch target {
	case "all", "broadcast", "*":
		return api.TargetID{}, true
	case "director":
		peer := m.directorPeerID()
		if peer.IsZero() {
			return api.TargetID{}, false
		}
		return api.TargetIDFromRuntime(peer.RuntimeID), true
	case "manager":
		peer, ok := m.peerForHost(m.localHostID())
		if !ok || peer.IsZero() {
			return api.TargetID{}, false
		}
		return api.TargetIDFromRuntime(peer.RuntimeID), true
	}
	if strings.HasPrefix(target, "manager@") {
		rawHost := strings.TrimPrefix(target, "manager@")
		hostID, err := api.ParseHostID(rawHost)
		if err != nil {
			return api.TargetID{}, false
		}
		peer, ok := m.peerForHost(hostID)
		if !ok || peer.IsZero() {
			return api.TargetID{}, false
		}
		return api.TargetIDFromRuntime(peer.RuntimeID), true
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, state := range m.agents {
		if state == nil {
			continue
		}
		if strings.EqualFold(state.slug, slug) {
			return api.TargetIDFromRuntime(id.RuntimeID), true
		}
	}
	return api.TargetID{}, false
}

func (m *Manager) publishAgentMessage(sender api.AgentID, msg api.AgentMessage) {
	if m == nil || m.dispatcher == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	prefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	hostID := m.localHostID()
	senderSubject := remote.AgentCommSubject(prefix, hostID, sender)
	if msg.To.IsBroadcast() {
		m.publishComm(senderSubject, data)
		m.publishComm(remote.BroadcastCommSubject(prefix), data)
		return
	}
	m.publishComm(senderSubject, data)
	if msg.To.Value() == sender.Value() {
		return
	}
	recipient := m.commSubjectForTarget(msg.To)
	if recipient != "" {
		m.publishComm(recipient, data)
	}
}

func (m *Manager) commSubjectForTarget(target api.TargetID) string {
	prefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	localHost := m.localHostID()
	if localHost == "" {
		return ""
	}
	peer, ok := m.peerForHost(localHost)
	if ok && target.Value() == peer.Value() {
		return remote.ManagerCommSubject(prefix, localHost)
	}
	director := m.directorPeerID()
	if !director.IsZero() && target.Value() == director.Value() {
		return remote.DirectorCommSubject(prefix)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, state := range m.agents {
		if state == nil {
			continue
		}
		if target.Value() != id.Value() {
			continue
		}
		hostID := localHost
		if state.remote && state.remoteHost != "" {
			hostID = state.remoteHost
		}
		return remote.AgentCommSubject(prefix, hostID, id)
	}
	return ""
}

func (m *Manager) publishComm(subject string, payload []byte) {
	if subject == "" || m == nil || m.dispatcher == nil {
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), subject, payload, "")
}

func (m *Manager) handleCommMessage(msg protocol.Message) {
	if len(msg.Data) == 0 {
		return
	}
	var payload api.AgentMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return
	}
	if payload.To.IsBroadcast() {
		m.deliverBroadcast(payload)
		m.emitMessageEvent("message.broadcast", payload)
		return
	}
	if m.deliverToTarget(payload) {
		m.emitMessageEvent("message.inbound", payload)
	}
}

func (m *Manager) deliverBroadcast(payload api.AgentMessage) {
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
		m.deliverMessage(state, payload)
	}
}

func (m *Manager) deliverToTarget(payload api.AgentMessage) bool {
	m.mu.Lock()
	var target *agentState
	for id, state := range m.agents {
		if state == nil || state.remote {
			continue
		}
		if payload.To.Value() == id.Value() {
			target = state
			break
		}
	}
	m.mu.Unlock()
	if target == nil {
		return false
	}
	m.deliverMessage(target, payload)
	return true
}

func (m *Manager) deliverMessage(state *agentState, payload api.AgentMessage) {
	if state == nil || state.session == nil {
		return
	}
	formatter := state.formatter
	if formatter == nil {
		_ = state.session.Send([]byte(payload.Content))
		return
	}
	formatted, err := formatter.Format(context.Background(), payload.Content)
	if err != nil || formatted == "" {
		return
	}
	_ = state.session.Send([]byte(formatted))
}

func (m *Manager) emitMessageEvent(name string, payload api.AgentMessage) {
	if m == nil || m.dispatcher == nil {
		return
	}
	event := protocol.Event{
		Name:       name,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	}
	_ = m.dispatcher.Publish(context.Background(), protocol.Subject("events", "message"), event)
}

func (m *Manager) localHostID() api.HostID {
	if m == nil || m.remoteDirector == nil {
		return ""
	}
	return m.remoteDirector.HostID()
}

func (m *Manager) directorPeerID() api.PeerID {
	if m == nil || m.remoteDirector == nil {
		return api.PeerID{}
	}
	return m.remoteDirector.PeerID()
}

func (m *Manager) peerForHost(hostID api.HostID) (api.PeerID, bool) {
	if m == nil || m.remoteDirector == nil || hostID == "" {
		return api.PeerID{}, false
	}
	if hostID == m.remoteDirector.HostID() {
		peer := m.remoteDirector.PeerID()
		if peer.IsZero() {
			return api.PeerID{}, false
		}
		return peer, true
	}
	if snap, ok := m.remoteDirector.HostSnapshot(hostID); ok && !snap.PeerID.IsZero() {
		return snap.PeerID, true
	}
	return api.PeerID{}, false
}
