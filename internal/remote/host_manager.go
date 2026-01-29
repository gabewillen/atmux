package remote

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

type remoteSession struct {
	agentID    api.AgentID
	sessionID  api.SessionID
	slug       string
	repoPath   string
	worktree   string
	runtime    *session.LocalSession
	buffer     *ReplayBuffer
	replayGate bool
	replaying  bool
	pending    [][]byte
	mu         sync.Mutex
}

// HostManager runs sessions and responds to remote control requests.
type HostManager struct {
	cfg           config.Config
	resolver      *paths.Resolver
	dispatcher    protocol.Dispatcher
	subjectPrefix string
	hostID        api.HostID
	peerID        api.PeerID
	bufferSize    int
	outbox        *Outbox
	mu            sync.Mutex
	sessions      map[api.SessionID]*remoteSession
	agentIndex    map[api.AgentID]*remoteSession
	ready         bool
	connected     bool
	everConnected bool
}

// NewHostManager constructs a host manager.
func NewHostManager(cfg config.Config, resolver *paths.Resolver) (*HostManager, error) {
	hostID := strings.TrimSpace(cfg.Remote.Manager.HostID)
	if hostID == "" {
		name, err := os.Hostname()
		if err != nil || name == "" {
			hostID = "manager"
		} else {
			hostID = strings.ToLower(name)
		}
	}
	parsedHostID, err := api.ParseHostID(hostID)
	if err != nil {
		return nil, fmt.Errorf("host manager: %w", err)
	}
	bufferSize := int(cfg.Remote.BufferSize)
	if bufferSize < 0 {
		bufferSize = 0
	}
	return &HostManager{
		cfg:           cfg,
		resolver:      resolver,
		subjectPrefix: SubjectPrefix(cfg.Remote.NATS.SubjectPrefix),
		hostID:        parsedHostID,
		peerID:        api.NewPeerID(),
		bufferSize:    bufferSize,
		outbox:        NewOutbox(bufferSize),
		sessions:      make(map[api.SessionID]*remoteSession),
		agentIndex:    make(map[api.AgentID]*remoteSession),
	}, nil
}

// Start connects to NATS and begins serving control requests.
func (m *HostManager) Start(ctx context.Context) error {
	attempts := 0
	for {
		if err := m.connect(ctx); err == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-m.dispatcher.Closed():
				m.markDisconnected("io_error")
			}
		} else {
			m.markDisconnected("io_error")
		}
		attempts++
		if m.cfg.Remote.ReconnectMaxAttempts > 0 && attempts >= m.cfg.Remote.ReconnectMaxAttempts {
			return fmt.Errorf("host manager: reconnect attempts exceeded")
		}
		backoff := reconnectDelay(m.cfg, attempts)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
	}
}

func (m *HostManager) connect(ctx context.Context) error {
	creds, err := LoadCredential(m.cfg.Remote.NATS.CredsPath)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	dispatcher, err := protocol.NewNATSDispatcher(ctx, hubURL(m.cfg), protocol.NATSOptions{Token: creds.Token})
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	m.mu.Lock()
	m.dispatcher = dispatcher
	m.connected = true
	wasConnected := m.everConnected
	m.everConnected = true
	m.mu.Unlock()
	if err := m.subscribeControl(ctx); err != nil {
		_ = dispatcher.Close(ctx)
		return err
	}
	if err := m.performHandshake(ctx, wasConnected); err != nil {
		_ = dispatcher.Close(ctx)
		return err
	}
	m.flushOutbox()
	return nil
}

func (m *HostManager) performHandshake(ctx context.Context, recovered bool) error {
	payload := HandshakePayload{
		Protocol: 1,
		PeerID:   m.peerID.String(),
		Role:     "daemon",
		HostID:   m.hostID.String(),
	}
	msg, err := EncodePayload("handshake", payload)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	subject := HandshakeSubject(m.subjectPrefix, m.hostID)
	reply, err := m.dispatcher.Request(ctx, subject, data, m.cfg.Remote.RequestTimeout)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	resp, err := DecodeControlMessage(reply.Data)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	if resp.Type == "error" {
		var payload ErrorPayload
		if err := DecodePayload(resp, &payload); err != nil {
			return fmt.Errorf("host manager: %w", err)
		}
		return fmt.Errorf("host manager: %w", ErrNotReady)
	}
	m.mu.Lock()
	m.ready = true
	m.mu.Unlock()
	if recovered {
		m.publishConnectionEvent(ctx, "connection.recovered", ConnectionRecoveredPayload{
			PeerID:    m.peerID.String(),
			Timestamp: NowRFC3339(),
		})
	} else {
		m.publishConnectionEvent(ctx, "connection.established", ConnectionEstablishedPayload{
			PeerID:    m.peerID.String(),
			Timestamp: NowRFC3339(),
		})
	}
	return nil
}

