// Package manager implements the manager-role daemon for amux remote agents.
//
// The manager runs on a remote host and manages PTY sessions for agents.
// It connects to the director's hub NATS server via a leaf connection,
// performs a handshake exchange, and handles control requests (spawn/kill/replay).
//
// Key responsibilities:
//   - Own PTYs on the host (one per agent)
//   - Stream PTY output to the director over NATS
//   - Receive PTY input from the director over NATS
//   - Maintain per-session replay buffers
//   - Handle connection recovery with replay-before-live semantics
//
// See spec §5.5.4 and §5.5.5 for manager daemon requirements.
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote/buffer"
	"github.com/agentflare-ai/amux/internal/remote/natsconn"
)

// Manager implements the manager-role daemon on a remote host.
type Manager struct {
	mu     sync.RWMutex
	conn   *natsconn.Conn
	cfg    *config.Config
	prefix string
	hostID string
	peerID string

	// handshakeComplete indicates whether the handshake exchange is done.
	handshakeComplete bool

	// sessions maps agent_id (base-10 string) to active sessions.
	sessions map[string]*ManagedSession

	// sessionsByID maps session_id (base-10 string) to sessions.
	sessionsByID map[string]*ManagedSession

	dispatcher event.Dispatcher
	resolver   *paths.Resolver
	bufferSize int64

	// hubConnected tracks whether the hub connection is active.
	hubConnected bool

	// outboundBuffer holds cross-host publications buffered during disconnection.
	outboundBuffer *OutboundBuffer

	// subs holds active NATS subscriptions.
	subs []*nats.Subscription

	cancel context.CancelFunc
}

// ManagedSession represents a PTY session managed by this host manager.
type ManagedSession struct {
	mu sync.Mutex

	// SessionID is the unique session identifier (base-10 string).
	SessionID string

	// AgentID is the agent this session belongs to (base-10 string).
	AgentID string

	// AgentSlug is the normalized agent slug.
	AgentSlug string

	// RepoPath is the git repository root on this host.
	RepoPath string

	// ReplayBuf is the per-session PTY output replay buffer.
	ReplayBuf *buffer.Ring

	// cmd is the running process.
	cmd *exec.Cmd

	// ptyMaster is the PTY master file descriptor for I/O.
	ptyMaster io.ReadWriteCloser

	// done is closed when the session exits.
	done chan struct{}

	// running indicates whether the session is active.
	running bool

	// replayPending indicates whether a replay is in progress.
	// While true, live PTY output MUST NOT be published.
	replayPending bool

	// liveBuf holds PTY output produced during a replay operation.
	liveBuf []byte
}

// New creates a new Manager with the given NATS connection and configuration.
func New(conn *natsconn.Conn, cfg *config.Config, hostID string, dispatcher event.Dispatcher) *Manager {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	prefix := cfg.Remote.NATS.SubjectPrefix
	if prefix == "" {
		prefix = "amux"
	}
	bufSize := cfg.Remote.BufferSize.Bytes
	if bufSize == 0 {
		bufSize = 10 * 1024 * 1024 // 10MB default
	}
	return &Manager{
		conn:           conn,
		cfg:            cfg,
		prefix:         prefix,
		hostID:         hostID,
		peerID:         ids.EncodeID(ids.NewID()),
		sessions:       make(map[string]*ManagedSession),
		sessionsByID:   make(map[string]*ManagedSession),
		dispatcher:     dispatcher,
		resolver:       paths.DefaultResolver,
		bufferSize:     bufSize,
		hubConnected:   true,
		outboundBuffer: NewOutboundBuffer(bufSize),
	}
}

// Start performs the handshake and begins listening for control requests.
//
// Per spec §5.5.7.6: daemon MUST:
// 1. Connect to NATS
// 2. Perform handshake on P.handshake.<host_id>
// 3. Start listening on P.ctl.<host_id> and P.pty.<host_id>.*.in
func (m *Manager) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	// Step 1: Perform handshake
	if err := m.performHandshake(ctx); err != nil {
		cancel()
		return fmt.Errorf("manager start: %w", err)
	}

	// Step 2: Subscribe to control requests
	ctlSub, err := m.conn.NC().Subscribe(
		protocol.ControlSubject(m.prefix, m.hostID),
		m.handleControlRequest,
	)
	if err != nil {
		cancel()
		return fmt.Errorf("manager subscribe control: %w", err)
	}
	m.subs = append(m.subs, ctlSub)

	// Step 3: Subscribe to PTY input (wildcard for all sessions)
	ptyInSub, err := m.conn.NC().Subscribe(
		protocol.PTYInputWildcard(m.prefix, m.hostID),
		m.handlePTYInput,
	)
	if err != nil {
		cancel()
		return fmt.Errorf("manager subscribe pty input: %w", err)
	}
	m.subs = append(m.subs, ptyInSub)

	// Emit connection.established event
	m.publishEvent("connection.established", &protocol.ConnectionEstablishedEvent{
		PeerID:    m.peerID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	})

	return nil
}

