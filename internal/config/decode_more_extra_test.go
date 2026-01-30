package config

import (
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestApplyShutdownAndTelemetryErrors(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"shutdown": map[string]any{
			"drain_timeout": "bad",
		},
	}
	if err := applyShutdown(&cfg, raw); err == nil {
		t.Fatalf("expected shutdown error")
	}
	raw = map[string]any{
		"telemetry": map[string]any{
			"metrics": map[string]any{
				"interval": "bad",
			},
		},
	}
	if err := applyTelemetry(&cfg, raw); err == nil {
		t.Fatalf("expected telemetry error")
	}
}

func TestApplyNATSAndPlugins(t *testing.T) {
	resolver, err := paths.NewResolverOptionalRepo(t.TempDir())
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := Config{}
	raw := map[string]any{
		"nats": map[string]any{
			"mode":               "embedded",
			"topology":           "hub",
			"hub_url":            "nats://hub",
			"listen":             "127.0.0.1:4222",
			"leaf_listen":        "127.0.0.1:7422",
			"advertise_url":      "nats://adv",
			"leaf_advertise_url": "nats://leaf",
			"jetstream_dir":      "~/jetstream",
		},
		"daemon": map[string]any{
			"socket_path": "~/amuxd.sock",
			"autostart":   true,
		},
		"plugins": map[string]any{
			"dir":          "~/plugins",
			"allow_remote": true,
		},
	}
	if err := applyNATS(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply nats: %v", err)
	}
	if err := applyDaemon(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply daemon: %v", err)
	}
	if err := applyPlugins(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply plugins: %v", err)
	}
	if cfg.NATS.JetStreamDir == "" || cfg.Daemon.SocketPath == "" || cfg.Plugins.Dir == "" {
		t.Fatalf("expected expanded paths")
	}
}

func TestApplyTelemetrySuccess(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"telemetry": map[string]any{
			"enabled":      true,
			"service_name": "amux",
			"exporter": map[string]any{
				"endpoint": "http://localhost",
				"protocol": "http",
			},
			"traces": map[string]any{
				"enabled":     true,
				"sampler":     "parentbased_traceidratio",
				"sampler_arg": 0.5,
			},
			"metrics": map[string]any{
				"enabled":  true,
				"interval": "5s",
			},
			"logs": map[string]any{
				"enabled": true,
				"level":   "debug",
			},
		},
	}
	if err := applyTelemetry(&cfg, raw); err != nil {
		t.Fatalf("apply telemetry: %v", err)
	}
	if !cfg.Telemetry.Enabled || cfg.Telemetry.Traces.SamplerArg != 0.5 {
		t.Fatalf("unexpected telemetry config")
	}
	if cfg.Telemetry.Metrics.Interval == 0 {
		t.Fatalf("expected metrics interval")
	}
}

func TestValidateAdapterConstraintError(t *testing.T) {
	section := map[string]any{
		"cli": map[string]any{
			"constraint": "not-a-semver",
		},
	}
	if err := validateAdapterConstraint("stub", section); err == nil {
		t.Fatalf("expected adapter constraint error")
	}
}

func TestApplyAdaptersAndAgents(t *testing.T) {
	resolver, err := paths.NewResolverOptionalRepo(t.TempDir())
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := Config{}
	raw := map[string]any{
		"adapters": map[string]any{
			"stub": map[string]any{
				"enabled": true,
			},
		},
		"agents": []any{
			map[string]any{
				"name":    "alpha",
				"adapter": "stub",
				"location": map[string]any{
					"type":      "local",
					"repo_path":  "~/repo",
					"host":       "example",
				},
				"listen_channels": []any{"alerts", "logs"},
			},
		},
	}
	if err := applyAdapters(&cfg, raw); err != nil {
		t.Fatalf("apply adapters: %v", err)
	}
	if err := applyAgents(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply agents: %v", err)
	}
	if len(cfg.Adapters) != 1 || len(cfg.Agents) != 1 {
		t.Fatalf("expected adapters and agents")
	}
	if cfg.Agents[0].ListenChannels == nil {
		t.Fatalf("expected listen channels")
	}
}

func TestApplyShutdownSuccess(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"shutdown": map[string]any{
			"drain_timeout":    "5s",
			"cleanup_worktrees": true,
		},
	}
	if err := applyShutdown(&cfg, raw); err != nil {
		t.Fatalf("apply shutdown: %v", err)
	}
	if cfg.Shutdown.DrainTimeout != 5*time.Second || !cfg.Shutdown.CleanupWorktrees {
		t.Fatalf("unexpected shutdown config")
	}
}
