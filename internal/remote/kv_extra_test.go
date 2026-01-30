package remote

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
)

func TestKVStoreGetAndListKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	t.Cleanup(func() {
		server.Shutdown()
	})
	dispatcher, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() {
		_ = dispatcher.Close(context.Background())
	})
	kv, err := NewKVStore(dispatcher.JetStream(), "kv")
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	if err := kv.Put(ctx, "hosts/a/info", []byte("a")); err != nil {
		t.Fatalf("put: %v", err)
	}
	if _, err := kv.Get(ctx, "missing"); err != nil {
		t.Fatalf("get missing: %v", err)
	}
	data, err := kv.Get(ctx, "hosts/a/info")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(data) != "a" {
		t.Fatalf("unexpected data: %s", string(data))
	}
	keys, err := kv.ListKeys(ctx, "hosts/")
	if err != nil {
		t.Fatalf("list keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatalf("expected keys")
	}
}
