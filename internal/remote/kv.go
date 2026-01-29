package remote

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

// KVStore provides access to a JetStream KV bucket.
type KVStore struct {
	kv nats.KeyValue
}

// NewKVStore ensures the KV bucket exists.
func NewKVStore(js nats.JetStreamContext, bucket string) (*KVStore, error) {
	if js == nil {
		return nil, fmt.Errorf("kv store: jetstream unavailable")
	}
	if bucket == "" {
		return nil, fmt.Errorf("kv store: bucket is empty")
	}
	kv, err := js.KeyValue(bucket)
	if err != nil {
		if errors.Is(err, nats.ErrBucketNotFound) {
			kv, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: bucket})
		}
		if err != nil {
			return nil, fmt.Errorf("kv store: %w", err)
		}
	}
	return &KVStore{kv: kv}, nil
}

// Put writes a key-value entry as UTF-8 bytes.
func (k *KVStore) Put(ctx context.Context, key string, value []byte) error {
	if ctx.Err() != nil {
		return fmt.Errorf("kv put: %w", ctx.Err())
	}
	if k == nil || k.kv == nil {
		return fmt.Errorf("kv put: store is nil")
	}
	if _, err := k.kv.Put(key, value); err != nil {
		return fmt.Errorf("kv put: %w", err)
	}
	return nil
}

// Get loads a key value if it exists.
func (k *KVStore) Get(ctx context.Context, key string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("kv get: %w", ctx.Err())
	}
	if k == nil || k.kv == nil {
		return nil, fmt.Errorf("kv get: store is nil")
	}
	entry, err := k.kv.Get(key)
	if err != nil {
		if errors.Is(err, nats.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("kv get: %w", err)
	}
	if entry == nil {
		return nil, nil
	}
	return entry.Value(), nil
}

// ListKeys returns all keys with the given prefix.
func (k *KVStore) ListKeys(ctx context.Context, prefix string) ([]string, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("kv list: %w", ctx.Err())
	}
	if k == nil || k.kv == nil {
		return nil, fmt.Errorf("kv list: store is nil")
	}
	lister, err := k.kv.ListKeys()
	if err != nil {
		return nil, fmt.Errorf("kv list: %w", err)
	}
	defer func() {
		_ = lister.Stop()
	}()
	keys := []string{}
	for key := range lister.Keys() {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	for err := range lister.Error() {
		if err != nil {
			return nil, fmt.Errorf("kv list: %w", err)
		}
	}
	return keys, nil
}