// Stop gracefully shuts down the manager.
func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}

	for _, sub := range m.subs {
		_ = sub.Unsubscribe()
	}

	// Stop all sessions
	m.mu.Lock()
	for _, sess := range m.sessions {
		sess.stop()
	}
	m.mu.Unlock()

	return nil
}

// performHandshake sends a handshake request to the director and waits for a reply.
//
// Per spec §5.5.7.3: the daemon MUST send a handshake request after establishing
// a NATS connection and MUST NOT accept spawn/kill/replay until complete.
func (m *Manager) performHandshake(ctx context.Context) error {
	payload := &protocol.HandshakePayload{
		Protocol: protocol.ProtocolVersion,
		PeerID:   m.peerID,
		Role:     "daemon",
		HostID:   m.hostID,
	}

	ctlMsg, err := protocol.NewControlMessage(protocol.TypeHandshake, payload)
	if err != nil {
		return fmt.Errorf("handshake: marshal: %w", err)
	}

	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return fmt.Errorf("handshake: encode: %w", err)
	}

	timeout := m.cfg.Remote.RequestTimeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	reply, err := m.conn.Request(
		protocol.HandshakeSubject(m.prefix, m.hostID),
		data, timeout,
	)
	if err != nil {
		return fmt.Errorf("handshake: request: %w", err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		return fmt.Errorf("handshake: decode response: %w", err)
	}

	if respMsg.Type == protocol.TypeError {
		var errPayload protocol.ErrorPayload
		_ = respMsg.DecodePayload(&errPayload)
		return fmt.Errorf("handshake rejected [%s]: %s", errPayload.Code, errPayload.Message)
	}

	if respMsg.Type != protocol.TypeHandshake {
		return fmt.Errorf("handshake: unexpected response type %q", respMsg.Type)
	}

	m.mu.Lock()
	m.handshakeComplete = true
	m.mu.Unlock()

	return nil
}

// handleControlRequest processes control requests from the director.
func (m *Manager) handleControlRequest(msg *nats.Msg) {
	var ctlMsg protocol.ControlMessage
	if err := json.Unmarshal(msg.Data, &ctlMsg); err != nil {
		m.replyError(msg, "unknown", protocol.CodeProtocolError,
			"invalid control message: "+err.Error())
		return
	}

	// Check handshake before processing spawn/kill/replay
	m.mu.RLock()
	ready := m.handshakeComplete
	m.mu.RUnlock()

	if !ready && ctlMsg.Type != protocol.TypeHandshake {
		m.replyError(msg, ctlMsg.Type, protocol.CodeNotReady,
			"handshake not yet complete")
		return
	}

	switch ctlMsg.Type {
	case protocol.TypeSpawn:
		m.handleSpawn(msg, &ctlMsg)
	case protocol.TypeKill:
		m.handleKill(msg, &ctlMsg)
	case protocol.TypeReplay:
		m.handleReplay(msg, &ctlMsg)
	case protocol.TypePing:
		m.handlePing(msg, &ctlMsg)
	default:
		m.replyError(msg, ctlMsg.Type, protocol.CodeProtocolError,
			"unknown control message type: "+ctlMsg.Type)
	}
}

