package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestIntegrationConfigActorReload(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".amux"), 0o755); err != nil {
		t.Fatalf("mkdir amux dir: %v", err)
	}
	cfgPath := filepath.Join(repo, ".amux", "config.toml")
	if err := os.WriteFile(cfgPath, []byte("[timeouts]\nidle = \"30s\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	actor, err := StartConfigActor(ctx, LoadOptions{
		Resolver:          resolver,
		Env:               map[string]string{},
		WatchPollInterval: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("start actor: %v", err)
	}
	waitStart := time.Now()
	for {
		if actor.Current().Timeouts.Idle == 30*time.Second {
			break
		}
		if time.Since(waitStart) > time.Second {
			t.Fatalf("initial config not loaded")
		}
		time.Sleep(5 * time.Millisecond)
	}
	changes := make(chan ConfigChange, 1)
	actor.Subscribe(func(change ConfigChange) {
		if change.Path == "timeouts.idle" {
			select {
			case changes <- change:
			default:
			}
		}
	})
	if err := os.WriteFile(cfgPath, []byte("[timeouts]\nidle = \"45s\"\n"), 0o644); err != nil {
		t.Fatalf("write updated config: %v", err)
	}
	advance := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(cfgPath, advance, advance); err != nil {
		t.Fatalf("touch config: %v", err)
	}
	select {
	case <-changes:
	case <-time.After(2 * time.Second):
		t.Fatalf("config reload not observed")
	}
	current := actor.Current()
	if current.Timeouts.Idle != 45*time.Second {
		t.Fatalf("expected idle=45s, got %s", current.Timeouts.Idle)
	}
}
