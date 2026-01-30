package config

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestApplyTimeoutsAndProcess(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"timeouts": map[string]any{
			"idle":  "2s",
			"stuck": "5s",
		},
		"process": map[string]any{
			"capture_mode":      "both",
			"stream_buffer_size": "1024",
			"hook_mode":         "preload",
			"poll_interval":     "1s",
			"hook_socket_dir":   "~/hooks",
		},
	}
	resolver, err := paths.NewResolverOptionalRepo(t.TempDir())
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if err := applyTimeouts(&cfg, raw); err != nil {
		t.Fatalf("apply timeouts: %v", err)
	}
	if err := applyProcess(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply process: %v", err)
	}
	if cfg.Timeouts.Idle == 0 || cfg.Process.StreamBufferSize == 0 {
		t.Fatalf("expected parsed values")
	}
}

func TestApplyTimeoutsError(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"timeouts": map[string]any{
			"idle": "not-a-duration",
		},
	}
	if err := applyTimeouts(&cfg, raw); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestApplyProcessError(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"process": map[string]any{
			"stream_buffer_size": "bad",
		},
	}
	if err := applyProcess(&cfg, raw, nil); err == nil {
		t.Fatalf("expected process error")
	}
}

func TestApplyGit(t *testing.T) {
	cfg := Config{}
	raw := map[string]any{
		"git": map[string]any{
			"merge": map[string]any{
				"strategy":      "squash",
				"allow_dirty":   true,
				"target_branch": "main",
			},
		},
	}
	if err := applyGit(&cfg, raw); err != nil {
		t.Fatalf("apply git: %v", err)
	}
	if cfg.Git.Merge.Strategy != "squash" || !cfg.Git.Merge.AllowDirty {
		t.Fatalf("unexpected git config")
	}
}