// handleSpawn creates a new PTY session for an agent.
//
// Per spec §5.5.7.3: spawn MUST be idempotent for a given agent_id.
func (m *Manager) handleSpawn(msg *nats.Msg, ctlMsg *protocol.ControlMessage) {
	var req protocol.SpawnRequest
	if err := ctlMsg.DecodePayload(&req); err != nil {
		m.replyError(msg, protocol.TypeSpawn, protocol.CodeProtocolError,
			"invalid spawn payload: "+err.Error())
		return
	}

	m.mu.Lock()

	// Check for existing session (idempotency)
	if existing, ok := m.sessions[req.AgentID]; ok {
		// Check for session_conflict
		if existing.AgentSlug != req.AgentSlug || existing.RepoPath != req.RepoPath {
			m.mu.Unlock()
			m.replyError(msg, protocol.TypeSpawn, protocol.CodeSessionConflict,
				fmt.Sprintf("session exists with different slug (%q vs %q) or repo_path (%q vs %q)",
					existing.AgentSlug, req.AgentSlug, existing.RepoPath, req.RepoPath))
			return
		}
		// Return existing session
		sessionID := existing.SessionID
		m.mu.Unlock()
		m.replyControl(msg, protocol.TypeSpawn, &protocol.SpawnResponse{
			AgentID:   req.AgentID,
			SessionID: sessionID,
		})
		return
	}

	m.mu.Unlock()

	// Expand repo_path if it starts with ~/
	repoPath := paths.ExpandHome(req.RepoPath)

	// Validate repo_path is a git repo
	repoRoot, err := paths.FindRepoRoot(repoPath)
	if err != nil || repoRoot == "" {
		m.replyError(msg, protocol.TypeSpawn, protocol.CodeInvalidRepo,
			fmt.Sprintf("repo_path %q is not a git repository", req.RepoPath))
		return
	}

	// Create or reuse worktree at .amux/worktrees/{agent_slug}/
	// (worktree operations are done inline; the manager handles its own worktrees)
	worktreePath := repoRoot + "/.amux/worktrees/" + req.AgentSlug

	// Build the command
	if len(req.Command) == 0 {
		m.replyError(msg, protocol.TypeSpawn, protocol.CodeInvalidAgent,
			"command must not be empty")
		return
	}

	sessionID := ids.EncodeID(ids.NewID())

	// Create PTY and start process
	cmd := exec.Command(req.Command[0], req.Command[1:]...)
	cmd.Dir = worktreePath
	if req.Env != nil {
		for k, v := range req.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	ptyMaster, err := startPTY(cmd)
	if err != nil {
		m.replyError(msg, protocol.TypeSpawn, protocol.CodeInternalError,
			"pty start failed: "+err.Error())
		return
	}

	sess := &ManagedSession{
		SessionID: sessionID,
		AgentID:   req.AgentID,
		AgentSlug: req.AgentSlug,
		RepoPath:  req.RepoPath,
		ReplayBuf: buffer.NewRing(m.bufferSize),
		cmd:       cmd,
		ptyMaster: ptyMaster,
		done:      make(chan struct{}),
		running:   true,
	}

	m.mu.Lock()
	m.sessions[req.AgentID] = sess
	m.sessionsByID[sessionID] = sess
	m.mu.Unlock()

	// Start PTY output reader goroutine
	go m.readPTYOutput(sess)

	// Start process monitor goroutine
	go m.watchSession(sess)

	// Reply with success
	m.replyControl(msg, protocol.TypeSpawn, &protocol.SpawnResponse{
		AgentID:   req.AgentID,
		SessionID: sessionID,
	})
}

// handleKill terminates a session.
func (m *Manager) handleKill(msg *nats.Msg, ctlMsg *protocol.ControlMessage) {
	var req protocol.KillRequest
	if err := ctlMsg.DecodePayload(&req); err != nil {
		m.replyError(msg, protocol.TypeKill, protocol.CodeProtocolError,
			"invalid kill payload: "+err.Error())
		return
	}

	m.mu.RLock()
	sess, ok := m.sessionsByID[req.SessionID]
	m.mu.RUnlock()

	if !ok {
		m.replyControl(msg, protocol.TypeKill, &protocol.KillResponse{
			SessionID: req.SessionID,
			Killed:    false,
		})
		return
	}

	sess.stop()

	m.replyControl(msg, protocol.TypeKill, &protocol.KillResponse{
		SessionID: req.SessionID,
		Killed:    true,
	})
}

// handleReplay replays buffered PTY output for a session.
//
// Per spec §5.5.7.3: the daemon MUST publish all replay bytes before
// any subsequently produced live PTY output bytes.
func (m *Manager) handleReplay(msg *nats.Msg, ctlMsg *protocol.ControlMessage) {
	var req protocol.ReplayRequest
	if err := ctlMsg.DecodePayload(&req); err != nil {
		m.replyError(msg, protocol.TypeReplay, protocol.CodeProtocolError,
			"invalid replay payload: "+err.Error())
		return
	}

	m.mu.RLock()
	sess, ok := m.sessionsByID[req.SessionID]
	m.mu.RUnlock()

	if !ok || !sess.ReplayBuf.Enabled() {
		m.replyControl(msg, protocol.TypeReplay, &protocol.ReplayResponse{
			SessionID: req.SessionID,
			Accepted:  false,
		})
		return
	}

	// Take a snapshot of the replay buffer
	snapshot := sess.ReplayBuf.Snapshot()

	// Mark replay as pending to gate live output
	sess.mu.Lock()
	sess.replayPending = true
	sess.mu.Unlock()

	// Reply with accepted
	m.replyControl(msg, protocol.TypeReplay, &protocol.ReplayResponse{
		SessionID: req.SessionID,
		Accepted:  true,
	})

	// Publish replay bytes in chunks
	if snapshot != nil {
		subject := protocol.PTYOutputSubject(m.prefix, m.hostID, req.SessionID)
		m.publishChunked(subject, snapshot)
	}

	// Release the live output gate
	sess.mu.Lock()
	// Flush any live output that was buffered during replay
	liveBuf := sess.liveBuf
	sess.liveBuf = nil
	sess.replayPending = false
	sess.mu.Unlock()

	// Publish buffered live output
	if len(liveBuf) > 0 {
		subject := protocol.PTYOutputSubject(m.prefix, m.hostID, req.SessionID)
		m.publishChunked(subject, liveBuf)
	}
}

// handlePing responds with a pong.
func (m *Manager) handlePing(msg *nats.Msg, ctlMsg *protocol.ControlMessage) {
	var ping protocol.PingPayload
	if err := ctlMsg.DecodePayload(&ping); err != nil {
		m.replyError(msg, protocol.TypePing, protocol.CodeProtocolError,
			"invalid ping payload: "+err.Error())
		return
	}

	m.replyControl(msg, protocol.TypePong, &protocol.PongPayload{
		TSUnixMs: ping.TSUnixMs,
	})
}

// handlePTYInput receives PTY input from the director and writes it to the session.
func (m *Manager) handlePTYInput(msg *nats.Msg) {
	// Extract session_id from subject: P.pty.<host_id>.<session_id>.in
	sessionID := extractSessionIDFromPTYSubject(msg.Subject, m.prefix, m.hostID)
	if sessionID == "" {
		return
	}

	m.mu.RLock()
	sess, ok := m.sessionsByID[sessionID]
	m.mu.RUnlock()

	if !ok || !sess.running {
		return
	}

	// Write input to PTY master
	_, _ = sess.ptyMaster.Write(msg.Data)
}

// readPTYOutput continuously reads PTY output and publishes it to NATS.
//
// Per spec §5.5.7.3: the replay buffer MUST be updated for all PTY output
// bytes regardless of hub connectivity.
func (m *Manager) readPTYOutput(sess *ManagedSession) {
	buf := make([]byte, 32*1024) // 32KB read buffer
	subject := protocol.PTYOutputSubject(m.prefix, m.hostID, sess.SessionID)

	for {
		n, err := sess.ptyMaster.Read(buf)
		if n > 0 {
			data := buf[:n]

			// Always update replay buffer regardless of connectivity
			sess.ReplayBuf.Write(data)

			// Check if replay is in progress
			sess.mu.Lock()
			if sess.replayPending {
				// Buffer live output during replay per spec §5.5.7.3
				sess.liveBuf = append(sess.liveBuf, data...)
				sess.mu.Unlock()
				continue
			}
			sess.mu.Unlock()

			// Check hub connectivity
			m.mu.RLock()
			connected := m.hubConnected
			m.mu.RUnlock()

			if connected {
				// Publish PTY output, chunking if necessary
				m.publishChunked(subject, data)
			}
			// If not connected, data is retained in the replay buffer
		}
		if err != nil {
			break
		}
	}
}

// watchSession monitors a session and emits events when it exits.
func (m *Manager) watchSession(sess *ManagedSession) {
	if sess.cmd == nil || sess.cmd.Process == nil {
		return
	}
	err := sess.cmd.Wait()

	sess.mu.Lock()
	sess.running = false
	sess.mu.Unlock()

	close(sess.done)

	// Close PTY
	_ = sess.ptyMaster.Close()

	// Emit process event
	eventName := "process.completed"
	if err != nil {
		eventName = "process.failed"
	}
	m.publishEvent(eventName, &protocol.ProcessCompletedEvent{
		AgentID:   sess.AgentID,
		Command:   sess.cmd.Path,
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano),
		EndedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	})

	// Emit agent exit event visible to director
	_ = m.dispatcher.Dispatch(context.Background(), event.NewEvent(
		event.TypeAgentTerminated, muid.MUID(0),
		map[string]any{
			"session_id": sess.SessionID,
			"agent_id":   sess.AgentID,
		},
	))
}

