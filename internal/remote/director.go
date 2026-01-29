// director.go implements director-side remote orchestration: hub NATS, handshake handling,
// request-reply spawn/kill/replay with timeout and fail-fast semantics per spec §5.5.6.1, §5.5.7, §5.5.7.2.1.
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

// Director manages the director role: hub connectivity, handshake handling, and control request-reply.
type Director struct {
	cfg    *config.RemoteConfig
	nc     *nats.Conn
	js     nats.JetStreamContext
	kv     nats.KeyValue
	prefix string

	mu       sync.RWMutex
	ready    map[string]struct{} // host_id that have completed handshake
	peerByHost map[string]string // host_id -> peer_id (base-10)
}

// NewDirector creates a director that will use the given remote config.
func NewDirector(cfg *config.RemoteConfig) *Director {
	if cfg == nil {
		cfg = &config.RemoteConfig{}
	}
	prefix := cfg.NATS.SubjectPrefix
	if prefix == "" {
		prefix = "amux"
	}
	return &Director{
		cfg:       cfg,
		prefix:    prefix,
		ready:     make(map[string]struct{}),
		peerByHost: make(map[string]string),
	}
}

// Connect establishes the NATS connection to the hub and provisions the JetStream KV bucket.
func (d *Director) Connect(ctx context.Context, url string) error {
	if url == "" {
		url = d.cfg.NATS.URL
	}
	if url == "" {
		return fmt.Errorf("NATS URL is required")
	}
	opts := []nats.Option{
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			// Spec §5.5.7.2.1: fail fast without issuing NATS request when disconnected.
			d.mu.Lock()
			d.ready = make(map[string]struct{})
			d.peerByHost = make(map[string]string)
			d.mu.Unlock()
		}),
	}
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("connect to hub: %w", err)
	}
	d.nc = nc
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("jetstream: %w", err)
	}
	d.js = js
	kv, err := EnsureKVBucket(ctx, js, d.cfg.NATS.KVBucket)
	if err != nil {
		nc.Close()
		return fmt.Errorf("kv bucket: %w", err)
	}
	d.kv = kv
	return nil
}

// Close closes the NATS connection.
func (d *Director) Close() {
	if d.nc != nil {
		d.nc.Close()
		d.nc = nil
	}
	d.js = nil
	d.kv = nil
	d.mu.Lock()
	d.ready = make(map[string]struct{})
	d.peerByHost = make(map[string]string)
	d.mu.Unlock()
}

