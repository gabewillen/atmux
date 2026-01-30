package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestAdapterDefaultsFallbackFile(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	adapterDir := filepath.Join(resolver.ProjectAdaptersDir(), "stub")
	if err := os.MkdirAll(adapterDir, 0o755); err != nil {
		t.Fatalf("mkdir adapter: %v", err)
	}
	wasmPath := filepath.Join(adapterDir, "stub.wasm")
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(wasmPath, wasm, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	defaultPath := filepath.Join(adapterDir, "config.default.toml")
	if err := os.WriteFile(defaultPath, []byte("[adapters.stub]\n"), 0o644); err != nil {
		t.Fatalf("write default: %v", err)
	}
	provider := NewDefaultsProvider(resolver)
	defaults, err := provider.AdapterDefaults()
	if err != nil {
		t.Fatalf("adapter defaults: %v", err)
	}
	if len(defaults) != 1 {
		t.Fatalf("expected defaults")
	}
	if defaults[0].Source != defaultPath {
		t.Fatalf("unexpected source: %s", defaults[0].Source)
	}
	if len(defaults[0].Data) == 0 {
		t.Fatalf("expected default data")
	}
}
