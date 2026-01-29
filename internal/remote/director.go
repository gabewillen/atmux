package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

// DirectorOptions configures the director runtime.
type DirectorOptions struct {
	Version      string
	HostID       api.HostID
	Bootstrapper *Bootstrapper
}

type hostState struct {
	hostID    api.HostID
	peerID    api.PeerID
	connected bool
	ready     bool
}

// HostSnapshot captures the director's view of a host manager.
type HostSnapshot struct {
	HostID    api.HostID
	PeerID    api.PeerID
	Connected bool
	Ready     bool
}

// Director orchestrates remote hosts via NATS.
type Director struct {
	cfg            config.Config
	dispatcher     protocol.Dispatcher
	subjectPrefix  string
	requestTimeout time.Duration
	kv             *KVStore
	creds          *CredentialStore
	bootstrapper   *Bootstrapper
	hostID         api.HostID
	peerID         api.PeerID
	version        string
	mu             sync.Mutex
	hosts          map[api.HostID]*hostState
	peerIndex      map[string]api.HostID
}

// HostID returns the director's host ID.
func (d *Director) HostID() api.HostID {
	if d == nil {
		return ""
	}
	return d.hostID
}

// PeerID returns the director's peer ID.
func (d *Director) PeerID() api.PeerID {
	if d == nil {
		return api.PeerID{}
	}
	return d.peerID
}

// HostSnapshot returns a snapshot of a host manager's connection state.
func (d *Director) HostSnapshot(hostID api.HostID) (HostSnapshot, bool) {
	if d == nil {
		return HostSnapshot{}, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.hosts[hostID]
	if !ok {
		return HostSnapshot{}, false
	}
	return HostSnapshot{
		HostID:    state.hostID,
		PeerID:    state.peerID,
		Connected: state.connected,
		Ready:     state.ready,
	}, true
}

// Hosts returns snapshots of all known host managers.
func (d *Director) Hosts() []HostSnapshot {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	hosts := make([]HostSnapshot, 0, len(d.hosts))
	for _, state := range d.hosts {
		if state == nil {
			continue
		}
		hosts = append(hosts, HostSnapshot{
			HostID:    state.hostID,
			PeerID:    state.peerID,
			Connected: state.connected,
			Ready:     state.ready,
		})
	}
	return hosts
}

// NewDirector constructs a director orchestrator.
func NewDirector(cfg config.Config, dispatcher protocol.Dispatcher, options DirectorOptions) (*Director, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("remote director: dispatcher required")
	}
	hostID := options.HostID
	if hostID == "" {
		resolved, err := api.ParseHostID(strings.ToLower(hostnameFallback()))
		if err != nil {
			resolved = api.HostID("director")
		}
		hostID = resolved
	}
	peerID, err := LoadOrCreatePeerID(cfg.NATS.JetStreamDir)
	if err != nil {
		return nil, fmt.Errorf("remote director: %w", err)
	}
	kv, err := NewKVStore(dispatcher.JetStream(), cfg.Remote.NATS.KVBucket)
	if err != nil {
		return nil, fmt.Errorf("remote director: %w", err)
	}
	creds, err := NewCredentialStore(cfg.NATS.JetStreamDir)
	if err != nil {
		return nil, fmt.Errorf("remote director: %w", err)
	}
	bootstrapper := options.Bootstrapper
	if bootstrapper == nil {
		bootstrapper = &Bootstrapper{}
	}
	return &Director{
		cfg:            cfg,
		dispatcher:     dispatcher,
		subjectPrefix:  SubjectPrefix(cfg.Remote.NATS.SubjectPrefix),
		requestTimeout: cfg.Remote.RequestTimeout,
		kv:             kv,
		creds:          creds,
		bootstrapper:   bootstrapper,
		hostID:         hostID,
		peerID:         peerID,
		version:        options.Version,
		hosts:          make(map[api.HostID]*hostState),
		peerIndex:      make(map[string]api.HostID),
	}, nil
}

// Start subscribes to handshake and host events subjects.
func (d *Director) Start(ctx context.Context) error {
	handshakeSubject := protocol.Subject(d.subjectPrefix, "handshake", "*")
	_, err := d.dispatcher.SubscribeRaw(ctx, handshakeSubject, d.handleHandshake)
	if err != nil {
		return fmt.Errorf("remote director: %w", err)
	}
	eventsSubject := protocol.Subject(d.subjectPrefix, "events", "*")
	_, err = d.dispatcher.SubscribeRaw(ctx, eventsSubject, d.handleHostEvent)
	if err != nil {
		return fmt.Errorf("remote director: %w", err)
	}
	commSubject := protocol.Subject(d.subjectPrefix, "comm", ">")
	_, err = d.dispatcher.SubscribeRaw(ctx, commSubject, d.handleCommMessage)
	if err != nil {
		return fmt.Errorf("remote director: %w", err)
	}
	return nil
}

