package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/agentflare-ai/amux/internal/remote/protocol"
	"github.com/nats-io/nats.go/jetstream"
)

// EnsureBucket ensures the KV bucket exists.
func EnsureBucket(ctx context.Context, js jetstream.JetStream, bucket string) (jetstream.KeyValue, error) {
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucket,
		Description: "Amux Remote State",
		Compression: true,
		TTL:         24 * time.Hour, // Reasonable default for ephemeral state
	})
	if err != nil {
		return nil, fmt.Errorf("create kv bucket %s: %w", bucket, err)
	}
	return kv, nil
}

// PutHostInfo writes host info to KV.
func PutHostInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.HostInfo) error {
	key := fmt.Sprintf(protocol.KVHostInfoTemplate, info.ID)
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal host info: %w", err)
	}
	if _, err := kv.Put(ctx, key, data); err != nil {
		return fmt.Errorf("put host info: %w", err)
	}
	return nil
}

// PutHeartbeat writes a heartbeat to KV.
func PutHeartbeat(ctx context.Context, kv jetstream.KeyValue, hb protocol.Heartbeat) error {
	key := fmt.Sprintf(protocol.KVHostHeartbeatTemplate, hb.HostID)
	data, err := json.Marshal(hb)
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}
	if _, err := kv.Put(ctx, key, data); err != nil {
		return fmt.Errorf("put heartbeat: %w", err)
	}
	return nil
}

// PutSessionInfo writes session info to KV.
func PutSessionInfo(ctx context.Context, kv jetstream.KeyValue, info protocol.SessionInfo) error {
	key := fmt.Sprintf(protocol.KVSessionTemplate, info.HostID, info.SessionID)
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal session info: %w", err)
	}
	if _, err := kv.Put(ctx, key, data); err != nil {
		return fmt.Errorf("put session info: %w", err)
	}
	return nil
}

// GetHostInfo retrieves host info.
func GetHostInfo(ctx context.Context, kv jetstream.KeyValue, hostID string) (*protocol.HostInfo, error) {
	key := fmt.Sprintf(protocol.KVHostInfoTemplate, hostID)
	entry, err := kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get host info: %w", err)
	}
	var info protocol.HostInfo
	if err := json.Unmarshal(entry.Value(), &info); err != nil {
		return nil, fmt.Errorf("unmarshal host info: %w", err)
	}
	return &info, nil
}
