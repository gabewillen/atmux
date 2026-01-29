package adapter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestWazeroRegistryMissingAdapter(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	registry, err := NewWazeroRegistry(context.Background(), resolver)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	_, err = registry.Load(context.Background(), "missing")
	if err == nil {
		t.Fatalf("expected error for missing adapter")
	}
	if !errors.Is(err, ErrAdapterNotFound) {
		t.Fatalf("expected ErrAdapterNotFound, got %v", err)
	}
}

func TestWazeroRegistryRejectsMissingExports(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	adapterName := "minimal"
	wasmPath := resolver.ProjectAdapterWasmPath(adapterName)
	if err := os.MkdirAll(filepath.Dir(wasmPath), 0o755); err != nil {
		t.Fatalf("mkdir wasm dir: %v", err)
	}
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if err := os.WriteFile(wasmPath, wasm, 0o644); err != nil {
		t.Fatalf("write wasm: %v", err)
	}
	registry, err := NewWazeroRegistry(context.Background(), resolver)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	_, err = registry.Load(context.Background(), adapterName)
	if err == nil {
		t.Fatalf("expected error for missing exports")
	}
	if !errors.Is(err, ErrAdapterMissingExport) {
		t.Fatalf("expected ErrAdapterMissingExport, got %v", err)
	}
}

func TestNoopAdapterBehaves(t *testing.T) {
	adapter := NewNoopAdapter("noop")
	if adapter.Name() != "noop" {
		t.Fatalf("unexpected name: %s", adapter.Name())
	}
	matcher := adapter.Matcher()
	matches, err := matcher.Match(context.Background(), []byte("output"))
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no matches")
	}
	formatter := adapter.Formatter()
	formatted, err := formatter.Format(context.Background(), "input")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if formatted != "input" {
		t.Fatalf("expected input, got %s", formatted)
	}
}
