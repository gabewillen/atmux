// Package director implements the director-side remote orchestration for amux.
//
// The director is responsible for:
//   - Managing remote host connections via NATS hub
//   - Tracking connected hosts and their state
//   - Processing handshake exchanges with manager-role daemons
//   - Routing control operations (spawn/kill/replay) to remote hosts
//   - Subscribing to PTY output and publishing PTY input
//   - Handling connection recovery and replay-before-live semantics
//
// See spec §5.5 for remote agent architecture.
package director

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/ids"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote/natsconn"
)

// Director manages remote host orchestration from the director-role node.
type Director struct {
	mu     sync.RWMutex
	conn   *natsconn.Conn
	kv     *natsconn.KVStore
	cfg    *config.Config
	prefix string
	peerID string

	// hosts tracks connected remote hosts by host_id.
	hosts map[string]*HostState

	// sessions tracks active remote sessions by session_id.
	sessions map[string]*RemoteSession

	dispatcher event.Dispatcher

	// subs holds active NATS subscriptions for cleanup.
	subs []*nats.Subscription

	cancel context.CancelFunc
}

// HostState tracks the state of a connected remote host.
type HostState struct {
	// HostID is the unique host identifier.
	HostID string

	// PeerID is the remote daemon's peer identifier.
	PeerID string

	// Connected indicates whether the host is currently connected.
	Connected bool

	// HandshakeComplete indicates whether the handshake exchange is done.
	HandshakeComplete bool

	// ConnectedAt is when the host last connected.
	ConnectedAt time.Time

	// Sessions is the set of session IDs running on this host.
	Sessions map[string]bool
}

// RemoteSession tracks a remote PTY session.
type RemoteSession struct {
	SessionID string
	AgentID   string
	HostID    string
	AgentSlug string
	RepoPath  string

	// ptyOutSub is the subscription for PTY output from this session.
	ptyOutSub *nats.Subscription
}

// New creates a new Director with the given NATS connection and configuration.
func New(conn *natsconn.Conn, cfg *config.Config, dispatcher event.Dispatcher) *Director {
	if dispatcher == nil {
		dispatcher = event.NewNoopDispatcher()
	}
	prefix := cfg.Remote.NATS.SubjectPrefix
	if prefix == "" {
		prefix = "amux"
	}
	return &Director{
		conn:       conn,
		cfg:        cfg,
		prefix:     prefix,
		peerID:     ids.EncodeID(ids.NewID()),
		hosts:      make(map[string]*HostState),
		sessions:   make(map[string]*RemoteSession),
		dispatcher: dispatcher,
	}
}

// Start begins listening for handshake requests and host events.
func (d *Director) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	// Initialize JetStream KV
	bucket := d.cfg.Remote.NATS.KVBucket
	if bucket == "" {
		bucket = "AMUX_KV"
	}
	kv, err := natsconn.InitKV(ctx, d.conn, bucket)
	if err != nil {
		cancel()
		return fmt.Errorf("director start: %w", err)
	}
	d.kv = kv

	// Subscribe to handshake requests from all hosts
	// Handshake uses P.handshake.* (wildcard for all host_ids)
	handshakeSub, err := d.conn.NC().Subscribe(
		d.prefix+".handshake.*",
		d.handleHandshake,
	)
	if err != nil {
		cancel()
		return fmt.Errorf("director subscribe handshake: %w", err)
	}
	d.subs = append(d.subs, handshakeSub)

	// Subscribe to events from all hosts
	eventsSub, err := d.conn.NC().Subscribe(
		d.prefix+".events.*",
		d.handleHostEvent,
	)
	if err != nil {
		cancel()
		return fmt.Errorf("director subscribe events: %w", err)
	}
	d.subs = append(d.subs, eventsSub)

	return nil
}

// Stop gracefully shuts down the director.
func (d *Director) Stop() error {
	if d.cancel != nil {
		d.cancel()
	}

	// Unsubscribe all
	for _, sub := range d.subs {
		_ = sub.Unsubscribe()
	}

	d.mu.Lock()
	// Unsubscribe per-session PTY output subscriptions
	for _, sess := range d.sessions {
		if sess.ptyOutSub != nil {
			_ = sess.ptyOutSub.Unsubscribe()
		}
	}
	d.mu.Unlock()

	return nil
}