// EnsureHost bootstraps the remote host and returns its host ID.
func (d *Director) EnsureHost(ctx context.Context, location api.Location, adapters []AdapterBundle) (api.HostID, Credential, error) {
	hostID, err := HostIDFromLocation(location)
	if err != nil {
		return "", Credential{}, fmt.Errorf("ensure host: %w", err)
	}
	cred, err := d.creds.GetOrCreate(hostID.String(), d.subjectPrefix, d.cfg.Remote.NATS.KVBucket)
	if err != nil {
		return "", Credential{}, fmt.Errorf("ensure host: %w", err)
	}
	hubClient, err := hubClientURL(d.cfg)
	if err != nil {
		return "", Credential{}, fmt.Errorf("ensure host: %w", err)
	}
	req := BootstrapRequest{
		HostID:        hostID,
		Location:      location,
		LeafURL:       hubURL(d.cfg),
		HubClientURL:  hubClient,
		CredsPath:     d.cfg.Remote.NATS.CredsPath,
		SubjectPrefix: d.subjectPrefix,
		KVBucket:      d.cfg.Remote.NATS.KVBucket,
		ManagerModel:  d.cfg.Remote.Manager.Model,
		Adapters:      adapters,
	}
	if err := d.bootstrapper.Bootstrap(ctx, req, cred); err != nil {
		return "", Credential{}, fmt.Errorf("ensure host: %w", err)
	}
	return hostID, cred, nil
}

// Spawn requests a remote spawn for the host.
func (d *Director) Spawn(ctx context.Context, hostID api.HostID, req SpawnRequest) (SpawnResponse, error) {
	msg, err := EncodePayload("spawn", req)
	if err != nil {
		return SpawnResponse{}, fmt.Errorf("spawn: %w", err)
	}
	resp, err := d.sendControl(ctx, hostID, msg)
	if err != nil {
		return SpawnResponse{}, err
	}
	var payload SpawnResponse
	if err := DecodePayload(resp, &payload); err != nil {
		return SpawnResponse{}, fmt.Errorf("spawn: %w", err)
	}
	if err := d.writeSessionKV(ctx, hostID, payload.SessionID, req); err != nil {
		return SpawnResponse{}, fmt.Errorf("spawn: %w", err)
	}
	return payload, nil
}

// Kill requests a remote kill for the host.
func (d *Director) Kill(ctx context.Context, hostID api.HostID, req KillRequest) (KillResponse, error) {
	msg, err := EncodePayload("kill", req)
	if err != nil {
		return KillResponse{}, fmt.Errorf("kill: %w", err)
	}
	resp, err := d.sendControl(ctx, hostID, msg)
	if err != nil {
		return KillResponse{}, err
	}
	var payload KillResponse
	if err := DecodePayload(resp, &payload); err != nil {
		return KillResponse{}, fmt.Errorf("kill: %w", err)
	}
	return payload, nil
}

// Replay requests a replay for the host.
func (d *Director) Replay(ctx context.Context, hostID api.HostID, req ReplayRequest) (ReplayResponse, error) {
	msg, err := EncodePayload("replay", req)
	if err != nil {
		return ReplayResponse{}, fmt.Errorf("replay: %w", err)
	}
	resp, err := d.sendControl(ctx, hostID, msg)
	if err != nil {
		return ReplayResponse{}, err
	}
	var payload ReplayResponse
	if err := DecodePayload(resp, &payload); err != nil {
		return ReplayResponse{}, fmt.Errorf("replay: %w", err)
	}
	return payload, nil
}

// AttachPTY opens a PTY connection via NATS subjects.
func (d *Director) AttachPTY(ctx context.Context, hostID api.HostID, sessionID api.SessionID) (net.Conn, error) {
	if err := d.ensureConnected(hostID); err != nil {
		return nil, err
	}
	conn, err := NewPTYConn(ctx, d.dispatcher, d.subjectPrefix, hostID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("attach pty: %w", err)
	}
	return conn, nil
}

