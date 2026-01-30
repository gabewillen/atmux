package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestDeriveHubURLFromLeaf(t *testing.T) {
	if _, err := deriveHubURLFromLeaf("bad"); err == nil {
		t.Fatalf("expected parse error")
	}
	if _, err := deriveHubURLFromLeaf("nats://host"); err == nil {
		t.Fatalf("expected missing port error")
	}
	got, err := deriveHubURLFromLeaf("nats://127.0.0.1:7422")
	if err != nil {
		t.Fatalf("derive: %v", err)
	}
	if got == "" {
		t.Fatalf("expected hub url")
	}
}

func TestHubClientURL(t *testing.T) {
	cfg := config.Config{NATS: config.NATSConfig{HubURL: "nats://hub:4222"}}
	if got, err := hubClientURL(cfg); err != nil || got != "nats://hub:4222" {
		t.Fatalf("expected hub url")
	}
	cfg = config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{URL: "nats://leaf:7422"}}}
	if got, err := hubClientURL(cfg); err != nil || got == "" {
		t.Fatalf("expected derived hub url")
	}
}
