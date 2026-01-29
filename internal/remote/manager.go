// Package remote implements Phase 3 remote agent orchestration.
// This file implements the manager role per spec §5.5.5.
package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"

	"github.com/stateforward/amux/internal/agent"
	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/errors"
	"github.com/stateforward/amux/pkg/api"
)

// Manager implements the manager role for a remote host per spec §5.5.5.
//
// A manager:
// - Connects to the director's hub NATS server via a leaf-mode connection
// - Owns PTY sessions for agents on this host
// - Handles spawn/kill/replay control requests from the director
// - Streams PTY output to the director
// - Maintains per-session replay buffers
type Manager struct {
	cfg        *config.Config
	hostID     string
	peerID     muid.MUID
	nc         *nats.Conn
	subjects   SubjectBuilder
	handshaken bool

	sessionsMu sync.RWMutex
	sessions   map[muid.MUID]*RemoteSession // sessionID -> session
}

// RemoteSession represents a single agent PTY session managed by the manager role.
type RemoteSession struct {
	SessionID muid.MUID
	AgentID   muid.MUID
	AgentSlug string
	RepoPath  string
	Cmd       []string
	Env       map[string]string

	PTY       *os.File
	LocalSess *agent.LocalSession

	// Replay buffer per spec §5.5.7.3
	replayMu     sync.Mutex
	replayBuffer *RingBuffer
	replayActive bool // true while replaying, or when gating live output until replay
}

// SessionExitEvent is published on the host events subject when a PTY session
// ends, allowing the director to observe session exit per spec §5.5.9.
type SessionExitEvent struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Reason    string `json:"reason"`
}

// RingBuffer implements a ring buffer for PTY output replay per spec §5.5.7.3.
type RingBuffer struct {
	data []byte
	cap  int
	head int
	tail int
	full bool
}

// NewRingBuffer creates a new ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data: make([]byte, capacity),
		cap:  capacity,
	}
}

// Write appends data to the ring buffer, overwriting oldest data if at capacity.
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	if rb.cap == 0 {
		return 0, nil
	}

	for _, b := range p {
		rb.data[rb.tail] = b
		rb.tail = (rb.tail + 1) % rb.cap

		if rb.full {
			rb.head = (rb.head + 1) % rb.cap
		}

		if rb.tail == rb.head {
			rb.full = true
		}
	}

	return len(p), nil
}

// Snapshot returns a snapshot of the current buffer contents in oldest-to-newest order.
func (rb *RingBuffer) Snapshot() []byte {
	if !rb.full && rb.head == rb.tail {
		return nil
	}

	var buf bytes.Buffer

	if rb.full {
		// Wraparound: read from head to end, then from start to tail
		buf.Write(rb.data[rb.head:])
		if rb.tail > 0 {
			buf.Write(rb.data[:rb.tail])
		}
	} else {
		// No wraparound: read from head to tail
		buf.Write(rb.data[rb.head:rb.tail])
	}

	return buf.Bytes()
}

// NewManager creates a new manager instance.
//
// The manager connects to the hub using cfg.Remote.NATS.URL and cfg.Remote.NATS.CredsPath.
// The hostID MUST be unique among concurrently connected hosts.
func NewManager(ctx context.Context, cfg *config.Config, hostID string, peerID muid.MUID) (*Manager, error) {
	if cfg == nil {
		return nil, errors.Wrap(errors.ErrInvalidInput, "config must not be nil")
	}
	if hostID == "" {
		return nil, errors.Wrap(errors.ErrInvalidInput, "host_id must not be empty")
	}
	if peerID == 0 {
		return nil, errors.Wrap(errors.ErrInvalidInput, "peer_id must not be zero")
	}

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("amux-manager-%s", hostID)),
		nats.ReconnectWait(time.Duration(cfg.Remote.ReconnectBackoffBase)),
		nats.MaxReconnects(cfg.Remote.ReconnectMaxAttempts),
	}

	// Add credentials if provided
	if cfg.Remote.NATS.CredsPath != "" {
		credsPath := api.ExpandHomeDir(cfg.Remote.NATS.CredsPath)
		opts = append(opts, nats.UserCredentials(credsPath))
	}

	nc, err := nats.Connect(cfg.Remote.NATS.URL, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to NATS hub at %s", cfg.Remote.NATS.URL)
	}

	m := &Manager{
		cfg:      cfg,
		hostID:   hostID,
		peerID:   peerID,
		nc:       nc,
		subjects: SubjectBuilder{Prefix: cfg.Remote.NATS.SubjectPrefix},
		sessions: make(map[muid.MUID]*RemoteSession),
	}

	// Subscribe to control requests
	if err := m.subscribeControl(ctx); err != nil {
		nc.Close()
		return nil, err
	}

	// Perform handshake
	if err := m.handshake(ctx); err != nil {
		nc.Close()
		return nil, err
	}

	return m, nil
}