func (m *HostManager) subscribeControl(ctx context.Context) error {
	ctlSubject := ControlSubject(m.subjectPrefix, m.hostID)
	_, err := m.dispatcher.SubscribeRaw(ctx, ctlSubject, m.handleControl)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	ptySubject := protocol.Subject(m.subjectPrefix, "pty", m.hostID.String(), "*", "in")
	_, err = m.dispatcher.SubscribeRaw(ctx, ptySubject, m.handlePTYInput)
	if err != nil {
		return fmt.Errorf("host manager: %w", err)
	}
	return nil
}

func (m *HostManager) handleControl(msg protocol.Message) {
	if msg.Reply == "" {
		return
	}
	control, err := DecodeControlMessage(msg.Data)
	requestType := "unknown"
	if err == nil && control.Type != "" {
		requestType = control.Type
	}
	if !m.isReady() {
		_ = m.replyError(msg.Reply, requestType, "not_ready", "handshake incomplete")
		return
	}
	if err != nil {
		_ = m.replyError(msg.Reply, "unknown", "invalid_request", "invalid control message")
		return
	}
	switch control.Type {
	case "spawn":
		m.handleSpawn(msg.Reply, control)
	case "kill":
		m.handleKill(msg.Reply, control)
	case "replay":
		m.handleReplay(msg.Reply, control)
	case "ping":
		m.handlePing(msg.Reply, control)
	default:
		_ = m.replyError(msg.Reply, "unknown", "unknown", "unknown request")
	}
}

func (m *HostManager) handleSpawn(reply string, control ControlMessage) {
	var req SpawnRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid spawn")
		return
	}
	if len(req.Command) == 0 {
		_ = m.replyError(reply, "spawn", "invalid_request", "missing command")
		return
	}
	agentID, err := api.ParseAgentID(req.AgentID)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid agent id")
		return
	}
	if strings.TrimSpace(req.AgentSlug) == "" {
		_ = m.replyError(reply, "spawn", "invalid_request", "missing agent slug")
		return
	}
	repoRoot := m.expandPath(req.RepoPath)
	if repoRoot == "" {
		_ = m.replyError(reply, "spawn", "invalid_repo", "repo path required")
		return
	}
	if err := ensureRepo(repoRoot); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_repo", err.Error())
		return
	}
	m.mu.Lock()
	if existing, ok := m.agentIndex[agentID]; ok {
		if existing.slug != req.AgentSlug || existing.repoPath != repoRoot {
			m.mu.Unlock()
			_ = m.replyError(reply, "spawn", "session_conflict", "session conflict")
			return
		}
		sessionID := existing.sessionID
		m.mu.Unlock()
		m.replySpawn(reply, agentID, sessionID)
		return
	}
	m.mu.Unlock()
	worktree := paths.WorktreePathForRepo(repoRoot, req.AgentSlug)
	runner := git.NewRunner()
	if _, err := runner.EnsureWorktree(context.Background(), repoRoot, req.AgentSlug); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_repo", "failed to create worktree")
		return
	}
	location := api.Location{Type: api.LocationSSH, Host: m.hostID.String(), RepoPath: repoRoot}
	sessionMeta, err := api.NewSession(agentID, repoRoot, worktree, location)
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "invalid session")
		return
	}
	cmd := session.Command{Argv: req.Command}
	if len(req.Env) > 0 {
		env := make([]string, 0, len(req.Env))
		for key, value := range req.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}
	sess, err := session.NewLocalSession(sessionMeta, nil, cmd, worktree, &adapter.NoopMatcher{}, m.dispatcher, session.Config{DrainTimeout: m.cfg.Shutdown.DrainTimeout})
	if err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to start session")
		return
	}
	if err := sess.Start(context.Background()); err != nil {
		_ = m.replyError(reply, "spawn", "invalid_request", "failed to start session")
		return
	}
	remoteSess := &remoteSession{
		agentID:   agentID,
		sessionID: sessionMeta.ID,
		slug:      req.AgentSlug,
		repoPath:  repoRoot,
		worktree:  worktree,
		runtime:   sess,
		buffer:    NewReplayBuffer(m.bufferSize),
	}
	m.mu.Lock()
	m.sessions[sessionMeta.ID] = remoteSess
	m.agentIndex[agentID] = remoteSess
	m.mu.Unlock()
	sess.AddOutputObserver(func(chunk []byte) {
		m.handleOutput(remoteSess, chunk)
	})
	go m.observeSession(remoteSess)
	m.replySpawn(reply, agentID, sessionMeta.ID)
}

