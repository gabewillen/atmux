// Package remote implements Phase 3 remote agent orchestration.
// This file implements the director role per spec §5.5.4, §5.5.6, §5.5.7.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stateforward/hsm-go/muid"

	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/errors"
)

// Director implements the director role per spec §5.5.6, §5.5.7.
//
// A director:
// - Runs (or connects to) a hub-mode NATS server with JetStream enabled
// - Provisions JetStream KV bucket for durable remote control-plane state
// - Accepts handshake requests from manager-role nodes
// - Sends spawn/kill/replay control requests to managers
// - Subscribes to PTY output and host events from managers
type Director struct {
	cfg      *config.Config
	peerID   muid.MUID
	nc       *nats.Conn
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	subjects SubjectBuilder

	hostsMu sync.RWMutex
	hosts   map[string]*HostState // hostID -> state
}

// HostState tracks the state of a connected manager-role host.
type HostState struct {
	HostID     string
	PeerID     muid.MUID
	Connected  bool
	Handshaken bool
}

// NewDirector creates a new director instance.
//
// The director connects to the hub NATS server at cfg.NATS.Listen (or cfg.Remote.NATS.URL)
// and provisions the required JetStream KV bucket per spec §5.5.6.3.
func NewDirector(ctx context.Context, cfg *config.Config, peerID muid.MUID) (*Director, error) {
	if cfg == nil {
		return nil, errors.Wrap(errors.ErrInvalidInput, "config must not be nil")
	}
	if peerID == 0 {
		return nil, errors.Wrap(errors.ErrInvalidInput, "peer_id must not be zero")
	}

	// Connect to hub NATS server
	opts := []nats.Option{
		nats.Name("amux-director"),
	}

	// For embedded mode, connect to the configured listen address
	url := cfg.Remote.NATS.URL
	if cfg.NATS.Mode == "embedded" && cfg.NATS.Topology == "hub" {
		url = fmt.Sprintf("nats://%s", cfg.NATS.Listen)
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to NATS hub at %s", url)
	}

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, errors.Wrap(err, "create JetStream context")
	}

	d := &Director{
		cfg:      cfg,
		peerID:   peerID,
		nc:       nc,
		js:       js,
		subjects: SubjectBuilder{Prefix: cfg.Remote.NATS.SubjectPrefix},
		hosts:    make(map[string]*HostState),
	}

	// Provision KV bucket per spec §5.5.6.3
	if err := d.provisionKV(ctx); err != nil {
		nc.Close()
		return nil, err
	}

	// Subscribe to handshake requests
	if err := d.subscribeHandshake(ctx); err != nil {
		nc.Close()
		return nil, err
	}

	return d, nil
}

// provisionKV provisions the JetStream KV bucket per spec §5.5.6.3.
func (d *Director) provisionKV(ctx context.Context) error {
	bucketName := d.cfg.Remote.NATS.KVBucket

	// Try to get existing bucket
	kv, err := d.js.KeyValue(ctx, bucketName)
	if err == nil {
		d.kv = kv
		return nil
	}

	// Create bucket if it doesn't exist
	kv, err = d.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "amux remote control-plane durable state",
		Storage:     jetstream.FileStorage,
	})
	if err != nil {
		return errors.Wrapf(err, "create JetStream KV bucket %s", bucketName)
	}

	d.kv = kv
	return nil
}

// subscribeHandshake subscribes to handshake requests on P.handshake.* per spec §5.5.7.3.
func (d *Director) subscribeHandshake(ctx context.Context) error {
	subj := d.subjects.Handshake("*")

	_, err := d.nc.Subscribe(subj, func(msg *nats.Msg) {
		d.handleHandshake(ctx, msg)
	})
	if err != nil {
		return errors.Wrapf(err, "subscribe to handshake subject %s", subj)
	}

	return nil
}

