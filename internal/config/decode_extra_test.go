package config

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestApplyEventsAndRemote(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"events": map[string]any{
			"batch_window":     "1s",
			"batch_max_events": int64(5),
			"batch_max_bytes":  "1KB",
			"batch_idle_flush": "2s",
			"coalesce": map[string]any{
				"io_streams": true,
				"presence":   true,
				"activity":   false,
			},
		},
		"remote": map[string]any{
			"transport":            "nats",
			"buffer_size":          "4096",
			"request_timeout":      "3s",
			"reconnect_max_attempts": int64(2),
			"reconnect_backoff_base": "1s",
			"reconnect_backoff_max":  "5s",
			"nats": map[string]any{
				"url":            "nats://localhost:4222",
				"creds_path":     "~/.creds",
				"subject_prefix": "amux",
				"kv_bucket":      "kv",
			},
		},
	}
	resolver, err := paths.NewResolverOptionalRepo(t.TempDir())
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if err := applyEvents(&cfg, raw); err != nil {
		t.Fatalf("apply events: %v", err)
	}
	if err := applyRemote(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply remote: %v", err)
	}
	if cfg.Events.BatchMaxEvents != 5 || cfg.Remote.Transport != "nats" {
		t.Fatalf("unexpected config values")
	}
}

func TestApplyEventsError(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"events": map[string]any{
			"batch_window": "not-a-duration",
		},
	}
	if err := applyEvents(&cfg, raw); err == nil {
		t.Fatalf("expected events error")
	}
}

func TestApplyRemoteError(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"remote": map[string]any{
			"buffer_size": "not-a-size",
		},
	}
	if err := applyRemote(&cfg, raw, nil); err == nil {
		t.Fatalf("expected remote error")
	}
}