func (m *HostManager) replySpawn(reply string, agentID api.AgentID, sessionID api.SessionID) {
	payload := SpawnResponse{AgentID: agentID.String(), SessionID: sessionID.String()}
	msg, err := EncodePayload("spawn", payload)
	if err != nil {
		_ = m.replyError(reply, "spawn", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "spawn", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handleKill(reply string, control ControlMessage) {
	var req KillRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "kill", "invalid_request", "invalid kill")
		return
	}
	sessionID, err := api.ParseSessionID(req.SessionID)
	if err != nil {
		_ = m.replyError(reply, "kill", "invalid_request", "invalid session id")
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	killed := false
	if session != nil {
		if err := session.runtime.Kill(context.Background()); err == nil {
			killed = true
			m.mu.Lock()
			delete(m.sessions, sessionID)
			delete(m.agentIndex, session.agentID)
			m.mu.Unlock()
		}
	}
	payload := KillResponse{SessionID: sessionID.String(), Killed: killed}
	msg, err := EncodePayload("kill", payload)
	if err != nil {
		_ = m.replyError(reply, "kill", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "kill", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handleReplay(reply string, control ControlMessage) {
	var req ReplayRequest
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "replay", "invalid_request", "invalid replay")
		return
	}
	sessionID, err := api.ParseSessionID(req.SessionID)
	if err != nil {
		_ = m.replyError(reply, "replay", "invalid_request", "invalid session id")
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	accepted := false
	if session != nil && session.buffer.Enabled() {
		accepted = true
		m.replaySession(session)
	} else if session != nil {
		session.mu.Lock()
		session.replayGate = false
		session.mu.Unlock()
	}
	payload := ReplayResponse{SessionID: sessionID.String(), Accepted: accepted}
	msg, err := EncodePayload("replay", payload)
	if err != nil {
		_ = m.replyError(reply, "replay", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "replay", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handlePing(reply string, control ControlMessage) {
	var req PingPayload
	if err := DecodePayload(control, &req); err != nil {
		_ = m.replyError(reply, "ping", "invalid_request", "invalid ping")
		return
	}
	payload := PingPayload{UnixMS: req.UnixMS}
	msg, err := EncodePayload("pong", payload)
	if err != nil {
		_ = m.replyError(reply, "ping", "internal", "failed to encode response")
		return
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		_ = m.replyError(reply, "ping", "internal", "failed to encode response")
		return
	}
	_ = m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) handlePTYInput(msg protocol.Message) {
	hostID, sessionID, direction, err := ParseSessionSubject(m.subjectPrefix, msg.Subject)
	if err != nil {
		return
	}
	if hostID != m.hostID || direction != "in" {
		return
	}
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	if session == nil {
		return
	}
	_ = session.runtime.Send(msg.Data)
}

func (m *HostManager) handleOutput(session *remoteSession, chunk []byte) {
	if session == nil || len(chunk) == 0 {
		return
	}
	session.buffer.Add(chunk)
	m.mu.Lock()
	connected := m.connected
	m.mu.Unlock()
	session.mu.Lock()
	if session.replayGate || !connected {
		session.mu.Unlock()
		return
	}
	if session.replaying {
		session.pending = append(session.pending, append([]byte(nil), chunk...))
		session.mu.Unlock()
		return
	}
	session.mu.Unlock()
	m.publishPTY(session.sessionID, chunk)
}

func (m *HostManager) replaySession(session *remoteSession) {
	snapshot := session.buffer.Snapshot()
	session.mu.Lock()
	session.replaying = true
	session.replayGate = false
	session.mu.Unlock()
	if len(snapshot) > 0 {
		m.publishPTY(session.sessionID, snapshot)
	}
	session.mu.Lock()
	pending := session.pending
	session.pending = nil
	session.replaying = false
	session.mu.Unlock()
	for _, chunk := range pending {
		m.publishPTY(session.sessionID, chunk)
	}
}

func (m *HostManager) publishPTY(sessionID api.SessionID, data []byte) {
	subject := PtyOutSubject(m.subjectPrefix, m.hostID, sessionID)
	chunks := chunkBytes(m.dispatcher.MaxPayload(), data)
	for _, chunk := range chunks {
		_ = m.dispatcher.PublishRaw(context.Background(), subject, chunk, "")
	}
}

func (m *HostManager) publishConnectionEvent(ctx context.Context, name string, payload any) {
	event, err := EncodeEventMessage(name, payload)
	if err != nil {
		return
	}
	data, err := EncodeEventMessageJSON(event)
	if err != nil {
		return
	}
	m.publishEvent(ctx, data)
}

func (m *HostManager) publishHostEvent(ctx context.Context, name string, payload any) {
	m.publishConnectionEvent(ctx, name, payload)
}

func (m *HostManager) publishEvent(ctx context.Context, payload []byte) {
	subject := EventsSubject(m.subjectPrefix, m.hostID)
	m.mu.Lock()
	connected := m.connected
	m.mu.Unlock()
	if !connected {
		m.outbox.Enqueue(subject, payload)
		return
	}
	_ = m.dispatcher.PublishRaw(ctx, subject, payload, "")
}

func (m *HostManager) flushOutbox() {
	entries := m.outbox.Drain()
	for _, entry := range entries {
		_ = m.dispatcher.PublishRaw(context.Background(), entry.subject, entry.payload, "")
	}
}

func (m *HostManager) markDisconnected(reason string) {
	m.mu.Lock()
	m.connected = false
	m.ready = false
	for _, sess := range m.sessions {
		sess.mu.Lock()
		sess.replayGate = true
		sess.mu.Unlock()
	}
	m.mu.Unlock()
	m.publishConnectionEvent(context.Background(), "connection.lost", ConnectionLostPayload{
		PeerID:    m.peerID.String(),
		Timestamp: NowRFC3339(),
		Reason:    reason,
	})
}

func (m *HostManager) observeSession(session *remoteSession) {
	if session == nil || session.runtime == nil {
		return
	}
	done := session.runtime.Done()
	if done == nil {
		return
	}
	err := <-done
	payload := map[string]any{
		"agent_id":   session.agentID.String(),
		"session_id": session.sessionID.String(),
	}
	if err != nil {
		payload["error"] = err.Error()
	}
	m.publishHostEvent(context.Background(), "session.exited", payload)
}

func (m *HostManager) isReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ready
}

func (m *HostManager) replyError(reply, requestType, code, message string) error {
	msg, err := NewErrorMessage(requestType, code, message)
	if err != nil {
		return err
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return err
	}
	return m.dispatcher.PublishRaw(context.Background(), reply, data, "")
}

func (m *HostManager) expandPath(path string) string {
	if m.resolver == nil {
		return path
	}
	return m.resolver.ExpandHome(path)
}

func ensureRepo(repoRoot string) error {
	gitDir := filepath.Join(repoRoot, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return fmt.Errorf("repo missing: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	if info.Mode().IsRegular() {
		return nil
	}
	return fmt.Errorf("repo missing: %w", errors.New("invalid git dir"))
}

func reconnectDelay(cfg config.Config, attempt int) time.Duration {
	base := cfg.Remote.ReconnectBackoffBase
	if base <= 0 {
		base = time.Second
	}
	max := cfg.Remote.ReconnectBackoffMax
	if max <= 0 {
		max = 30 * time.Second
	}
	delay := base
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= max {
			return max
		}
	}
	if delay > max {
		return max
	}
	return delay
}
