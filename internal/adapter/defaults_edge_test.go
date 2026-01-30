package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestDiscoverAdapterModulesDedupAndIgnoreFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	projectDir := resolver.ProjectAdaptersDir()
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "notadir"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	alphaDir := filepath.Join(projectDir, "alpha")
	if err := os.MkdirAll(alphaDir, 0o755); err != nil {
		t.Fatalf("mkdir alpha: %v", err)
	}
	alphaWasm := filepath.Join(alphaDir, "alpha.wasm")
	if err := os.WriteFile(alphaWasm, []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	userDir := resolver.UserAdaptersDir()
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatalf("mkdir user: %v", err)
	}
	alphaUser := filepath.Join(userDir, "alpha")
	if err := os.MkdirAll(alphaUser, 0o755); err != nil {
		t.Fatalf("mkdir alpha user: %v", err)
	}
	if err := os.WriteFile(filepath.Join(alphaUser, "alpha.wasm"), []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	betaDir := filepath.Join(userDir, "beta")
	if err := os.MkdirAll(betaDir, 0o755); err != nil {
		t.Fatalf("mkdir beta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(betaDir, "beta.wasm"), []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	modules, err := discoverAdapterModules(resolver)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("expected two modules, got %d", len(modules))
	}
	var alphaSource string
	var betaFound bool
	for _, mod := range modules {
		if mod.name == "alpha" {
			alphaSource = mod.source
		}
		if mod.name == "beta" {
			betaFound = true
		}
	}
	if alphaSource != "project" || !betaFound {
		t.Fatalf("unexpected modules: %#v", modules)
	}
}

func TestAdapterDefaultsRejectsInvalidUTF8(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
	if err := os.WriteFile(filepath.Join(adapterDir, "stub.wasm"), []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	if err := os.WriteFile(filepath.Join(adapterDir, "config.default.toml"), []byte{0xff, 0xfe}, 0o644); err != nil {
		t.Fatalf("write default: %v", err)
	}
	provider := NewDefaultsProvider(resolver)
	if _, err := provider.AdapterDefaults(); err == nil {
		t.Fatalf("expected utf-8 error")
	}
}

func TestScanAdapterDirReadDirError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notadir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := scanAdapterDir(file, "test", map[string]struct{}{}); err == nil {
		t.Fatalf("expected readdir error")
	}
}