// RunHandshakeHandler subscribes to P.handshake.> and handles handshake requests.
// On success it records the host as ready and replies with director handshake payload.
// Call once after Connect; it runs until the subscription is closed or context is done.
func (d *Director) RunHandshakeHandler(ctx context.Context) error {
	if d.nc == nil {
		return fmt.Errorf("not connected")
	}
	sub, err := d.nc.Subscribe(SubjectHandshakeAll(d.prefix), func(msg *nats.Msg) {
		// Reply subject is set by NATS for request-reply
		hostID := extractHostIDFromSubject(d.prefix+".handshake.", msg.Subject)
		if hostID == "" {
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"invalid_subject","message":"missing host_id"}}`))
			return
		}
		var cm ControlMessage
		if err := json.Unmarshal(msg.Data, &cm); err != nil {
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"invalid","message":"invalid JSON"}}`))
			return
		}
		if cm.Type != ControlTypeHandshake {
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"invalid","message":"expected handshake"}}`))
			return
		}
		var hp HandshakePayload
		if err := json.Unmarshal(cm.Payload, &hp); err != nil {
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"invalid","message":"invalid payload"}}`))
			return
		}
		if hp.HostID != hostID {
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"invalid","message":"host_id mismatch"}}`))
			return
		}
		d.mu.Lock()
		if _, exists := d.peerByHost[hostID]; exists {
			d.mu.Unlock()
			_ = msg.Respond([]byte(`{"type":"error","payload":{"request_type":"handshake","code":"collision","message":"host already connected"}}`))
			return
		}
		directorPeerID := api.EncodeID(api.NextRuntimeID())
		d.peerByHost[hostID] = hp.PeerID
		d.ready[hostID] = struct{}{}
		d.mu.Unlock()
		reply := HandshakePayload{
			Protocol: 1,
			PeerID:   directorPeerID,
			Role:     "director",
			HostID:   hostID,
		}
		replyPayload, _ := json.Marshal(reply)
		replyMsg := ControlMessage{Type: ControlTypeHandshake, Payload: replyPayload}
		replyData, _ := json.Marshal(replyMsg)
		_ = msg.Respond(replyData)
	})
	if err != nil {
		return fmt.Errorf("subscribe handshake: %w", err)
	}
	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
	}()
	return nil
}

func extractHostIDFromSubject(prefix, subject string) string {
	if len(subject) <= len(prefix) {
		return ""
	}
	return subject[len(prefix):]
}

// IsReady returns true if the host has completed handshake (spec §5.5.7.2.1 fail-fast).
func (d *Director) IsReady(hostID string) bool {
	d.mu.RLock()
	_, ok := d.ready[hostID]
	d.mu.RUnlock()
	return ok
}

// RequestTimeout returns the configured remote.request_timeout duration.
func (d *Director) RequestTimeout() time.Duration {
	t := d.cfg.RequestTimeout
	if t == "" {
		t = "5s"
	}
	dur, _ := time.ParseDuration(t)
	if dur <= 0 {
		dur = 5 * time.Second
	}
	return dur
}

// Spawn sends a spawn request to P.ctl.<host_id> and waits for response with request_timeout.
// Fail-fast: if host is not ready, returns error without sending (spec §5.5.7.2.1).
func (d *Director) Spawn(ctx context.Context, hostID string, req SpawnPayloadRequest) (SpawnPayloadResponse, error) {
	if !d.IsReady(hostID) {
		return SpawnPayloadResponse{}, fmt.Errorf("host %q not ready (handshake incomplete)", hostID)
	}
	if d.nc == nil {
		return SpawnPayloadResponse{}, fmt.Errorf("not connected")
	}
	payload, _ := json.Marshal(req)
	msg := ControlMessage{Type: ControlTypeSpawn, Payload: payload}
	data, _ := json.Marshal(msg)
	subject := SubjectCtl(d.prefix, hostID)
	reqCtx, cancel := context.WithTimeout(ctx, d.RequestTimeout())
	defer cancel()
	reply, err := d.nc.RequestWithContext(reqCtx, subject, data)
	if err != nil {
		return SpawnPayloadResponse{}, fmt.Errorf("spawn request: %w", err)
	}
	var cm ControlMessage
	if err := json.Unmarshal(reply.Data, &cm); err != nil {
		return SpawnPayloadResponse{}, fmt.Errorf("spawn response: %w", err)
	}
	if cm.Type == ControlTypeError {
		var ep ErrorPayload
		_ = json.Unmarshal(cm.Payload, &ep)
		return SpawnPayloadResponse{}, fmt.Errorf("spawn error %s: %s", ep.Code, ep.Message)
	}
	if cm.Type != ControlTypeSpawn {
		return SpawnPayloadResponse{}, fmt.Errorf("unexpected response type %q", cm.Type)
	}
	var resp SpawnPayloadResponse
	if err := json.Unmarshal(cm.Payload, &resp); err != nil {
		return SpawnPayloadResponse{}, fmt.Errorf("spawn response payload: %w", err)
	}
	return resp, nil
}

// Kill sends a kill request to P.ctl.<host_id> and waits for response.
func (d *Director) Kill(ctx context.Context, hostID string, sessionID string) (KillPayloadResponse, error) {
	if !d.IsReady(hostID) {
		return KillPayloadResponse{}, fmt.Errorf("host %q not ready", hostID)
	}
	if d.nc == nil {
		return KillPayloadResponse{}, fmt.Errorf("not connected")
	}
	req := KillPayloadRequest{SessionID: sessionID}
	payload, _ := json.Marshal(req)
	msg := ControlMessage{Type: ControlTypeKill, Payload: payload}
	data, _ := json.Marshal(msg)
	reqCtx, cancel := context.WithTimeout(ctx, d.RequestTimeout())
	defer cancel()
	reply, err := d.nc.RequestWithContext(reqCtx, SubjectCtl(d.prefix, hostID), data)
	if err != nil {
		return KillPayloadResponse{}, fmt.Errorf("kill request: %w", err)
	}
	var cm ControlMessage
	if err := json.Unmarshal(reply.Data, &cm); err != nil {
		return KillPayloadResponse{}, fmt.Errorf("kill response: %w", err)
	}
	if cm.Type == ControlTypeError {
		var ep ErrorPayload
		_ = json.Unmarshal(cm.Payload, &ep)
		return KillPayloadResponse{}, fmt.Errorf("kill error %s: %s", ep.Code, ep.Message)
	}
	if cm.Type != ControlTypeKill {
		return KillPayloadResponse{}, fmt.Errorf("unexpected response type %q", cm.Type)
	}
	var resp KillPayloadResponse
	if err := json.Unmarshal(cm.Payload, &resp); err != nil {
		return KillPayloadResponse{}, fmt.Errorf("kill response payload: %w", err)
	}
	return resp, nil
}

// Replay sends a replay request to P.ctl.<host_id> and waits for response.
func (d *Director) Replay(ctx context.Context, hostID string, sessionID string) (ReplayPayloadResponse, error) {
	if !d.IsReady(hostID) {
		return ReplayPayloadResponse{}, fmt.Errorf("host %q not ready", hostID)
	}
	if d.nc == nil {
		return ReplayPayloadResponse{}, fmt.Errorf("not connected")
	}
	req := ReplayPayloadRequest{SessionID: sessionID}
	payload, _ := json.Marshal(req)
	msg := ControlMessage{Type: ControlTypeReplay, Payload: payload}
	data, _ := json.Marshal(msg)
	reqCtx, cancel := context.WithTimeout(ctx, d.RequestTimeout())
	defer cancel()
	reply, err := d.nc.RequestWithContext(reqCtx, SubjectCtl(d.prefix, hostID), data)
	if err != nil {
		return ReplayPayloadResponse{}, fmt.Errorf("replay request: %w", err)
	}
	var cm ControlMessage
	if err := json.Unmarshal(reply.Data, &cm); err != nil {
		return ReplayPayloadResponse{}, fmt.Errorf("replay response: %w", err)
	}
	if cm.Type == ControlTypeError {
		var ep ErrorPayload
		_ = json.Unmarshal(cm.Payload, &ep)
		return ReplayPayloadResponse{}, fmt.Errorf("replay error %s: %s", ep.Code, ep.Message)
	}
	if cm.Type != ControlTypeReplay {
		return ReplayPayloadResponse{}, fmt.Errorf("unexpected response type %q", cm.Type)
	}
	var resp ReplayPayloadResponse
	if err := json.Unmarshal(cm.Payload, &resp); err != nil {
		return ReplayPayloadResponse{}, fmt.Errorf("replay response payload: %w", err)
	}
	return resp, nil
}
