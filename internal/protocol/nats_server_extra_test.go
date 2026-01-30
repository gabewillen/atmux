package protocol

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveLeafListen(t *testing.T) {
	cfg := HubServerConfig{LeafListen: "127.0.0.1:-1"}
	host, port, err := resolveLeafListen(cfg, "127.0.0.1", 4222)
	if err != nil {
		t.Fatalf("resolve leaf listen: %v", err)
	}
	if host == "" || port == 0 {
		t.Fatalf("expected allocated leaf host/port")
	}
}

func TestBuildLeafURLAdvertise(t *testing.T) {
	got := buildLeafURL("127.0.0.1", 7422, "leaf.example.com:7422", false)
	if got != "nats://leaf.example.com:7422" {
		t.Fatalf("unexpected advertise url: %s", got)
	}
	got = buildLeafURL("127.0.0.1", 7422, "tls://leaf.example.com:7422", true)
	if got != "tls://leaf.example.com:7422" {
		t.Fatalf("unexpected advertise url with scheme: %s", got)
	}
}

func TestStartLeafServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	jetDir := filepath.Join(t.TempDir(), "jetstream")
	hub, err := StartHubServer(ctx, HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: jetDir,
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	t.Cleanup(func() { _ = hub.Close() })
	leaf, err := StartLeafServer(ctx, LeafServerConfig{
		Listen: "127.0.0.1:-1",
		HubURL: hub.LeafURL(),
	})
	if err != nil {
		t.Fatalf("start leaf: %v", err)
	}
	t.Cleanup(func() { _ = leaf.Close() })
	if leaf.URL() == "" {
		t.Fatalf("expected leaf url")
	}
}
