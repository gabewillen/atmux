package remote

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/monitor"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"
)

// Manager runs on the remote host and manages agents/sessions via NATS.
type Manager struct {
	HostID api.HostID
	Config *config.Config
	NC     *nats.Conn
	Bus    *agent.EventBus
	
	// State
	agents   map[api.AgentID]*agent.Agent
	sessions map[api.SessionID]*RemoteSession
}

type RemoteSession struct {
	*agent.Session
	Replay *ReplayBuffer
	Monitor *monitor.Monitor
}

// NewManager creates a new Manager.
func NewManager(cfg *config.Config, hostID api.HostID) *Manager {
	return &Manager{
		HostID:   hostID,
		Config:   cfg,
		Bus:      agent.NewEventBus(),
		agents:   make(map[api.AgentID]*agent.Agent),
		sessions: make(map[api.SessionID]*RemoteSession),
	}
}

// Start connects to NATS (if not provided) and starts the control loop.
// Note: NC might be provided if we are reusing a connection.
func (m *Manager) Start(ctx context.Context, nc *nats.Conn) error {
	if nc == nil {
		// Connect logic would be here if not provided
		return fmt.Errorf("nats connection required")
	}
	m.NC = nc

	// 1. Handshake
	if err := m.performHandshake(ctx); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	// 2. Subscribe to Control
	ctlSubject := protocol.SubjectForCtl(m.Config.Remote.NATS.SubjectPrefix, m.HostID)
	if _, err := m.NC.Subscribe(ctlSubject, m.handleControl); err != nil {
		return fmt.Errorf("failed to subscribe to control subject: %w", err)
	}

	// 3. Subscribe to PTY In (for input injection)
	// P.pty.<host_id>.*.in
	ptyInSubject := fmt.Sprintf("%s.pty.%s.*.in", m.Config.Remote.NATS.SubjectPrefix, m.HostID)
	if _, err := m.NC.Subscribe(ptyInSubject, m.handlePTYInput); err != nil {
		return fmt.Errorf("failed to subscribe to pty input: %w", err)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

func (m *Manager) performHandshake(ctx context.Context) error {
	req := protocol.HandshakeRequest{
		Protocol: 1,
		Role:     "manager",
		HostID:   m.HostID,
		PeerID:   api.PeerID(muid.Make()), // Generate ephemeral peer ID
	}
	
	data, _ := json.Marshal(req)
	
	// Publish request
	subject := protocol.SubjectForHandshake(m.Config.Remote.NATS.SubjectPrefix, m.HostID)
	
	// We expect a reply? Spec says "daemon sends handshake request on connect... director validates".
	// Usually handshake is Request-Reply.
	
	msg, err := m.NC.RequestWithContext(ctx, subject, data)
	if err != nil {
		return err
	}
	
	var resp protocol.HandshakeResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return err
	}
	
	if resp.Error != nil {
		return fmt.Errorf("handshake rejected: %s", resp.Error.Message)
	}
	
	return nil
}

func (m *Manager) handleControl(msg *nats.Msg) {
	var req protocol.ControlRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		m.replyError(msg, "invalid_request", err.Error())
		return
	}

	var resp protocol.ControlResponse
	resp.Type = req.Type

	var err error
	switch req.Type {
	case "spawn":
		var payload protocol.SpawnPayload
		if err = json.Unmarshal(req.Payload, &payload); err == nil {
			resp.Payload, err = m.handleSpawn(payload)
		}
	case "replay":
		// Replay is usually per session. Payload should contain session_id.
		// Spec says "Replay buffer... replay publishes snapshot".
		// We expect payload to specify session.
		// For now assume generic session lookup or payload structure.
		// Let's assume payload is just session ID string or struct.
		// Spec doesn't strictly define ReplayPayload in types.go, assuming simple map or string.
		// Let's implement basics.
		err = fmt.Errorf("replay not fully implemented in this handler yet")
	default:
		err = fmt.Errorf("unknown command: %s", req.Type)
	}

	if err != nil {
		m.replyError(msg, "execution_failed", err.Error())
		return
	}

	respData, _ := json.Marshal(resp)
	msg.Respond(respData)
}

func (m *Manager) handleSpawn(p protocol.SpawnPayload) (json.RawMessage, error) {
	// 1. Create/Get Agent
	a, exists := m.agents[p.AgentID]
	if !exists {
		// We need to create it.
		// Config comes from payload? Or we use default?
		// For remote spawn, usually some config is passed or we infer it.
		// Using a dummy config for now based on payload.
		cfg := config.AgentConfig{
			Name: string(p.Slug),
			Location: config.LocationConfig{
				RepoPath: p.RepoPath,
			},
		}
		var err error
		a, err = agent.NewAgent(cfg, api.RepoRoot(p.RepoPath))
		if err != nil {
			return nil, err
		}
		m.agents[p.AgentID] = a
	}

	// 2. Spawn
	// TODO: Pass context with timeout?
	ctx := context.Background()
	if err := agent.SpawnAgent(ctx, a); err != nil {
		return nil, err
	}

	// 3. Get Session
	// agent.SpawnAgent creates a session. We need to find the one just created.
	// Hack: get the latest one or refactor SpawnAgent to return it.
	// For now, iterate and find the one with PTY that is not nil.
	var session *agent.Session
	for _, s := range a.Sessions {
		if s.PTY != nil {
			session = s
			break
		}
	}
	if session == nil {
		return nil, fmt.Errorf("session not created")
	}

	// 4. Setup Remote Session (Monitor + Replay)
	rs := &RemoteSession{
		Session: session,
		Replay:  NewReplayBuffer(10 * 1024 * 1024), // 10MB default
	}
	
	// 5. Start Monitor with Hooks
	mon := monitor.NewMonitor(a.ID, m.Bus, session.PTY) // Pass Manager Bus
	// Note: NewMonitor expects a Bus. We should create one or mock it.
	// For Phase 3 Remote, we mainly care about PTY streaming.
	
	// Hook to stream to NATS
	subject := protocol.SubjectForPTYOut(m.Config.Remote.NATS.SubjectPrefix, m.HostID, session.ID)
	streamHook := func(data []byte) {
		// 1. Write to Replay
		rs.Replay.Write(data)
		
		// 2. Publish to NATS
		// Spec says "Payload chunking...". NATS handles some, but we should be careful.
		// pty.Start usually gives small chunks.
		m.NC.Publish(subject, data)
	}
	
	mon.Start(context.Background(), streamHook)
	rs.Monitor = mon
	m.sessions[session.ID] = rs

	// Response
	resp := protocol.SpawnResponsePayload{
		AgentID:   a.ID,
		SessionID: session.ID,
	}
	return json.Marshal(resp)
}

func (m *Manager) handlePTYInput(msg *nats.Msg) {
	// Parse Subject to get SessionID
	// pty.<host>.<session>.*
	// We can't easily parse without regex or token splitting.
	// Assume we can map it or payload has it.
	// Actually, sub is wildcard `*.in`.
	// Use msg.Subject to extract SessionID.
	
	// TODO: Input implementation
}

func (m *Manager) replyError(msg *nats.Msg, code, message string) {
	resp := protocol.ControlResponse{
		Error: &protocol.Error{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	msg.Respond(data)
}
