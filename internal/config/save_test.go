package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestSaveProjectConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "amux-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := DefaultConfig()
	cfg.General.LogLevel = "debug"

	// Save to temp project dir
	if err := SaveProjectConfig(&cfg, tempDir); err != nil {
		t.Fatalf("SaveProjectConfig failed: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(tempDir, ".amux", "config.toml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Config file not created at %s", expectedPath)
	}

	// Load it back
	loadedCfg, err := Load(tempDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedCfg.General.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got %q", loadedCfg.General.LogLevel)
	}
}

func TestSaveUserConfig_Integration(t *testing.T) {
	// Mock user config dir via internal/paths or by setting HOME env var for this test?
	// internal/paths uses os.UserHomeDir().
	// We can skip this or mock it if we really need to.
	// For now, let's rely on SaveProjectConfig test for the core logic (saveToFile).
	// If we really want to test SaveUserConfig, we'd need to mock paths.DefaultConfigDir
	// or change HOME.
	
	origHome, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home dir")
	}
	
	// Create a temp home
	tempHome, err := os.MkdirTemp("", "amux-home-test")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)
	
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	cfg := DefaultConfig()
	cfg.General.LogFormat = "json"

	if err := SaveUserConfig(&cfg); err != nil {
		t.Fatalf("SaveUserConfig failed: %v", err)
	}
	
	expectedDir, _ := paths.DefaultConfigDir()
	expectedPath := filepath.Join(expectedDir, "config.toml")
	
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("User config file not created at %s", expectedPath)
	}
}