// expandHomeDir expands ~/ to the user's home directory.
func (m *Manager) expandHomeDir(path string) string {
	return api.ExpandHomeDir(path)
}

// handshake performs the initial handshake with the director per spec §5.5.7.3.
func (m *Manager) handshake(ctx context.Context) error {
	req := HandshakePayload{
		Protocol: 1,
		PeerID:   FormatID(m.peerID),
		Role:     "daemon",
		HostID:   m.hostID,
	}

	reqData, err := MarshalControlMessage("handshake", req)
	if err != nil {
		return errors.Wrap(err, "marshal handshake request")
	}

	// Send handshake to P.handshake.<host_id> per spec §5.5.7.3
	subj := m.subjects.Handshake(m.hostID)

	resp, err := m.nc.RequestWithContext(ctx, subj, reqData)
	if err != nil {
		return errors.Wrapf(err, "send handshake to %s", subj)
	}

	var handshakeResp HandshakePayload
	msgType, err := UnmarshalControlMessage(resp.Data, &handshakeResp)
	if err != nil {
		return errors.Wrap(err, "unmarshal handshake response")
	}

	if msgType == "error" {
		var errPayload ErrorPayload
		_ = json.Unmarshal(resp.Data, &errPayload)
		return errors.Wrapf(errors.ErrRemote, "handshake rejected: %s (%s)", errPayload.Message, errPayload.Code)
	}

	if msgType != "handshake" {
		return errors.Wrapf(errors.ErrInvalidInput, "unexpected handshake response type: %s", msgType)
	}

	m.handshaken = true
	return nil
}

// subscribeControl subscribes to control requests on P.ctl.<host_id> per spec §5.5.7.2.
func (m *Manager) subscribeControl(ctx context.Context) error {
	subj := m.subjects.Control(m.hostID)

	_, err := m.nc.Subscribe(subj, func(msg *nats.Msg) {
		m.handleControlRequest(ctx, msg)
	})
	if err != nil {
		return errors.Wrapf(err, "subscribe to control subject %s", subj)
	}

	return nil
}

// handleControlRequest handles incoming control requests from the director.
func (m *Manager) handleControlRequest(ctx context.Context, msg *nats.Msg) {
	var ctrl ControlMessage
	if err := json.Unmarshal(msg.Data, &ctrl); err != nil {
		_ = m.replyError(msg, "unknown", "invalid_request", "malformed JSON")
		return
	}

	// Reject pre-handshake requests per spec §5.5.7.3
	if !m.handshaken && (ctrl.Type == "spawn" || ctrl.Type == "kill" || ctrl.Type == "replay") {
		_ = m.replyError(msg, ctrl.Type, "not_ready", "handshake not complete")
		return
	}

	switch ctrl.Type {
	case "spawn":
		m.handleSpawn(ctx, msg, ctrl.Payload)
	case "kill":
		m.handleKill(ctx, msg, ctrl.Payload)
	case "replay":
		m.handleReplay(ctx, msg, ctrl.Payload)
	case "ping":
		m.handlePing(ctx, msg, ctrl.Payload)
	default:
		_ = m.replyError(msg, ctrl.Type, "unsupported", fmt.Sprintf("unsupported control type: %s", ctrl.Type))
	}
}

