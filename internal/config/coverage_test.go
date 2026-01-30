package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

type stubDefaults struct {
	items []AdapterDefault
}

func (s stubDefaults) AdapterDefaults() ([]AdapterDefault, error) {
	return s.items, nil
}

func TestParseTOMLComplex(t *testing.T) {
	doc := `
title = "example"
[timeouts]
idle = "10s"
[[agents]]
name = "alpha"
[[agents]]
name = "beta"
[adapters.stub]
patterns = { prompt = "ready", message = "MSG:" }
`
	parsed, err := ParseTOML([]byte(doc))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed["title"] != "example" {
		t.Fatalf("unexpected title")
	}
}

func TestEnvOverridesAndMerge(t *testing.T) {
	env := map[string]string{
		"AMUX__GENERAL__LOG_LEVEL": "debug",
		"AMUX__ADAPTERS__STUB__ENABLED": "true",
	}
	overrides, err := EnvOverrides(env)
	if err != nil {
		t.Fatalf("env overrides: %v", err)
	}
	base := map[string]any{"general": map[string]any{"log_level": "info"}}
	if err := MergeMaps(base, overrides); err != nil {
		t.Fatalf("merge: %v", err)
	}
}

func TestLoadAndActor(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	configPath := resolver.ProjectConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("timeouts = { idle = \"5s\" }\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	defaults := stubDefaults{items: []AdapterDefault{{
		Name:   "stub",
		Source: "test",
		Data:   []byte("[adapters.stub]\npatterns = { prompt = \"ready\" }\n"),
	}}}
	cfg, err := Load(LoadOptions{Resolver: resolver, AdapterDefaults: defaults, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Timeouts.Idle <= 0 {
		t.Fatalf("expected idle timeout")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	actor, err := StartConfigActor(ctx, LoadOptions{Resolver: resolver, AdapterDefaults: defaults, Env: map[string]string{}, WatchPollInterval: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("start actor: %v", err)
	}
	changes := make(chan ConfigChange, 1)
	actor.Subscribe(func(change ConfigChange) {
		select {
		case changes <- change:
		default:
		}
	})
	timer := time.NewTimer(1100 * time.Millisecond)
	<-timer.C
	if err := os.WriteFile(configPath, []byte("timeouts = { idle = \"6s\" }\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	select {
	case <-changes:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected change")
	}
}

func TestConfigFileHelpers(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.toml")
	if err := WriteConfigFile(path, map[string]any{"general": map[string]any{"log_level": "debug"}}); err != nil {
		t.Fatalf("write config: %v", err)
	}
	data, err := LoadConfigFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if data == nil {
		t.Fatalf("expected data")
	}
	if _, err := LoadConfigFile(""); err != nil {
		t.Fatalf("load empty path: %v", err)
	}
}

func TestValueParsers(t *testing.T) {
	if _, err := ParseByteSize("bad"); err == nil {
		t.Fatalf("expected byte size error")
	}
	if size, err := ParseByteSize("10MB"); err != nil || size == 0 {
		t.Fatalf("unexpected size: %v %v", size, err)
	}
	if _, err := ParseByteSizeValue(1.5); err != nil {
		t.Fatalf("parse byte size value: %v", err)
	}
	if _, err := parseDurationValue("bad"); err == nil {
		t.Fatalf("expected duration error")
	}
}

func TestSensitiveKeyHelpers(t *testing.T) {
	node := map[string]any{"token": "secret", "nested": map[string]any{"api_key": "value"}}
	keys := FindSensitiveKeys(node, "")
	if len(keys) == 0 {
		t.Fatalf("expected sensitive keys")
	}
	redacted := RedactSensitive(node)
	if redacted["token"] != "[redacted]" {
		t.Fatalf("expected redaction")
	}
}

func TestResolveConfigPath(t *testing.T) {
	root := t.TempDir()
	resolved, err := ResolveConfigPath(root, "config.toml")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.HasPrefix(resolved, root) {
		t.Fatalf("unexpected path: %s", resolved)
	}
}

func TestDiffConfig(t *testing.T) {
	oldCfg := DefaultConfig(nil)
	newCfg := oldCfg
	newCfg.Timeouts.Idle = 2 * time.Second
	newCfg.Telemetry.Enabled = true
	changes := DiffConfig(oldCfg, newCfg)
	if len(changes) == 0 {
		t.Fatalf("expected changes")
	}
}
