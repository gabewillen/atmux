package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestApplyAgents(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	raw := map[string]any{
		"agents": []any{
			map[string]any{
				"name":    "alpha",
				"about":   "test",
				"adapter": "stub",
				"listen_channels": []any{
					"chan-a",
					"chan-b",
				},
				"location": map[string]any{
					"type":      "ssh",
					"host":      "example.com",
					"repo_path": "~/repo",
				},
			},
		},
	}
	cfg := Config{}
	if err := applyAgents(&cfg, raw, resolver); err != nil {
		t.Fatalf("apply agents: %v", err)
	}
	if len(cfg.Agents) != 1 {
		t.Fatalf("expected 1 agent")
	}
	agent := cfg.Agents[0]
	if agent.Name != "alpha" || agent.Adapter != "stub" || agent.Location.Host != "example.com" {
		t.Fatalf("unexpected agent: %#v", agent)
	}
	if agent.Location.RepoPath == "" || agent.Location.RepoPath == "~/repo" {
		t.Fatalf("expected expanded repo path")
	}
}