// handleHandshake handles a handshake request from a manager per spec §5.5.7.3.
func (d *Director) handleHandshake(ctx context.Context, msg *nats.Msg) {
	var req HandshakePayload
	msgType, err := UnmarshalControlMessage(msg.Data, &req)
	if err != nil || msgType != "handshake" {
		_ = d.replyHandshakeError(msg, "invalid_request", "malformed handshake")
		return
	}

	// Extract host_id from subject per spec §5.5.7.3
	// Subject format: P.handshake.<host_id>
	// We need to parse the last token from the subject
	subjectHostID := extractHostIDFromSubject(msg.Subject, d.cfg.Remote.NATS.SubjectPrefix)
	if subjectHostID == "" {
		_ = d.replyHandshakeError(msg, "invalid_request", "cannot parse host_id from subject")
		return
	}

	// Validate that payload host_id matches subject host_id per spec §5.5.7.3
	if req.HostID != subjectHostID {
		_ = d.replyHandshakeError(msg, "host_id_mismatch", fmt.Sprintf("payload host_id %s does not match subject host_id %s", req.HostID, subjectHostID))
		return
	}

	// Validate protocol version
	if req.Protocol != 1 {
		_ = d.replyHandshakeError(msg, "unsupported_protocol", fmt.Sprintf("protocol version %d not supported", req.Protocol))
		return
	}

	// Parse peer_id
	peerID, err := ParseID(req.PeerID)
	if err != nil {
		_ = d.replyHandshakeError(msg, "invalid_peer_id", "peer_id must be base-10 integer")
		return
	}

	// Check for collisions per spec §5.5.7.3
	d.hostsMu.Lock()
	if existing, found := d.hosts[req.HostID]; found && existing.Connected {
		if existing.PeerID != peerID {
			d.hostsMu.Unlock()
			_ = d.replyHandshakeError(msg, "host_id_collision", fmt.Sprintf("host_id %s already connected with different peer_id", req.HostID))
			return
		}
	}

	// Accept handshake
	d.hosts[req.HostID] = &HostState{
		HostID:     req.HostID,
		PeerID:     peerID,
		Connected:  true,
		Handshaken: true,
	}
	d.hostsMu.Unlock()

	// Store host info in KV per spec §5.5.6.3
	hostInfo := map[string]any{
		"host_id":   req.HostID,
		"peer_id":   req.PeerID,
		"role":      req.Role,
		"connected": time.Now().UTC().Format(time.RFC3339),
	}
	hostInfoData, _ := json.Marshal(hostInfo)
	_, _ = d.kv.Put(ctx, fmt.Sprintf("hosts/%s/info", req.HostID), hostInfoData)

	// Reply with director handshake
	resp := HandshakePayload{
		Protocol: 1,
		PeerID:   FormatID(d.peerID),
		Role:     "director",
		HostID:   "amux-host", // Director's host_id (could be configurable)
	}
	respData, _ := MarshalControlMessage("handshake", resp)
	_ = msg.Respond(respData)
}

// replyHandshakeError sends an error response to a handshake request.
func (d *Director) replyHandshakeError(msg *nats.Msg, code, message string) error {
	errPayload := ErrorPayload{
		RequestType: "handshake",
		Code:        code,
		Message:     message,
	}
	errData, _ := MarshalControlMessage("error", errPayload)
	return msg.Respond(errData)
}

// extractHostIDFromSubject extracts the host_id from a NATS subject.
// For P.handshake.<host_id>, returns <host_id>.
func extractHostIDFromSubject(subject, prefix string) string {
	// Expected format: prefix.handshake.<host_id>
	// Strip prefix and "handshake." to get host_id
	expected := prefix + ".handshake."
	if len(subject) > len(expected) {
		return subject[len(expected):]
	}
	return ""
}

// Spawn sends a spawn control request to the target host per spec §5.5.7.2, §5.5.7.3.
func (d *Director) Spawn(ctx context.Context, hostID string, req SpawnRequestPayload) (*SpawnResponsePayload, error) {
	// Fail fast if host is not connected per spec §5.5.7.2.1
	d.hostsMu.RLock()
	host, found := d.hosts[hostID]
	if !found || !host.Connected || !host.Handshaken {
		d.hostsMu.RUnlock()
		return nil, errors.Wrapf(errors.ErrRemote, "host %s not connected or handshake incomplete", hostID)
	}
	d.hostsMu.RUnlock()

	reqData, err := MarshalControlMessage("spawn", req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal spawn request")
	}

	subj := d.subjects.Control(hostID)

	resp, err := d.nc.RequestWithContext(ctx, subj, reqData)
	if err != nil {
		return nil, errors.Wrapf(err, "send spawn request to %s", hostID)
	}

	var spawnResp SpawnResponsePayload
	msgType, err := UnmarshalControlMessage(resp.Data, &spawnResp)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal spawn response")
	}

	if msgType == "error" {
		var errPayload ErrorPayload
		_ = json.Unmarshal(resp.Data, &errPayload)
		return nil, errors.Wrapf(errors.ErrRemote, "spawn failed: %s (%s)", errPayload.Message, errPayload.Code)
	}

	if msgType != "spawn" {
		return nil, errors.Wrapf(errors.ErrInvalidInput, "unexpected spawn response type: %s", msgType)
	}

	return &spawnResp, nil
}

