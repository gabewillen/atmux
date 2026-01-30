package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestFindWasmPath(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if _, err := FindWasmPath(resolver, ""); err == nil {
		t.Fatalf("expected invalid name error")
	}
	if _, err := FindWasmPath(resolver, "missing"); err == nil {
		t.Fatalf("expected not found error")
	}
	adapterDir := resolver.ProjectAdaptersDir()
	if err := os.MkdirAll(filepath.Join(adapterDir, "stub"), 0o755); err != nil {
		t.Fatalf("mkdir adapter: %v", err)
	}
	wasmPath := filepath.Join(adapterDir, "stub", "stub.wasm")
	if err := os.WriteFile(wasmPath, []byte("not-wasm"), 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	found, err := FindWasmPath(resolver, "stub")
	if err != nil {
		t.Fatalf("find wasm: %v", err)
	}
	if found != wasmPath {
		t.Fatalf("unexpected path: %s", found)
	}
}

func TestDefaultsProviderErrors(t *testing.T) {
	provider := NewDefaultsProvider(nil)
	if _, err := provider.AdapterDefaults(); err == nil {
		t.Fatalf("expected resolver error")
	}
	if _, err := discoverAdapterModules(nil); err == nil {
		t.Fatalf("expected discover error")
	}
}

func TestDiscoverAdapterModules(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	project := resolver.ProjectAdaptersDir()
	if err := os.MkdirAll(filepath.Join(project, "stub"), 0o755); err != nil {
		t.Fatalf("mkdir adapter: %v", err)
	}
	wasmPath := filepath.Join(project, "stub", "stub.wasm")
	if err := os.WriteFile(wasmPath, []byte("not-wasm"), 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	modules, err := discoverAdapterModules(resolver)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected module")
	}
}

func TestLoadAdapterDefaultsErrors(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	project := resolver.ProjectAdaptersDir()
	if err := os.MkdirAll(filepath.Join(project, "stub"), 0o755); err != nil {
		t.Fatalf("mkdir adapter: %v", err)
	}
	wasmPath := filepath.Join(project, "stub", "stub.wasm")
	if err := os.WriteFile(wasmPath, []byte("not-wasm"), 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	provider := NewDefaultsProvider(resolver)
	if _, err := provider.AdapterDefaults(); err == nil {
		t.Fatalf("expected adapter defaults error")
	}
	if _, _, err := loadAdapterDefaults(context.Background(), nil, adapterModule{name: "stub", path: wasmPath}); err == nil {
		t.Fatalf("expected runtime error")
	}
	if _, err := callConfigDefault(context.Background(), nil, nil, nil); err == nil {
		t.Fatalf("expected config default error")
	}
	if _, ok := firstResult(nil); ok {
		t.Fatalf("expected no result")
	}
}
