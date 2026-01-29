// manager.go implements manager-role daemon behaviors: leaf NATS connection, handshake client,
// control request handler (spawn/kill/replay), per-session replay buffer, and PTY I/O subjects per spec §5.5.5, §5.5.7, §5.5.8, §5.5.9.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

// Manager runs the manager role: leaf connection, handshake, and control/PTY handling.
type Manager struct {
	cfg    *config.RemoteConfig
	hostID string
	nc     *nats.Conn
	prefix string
	bufCap int

	mu       sync.RWMutex
	handshakeDone bool
	sessions  map[string]*ManagedSession // session_id -> session
	agentByID map[string]string          // agent_id -> session_id (for spawn idempotency)
}

// ManagedSession holds a remote session's replay buffer and metadata (spec §5.5.9).
type ManagedSession struct {
	ID       string
	AgentID  string
	AgentSlug string
	RepoPath string
	Buffer   *RingBuffer
	liveGate sync.Mutex // guards live output until replay request handled
	replayDone bool
}

// NewManager creates a manager for the given host_id and remote config.
func NewManager(cfg *config.RemoteConfig, hostID string) *Manager {
	if cfg == nil {
		cfg = &config.RemoteConfig{}
	}
	prefix := cfg.NATS.SubjectPrefix
	if prefix == "" {
		prefix = "amux"
	}
	bufCap := parseBufferSize(cfg.BufferSize)
	return &Manager{
		cfg:       cfg,
		hostID:    hostID,
		prefix:    prefix,
		bufCap:    bufCap,
		sessions:  make(map[string]*ManagedSession),
		agentByID: make(map[string]string),
	}
}

// parseBufferSize parses remote.buffer_size per spec §4.2.8 (byte size: integer or NKB/NMB/NGB, binary).
func parseBufferSize(s string) int {
	if s == "" {
		return 10 * 1024 * 1024 // 10MB default
	}
	var n int
	var unit string
	_, _ = fmt.Sscanf(s, "%d%s", &n, &unit)
	switch unit {
	case "KB", "kb":
		return n * 1024
	case "MB", "mb":
		return n * 1024 * 1024
	case "GB", "gb":
		return n * 1024 * 1024 * 1024
	case "B", "b", "":
		return n
	default:
		return n
	}
}

// Connect establishes the NATS connection to the hub (optionally with creds).
func (m *Manager) Connect(ctx context.Context, url, credsPath string) error {
	if url == "" {
		url = m.cfg.NATS.URL
	}
	if url == "" {
		return fmt.Errorf("NATS URL is required")
	}
	opts := []nats.Option{
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(time.Second),
	}
	if credsPath != "" {
		opts = append(opts, nats.UserCredentials(credsPath))
	}
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("connect to hub: %w", err)
	}
	m.nc = nc
	return nil
}

// Close closes the NATS connection and clears session state.
func (m *Manager) Close() {
	if m.nc != nil {
		m.nc.Close()
		m.nc = nil
	}
	m.mu.Lock()
	m.handshakeDone = false
	m.sessions = make(map[string]*ManagedSession)
	m.agentByID = make(map[string]string)
	m.mu.Unlock()
}

// Handshake sends the handshake request to P.handshake.<host_id> and waits for director reply.
// Must be called after Connect and before accepting spawn/kill/replay (spec §5.5.7.3).
func (m *Manager) Handshake(ctx context.Context) error {
	if m.nc == nil {
		return fmt.Errorf("not connected")
	}
	peerID := api.EncodeID(api.NextRuntimeID())
	req := HandshakePayload{
		Protocol: 1,
		PeerID:   peerID,
		Role:     "daemon",
		HostID:   m.hostID,
	}
	payload, _ := json.Marshal(req)
	msg := ControlMessage{Type: ControlTypeHandshake, Payload: payload}
	data, _ := json.Marshal(msg)
	subject := SubjectHandshake(m.prefix, m.hostID)
	reqCtx, cancel := context.WithTimeout(ctx, m.requestTimeout())
	defer cancel()
	reply, err := m.nc.RequestWithContext(reqCtx, subject, data)
	if err != nil {
		return fmt.Errorf("handshake request: %w", err)
	}
	var cm ControlMessage
	if err := json.Unmarshal(reply.Data, &cm); err != nil {
		return fmt.Errorf("handshake response: %w", err)
	}
	if cm.Type == ControlTypeError {
		var ep ErrorPayload
		_ = json.Unmarshal(cm.Payload, &ep)
		return fmt.Errorf("handshake error %s: %s", ep.Code, ep.Message)
	}
	if cm.Type != ControlTypeHandshake {
		return fmt.Errorf("unexpected handshake response type %q", cm.Type)
	}
	m.mu.Lock()
	m.handshakeDone = true
	m.mu.Unlock()
	return nil
}