func (d *Director) sendControl(ctx context.Context, hostID api.HostID, msg ControlMessage) (ControlMessage, error) {
	if err := d.ensureConnected(hostID); err != nil {
		return ControlMessage{}, err
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return ControlMessage{}, fmt.Errorf("control: %w", err)
	}
	subject := ControlSubject(d.subjectPrefix, hostID)
	reply, err := d.dispatcher.Request(ctx, subject, data, d.requestTimeout)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrNoResponders) {
			d.setReady(hostID, false)
			d.setConnected(hostID, false)
		}
		return ControlMessage{}, fmt.Errorf("control: %w", err)
	}
	resp, err := DecodeControlMessage(reply.Data)
	if err != nil {
		return ControlMessage{}, fmt.Errorf("control: %w", err)
	}
	if resp.Type == "error" {
		var payload ErrorPayload
		if err := DecodePayload(resp, &payload); err != nil {
			return ControlMessage{}, fmt.Errorf("control: %w", err)
		}
		if payload.Code == "not_ready" {
			d.setReady(hostID, false)
			return ControlMessage{}, fmt.Errorf("control: %w", ErrNotReady)
		}
		return ControlMessage{}, fmt.Errorf("control: %w", fmt.Errorf("%s", payload.Message))
	}
	return resp, nil
}

func (d *Director) handleHandshake(msg protocol.Message) {
	if msg.Reply == "" {
		return
	}
	hostID, err := ParseHandshakeSubject(d.subjectPrefix, msg.Subject)
	if err != nil {
		return
	}
	control, err := DecodeControlMessage(msg.Data)
	if err != nil || control.Type != "handshake" {
		_ = d.replyError(msg.Reply, "handshake", "invalid_handshake", "invalid handshake")
		return
	}
	var payload HandshakePayload
	if err := DecodePayload(control, &payload); err != nil {
		_ = d.replyError(msg.Reply, "handshake", "invalid_handshake", "invalid handshake")
		return
	}
	if payload.HostID != hostID.String() {
		_ = d.replyError(msg.Reply, "handshake", "invalid_host", "host id mismatch")
		return
	}
	if payload.Protocol != 1 {
		_ = d.replyError(msg.Reply, "handshake", "unsupported_protocol", "unsupported protocol")
		return
	}
	peerID, err := api.ParsePeerID(payload.PeerID)
	if err != nil {
		_ = d.replyError(msg.Reply, "handshake", "invalid_peer", "invalid peer")
		return
	}
	d.mu.Lock()
	if existing, ok := d.hosts[hostID]; ok {
		if existing.peerID.String() != peerID.String() {
			if existing.connected {
				d.mu.Unlock()
				_ = d.replyError(msg.Reply, "handshake", "host_conflict", "host already connected")
				return
			}
			delete(d.peerIndex, existing.peerID.String())
		}
	}
	if owner, ok := d.peerIndex[peerID.String()]; ok && owner != hostID {
		d.mu.Unlock()
		_ = d.replyError(msg.Reply, "handshake", "peer_conflict", "peer already connected")
		return
	}
	state := &hostState{hostID: hostID, peerID: peerID, connected: true, ready: true}
	d.hosts[hostID] = state
	d.peerIndex[peerID.String()] = hostID
	d.mu.Unlock()
	if err := d.writeHostKV(context.Background(), hostID, peerID); err != nil {
		_ = d.replyError(msg.Reply, "handshake", "kv_error", "failed to write host info")
		return
	}
	response, err := EncodePayload("handshake", HandshakePayload{
		Protocol: 1,
		PeerID:   d.peerID.String(),
		Role:     "director",
		HostID:   d.hostID.String(),
	})
	if err != nil {
		_ = d.replyError(msg.Reply, "handshake", "internal", "failed to encode handshake")
		return
	}
	data, err := EncodeControlMessage(response)
	if err != nil {
		_ = d.replyError(msg.Reply, "handshake", "internal", "failed to encode handshake")
		return
	}
	_ = d.dispatcher.PublishRaw(context.Background(), msg.Reply, data, "")
}

func (d *Director) handleHostEvent(msg protocol.Message) {
	prefixParts := strings.Split(d.subjectPrefix, ".")
	subjectParts := strings.Split(msg.Subject, ".")
	if len(subjectParts) != len(prefixParts)+2 {
		return
	}
	for i, part := range prefixParts {
		if subjectParts[i] != part {
			return
		}
	}
	if subjectParts[len(prefixParts)] != "events" {
		return
	}
	hostID, err := api.ParseHostID(subjectParts[len(prefixParts)+1])
	if err != nil {
		return
	}
	var event EventMessage
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return
	}
	name := event.Event.Name
	if name == "connection.established" || name == "connection.recovered" {
		d.setConnected(hostID, true)
		d.setReady(hostID, true)
		go d.requestReplay(context.Background(), hostID)
		return
	}
	if name == "connection.lost" {
		d.setReady(hostID, false)
		d.setConnected(hostID, false)
	}
}

func (d *Director) handleCommMessage(msg protocol.Message) {
	_ = msg
}

