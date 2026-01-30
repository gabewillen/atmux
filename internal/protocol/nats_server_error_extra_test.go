package protocol

import (
	"context"
	"testing"
	"time"
)

func TestStartHubServerErrorsExtra(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	if _, err := StartHubServer(ctx, HubServerConfig{Listen: "bad", JetStreamDir: "/tmp"}); err == nil {
		t.Fatalf("expected listen error")
	}
	if _, err := StartHubServer(ctx, HubServerConfig{Listen: "127.0.0.1:-1"}); err == nil {
		t.Fatalf("expected jetstream dir error")
	}
}

func TestStartLeafServerErrorsExtra(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	if _, err := StartLeafServer(ctx, LeafServerConfig{Listen: "bad", HubURL: "nats://host:4222"}); err == nil {
		t.Fatalf("expected listen error")
	}
	if _, err := StartLeafServer(ctx, LeafServerConfig{Listen: "127.0.0.1:-1", HubURL: "://bad"}); err == nil {
		t.Fatalf("expected hub url parse error")
	}
}
