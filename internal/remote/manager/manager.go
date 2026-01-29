package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"runtime"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/remote/conn"
	"github.com/agentflare-ai/amux/internal/remote/kv"
	"github.com/agentflare-ai/amux/internal/remote/protocol"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/creack/pty"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stateforward/hsm-go/muid"
)

// Options for Manager.
type Options struct {
	Config   config.Config
	HostID   string
	Worktree *worktree.Manager
}

// Manager manages remote sessions on this host.
type Manager struct {
	opts     Options
	nc       *nats.Conn
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	sessions map[string]*Session
	mu       sync.RWMutex
	ctx      context.Context
	cancel   func()
}

// Session represents a running remote session.
type Session struct {
	ID        string
	Agent     *agent.AgentActor
	Buffer    *ReplayBuffer
	Cancel    func()
	CreatedAt time.Time
}

// New creates a new Manager.
func New(opts Options) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		opts:     opts,
		sessions: make(map[string]*Session),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start connects to NATS and begins listening for control messages.
func (m *Manager) Start(ctx context.Context) error {
	// Connect to NATS
	nc, err := conn.Connect(conn.Options{
		URL:           m.opts.Config.Remote.NATS.URL,
		Name:          fmt.Sprintf("amux-node-%s", m.opts.HostID),
		CredsPath:     m.opts.Config.Remote.NATS.CredsPath,
		ReconnectWait: 2 * time.Second, // TODO parse config duration
		MaxReconnects: 60,
	})
	if err != nil {
		return err
	}
	m.nc = nc

	// Init JetStream
	js, err := conn.JetStream(nc)
	if err != nil {
		nc.Close()
		return err
	}
	m.js = js

	// Obtain KV handle
	kvBucket, err := kv.EnsureBucket(ctx, js, protocol.KVBucketDefault)
	if err != nil {
		nc.Close()
		return err
	}
	m.kv = kvBucket

	// Register Host Info
	hostname, _ := os.Hostname()
	info := protocol.HostInfo{
		ID:           m.opts.HostID,
		Hostname:     hostname,
		Platform:     runtime.GOOS,
		Arch:         runtime.GOARCH,
		Version:      "v2.4.0", // TODO: Build info
		FirstSeenAt:  time.Now().UTC(),
		Capabilities: []string{"manager", "pty"},
	}
	if err := kv.PutHostInfo(ctx, m.kv, info); err != nil {
		nc.Close()
		return fmt.Errorf("register host info: %w", err)
	}

	// Subscribe to control subject
	subject := fmt.Sprintf(protocol.ControlSubjectTemplate,
		m.opts.Config.Remote.NATS.SubjectPrefix,
		m.opts.HostID, "*") // Listen for all ops

	if _, err := nc.Subscribe(subject, m.handleControl); err != nil {
		nc.Close()
		return fmt.Errorf("subscribe control: %w", err)
	}

	// Start Heartbeat loop
	go m.heartbeatLoop()

	return nil
}

// Stop shuts down the manager and all sessions.
func (m *Manager) Stop() {
	m.cancel()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		s.Agent.Stop()
		s.Cancel()
	}
	if m.nc != nil {
		m.nc.Close()
	}
}

func (m *Manager) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second) // TODO config
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Publish heartbeat
			hb := protocol.Heartbeat{
				HostID:    m.opts.HostID,
				Timestamp: time.Now().UTC(),
				Sessions:  len(m.sessions),
			}
			kv.PutHeartbeat(m.ctx, m.kv, hb)
		}
	}
}

func (m *Manager) handleControl(msg *nats.Msg) {
	// Parse subject to get op? Or rely on payload?
	// Envelope has Op.
	var req protocol.ControlRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		m.replyError(msg, "", "invalid_request", "json unmarshal failed")
		return
	}

	switch req.Op {
	case protocol.OpSpawn:
		m.handleSpawn(msg, req)
	case protocol.OpSignal:
		m.handleSignal(msg, req)
	case protocol.OpResize:
		m.handleResize(msg, req)
	case protocol.OpReplay:
		m.handleReplay(msg, req)
	case protocol.OpHandshake:
		m.handleHandshake(msg, req)
	default:
		m.replyError(msg, req.RequestID, "unknown_op", fmt.Sprintf("op %s not supported", req.Op))
	}
}

