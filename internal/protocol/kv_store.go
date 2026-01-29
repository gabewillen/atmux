// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// KVStore provides access to NATS JetStream Key-Value store
type KVStore struct {
	kv jetstream.KeyValue
}

// NewKVStore creates a new KV store instance
func NewKVStore(nc *nats.Conn, bucketName string) (*KVStore, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Create or get the KV bucket
	kv, err := js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:      bucketName,
		Description: "amux remote state storage",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create/get KV bucket: %w", err)
	}

	return &KVStore{
		kv: kv,
	}, nil
}

// HostInfo represents host metadata stored in KV
type HostInfo struct {
	Version     string    `json:"version"`
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	PeerID      string    `json:"peer_id"`
	StartupTime time.Time `json:"startup_time"`
}

// SessionMetadata represents session metadata stored in KV
type SessionMetadata struct {
	AgentID   string `json:"agent_id"`
	AgentSlug string `json:"agent_slug"`
	RepoPath  string `json:"repo_path"`
	State     string `json:"state"`
}

// PutHostInfo stores host information in the KV store
func (kvs *KVStore) PutHostInfo(ctx context.Context, hostID string, info HostInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal host info: %w", err)
	}

	key := fmt.Sprintf("hosts/%s/info", hostID)
	_, err = kvs.kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to put host info: %w", err)
	}

	return nil
}

// GetHostInfo retrieves host information from the KV store
func (kvs *KVStore) GetHostInfo(ctx context.Context, hostID string) (*HostInfo, error) {
	key := fmt.Sprintf("hosts/%s/info", hostID)
	entry, err := kvs.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	var info HostInfo
	if err := json.Unmarshal(entry.Value(), &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal host info: %w", err)
	}

	return &info, nil
}

// PutHeartbeat stores a heartbeat timestamp for a host
func (kvs *KVStore) PutHeartbeat(ctx context.Context, hostID string) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	key := fmt.Sprintf("hosts/%s/heartbeat", hostID)
	_, err := kvs.kv.Put(ctx, key, []byte(timestamp))
	if err != nil {
		return fmt.Errorf("failed to put heartbeat: %w", err)
	}

	return nil
}

// GetHeartbeat retrieves the last heartbeat timestamp for a host
func (kvs *KVStore) GetHeartbeat(ctx context.Context, hostID string) (*time.Time, error) {
	key := fmt.Sprintf("hosts/%s/heartbeat", hostID)
	entry, err := kvs.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get heartbeat: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, string(entry.Value()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse heartbeat timestamp: %w", err)
	}

	return &timestamp, nil
}

// PutSessionMetadata stores session metadata in the KV store
func (kvs *KVStore) PutSessionMetadata(ctx context.Context, hostID, sessionID string, metadata SessionMetadata) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal session metadata: %w", err)
	}

	key := fmt.Sprintf("sessions/%s/%s", hostID, sessionID)
	_, err = kvs.kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to put session metadata: %w", err)
	}

	return nil
}

// GetSessionMetadata retrieves session metadata from the KV store
func (kvs *KVStore) GetSessionMetadata(ctx context.Context, hostID, sessionID string) (*SessionMetadata, error) {
	key := fmt.Sprintf("sessions/%s/%s", hostID, sessionID)
	entry, err := kvs.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session metadata: %w", err)
	}

	var metadata SessionMetadata
	if err := json.Unmarshal(entry.Value(), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session metadata: %w", err)
	}

	return &metadata, nil
}

// DeleteSessionMetadata removes session metadata from the KV store
func (kvs *KVStore) DeleteSessionMetadata(ctx context.Context, hostID, sessionID string) error {
	key := fmt.Sprintf("sessions/%s/%s", hostID, sessionID)
	err := kvs.kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete session metadata: %w", err)
	}

	return nil
}