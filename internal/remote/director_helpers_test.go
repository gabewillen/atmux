package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHostIDFromLocation(t *testing.T) {
	if _, err := HostIDFromLocation(api.Location{}); err == nil {
		t.Fatalf("expected host id error")
	}
	host, err := HostIDFromLocation(api.Location{Host: "host"})
	if err != nil || host == "" {
		t.Fatalf("unexpected host id")
	}
}

func TestHubURL(t *testing.T) {
	cfg := config.Config{}
	cfg.Remote.NATS.URL = "nats://remote"
	if got := hubURL(cfg); got != "nats://remote" {
		t.Fatalf("unexpected hub url")
	}
	cfg.Remote.NATS.URL = ""
	cfg.Node.Role = "director"
	cfg.NATS.LeafAdvertiseURL = "nats://leaf"
	cfg.NATS.HubURL = "nats://hub"
	if got := hubURL(cfg); got != "nats://leaf" {
		t.Fatalf("unexpected hub url fallback")
	}
}

func TestHostnameFallback(t *testing.T) {
	if hostnameFallback() == "" {
		t.Fatalf("expected hostname fallback")
	}
}