func (m *Manager) handleHandshake(msg *nats.Msg, req protocol.ControlRequest) {
	// Simple ACK with host info
	resp := protocol.ControlResponse{
		RequestID: req.RequestID,
		Status:    "ok",
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}

func (m *Manager) handleSpawn(msg *nats.Msg, req protocol.ControlRequest) {
	var payload protocol.SpawnPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		m.replyError(msg, req.RequestID, "invalid_payload", "spawn payload unmarshal failed")
		return
	}

	// Create Session
	sessionID := muid.Make().String()

	// Create Agent
	// TODO: Handle agent config properly. For now simple mapping.
	// If agent_slug is provided, use it.
	agentName := payload.AgentSlug
	if agentName == "" {
		agentName = "session-" + sessionID
	}

	// Assuming empty adapter for now if not specified?
	// Or we need a default shell adapter.
	// Phase 0 plan said no built-in adapters, only WASM.
	// Phase 8 implements Adapter interface.
	// For now, assume "shell" or similar.

	a := agent.NewAgent(agentName, "shell", payload.RepoPath, m.opts.Worktree)

	// Start Agent
	a.Start()

	// Create Replay Buffer
	// 1MB default buffer
	rb := NewReplayBuffer(1024 * 1024)

	s := &Session{
		ID:        sessionID,
		Agent:     a,
		Buffer:    rb,
		CreatedAt: time.Now().UTC(),
	}

	// Setup PTY IO
	go m.streamPTY(s)

	m.mu.Lock()
	m.sessions[sessionID] = s
	m.mu.Unlock()

	// Register session in KV
	kv.PutSessionInfo(m.ctx, m.kv, protocol.SessionInfo{
		SessionID: sessionID,
		AgentID:   payload.AgentID, // Store requested AgentID, or generated one?
		HostID:    m.opts.HostID,
		State:     "running",
		CreatedAt: s.CreatedAt,
	})

	// Reply Success
	respPayload := protocol.SpawnResponsePayload{
		SessionID: sessionID,
	}
	pBytes, _ := json.Marshal(respPayload)

	resp := protocol.ControlResponse{
		RequestID: req.RequestID,
		Status:    "ok",
		Payload:   pBytes,
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}

func (m *Manager) streamPTY(s *Session) {
	// Wait for PtyFile?
	// Poll for it
	var ptyFile *os.File
	for i := 0; i < 20; i++ { // 2 seconds timeout
		ptyFile = s.Agent.PtyFile()
		if ptyFile != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if ptyFile == nil {
		// Log error?
		return
	}

	// Stream output
	buf := make([]byte, 4096)
	subject := fmt.Sprintf(protocol.PTYOutSubjectTemplate, m.opts.Config.Remote.NATS.SubjectPrefix, m.opts.HostID, s.ID)

	for {
		n, err := ptyFile.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			// Append to replay buffer
			record := s.Buffer.Append(s.ID, data)

			// Publish
			// Construct PTYIO envelope
			msg := protocol.PTYIO{
				SessionID: s.ID,
				Data:      data,
				Seq:       record.Seq,
				Timestamp: record.Timestamp,
			}
			bytes, _ := json.Marshal(msg)
			m.nc.Publish(subject, bytes)
		}
		if err != nil {
			if err != io.EOF {
				// logical error
			}
			break
		}
	}
}

func (m *Manager) handleSignal(msg *nats.Msg, req protocol.ControlRequest) {
	var payload protocol.SignalPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		m.replyError(msg, req.RequestID, "invalid_payload", "signal payload unmarshal failed")
		return
	}

	m.mu.RLock()
	session, ok := m.sessions[payload.SessionID]
	m.mu.RUnlock()

	if !ok {
		m.replyError(msg, req.RequestID, "not_found", "session not found")
		return
	}

	// Assuming agent has Stop() which triggers cleanup.
	// For now map all signals to Stop.
	session.Agent.Stop()

	resp := protocol.ControlResponse{
		RequestID: req.RequestID,
		Status:    "ok",
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}

func (m *Manager) handleResize(msg *nats.Msg, req protocol.ControlRequest) {
	var payload protocol.ResizePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		m.replyError(msg, req.RequestID, "invalid_payload", "resize payload unmarshal failed")
		return
	}

	m.mu.RLock()
	session, ok := m.sessions[payload.SessionID]
	m.mu.RUnlock()

	if !ok {
		m.replyError(msg, req.RequestID, "not_found", "session not found")
		return
	}

	ptyFile := session.Agent.PtyFile()
	if ptyFile != nil {
		if err := pty.Setsize(ptyFile, &pty.Winsize{
			Rows: uint16(payload.Rows),
			Cols: uint16(payload.Cols),
		}); err != nil {
			m.replyError(msg, req.RequestID, "resize_failed", err.Error())
			return
		}
	}

	resp := protocol.ControlResponse{
		RequestID: req.RequestID,
		Status:    "ok",
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}

func (m *Manager) handleReplay(msg *nats.Msg, req protocol.ControlRequest) {
	var payload protocol.ReplayPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		m.replyError(msg, req.RequestID, "invalid_payload", "replay payload unmarshal failed")
		return
	}

	m.mu.RLock()
	session, ok := m.sessions[payload.SessionID]
	m.mu.RUnlock()

	if !ok {
		m.replyError(msg, req.RequestID, "not_found", "session not found")
		return
	}

	items := session.Buffer.Replay(payload.SinceSequence)

	subject := fmt.Sprintf(protocol.PTYOutSubjectTemplate, m.opts.Config.Remote.NATS.SubjectPrefix, m.opts.HostID, session.ID)

	for _, item := range items {
		bytes, _ := json.Marshal(item)
		m.nc.Publish(subject, bytes)
	}

	resp := protocol.ControlResponse{
		RequestID: req.RequestID,
		Status:    "ok",
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}

func (m *Manager) replyError(msg *nats.Msg, reqID string, code, message string) {
	resp := protocol.ControlResponse{
		RequestID: reqID,
		Status:    "error",
		Error: &protocol.Error{
			Code:    code,
			Message: message,
		},
		CreatedAt: time.Now().UTC(),
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}