// publishChunked publishes data to a NATS subject, splitting into chunks
// that don't exceed the maximum NATS payload size.
//
// Per spec §5.5.7.4: "Implementations MUST chunk PTY bytes such that no
// single NATS message payload exceeds the maximum supported NATS payload size."
func (m *Manager) publishChunked(subject string, data []byte) {
	maxPayload := int64(m.conn.NC().MaxPayload())
	if maxPayload <= 0 {
		maxPayload = 1024 * 1024 // 1MB default
	}

	for len(data) > 0 {
		chunk := data
		if int64(len(chunk)) > maxPayload {
			chunk = data[:maxPayload]
		}
		_ = m.conn.Publish(subject, chunk)
		data = data[len(chunk):]
	}
}

// publishEvent publishes an EventMessage on the host events subject.
func (m *Manager) publishEvent(name string, data any) {
	evtMsg, err := protocol.NewBroadcastEvent(name, data)
	if err != nil {
		return
	}
	evtData, err := json.Marshal(evtMsg)
	if err != nil {
		return
	}
	subject := protocol.EventsSubject(m.prefix, m.hostID)

	m.mu.RLock()
	connected := m.hubConnected
	m.mu.RUnlock()

	if connected {
		_ = m.conn.Publish(subject, evtData)
	} else {
		// Buffer for later flush
		m.outboundBuffer.Enqueue(subject, evtData)
	}
}

