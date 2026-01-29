package manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

type listenSubscription struct {
	subject string
	sub     protocol.Subscription
}

func (m *Manager) configureListen(ctx context.Context, id api.AgentID, state *agentState) {
	if m == nil || state == nil || state.remote {
		return
	}
	subjects := m.resolveListenSubjects(state.config.ListenChannels)
	m.updateListenTargets(ctx, id, subjects)
}

func (m *Manager) clearListen(id api.AgentID) {
	if m == nil {
		return
	}
	m.updateListenTargets(context.Background(), id, nil)
}

func (m *Manager) resolveListenSubjects(targets []string) []string {
	if len(targets) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(targets))
	subjects := make([]string, 0, len(targets))
	for _, raw := range targets {
		subject, ok := m.listenSubjectForTarget(raw)
		if !ok {
			continue
		}
		if _, exists := seen[subject]; exists {
			continue
		}
		seen[subject] = struct{}{}
		subjects = append(subjects, subject)
	}
	return subjects
}

func (m *Manager) listenSubjectForTarget(target string) (string, bool) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", false
	}
	prefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	if strings.HasPrefix(trimmed, "subject:") {
		subject := strings.TrimSpace(strings.TrimPrefix(trimmed, "subject:"))
		if subject == "" {
			return "", false
		}
		return subject, true
	}
	if strings.HasPrefix(trimmed, prefix+".") {
		return trimmed, true
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "all", "broadcast", "*":
		return remote.BroadcastCommSubject(prefix), true
	case "director":
		return remote.DirectorCommSubject(prefix), true
	case "manager":
		hostID := m.localHostID()
		if hostID == "" {
			return "", false
		}
		return remote.ManagerCommSubject(prefix, hostID), true
	}
	if strings.HasPrefix(lower, "manager@") {
		rawHost := strings.TrimPrefix(lower, "manager@")
		hostID, err := api.ParseHostID(rawHost)
		if err != nil {
			return "", false
		}
		return remote.ManagerCommSubject(prefix, hostID), true
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, state := range m.agents {
		if state == nil {
			continue
		}
		if !strings.EqualFold(state.slug, trimmed) {
			continue
		}
		hostID := m.localHostID()
		if state.remote && state.remoteHost != "" {
			hostID = state.remoteHost
		}
		if hostID == "" {
			return "", false
		}
		return remote.AgentCommSubject(prefix, hostID, id), true
	}
	return "", false
}

func (m *Manager) updateListenTargets(ctx context.Context, id api.AgentID, subjects []string) {
	if m == nil || m.dispatcher == nil {
		return
	}
	newSet := make(map[string]struct{}, len(subjects))
	for _, subject := range subjects {
		newSet[subject] = struct{}{}
	}
	var unsubscribe []protocol.Subscription
	var subscribe []string
	m.mu.Lock()
	state := m.agents[id]
	var oldSubjects []string
	if state != nil {
		oldSubjects = append([]string(nil), state.listenSubjects...)
		state.listenSubjects = subjects
	}
	oldSet := make(map[string]struct{}, len(oldSubjects))
	for _, subject := range oldSubjects {
		oldSet[subject] = struct{}{}
	}
	for subject := range oldSet {
		if _, ok := newSet[subject]; ok {
			continue
		}
		if targets, ok := m.listenTargets[subject]; ok {
			delete(targets, id)
			if len(targets) == 0 {
				delete(m.listenTargets, subject)
				if sub, ok := m.listenSubs[subject]; ok {
					unsubscribe = append(unsubscribe, sub.sub)
					delete(m.listenSubs, subject)
				}
			}
		}
	}
	for subject := range newSet {
		if targets, ok := m.listenTargets[subject]; ok {
			targets[id] = struct{}{}
			continue
		}
		m.listenTargets[subject] = map[api.AgentID]struct{}{id: {}}
		if !m.shouldSubscribeListenSubject(subject) {
			continue
		}
		if _, ok := m.listenSubs[subject]; ok {
			continue
		}
		subscribe = append(subscribe, subject)
	}
	m.mu.Unlock()
	for _, sub := range unsubscribe {
		_ = sub.Unsubscribe()
	}
	for _, subject := range subscribe {
		sub, err := m.dispatcher.SubscribeRaw(ctx, subject, m.handleCommMessage)
		if err != nil {
			m.warnf("listen subscribe failed: subject=%s error=%v", subject, err)
			continue
		}
		m.mu.Lock()
		if _, ok := m.listenTargets[subject]; !ok {
			m.mu.Unlock()
			_ = sub.Unsubscribe()
			continue
		}
		m.listenSubs[subject] = &listenSubscription{subject: subject, sub: sub}
		m.mu.Unlock()
	}
}

func (m *Manager) shouldSubscribeListenSubject(subject string) bool {
	if m == nil {
		return false
	}
	prefix := remote.SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	localHost := m.localHostID()
	if localHost == "" {
		return true
	}
	if subject == remote.ManagerCommSubject(prefix, localHost) {
		return false
	}
	if subject == remote.BroadcastCommSubject(prefix) {
		return false
	}
	agentPrefix := protocol.Subject(prefix, "comm", "agent", localHost.String()) + "."
	return !strings.HasPrefix(subject, agentPrefix)
}

func (m *Manager) mirrorListenedMessage(subject string, payload api.AgentMessage) {
	if m == nil || subject == "" {
		return
	}
	m.mu.Lock()
	listeners := m.listenTargets[subject]
	if len(listeners) == 0 {
		m.mu.Unlock()
		return
	}
	states := make([]*agentState, 0, len(listeners))
	for id := range listeners {
		state := m.agents[id]
		if state == nil || state.remote {
			continue
		}
		states = append(states, state)
	}
	m.mu.Unlock()
	for _, state := range states {
		m.mirrorMessageToState(subject, payload, state)
	}
}

func (m *Manager) mirrorMessageToState(subject string, payload api.AgentMessage, state *agentState) {
	if state == nil || state.session == nil {
		return
	}
	text := fmt.Sprintf("[%s] %s\n", subject, payload.Content)
	formatted := text
	if state.formatter != nil {
		if out, err := state.formatter.Format(context.Background(), text); err == nil && out != "" {
			formatted = out
		}
	}
	_ = state.session.Send([]byte(formatted))
}
