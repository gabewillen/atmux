package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

type listenSubscription struct {
	subject string
	sub     protocol.Subscription
}

func (m *HostManager) configureListen(ctx context.Context, session *remoteSession, targets []string) {
	if m == nil || session == nil {
		return
	}
	subjects := m.resolveListenSubjects(targets)
	m.updateListenTargets(ctx, session.agentID, subjects)
}

func (m *HostManager) clearListen(session *remoteSession) {
	if m == nil || session == nil {
		return
	}
	m.updateListenTargets(context.Background(), session.agentID, nil)
}

func (m *HostManager) resolveListenSubjects(targets []string) []string {
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

func (m *HostManager) listenSubjectForTarget(target string) (string, bool) {
	trimmed := strings.TrimSpace(target)
	if trimmed == "" {
		return "", false
	}
	prefix := SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
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
		return BroadcastCommSubject(prefix), true
	case "director":
		return DirectorCommSubject(prefix), true
	case "manager":
		if m.hostID == "" {
			return "", false
		}
		return ManagerCommSubject(prefix, m.hostID), true
	}
	if strings.HasPrefix(lower, "manager@") {
		rawHost := strings.TrimPrefix(lower, "manager@")
		hostID, err := api.ParseHostID(rawHost)
		if err != nil {
			return "", false
		}
		return ManagerCommSubject(prefix, hostID), true
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, sess := range m.agentIndex {
		if sess == nil {
			continue
		}
		if !strings.EqualFold(sess.slug, trimmed) {
			continue
		}
		return AgentCommSubject(prefix, m.hostID, id), true
	}
	return "", false
}

func (m *HostManager) updateListenTargets(ctx context.Context, id api.AgentID, subjects []string) {
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
	session := m.agentIndex[id]
	var oldSubjects []string
	if session != nil {
		oldSubjects = append([]string(nil), session.listenSubjects...)
		session.listenSubjects = subjects
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
			if m.logger != nil {
				m.logger.Printf("listen subscribe failed: subject=%s error=%v", subject, err)
			}
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

func (m *HostManager) shouldSubscribeListenSubject(subject string) bool {
	if m == nil {
		return false
	}
	prefix := SubjectPrefix(m.cfg.Remote.NATS.SubjectPrefix)
	if subject == ManagerCommSubject(prefix, m.hostID) {
		return false
	}
	if subject == BroadcastCommSubject(prefix) {
		return false
	}
	agentPrefix := protocol.Subject(prefix, "comm", "agent", m.hostID.String()) + "."
	return !strings.HasPrefix(subject, agentPrefix)
}

func (m *HostManager) mirrorListenedMessage(subject string, payload api.AgentMessage) {
	if m == nil || subject == "" {
		return
	}
	m.mu.Lock()
	listeners := m.listenTargets[subject]
	if len(listeners) == 0 {
		m.mu.Unlock()
		return
	}
	sessions := make([]*remoteSession, 0, len(listeners))
	for id := range listeners {
		if session := m.agentIndex[id]; session != nil {
			sessions = append(sessions, session)
		}
	}
	m.mu.Unlock()
	for _, session := range sessions {
		m.mirrorMessageToSession(subject, payload, session)
	}
}

func (m *HostManager) mirrorMessageToSession(subject string, payload api.AgentMessage, session *remoteSession) {
	if session == nil || session.runtime == nil {
		return
	}
	text := fmt.Sprintf("[%s] %s\n", subject, payload.Content)
	formatted := text
	if session.formatter != nil {
		if out, err := session.formatter.Format(context.Background(), text); err == nil && out != "" {
			formatted = out
		}
	}
	_ = session.runtime.Send([]byte(formatted))
}