// SetHubConnected updates the hub connection state.
// Called by the NATS disconnect/reconnect handlers.
func (m *Manager) SetHubConnected(connected bool) {
	m.mu.Lock()
	m.hubConnected = connected
	m.mu.Unlock()

	if connected {
		// Flush buffered publications per spec §5.5.8
		m.outboundBuffer.FlushTo(func(subject string, data []byte) {
			_ = m.conn.Publish(subject, data)
		})
	}
}

// replyControl sends a ControlMessage reply.
func (m *Manager) replyControl(msg *nats.Msg, msgType string, payload any) {
	ctlMsg, err := protocol.NewControlMessage(msgType, payload)
	if err != nil {
		return
	}
	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return
	}
	_ = msg.Respond(data)
}

// replyError sends an error ControlMessage reply.
func (m *Manager) replyError(msg *nats.Msg, requestType, code, message string) {
	errMsg, err := protocol.NewErrorMessage(requestType, code, message)
	if err != nil {
		return
	}
	data, err := json.Marshal(errMsg)
	if err != nil {
		return
	}
	_ = msg.Respond(data)
}

// stop gracefully terminates a managed session.
func (s *ManagedSession) stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// Kill the process
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}

	// Close PTY
	if s.ptyMaster != nil {
		_ = s.ptyMaster.Close()
	}

	// Wait for exit
	select {
	case <-s.done:
	case <-time.After(5 * time.Second):
	}
}

// extractSessionIDFromPTYSubject extracts the session_id from a PTY input subject.
// Subject format: P.pty.<host_id>.<session_id>.in
func extractSessionIDFromPTYSubject(subject, prefix, hostID string) string {
	// Expected: prefix.pty.hostID.sessionID.in
	expectedPrefix := prefix + ".pty." + hostID + "."
	if len(subject) <= len(expectedPrefix) {
		return ""
	}
	remainder := subject[len(expectedPrefix):]
	// remainder should be "sessionID.in"
	for i := len(remainder) - 1; i >= 0; i-- {
		if remainder[i] == '.' {
			return remainder[:i]
		}
	}
	return ""
}