// handleHandshake processes a handshake request from a manager-role daemon.
//
// Per spec §5.5.7.3: the director MUST treat the <host_id> token in the
// request subject as canonical. If the handshake payload contains a different
// host_id, the director MUST reject the handshake.
func (d *Director) handleHandshake(msg *nats.Msg) {
	// Extract host_id from subject: P.handshake.<host_id>
	subjectHostID := extractHostIDFromSubject(msg.Subject, d.prefix+".handshake.")

	var ctlMsg protocol.ControlMessage
	if err := json.Unmarshal(msg.Data, &ctlMsg); err != nil {
		d.replyError(msg, "handshake", protocol.CodeProtocolError,
			"invalid handshake message: "+err.Error())
		return
	}

	if ctlMsg.Type != protocol.TypeHandshake {
		d.replyError(msg, "handshake", protocol.CodeProtocolError,
			"expected handshake message, got "+ctlMsg.Type)
		return
	}

	var payload protocol.HandshakePayload
	if err := ctlMsg.DecodePayload(&payload); err != nil {
		d.replyError(msg, "handshake", protocol.CodeProtocolError,
			"invalid handshake payload: "+err.Error())
		return
	}

	// Validate host_id matches subject
	if payload.HostID != subjectHostID {
		d.replyError(msg, "handshake", protocol.CodeHostIDMismatch,
			fmt.Sprintf("host_id in payload (%q) does not match subject (%q)",
				payload.HostID, subjectHostID))
		return
	}

	// Validate protocol version
	if payload.Protocol != protocol.ProtocolVersion {
		d.replyError(msg, "handshake", protocol.CodeProtocolError,
			fmt.Sprintf("unsupported protocol version %d, expected %d",
				payload.Protocol, protocol.ProtocolVersion))
		return
	}

	// Check for peer_id collision
	d.mu.Lock()
	for _, host := range d.hosts {
		if host.PeerID == payload.PeerID && host.HostID != payload.HostID {
			d.mu.Unlock()
			d.replyError(msg, "handshake", protocol.CodePeerCollision,
				fmt.Sprintf("peer_id %q already in use by host %q",
					payload.PeerID, host.HostID))
			return
		}
	}

	// Register or update host
	// If an existing host reconnects with a new peer_id (daemon restart),
	// the update below handles it by overwriting the peer_id.
	host := d.hosts[payload.HostID]
	if host == nil {
		host = &HostState{
			HostID:   payload.HostID,
			Sessions: make(map[string]bool),
		}
		d.hosts[payload.HostID] = host
	}
	host.PeerID = payload.PeerID
	host.Connected = true
	host.HandshakeComplete = true
	host.ConnectedAt = time.Now().UTC()
	d.mu.Unlock()

	// Store host info in KV
	_ = d.kv.PutHostInfo(context.Background(), payload.HostID, &natsconn.HostInfo{
		PeerID:    payload.PeerID,
		StartedAt: host.ConnectedAt.Format(time.RFC3339Nano),
	})

	// Send handshake response
	resp := &protocol.HandshakePayload{
		Protocol: protocol.ProtocolVersion,
		PeerID:   d.peerID,
		Role:     "director",
		HostID:   d.conn.HostID(),
	}
	d.replyControl(msg, protocol.TypeHandshake, resp)

	// Emit connection.established event
	_ = d.dispatcher.Dispatch(context.Background(), event.NewEvent(
		event.TypeConnectionEstablished, muid.MUID(0),
		map[string]any{
			"peer_id":   payload.PeerID,
			"host_id":   payload.HostID,
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		},
	))
}

// handleHostEvent processes an event from a remote host.
func (d *Director) handleHostEvent(msg *nats.Msg) {
	var evtMsg protocol.EventMessage
	if err := json.Unmarshal(msg.Data, &evtMsg); err != nil {
		return // ignore malformed events
	}

	// Dispatch locally
	_ = d.dispatcher.Dispatch(context.Background(), event.NewEvent(
		event.Type(evtMsg.Event.Name), muid.MUID(0), evtMsg.Event.Data,
	))
}

