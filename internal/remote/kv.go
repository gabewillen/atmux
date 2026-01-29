package remote

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KVStore provides a simple bucketed key-value store.
type KVStore struct {
	baseDir string
	bucket  string
}

// NewKVStore ensures the KV bucket directory exists.
func NewKVStore(baseDir string, bucket string) (*KVStore, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("kv store: base dir is empty")
	}
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("kv store: bucket is empty")
	}
	bucketDir := filepath.Join(baseDir, "kv", bucket)
	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		return nil, fmt.Errorf("kv store: %w", err)
	}
	return &KVStore{baseDir: baseDir, bucket: bucket}, nil
}

// Put writes a key-value entry as UTF-8 bytes.
func (k *KVStore) Put(ctx context.Context, key string, value []byte) error {
	if ctx.Err() != nil {
		return fmt.Errorf("kv put: %w", ctx.Err())
	}
	if k == nil {
		return fmt.Errorf("kv put: store is nil")
	}
	path, err := k.keyPath(key)
	if err != nil {
		return fmt.Errorf("kv put: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("kv put: %w", err)
	}
	if err := os.WriteFile(path, value, 0o644); err != nil {
		return fmt.Errorf("kv put: %w", err)
	}
	return nil
}

// Get loads a key value if it exists.
func (k *KVStore) Get(ctx context.Context, key string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, fmt.Errorf("kv get: %w", ctx.Err())
	}
	if k == nil {
		return nil, fmt.Errorf("kv get: store is nil")
	}
	path, err := k.keyPath(key)
	if err != nil {
		return nil, fmt.Errorf("kv get: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("kv get: %w", err)
	}
	return data, nil
}

func (k *KVStore) keyPath(key string) (string, error) {
	clean := strings.TrimSpace(key)
	if clean == "" {
		return "", fmt.Errorf("kv key: empty")
	}
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("kv key: invalid")
	}
	bucketDir := filepath.Join(k.baseDir, "kv", k.bucket)
	return filepath.Join(bucketDir, filepath.FromSlash(clean)), nil
}
