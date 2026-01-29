// kv.go implements JetStream KV bucket provisioning and required durable state keys per spec §5.5.6.3.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// KVKeyHostInfo returns the key for host metadata: hosts/<host_id>/info.
func KVKeyHostInfo(hostID string) string {
	return fmt.Sprintf("hosts/%s/info", hostID)
}

// KVKeyHostHeartbeat returns the key for last-seen heartbeat: hosts/<host_id>/heartbeat.
func KVKeyHostHeartbeat(hostID string) string {
	return fmt.Sprintf("hosts/%s/heartbeat", hostID)
}

// KVKeySession returns the key for session metadata: sessions/<host_id>/<session_id>.
func KVKeySession(hostID, sessionID string) string {
	return fmt.Sprintf("sessions/%s/%s", hostID, sessionID)
}

// HostInfoValue is the UTF-8 JSON value for hosts/<host_id>/info (spec §5.5.6.3).
type HostInfoValue struct {
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	PeerID    string `json:"peer_id"`
	StartedAt string `json:"started_at"` // RFC 3339
}

// SessionKVValue is the UTF-8 JSON value for sessions/<host_id>/<session_id> (spec §5.5.6.3).
type SessionKVValue struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}

// EnsureKVBucket creates the JetStream KV bucket if it does not exist (spec §5.5.6.3).
// Bucket name is from config (default AMUX_KV). Returns the KV store and any error.
func EnsureKVBucket(ctx context.Context, js nats.JetStreamContext, bucket string) (nats.KeyValue, error) {
	if bucket == "" {
		bucket = "AMUX_KV"
	}
	kv, err := js.KeyValue(bucket)
	if err != nil {
		_, createErr := js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket:      bucket,
			Description: "amux remote control-plane state",
		})
		if createErr != nil {
			return nil, fmt.Errorf("create KV bucket %q: %w", bucket, createErr)
		}
		kv, err = js.KeyValue(bucket)
		if err != nil {
			return nil, fmt.Errorf("open KV bucket %q: %w", bucket, err)
		}
	}
	return kv, nil
}

// PutHostInfo writes hosts/<host_id>/info as UTF-8 JSON.
func PutHostInfo(kv nats.KeyValue, hostID string, v HostInfoValue) error {
	v.StartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal host info: %w", err)
	}
	_, err = kv.Put(KVKeyHostInfo(hostID), data)
	return err
}

// PutHostHeartbeat writes hosts/<host_id>/heartbeat (RFC 3339 timestamp).
func PutHostHeartbeat(kv nats.KeyValue, hostID string) error {
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := kv.Put(KVKeyHostHeartbeat(hostID), []byte(`"`+ts+`"`))
	return err
}

// PutSession writes sessions/<host_id>/<session_id> as UTF-8 JSON.
func PutSession(kv nats.KeyValue, hostID, sessionID string, v SessionKVValue) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	_, err = kv.Put(KVKeySession(hostID, sessionID), data)
	return err
}
