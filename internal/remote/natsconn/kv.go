// Package natsconn - kv.go provides JetStream Key-Value operations
// for durable remote control-plane state.
//
// See spec §5.5.6.3 for KV bucket requirements.
package natsconn

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// KVStore wraps a JetStream KV bucket for amux remote state.
type KVStore struct {
	kv     jetstream.KeyValue
	bucket string
}

// InitKV initializes the JetStream KV bucket, creating it if it does not exist.
//
// Per spec §5.5.6.3: "The director MUST create the bucket if it does not exist."
func InitKV(ctx context.Context, conn *Conn, bucket string) (*KVStore, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("init kv: %w", err)
	}

	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: bucket,
	})
	if err != nil {
		return nil, fmt.Errorf("create kv bucket %q: %w", bucket, err)
	}

	return &KVStore{kv: kv, bucket: bucket}, nil
}

// GetKV connects to an existing JetStream KV bucket.
func GetKV(ctx context.Context, conn *Conn, bucket string) (*KVStore, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("get kv: %w", err)
	}

	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("get kv bucket %q: %w", bucket, err)
	}

	return &KVStore{kv: kv, bucket: bucket}, nil
}

// --- Host info per spec §5.5.6.3: hosts/<host_id>/info ---

// HostInfo holds host metadata stored in JetStream KV.
type HostInfo struct {
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	PeerID    string `json:"peer_id"`
	StartedAt string `json:"started_at"`
}

// PutHostInfo writes host metadata to KV.
// Key: hosts/<host_id>/info
func (s *KVStore) PutHostInfo(ctx context.Context, hostID string, info *HostInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal host info: %w", err)
	}
	key := "hosts." + hostID + ".info"
	_, err = s.kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("put host info for %q: %w", hostID, err)
	}
	return nil
}

// GetHostInfo reads host metadata from KV.
func (s *KVStore) GetHostInfo(ctx context.Context, hostID string) (*HostInfo, error) {
	key := "hosts." + hostID + ".info"
	entry, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get host info for %q: %w", hostID, err)
	}
	var info HostInfo
	if err := json.Unmarshal(entry.Value(), &info); err != nil {
		return nil, fmt.Errorf("unmarshal host info: %w", err)
	}
	return &info, nil
}

// --- Host heartbeat per spec §5.5.6.3: hosts/<host_id>/heartbeat ---

// HostHeartbeat holds the last-seen heartbeat timestamp.
type HostHeartbeat struct {
	Timestamp string `json:"timestamp"`
}

// PutHeartbeat writes a heartbeat timestamp to KV.
// Key: hosts/<host_id>/heartbeat
func (s *KVStore) PutHeartbeat(ctx context.Context, hostID string) error {
	hb := &HostHeartbeat{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(hb)
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}
	key := "hosts." + hostID + ".heartbeat"
	_, err = s.kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("put heartbeat for %q: %w", hostID, err)
	}
	return nil
}

// GetHeartbeat reads the last heartbeat timestamp from KV.
func (s *KVStore) GetHeartbeat(ctx context.Context, hostID string) (*HostHeartbeat, error) {
	key := "hosts." + hostID + ".heartbeat"
	entry, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get heartbeat for %q: %w", hostID, err)
	}
	var hb HostHeartbeat
	if err := json.Unmarshal(entry.Value(), &hb); err != nil {
		return nil, fmt.Errorf("unmarshal heartbeat: %w", err)
	}
	return &hb, nil
}

// --- Session metadata per spec §5.5.6.3: sessions/<host_id>/<session_id> ---

// SessionMeta holds session metadata sufficient for reconnection.
type SessionMeta struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}

// PutSessionMeta writes session metadata to KV.
// Key: sessions.<host_id>.<session_id>
func (s *KVStore) PutSessionMeta(ctx context.Context, hostID, sessionID string, meta *SessionMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal session meta: %w", err)
	}
	key := "sessions." + hostID + "." + sessionID
	_, err = s.kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("put session meta for %q: %w", hostID, err)
	}
	return nil
}

// GetSessionMeta reads session metadata from KV.
func (s *KVStore) GetSessionMeta(ctx context.Context, hostID, sessionID string) (*SessionMeta, error) {
	key := "sessions." + hostID + "." + sessionID
	entry, err := s.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get session meta for %q: %w", hostID, err)
	}
	var meta SessionMeta
	if err := json.Unmarshal(entry.Value(), &meta); err != nil {
		return nil, fmt.Errorf("unmarshal session meta: %w", err)
	}
	return &meta, nil
}

// DeleteSessionMeta removes session metadata from KV.
func (s *KVStore) DeleteSessionMeta(ctx context.Context, hostID, sessionID string) error {
	key := "sessions." + hostID + "." + sessionID
	return s.kv.Delete(ctx, key)
}

// Bucket returns the bucket name.
func (s *KVStore) Bucket() string {
	return s.bucket
}
