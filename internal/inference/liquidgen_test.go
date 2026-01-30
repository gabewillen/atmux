package inference

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractVersion(t *testing.T) {
	line := "project(liquidgen VERSION 1.2.3)"
	if version := extractVersion(line); version != "1.2.3" {
		t.Fatalf("unexpected version: %s", version)
	}
	if version := extractVersion("project(liquidgen)"); version != "" {
		t.Fatalf("expected empty version")
	}
}

func TestLiquidgenEngineInferErrors(t *testing.T) {
	root := t.TempDir()
	cmake := filepath.Join(root, "CMakeLists.txt")
	if err := os.WriteFile(cmake, []byte("project(liquidgen VERSION 0.1.0)\n"), 0o644); err != nil {
		t.Fatalf("write cmake: %v", err)
	}
	engine, err := NewLiquidgenEngine(root, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	if _, err := engine.Infer(context.Background(), Request{ModelID: "missing"}); err == nil || !strings.Contains(err.Error(), ErrUnknownModel.Error()) {
		t.Fatalf("expected unknown model error, got %v", err)
	}
	engine.RegisterModel("model", "/tmp/model.bin")
	if _, err := engine.Infer(context.Background(), Request{ModelID: "model"}); err == nil || !strings.Contains(err.Error(), ErrInferenceUnavailable.Error()) {
		t.Fatalf("expected inference unavailable error, got %v", err)
	}
}

func TestLiquidgenEngineInferSuccess(t *testing.T) {
	root := t.TempDir()
	cmake := filepath.Join(root, "CMakeLists.txt")
	if err := os.WriteFile(cmake, []byte("project(liquidgen VERSION 0.2.0)\n"), 0o644); err != nil {
		t.Fatalf("write cmake: %v", err)
	}
	cliPath := filepath.Join(root, "liquidgen_cli")
	script := "#!/bin/sh\necho output\n"
	if err := os.WriteFile(cliPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write cli: %v", err)
	}
	engine, err := NewLiquidgenEngine(root, nil)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	engine.RegisterModel("model", "/tmp/model.bin")
	resp, err := engine.Infer(context.Background(), Request{ModelID: "model", Prompt: "ignored"})
	if err != nil {
		t.Fatalf("infer: %v", err)
	}
	if resp.Output != "output" {
		t.Fatalf("unexpected output: %s", resp.Output)
	}
	models, err := engine.Models(context.Background())
	if err != nil {
		t.Fatalf("models: %v", err)
	}
	if len(models) != 1 || models[0].ID != "model" {
		t.Fatalf("unexpected models: %#v", models)
	}
}

func TestNewDefaultEngine(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "third_party", "liquidgen"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cmake := filepath.Join(root, "third_party", "liquidgen", "CMakeLists.txt")
	if err := os.WriteFile(cmake, []byte("project(liquidgen VERSION 0.3.0)\n"), 0o644); err != nil {
		t.Fatalf("write cmake: %v", err)
	}
	if _, err := NewDefaultEngine(root, nil); err != nil {
		t.Fatalf("new default engine: %v", err)
	}
}
