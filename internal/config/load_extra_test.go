package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type errorDefaults struct{}

func (errorDefaults) AdapterDefaults() ([]AdapterDefault, error) {
	return nil, errors.New("boom")
}

type invalidDefaults struct{}

func (invalidDefaults) AdapterDefaults() ([]AdapterDefault, error) {
	return []AdapterDefault{{Name: "stub", Source: "test", Data: []byte("key = 1")}}, nil
}

func TestMergeFileErrors(t *testing.T) {
	dir := t.TempDir()
	if err := mergeFile(map[string]any{}, dir, log.New(os.Stderr, "test ", 0)); err == nil {
		t.Fatalf("expected directory error")
	}
	path := filepath.Join(t.TempDir(), "bad.toml")
	if err := os.WriteFile(path, []byte("["), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mergeFile(map[string]any{}, path, log.New(os.Stderr, "test ", 0)); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestMergeAdapterDefaultsErrors(t *testing.T) {
	if err := mergeAdapterDefaults(map[string]any{}, errorDefaults{}); err == nil {
		t.Fatalf("expected adapter defaults error")
	}
	if err := mergeAdapterDefaults(map[string]any{}, invalidDefaults{}); err == nil {
		t.Fatalf("expected invalid defaults error")
	}
}