// Spawn sends a spawn request to a remote host.
//
// Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.
func (d *Director) Spawn(ctx context.Context, hostID string, req *protocol.SpawnRequest) (*protocol.SpawnResponse, error) {
	d.mu.RLock()
	host, ok := d.hosts[hostID]
	d.mu.RUnlock()

	// Fail fast if host is not connected
	if !ok || !host.Connected || !host.HandshakeComplete {
		return nil, fmt.Errorf("spawn: host %q is not connected", hostID)
	}

	ctlMsg, err := protocol.NewControlMessage(protocol.TypeSpawn, req)
	if err != nil {
		return nil, fmt.Errorf("spawn: marshal request: %w", err)
	}

	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return nil, fmt.Errorf("spawn: encode message: %w", err)
	}

	// Send request with timeout
	timeout := d.cfg.Remote.RequestTimeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	reply, err := d.conn.Request(
		protocol.ControlSubject(d.prefix, hostID),
		data, timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("spawn: request to %q: %w", hostID, err)
	}

	// Decode response
	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("spawn: decode response: %w", err)
	}

	// Check for error response
	if respMsg.Type == protocol.TypeError {
		var errPayload protocol.ErrorPayload
		if err := respMsg.DecodePayload(&errPayload); err != nil {
			return nil, fmt.Errorf("spawn: decode error response: %w", err)
		}
		if errPayload.Code == protocol.CodeNotReady {
			return nil, fmt.Errorf("spawn: host %q not ready: %s", hostID, errPayload.Message)
		}
		return nil, fmt.Errorf("spawn: host %q error [%s]: %s",
			hostID, errPayload.Code, errPayload.Message)
	}

	if respMsg.Type != protocol.TypeSpawn {
		return nil, fmt.Errorf("spawn: unexpected response type %q", respMsg.Type)
	}

	var resp protocol.SpawnResponse
	if err := respMsg.DecodePayload(&resp); err != nil {
		return nil, fmt.Errorf("spawn: decode spawn response: %w", err)
	}

	// Track the session
	d.mu.Lock()
	session := &RemoteSession{
		SessionID: resp.SessionID,
		AgentID:   req.AgentID,
		HostID:    hostID,
		AgentSlug: req.AgentSlug,
		RepoPath:  req.RepoPath,
	}
	d.sessions[resp.SessionID] = session
	host.Sessions[resp.SessionID] = true
	d.mu.Unlock()

	// Store session metadata in KV
	_ = d.kv.PutSessionMeta(ctx, hostID, resp.SessionID, &natsconn.SessionMeta{
		AgentID:   req.AgentID,
		AgentSlug: req.AgentSlug,
		RepoPath:  req.RepoPath,
		State:     "running",
	})

	return &resp, nil
}

// Kill sends a kill request to a remote host.
//
// Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.
func (d *Director) Kill(ctx context.Context, hostID string, sessionID string) (*protocol.KillResponse, error) {
	d.mu.RLock()
	host, ok := d.hosts[hostID]
	d.mu.RUnlock()

	if !ok || !host.Connected || !host.HandshakeComplete {
		return nil, fmt.Errorf("kill: host %q is not connected", hostID)
	}

	req := &protocol.KillRequest{SessionID: sessionID}
	ctlMsg, err := protocol.NewControlMessage(protocol.TypeKill, req)
	if err != nil {
		return nil, fmt.Errorf("kill: marshal request: %w", err)
	}

	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return nil, fmt.Errorf("kill: encode message: %w", err)
	}

	timeout := d.cfg.Remote.RequestTimeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	reply, err := d.conn.Request(
		protocol.ControlSubject(d.prefix, hostID),
		data, timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("kill: request to %q: %w", hostID, err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("kill: decode response: %w", err)
	}

	if respMsg.Type == protocol.TypeError {
		var errPayload protocol.ErrorPayload
		_ = respMsg.DecodePayload(&errPayload)
		return nil, fmt.Errorf("kill: host %q error [%s]: %s",
			hostID, errPayload.Code, errPayload.Message)
	}

	var resp protocol.KillResponse
	if err := respMsg.DecodePayload(&resp); err != nil {
		return nil, fmt.Errorf("kill: decode response: %w", err)
	}

	if resp.Killed {
		d.mu.Lock()
		delete(d.sessions, sessionID)
		if host, ok := d.hosts[hostID]; ok {
			delete(host.Sessions, sessionID)
		}
		d.mu.Unlock()

		_ = d.kv.DeleteSessionMeta(ctx, hostID, sessionID)
	}

	return &resp, nil
}