// handleSpawn handles a spawn control request per spec §5.5.7.3.
func (m *Manager) handleSpawn(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage) {
	var req SpawnRequestPayload
	if err := json.Unmarshal(payloadRaw, &req); err != nil {
		_ = m.replyError(msg, "spawn", "invalid_request", "invalid spawn payload")
		return
	}

	agentID, err := ParseID(req.AgentID)
	if err != nil {
		_ = m.replyError(msg, "spawn", "invalid_request", fmt.Sprintf("invalid agent_id: %s", req.AgentID))
		return
	}

	// Idempotency check per spec §5.5.7.3: if a session already exists for agent_id, return existing session_id
	m.sessionsMu.Lock()
	for _, sess := range m.sessions {
		if sess.AgentID == agentID {
			// Check for session_conflict
			if sess.AgentSlug != req.AgentSlug || sess.RepoPath != req.RepoPath {
				m.sessionsMu.Unlock()
				_ = m.replyError(msg, "spawn", "session_conflict", "agent_id already has a session with different slug or repo_path")
				return
			}

			// Return existing session
			m.sessionsMu.Unlock()
			resp := SpawnResponsePayload{
				AgentID:   req.AgentID,
				SessionID: FormatID(sess.SessionID),
			}
			respData, _ := MarshalControlMessage("spawn", resp)
			_ = msg.Respond(respData)
			return
		}
	}
	m.sessionsMu.Unlock()

	// Create new session
	sessionID := api.GenerateID()
	repoPath := m.expandHomeDir(req.RepoPath)

	// Create or reuse worktree per spec §5.3.1
	ag := &api.Agent{
		ID:       agentID,
		Adapter:  "",
		Name:     req.AgentSlug,
		RepoRoot: repoPath,
		Worktree: fmt.Sprintf("%s/.amux/worktrees/%s", repoPath, req.AgentSlug),
	}

	// Ensure worktree exists (simplified for Phase 3)
	_ = os.MkdirAll(ag.Worktree, 0o755)

	// Start PTY session
	localSess, err := agent.StartLocalSession(ctx, ag, req.Command, req.Env)
	if err != nil {
		_ = m.replyError(msg, "spawn", "spawn_failed", fmt.Sprintf("failed to start PTY: %v", err))
		return
	}

	sess := &RemoteSession{
		SessionID:    sessionID,
		AgentID:      agentID,
		AgentSlug:    req.AgentSlug,
		RepoPath:     repoPath,
		Cmd:          req.Command,
		Env:          req.Env,
		PTY:          localSess.PTY,
		LocalSess:    localSess,
		replayBuffer: NewRingBuffer(int(m.cfg.Remote.BufferSize)),
	}

	m.sessionsMu.Lock()
	m.sessions[sessionID] = sess
	m.sessionsMu.Unlock()

	// Start PTY output streaming
	go m.streamPTYOutput(ctx, sess)

	// Reply with success
	resp := SpawnResponsePayload{
		AgentID:   req.AgentID,
		SessionID: FormatID(sessionID),
	}
	respData, _ := MarshalControlMessage("spawn", resp)
	_ = msg.Respond(respData)
}

// handleKill handles a kill control request per spec §5.5.7.3.
func (m *Manager) handleKill(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage) {
	var req KillRequestPayload
	if err := json.Unmarshal(payloadRaw, &req); err != nil {
		_ = m.replyError(msg, "kill", "invalid_request", "invalid kill payload")
		return
	}

	sessionID, err := ParseID(req.SessionID)
	if err != nil {
		_ = m.replyError(msg, "kill", "invalid_request", fmt.Sprintf("invalid session_id: %s", req.SessionID))
		return
	}

	m.sessionsMu.Lock()
	sess, found := m.sessions[sessionID]
	if found {
		delete(m.sessions, sessionID)
	}
	m.sessionsMu.Unlock()

	if found && sess.LocalSess != nil {
		_ = sess.LocalSess.Stop()
	}

	resp := KillResponsePayload{
		SessionID: req.SessionID,
		Killed:    found,
	}
	respData, _ := MarshalControlMessage("kill", resp)
	_ = msg.Respond(respData)
}

