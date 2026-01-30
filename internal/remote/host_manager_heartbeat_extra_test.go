package remote

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestWriteHeartbeatErrors(t *testing.T) {
	manager := &HostManager{hostID: api.MustParseHostID("host")}
	if err := manager.writeHeartbeat(context.Background()); err == nil {
		t.Fatalf("expected kv unavailable error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	manager.kv = &KVStore{}
	if err := manager.writeHeartbeat(ctx); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestHeartbeatLoopWritesKV(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	dispatcher, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() { _ = dispatcher.Close(context.Background()) })
	kv, err := NewKVStore(dispatcher.JetStream(), "kv")
	if err != nil {
		t.Fatalf("kv store: %v", err)
	}
	manager := &HostManager{
		cfg: config.Config{
			Remote: config.RemoteConfig{
				NATS: config.RemoteNATSConfig{
					HeartbeatInterval: 10 * time.Millisecond,
				},
			},
		},
		kv:     kv,
		hostID: api.MustParseHostID("host"),
		ready:  true,
	}
	hbCtx, hbCancel := context.WithCancel(context.Background())
	go manager.heartbeatLoop(hbCtx)
	defer hbCancel()
	var data []byte
	deadline := time.NewTimer(500 * time.Millisecond)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
waitLoop:
	for len(data) == 0 {
		select {
		case <-ticker.C:
			data, _ = kv.Get(context.Background(), "hosts/host/heartbeat")
		case <-deadline.C:
			break waitLoop
		}
	}
	if len(data) == 0 {
		t.Fatalf("expected heartbeat entry")
	}
}