func (d *Director) setReady(hostID api.HostID, ready bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.hosts[hostID]
	if !ok {
		return
	}
	state.ready = ready
}

func (d *Director) setConnected(hostID api.HostID, connected bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.hosts[hostID]
	if !ok {
		return
	}
	state.connected = connected
}

// HostReady reports whether the host is ready for control requests.
func (d *Director) HostReady(hostID api.HostID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.hosts[hostID]
	if !ok {
		return false
	}
	return state.ready
}

func (d *Director) ensureConnected(hostID api.HostID) error {
	d.mu.Lock()
	state, ok := d.hosts[hostID]
	d.mu.Unlock()
	if !ok || !state.connected {
		return fmt.Errorf("remote host: %w", ErrHostDisconnected)
	}
	if !state.ready {
		return fmt.Errorf("remote host: %w", ErrNotReady)
	}
	return nil
}

func (d *Director) replyError(reply, requestType, code, message string) error {
	msg, err := NewErrorMessage(requestType, code, message)
	if err != nil {
		return fmt.Errorf("reply error: %w", err)
	}
	data, err := EncodeControlMessage(msg)
	if err != nil {
		return fmt.Errorf("reply error: %w", err)
	}
	if err := d.dispatcher.PublishRaw(context.Background(), reply, data, ""); err != nil {
		return fmt.Errorf("reply error: %w", err)
	}
	return nil
}

func (d *Director) writeHostKV(ctx context.Context, hostID api.HostID, peerID api.PeerID) error {
	info := map[string]any{
		"version":    d.version,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"peer_id":    peerID.String(),
		"started_at": NowRFC3339(),
	}
	encoded, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("kv host: %w", err)
	}
	if err := d.kv.Put(ctx, fmt.Sprintf("hosts/%s/info", hostID.String()), encoded); err != nil {
		return fmt.Errorf("kv host: %w", err)
	}
	heartbeat := map[string]any{"timestamp": NowRFC3339()}
	beat, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("kv host: %w", err)
	}
	if err := d.kv.Put(ctx, fmt.Sprintf("hosts/%s/heartbeat", hostID.String()), beat); err != nil {
		return fmt.Errorf("kv host: %w", err)
	}
	return nil
}

func (d *Director) writeSessionKV(ctx context.Context, hostID api.HostID, sessionID string, req SpawnRequest) error {
	if sessionID == "" {
		return nil
	}
	data := map[string]any{
		"agent_id":   req.AgentID,
		"agent_slug": req.AgentSlug,
		"repo_path":  req.RepoPath,
		"state":      "running",
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("kv session: %w", err)
	}
	key := fmt.Sprintf("sessions/%s/%s", hostID.String(), sessionID)
	if err := d.kv.Put(ctx, key, encoded); err != nil {
		return fmt.Errorf("kv session: %w", err)
	}
	return nil
}

func (d *Director) requestReplay(ctx context.Context, hostID api.HostID) {
	if d == nil || d.kv == nil {
		return
	}
	prefix := fmt.Sprintf("sessions/%s/", hostID.String())
	keys, err := d.kv.ListKeys(ctx, prefix)
	if err != nil {
		return
	}
	for _, key := range keys {
		sessionID := strings.TrimPrefix(key, prefix)
		if sessionID == "" {
			continue
		}
		state := "running"
		if data, err := d.kv.Get(ctx, key); err == nil && len(data) > 0 {
			var payload struct {
				State string `json:"state"`
			}
			if err := json.Unmarshal(data, &payload); err == nil && payload.State != "" {
				state = payload.State
			}
		}
		if state != "running" {
			continue
		}
		_, _ = d.Replay(ctx, hostID, ReplayRequest{SessionID: sessionID})
	}
}

// HostIDFromLocation derives host_id from location.
func HostIDFromLocation(location api.Location) (api.HostID, error) {
	host := strings.TrimSpace(location.Host)
	if host == "" {
		return "", fmt.Errorf("host id: %w", ErrInvalidMessage)
	}
	return api.ParseHostID(host)
}

func hostnameFallback() string {
	name, err := os.Hostname()
	if err != nil {
		return "director"
	}
	if name == "" {
		return "director"
	}
	return name
}

func hubURL(cfg config.Config) string {
	url := strings.TrimSpace(cfg.Remote.NATS.URL)
	if url != "" {
		return url
	}
	role := strings.TrimSpace(cfg.Node.Role)
	if role != "manager" {
		if url = strings.TrimSpace(cfg.NATS.LeafAdvertiseURL); url != "" {
			return url
		}
	}
	return cfg.NATS.HubURL
}