// Replay sends a replay request to a remote host.
//
// Per spec §5.5.7.2.1: the director MUST fail fast if the host is disconnected.
func (d *Director) Replay(ctx context.Context, hostID string, sessionID string) (*protocol.ReplayResponse, error) {
	d.mu.RLock()
	host, ok := d.hosts[hostID]
	d.mu.RUnlock()

	if !ok || !host.Connected || !host.HandshakeComplete {
		return nil, fmt.Errorf("replay: host %q is not connected", hostID)
	}

	req := &protocol.ReplayRequest{SessionID: sessionID}
	ctlMsg, err := protocol.NewControlMessage(protocol.TypeReplay, req)
	if err != nil {
		return nil, fmt.Errorf("replay: marshal request: %w", err)
	}

	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return nil, fmt.Errorf("replay: encode message: %w", err)
	}

	timeout := d.cfg.Remote.RequestTimeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	reply, err := d.conn.Request(
		protocol.ControlSubject(d.prefix, hostID),
		data, timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("replay: request to %q: %w", hostID, err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("replay: decode response: %w", err)
	}

	if respMsg.Type == protocol.TypeError {
		var errPayload protocol.ErrorPayload
		_ = respMsg.DecodePayload(&errPayload)
		return nil, fmt.Errorf("replay: host %q error [%s]: %s",
			hostID, errPayload.Code, errPayload.Message)
	}

	var resp protocol.ReplayResponse
	if err := respMsg.DecodePayload(&resp); err != nil {
		return nil, fmt.Errorf("replay: decode response: %w", err)
	}

	return &resp, nil
}

// SubscribePTYOutput subscribes to PTY output for a session on a remote host.
// The handler receives raw PTY output bytes.
func (d *Director) SubscribePTYOutput(hostID, sessionID string, handler func(data []byte)) error {
	subject := protocol.PTYOutputSubject(d.prefix, hostID, sessionID)
	sub, err := d.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return fmt.Errorf("subscribe pty output: %w", err)
	}

	d.mu.Lock()
	if sess, ok := d.sessions[sessionID]; ok {
		sess.ptyOutSub = sub
	}
	d.mu.Unlock()

	return nil
}

// PublishPTYInput sends PTY input to a session on a remote host.
func (d *Director) PublishPTYInput(hostID, sessionID string, data []byte) error {
	subject := protocol.PTYInputSubject(d.prefix, hostID, sessionID)
	return d.conn.Publish(subject, data)
}

// SendPing sends a ping to a remote host via the control subject.
func (d *Director) SendPing(hostID string) (*protocol.PongPayload, error) {
	ctlMsg, err := protocol.NewControlMessage(protocol.TypePing, protocol.NewPingPayload())
	if err != nil {
		return nil, fmt.Errorf("ping: marshal: %w", err)
	}

	data, err := json.Marshal(ctlMsg)
	if err != nil {
		return nil, fmt.Errorf("ping: encode: %w", err)
	}

	timeout := d.cfg.Remote.RequestTimeout.Duration
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	reply, err := d.conn.Request(
		protocol.ControlSubject(d.prefix, hostID),
		data, timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("ping: request to %q: %w", hostID, err)
	}

	var respMsg protocol.ControlMessage
	if err := json.Unmarshal(reply.Data, &respMsg); err != nil {
		return nil, fmt.Errorf("ping: decode response: %w", err)
	}

	var pong protocol.PongPayload
	if err := respMsg.DecodePayload(&pong); err != nil {
		return nil, fmt.Errorf("ping: decode pong: %w", err)
	}

	return &pong, nil
}

// HostConnected returns true if the given host has completed its handshake.
func (d *Director) HostConnected(hostID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	host, ok := d.hosts[hostID]
	return ok && host.Connected && host.HandshakeComplete
}

// SetHostDisconnected marks a host as disconnected.
// Called when the NATS connection to a host is lost.
func (d *Director) SetHostDisconnected(hostID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if host, ok := d.hosts[hostID]; ok {
		host.Connected = false
		host.HandshakeComplete = false
	}
}

// ConnectedHosts returns the list of currently connected host IDs.
func (d *Director) ConnectedHosts() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var result []string
	for id, host := range d.hosts {
		if host.Connected {
			result = append(result, id)
		}
	}
	return result
}

// SessionsForHost returns the session IDs for a given host.
func (d *Director) SessionsForHost(hostID string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	host, ok := d.hosts[hostID]
	if !ok {
		return nil
	}
	var result []string
	for id := range host.Sessions {
		result = append(result, id)
	}
	return result
}

// replyControl sends a ControlMessage reply.
func (d *Director) replyControl(msg *nats.Msg, msgType string, payload any) {
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
func (d *Director) replyError(msg *nats.Msg, requestType, code, message string) {
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

// extractHostIDFromSubject extracts the host_id suffix from a NATS subject.
func extractHostIDFromSubject(subject, prefix string) string {
	if len(subject) <= len(prefix) {
		return ""
	}
	return subject[len(prefix):]
}