// handleReplay handles a replay control request per spec §5.5.7.3.
func (m *Manager) handleReplay(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage) {
	var req ReplayRequestPayload
	if err := json.Unmarshal(payloadRaw, &req); err != nil {
		_ = m.replyError(msg, "replay", "invalid_request", "invalid replay payload")
		return
	}

	sessionID, err := ParseID(req.SessionID)
	if err != nil {
		_ = m.replyError(msg, "replay", "invalid_request", fmt.Sprintf("invalid session_id: %s", req.SessionID))
		return
	}

	m.sessionsMu.RLock()
	sess, found := m.sessions[sessionID]
	m.sessionsMu.RUnlock()

	if !found || int(m.cfg.Remote.BufferSize) == 0 {
		resp := ReplayResponsePayload{
			SessionID: req.SessionID,
			Accepted:  false,
		}
		respData, _ := MarshalControlMessage("replay", resp)
		_ = msg.Respond(respData)
		return
	}

	// Take replay snapshot and publish
	sess.replayMu.Lock()
	snapshot := sess.replayBuffer.Snapshot()
	sess.replayActive = true
	sess.replayMu.Unlock()

	// Publish replay to PTY output subject
	if len(snapshot) > 0 {
		outSubj := m.subjects.PTYOut(m.hostID, sessionID)
		_ = m.publishChunked(outSubj, snapshot, 1024*1024) // 1MB chunks
	}

	sess.replayMu.Lock()
	sess.replayActive = false
	sess.replayMu.Unlock()

	resp := ReplayResponsePayload{
		SessionID: req.SessionID,
		Accepted:  true,
	}
	respData, _ := MarshalControlMessage("replay", resp)
	_ = msg.Respond(respData)
}

// handlePing handles a ping control request per spec §5.5.7.3.
func (m *Manager) handlePing(ctx context.Context, msg *nats.Msg, payloadRaw json.RawMessage) {
	var req PingPayload
	if err := json.Unmarshal(payloadRaw, &req); err != nil {
		_ = m.replyError(msg, "ping", "invalid_request", "invalid ping payload")
		return
	}

	resp := PongPayload(req)
	respData, _ := MarshalControlMessage("pong", resp)
	_ = msg.Respond(respData)
}

// replyError sends an error response.
func (m *Manager) replyError(msg *nats.Msg, requestType, code, message string) error {
	errPayload := ErrorPayload{
		RequestType: requestType,
		Code:        code,
		Message:     message,
	}
	errData, _ := MarshalControlMessage("error", errPayload)
	return msg.Respond(errData)
}

// streamPTYOutput streams PTY output to the director per spec §5.5.7.4.
func (m *Manager) streamPTYOutput(ctx context.Context, sess *RemoteSession) {
	outSubj := m.subjects.PTYOut(m.hostID, sess.SessionID)
	buf := make([]byte, 4096)

	for {
		n, err := sess.PTY.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			// Update replay buffer
			sess.replayMu.Lock()
			_, _ = sess.replayBuffer.Write(data)
			replayActive := sess.replayActive
			sess.replayMu.Unlock()

			// Block live output while replay is active per spec §5.5.7.3 and
			// while we are gating output after a publish error until replay
			// has been requested and handled.
			if !replayActive {
				if err := m.publishChunked(outSubj, data, 1024*1024); err != nil {
					// On publish error, enable replayActive gating so that the
					// director can request a replay before live output resumes.
					sess.replayMu.Lock()
					sess.replayActive = true
					sess.replayMu.Unlock()
				}
			}
		}

		if err != nil {
			break
		}
	}
+
+	// When PTY reading stops, emit a session exit event to the host events
+	// subject so the director can observe unexpected exits per spec §5.5.9.
+	out := SessionExitEvent{
+		SessionID: FormatID(sess.SessionID),
+		AgentID:   FormatID(sess.AgentID),
+		Reason:    "pty_closed",
+	}
+
+	if payload, err := json.Marshal(out); err == nil {
+		_ = m.nc.Publish(m.subjects.Events(m.hostID), payload)
+	}
 }

// publishChunked publishes data in chunks not exceeding maxChunkSize per spec §5.5.7.4.
func (m *Manager) publishChunked(subject string, data []byte, maxChunkSize int) error {
	for len(data) > 0 {
		chunkSize := len(data)
		if chunkSize > maxChunkSize {
			chunkSize = maxChunkSize
		}

		chunk := data[:chunkSize]
		if err := m.nc.Publish(subject, chunk); err != nil {
			return err
		}
		data = data[chunkSize:]
	}

	return nil
}

// Close closes the manager and all managed sessions.
func (m *Manager) Close() error {
	m.sessionsMu.Lock()
	defer m.sessionsMu.Unlock()

	for _, sess := range m.sessions {
		if sess.LocalSess != nil {
			_ = sess.LocalSess.Stop()
		}
	}

	if m.nc != nil {
		m.nc.Close()
	}

	return nil
}