// IsHandshakeDone returns true after Handshake has succeeded.
func (m *Manager) IsHandshakeDone() bool {
	m.mu.RLock()
	done := m.handshakeDone
	m.mu.RUnlock()
	return done
}

func (m *Manager) requestTimeout() time.Duration {
	t := m.cfg.RequestTimeout
	if t == "" {
		t = "5s"
	}
	dur, _ := time.ParseDuration(t)
	if dur <= 0 {
		dur = 5 * time.Second
	}
	return dur
}

// RunControlHandler subscribes to P.ctl.<host_id> and handles spawn/kill/replay requests.
// For spawn: idempotent by agent_id; session_conflict if repo_path or agent_slug differs (spec §5.5.7.3).
// For replay: publishes replay buffer to P.pty.<host_id>.<session_id>.out then allows live output.
func (m *Manager) RunControlHandler(ctx context.Context) error {
	if m.nc == nil {
		return fmt.Errorf("not connected")
	}
	subject := SubjectCtl(m.prefix, m.hostID)
	_, err := m.nc.Subscribe(subject, func(msg *nats.Msg) {
		var cm ControlMessage
		if err := json.Unmarshal(msg.Data, &cm); err != nil {
			m.respondError(msg, "unknown", "invalid", "invalid JSON")
			return
		}
		if !m.IsHandshakeDone() {
			m.respondError(msg, cm.Type, ErrorCodeNotReady, "handshake not complete")
			return
		}
		switch cm.Type {
		case ControlTypeSpawn:
			m.handleSpawn(msg, cm.Payload)
		case ControlTypeKill:
			m.handleKill(msg, cm.Payload)
		case ControlTypeReplay:
			m.handleReplay(msg, cm.Payload)
		default:
			m.respondError(msg, cm.Type, "unknown", "unknown request type")
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe ctl: %w", err)
	}
	return nil
}

func (m *Manager) respondError(msg *nats.Msg, requestType, code, message string) {
	if requestType == "" {
		requestType = "unknown"
	}
	ep := ErrorPayload{RequestType: requestType, Code: code, Message: message}
	payload, _ := json.Marshal(ep)
	cm := ControlMessage{Type: ControlTypeError, Payload: payload}
	data, _ := json.Marshal(cm)
	_ = msg.Respond(data)
}

func (m *Manager) handleSpawn(msg *nats.Msg, payload json.RawMessage) {
	var req SpawnPayloadRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		m.respondError(msg, ControlTypeSpawn, "invalid", "invalid payload")
		return
	}
	m.mu.Lock()
	if existingSessionID, exists := m.agentByID[req.AgentID]; exists {
		existing := m.sessions[existingSessionID]
		if existing == nil {
			m.mu.Unlock()
			m.respondError(msg, ControlTypeSpawn, "invalid", "session state inconsistent")
			return
		}
		if existing.AgentSlug != req.AgentSlug || existing.RepoPath != req.RepoPath {
			m.mu.Unlock()
			m.respondError(msg, ControlTypeSpawn, ErrorCodeSessionConflict, "agent_id already in use with different repo_path or agent_slug")
			return
		}
		resp := SpawnPayloadResponse{AgentID: req.AgentID, SessionID: existingSessionID}
		m.mu.Unlock()
		m.respondSpawn(msg, resp)
		return
	}
	sessionID := api.EncodeID(api.NextRuntimeID())
	buf := NewRingBuffer(m.bufCap)
	sess := &ManagedSession{
		ID:        sessionID,
		AgentID:   req.AgentID,
		AgentSlug: req.AgentSlug,
		RepoPath:  req.RepoPath,
		Buffer:    buf,
		replayDone: m.bufCap == 0,
	}
	m.sessions[sessionID] = sess
	m.agentByID[req.AgentID] = sessionID
	m.mu.Unlock()
	resp := SpawnPayloadResponse{AgentID: req.AgentID, SessionID: sessionID}
	m.respondSpawn(msg, resp)
}

func (m *Manager) respondSpawn(msg *nats.Msg, resp SpawnPayloadResponse) {
	payload, _ := json.Marshal(resp)
	cm := ControlMessage{Type: ControlTypeSpawn, Payload: payload}
	data, _ := json.Marshal(cm)
	_ = msg.Respond(data)
}

func (m *Manager) handleKill(msg *nats.Msg, payload json.RawMessage) {
	var req KillPayloadRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		m.respondError(msg, ControlTypeKill, "invalid", "invalid payload")
		return
	}
	m.mu.Lock()
	sess, exists := m.sessions[req.SessionID]
	killed := false
	if exists && sess != nil {
		delete(m.agentByID, sess.AgentID)
		delete(m.sessions, req.SessionID)
		killed = true
	}
	m.mu.Unlock()
	resp := KillPayloadResponse{SessionID: req.SessionID, Killed: killed}
	payloadOut, _ := json.Marshal(resp)
	cm := ControlMessage{Type: ControlTypeKill, Payload: payloadOut}
	data, _ := json.Marshal(cm)
	_ = msg.Respond(data)
}