// Kill sends a kill control request to the target host per spec §5.5.7.2, §5.5.7.3.
func (d *Director) Kill(ctx context.Context, hostID string, req KillRequestPayload) (*KillResponsePayload, error) {
	// Fail fast if host is not connected per spec §5.5.7.2.1
	d.hostsMu.RLock()
	host, found := d.hosts[hostID]
	if !found || !host.Connected || !host.Handshaken {
		d.hostsMu.RUnlock()
		return nil, errors.Wrapf(errors.ErrRemote, "host %s not connected or handshake incomplete", hostID)
	}
	d.hostsMu.RUnlock()

	reqData, err := MarshalControlMessage("kill", req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal kill request")
	}

	subj := d.subjects.Control(hostID)

	resp, err := d.nc.RequestWithContext(ctx, subj, reqData)
	if err != nil {
		return nil, errors.Wrapf(err, "send kill request to %s", hostID)
	}

	var killResp KillResponsePayload
	msgType, err := UnmarshalControlMessage(resp.Data, &killResp)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal kill response")
	}

	if msgType == "error" {
		var errPayload ErrorPayload
		_ = json.Unmarshal(resp.Data, &errPayload)
		return nil, errors.Wrapf(errors.ErrRemote, "kill failed: %s (%s)", errPayload.Message, errPayload.Code)
	}

	if msgType != "kill" {
		return nil, errors.Wrapf(errors.ErrInvalidInput, "unexpected kill response type: %s", msgType)
	}

	return &killResp, nil
}

// Replay sends a replay control request to the target host per spec §5.5.7.2, §5.5.7.3.
func (d *Director) Replay(ctx context.Context, hostID string, req ReplayRequestPayload) (*ReplayResponsePayload, error) {
	// Fail fast if host is not connected per spec §5.5.7.2.1
	d.hostsMu.RLock()
	host, found := d.hosts[hostID]
	if !found || !host.Connected || !host.Handshaken {
		d.hostsMu.RUnlock()
		return nil, errors.Wrapf(errors.ErrRemote, "host %s not connected or handshake incomplete", hostID)
	}
	d.hostsMu.RUnlock()

	reqData, err := MarshalControlMessage("replay", req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal replay request")
	}

	subj := d.subjects.Control(hostID)

	resp, err := d.nc.RequestWithContext(ctx, subj, reqData)
	if err != nil {
		return nil, errors.Wrapf(err, "send replay request to %s", hostID)
	}

	var replayResp ReplayResponsePayload
	msgType, err := UnmarshalControlMessage(resp.Data, &replayResp)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal replay response")
	}

	if msgType == "error" {
		var errPayload ErrorPayload
		_ = json.Unmarshal(resp.Data, &errPayload)
		return nil, errors.Wrapf(errors.ErrRemote, "replay failed: %s (%s)", errPayload.Message, errPayload.Code)
	}

	if msgType != "replay" {
		return nil, errors.Wrapf(errors.ErrInvalidInput, "unexpected replay response type: %s", msgType)
	}

	return &replayResp, nil
}

// SubscribePTYOutput subscribes to PTY output for a specific session per spec §5.5.7.4.
func (d *Director) SubscribePTYOutput(ctx context.Context, hostID string, sessionID muid.MUID, handler func([]byte)) error {
	subj := d.subjects.PTYOut(hostID, sessionID)

	_, err := d.nc.Subscribe(subj, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return errors.Wrapf(err, "subscribe to PTY output subject %s", subj)
	}

	return nil
}

// PublishPTYInput publishes PTY input for a specific session per spec §5.5.7.4.
func (d *Director) PublishPTYInput(ctx context.Context, hostID string, sessionID muid.MUID, data []byte) error {
	subj := d.subjects.PTYIn(hostID, sessionID)

	if err := d.nc.Publish(subj, data); err != nil {
		return errors.Wrapf(err, "publish PTY input to subject %s", subj)
	}

	return nil
}

// Close closes the director and releases resources.
func (d *Director) Close() error {
	if d.nc != nil {
		d.nc.Close()
	}
	return nil
}
