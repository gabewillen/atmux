package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFileErrors(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.toml")
	if err := os.WriteFile(path, []byte("key = [1 2]"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadConfigFile(path); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestWriteConfigFileErrors(t *testing.T) {
	if err := WriteConfigFile("", map[string]any{}); err == nil {
		t.Fatalf("expected empty path error")
	}
	if err := WriteConfigFile(filepath.Join(t.TempDir(), "config.toml"), map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatalf("expected encode error")
	}
	dirFile := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(dirFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	target := filepath.Join(dirFile, "config.toml")
	if err := WriteConfigFile(target, map[string]any{"ok": "v"}); err == nil {
		t.Fatalf("expected mkdir error")
	}
}
