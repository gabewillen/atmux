package director

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/remote/conn"
	"github.com/agentflare-ai/amux/internal/remote/kv"
	"github.com/agentflare-ai/amux/internal/remote/protocol"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stateforward/hsm-go/muid"
)

// Options for Director.
type Options struct {
	Config config.Config
}

// Director orchestrates remote hosts.
type Director struct {
	opts Options
	nc   *nats.Conn
	js   jetstream.JetStream
	kv   jetstream.KeyValue
}

// New creates a new Director.
func New(opts Options) *Director {
	return &Director{opts: opts}
}

// Start connects to NATS.
func (d *Director) Start(ctx context.Context) error {
	nc, err := conn.Connect(conn.Options{
		URL:           d.opts.Config.Remote.NATS.URL, // Use configured URL
		Name:          "amux-director",
		CredsPath:     d.opts.Config.Remote.NATS.CredsPath, // Optional if hub doesn't need it or keys in URL
		ReconnectWait: 2 * time.Second,
		MaxReconnects: -1, // Infinite
	})
	if err != nil {
		return err
	}
	d.nc = nc

	js, err := conn.JetStream(nc)
	if err != nil {
		nc.Close()
		return err
	}
	d.js = js

	kvBucket, err := kv.EnsureBucket(ctx, js, protocol.KVBucketDefault)
	if err != nil {
		nc.Close()
		return err
	}
	d.kv = kvBucket

	return nil
}

// Stop disconnects.
func (d *Director) Stop() {
	if d.nc != nil {
		d.nc.Close()
	}
}

// ListHosts returns known hosts from KV.
func (d *Director) ListHosts(ctx context.Context) ([]protocol.HostInfo, error) {
	// Key traversal is not directly efficient in NATS KV without keys listing.
	// KV Keys() returns all keys.
	keys, err := d.kv.Keys(ctx)
	if err != nil {
		if err == jetstream.ErrKeyNotFound {
			return nil, nil // No hosts
		}
		return nil, err
	}

	var hosts []protocol.HostInfo
	for _, key := range keys {
		entry, err := d.kv.Get(ctx, key)
		if err != nil {
			continue
		}
		var info protocol.HostInfo
		if err := json.Unmarshal(entry.Value(), &info); err == nil {
			hosts = append(hosts, info)
		}
	}
	return hosts, nil
}

// Spawn sends a spawn request to the target host.
func (d *Director) Spawn(ctx context.Context, hostID string, payload protocol.SpawnPayload) (string, error) {
	reqID := muid.Make().String()
	pBytes, _ := json.Marshal(payload)

	req := protocol.ControlRequest{
		Op:        protocol.OpSpawn,
		RequestID: reqID,
		Payload:   pBytes,
		CreatedAt: time.Now().UTC(),
	}

	resp, err := d.request(ctx, hostID, req)
	if err != nil {
		return "", err
	}

	var respPayload protocol.SpawnResponsePayload
	if err := json.Unmarshal(resp.Payload, &respPayload); err != nil {
		return "", fmt.Errorf("invalid spawn response: %w", err)
	}

	return respPayload.SessionID, nil
}

// Signal sends a signal to a session.
func (d *Director) Signal(ctx context.Context, hostID, sessionID, signal string) error {
	reqID := muid.Make().String()
	payload := protocol.SignalPayload{
		SessionID: sessionID,
		Signal:    signal,
	}
	pBytes, _ := json.Marshal(payload)

	req := protocol.ControlRequest{
		Op:        protocol.OpSignal,
		RequestID: reqID,
		Payload:   pBytes,
		CreatedAt: time.Now().UTC(),
	}

	_, err := d.request(ctx, hostID, req)
	return err
}

// Replay requests replay for a session.
func (d *Director) Replay(ctx context.Context, hostID, sessionID string, sinceSeq uint64) error {
	reqID := muid.Make().String()
	payload := protocol.ReplayPayload{
		SessionID:     sessionID,
		SinceSequence: sinceSeq,
	}
	pBytes, _ := json.Marshal(payload)

	req := protocol.ControlRequest{
		Op:        protocol.OpReplay,
		RequestID: reqID,
		Payload:   pBytes,
		CreatedAt: time.Now().UTC(),
	}

	_, err := d.request(ctx, hostID, req)
	return err
}

func (d *Director) request(ctx context.Context, hostID string, req protocol.ControlRequest) (*protocol.ControlResponse, error) {
	subject := fmt.Sprintf(protocol.ControlSubjectTemplate, d.opts.Config.Remote.NATS.SubjectPrefix, hostID, req.Op)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Use Configured timeout
	timeout, _ := time.ParseDuration(d.opts.Config.Remote.RequestTimeout)
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	msg, err := d.nc.Request(subject, reqBytes, timeout)
	if err != nil {
		return nil, fmt.Errorf("nats request: %w", err)
	}

	var resp protocol.ControlResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response json: %w", err)
	}

	if resp.Status != "ok" {
		msg := "unknown error"
		if resp.Error != nil {
			msg = fmt.Sprintf("[%s] %s", resp.Error.Code, resp.Error.Message)
		}
		return nil, fmt.Errorf("remote error: %s", msg)
	}

	return &resp, nil
}
