package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Set up a clean environment
	os.Clearenv()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.LogLevel != "info" {
		t.Errorf("expected default log_level info, got %s", cfg.General.LogLevel)
	}
	if cfg.Timeouts.Idle != "30s" {
		t.Errorf("expected default idle timeout 30s, got %s", cfg.Timeouts.Idle)
	}
}

func TestLoad_File(t *testing.T) {
	os.Clearenv()
	tmpDir := t.TempDir()
	
	// Create .amux/config.toml
	amuxDir := filepath.Join(tmpDir, ".amux")
	if err := os.MkdirAll(amuxDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	configContent := `
[general]
log_level = "debug"

[timeouts]
idle = "1m"
`
	if err := os.WriteFile(filepath.Join(amuxDir, "config.toml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected log_level debug, got %s", cfg.General.LogLevel)
	}
	if cfg.Timeouts.Idle != "1m" {
		t.Errorf("expected idle timeout 1m, got %s", cfg.Timeouts.Idle)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Clearenv()
	tmpDir := t.TempDir()

	// Set env vars
	t.Setenv("AMUX__GENERAL__LOG_LEVEL", "warn")
	t.Setenv("AMUX__TIMEOUTS__IDLE", "10m")
	// Test adapter special case
	t.Setenv("AMUX__ADAPTERS__CLAUDE_CODE__MODEL", "claude-3-opus")

	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.LogLevel != "warn" {
		t.Errorf("expected log_level warn, got %s", cfg.General.LogLevel)
	}
	if cfg.Timeouts.Idle != "10m" {
		t.Errorf("expected idle timeout 10m, got %s", cfg.Timeouts.Idle)
	}

	// Check adapter config
	claudeCfg, ok := cfg.Adapters["claude-code"]
	if !ok {
		t.Fatal("expected claude-code adapter config")
	}
	if val, ok := claudeCfg["model"]; !ok || val != "claude-3-opus" {
		t.Errorf("expected claude-code model claude-3-opus, got %v", val)
	}
}