func (m *Manager) handleReplay(msg *nats.Msg, payload json.RawMessage) {
	var req ReplayPayloadRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		m.respondError(msg, ControlTypeReplay, "invalid", "invalid payload")
		return
	}
	m.mu.RLock()
	sess := m.sessions[req.SessionID]
	bufCap := m.bufCap
	m.mu.RUnlock()
	accepted := false
	if sess != nil && bufCap > 0 {
		snapshot := sess.Buffer.Snapshot()
		if len(snapshot) > 0 {
			accepted = true
			subjectOut := SubjectPTYOut(m.prefix, m.hostID, req.SessionID)
			chunkSize := 64 * 1024
			for i := 0; i < len(snapshot); i += chunkSize {
				end := i + chunkSize
				if end > len(snapshot) {
					end = len(snapshot)
				}
				_ = m.nc.Publish(subjectOut, snapshot[i:end])
			}
		}
	}
	if sess != nil {
		sess.liveGate.Lock()
		sess.replayDone = true
		sess.liveGate.Unlock()
	}
	resp := ReplayPayloadResponse{SessionID: req.SessionID, Accepted: accepted}
	payloadOut, _ := json.Marshal(resp)
	cm := ControlMessage{Type: ControlTypeReplay, Payload: payloadOut}
	data, _ := json.Marshal(cm)
	_ = msg.Respond(data)
}

// WritePTYOut appends bytes to the session's replay buffer and publishes to P.pty.<host_id>.<session_id>.out
// only if replay for this session has been handled (spec §5.5.8: no live publish until replay handled).
func (m *Manager) WritePTYOut(sessionID string, p []byte) error {
	m.mu.RLock()
	sess := m.sessions[sessionID]
	m.mu.RUnlock()
	if sess == nil {
		return fmt.Errorf("session %q not found", sessionID)
	}
	sess.Buffer.Write(p)
	sess.liveGate.Lock()
	canLive := sess.replayDone
	sess.liveGate.Unlock()
	if !canLive {
		return nil
	}
	subject := SubjectPTYOut(m.prefix, m.hostID, sessionID)
	maxPayload := 64 * 1024
	for i := 0; i < len(p); i += maxPayload {
		end := i + maxPayload
		if end > len(p) {
			end = len(p)
		}
		if err := m.nc.Publish(subject, p[i:end]); err != nil {
			return fmt.Errorf("publish PTY out: %w", err)
		}
	}
	return nil
}

// SubscribePTYIn subscribes to P.pty.<host_id>.<session_id>.in for a session and calls fn for each message.
func (m *Manager) SubscribePTYIn(ctx context.Context, sessionID string, fn func([]byte)) error {
	if m.nc == nil {
		return fmt.Errorf("not connected")
	}
	subject := SubjectPTYIn(m.prefix, m.hostID, sessionID)
	_, err := m.nc.Subscribe(subject, func(msg *nats.Msg) {
		fn(msg.Data)
	})
	if err != nil {
		return fmt.Errorf("subscribe PTY in: %w", err)
	}
	return nil
}
