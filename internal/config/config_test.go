// Package config implements tests for the configuration subsystem
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadConfig tests loading configuration from file and environment
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "amux.toml")
	
	// Write a sample config file
	configContent := `
[core]
repo_root = "/tmp/test-repo"
debug = true

[server]
socket_path = "/tmp/amux.sock"
rpc_timeout = "60s"

[logging]
level = "debug"
format = "text"
file = "/tmp/amux.log"

[telemetry]
enabled = true
endpoint = "http://localhost:4318"
service_name = "test-service"

[remote]
enabled = true
nats_url = "nats://test:4222"
creds_path = "/tmp/creds.jwt"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test loading config from file
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values from file
	if config.Core.RepoRoot != "/tmp/test-repo" {
		t.Errorf("Expected repo_root '/tmp/test-repo', got '%s'", config.Core.RepoRoot)
	}
	if !config.Core.Debug {
		t.Error("Expected debug to be true")
	}
	if config.Server.SocketPath != "/tmp/amux.sock" {
		t.Errorf("Expected socket_path '/tmp/amux.sock', got '%s'", config.Server.SocketPath)
	}
	if config.Server.RPCTimeout != 60*time.Second {
		t.Errorf("Expected rpc_timeout 60s, got %v", config.Server.RPCTimeout)
	}
	if config.Logging.Level != "debug" {
		t.Errorf("Expected logging level 'debug', got '%s'", config.Logging.Level)
	}
	if config.Logging.Format != "text" {
		t.Errorf("Expected logging format 'text', got '%s'", config.Logging.Format)
	}
	if config.Logging.File != "/tmp/amux.log" {
		t.Errorf("Expected logging file '/tmp/amux.log', got '%s'", config.Logging.File)
	}
	if !config.Telemetry.Enabled {
		t.Error("Expected telemetry to be enabled")
	}
	if config.Telemetry.Endpoint != "http://localhost:4318" {
		t.Errorf("Expected telemetry endpoint 'http://localhost:4318', got '%s'", config.Telemetry.Endpoint)
	}
	if config.Telemetry.ServiceName != "test-service" {
		t.Errorf("Expected telemetry service name 'test-service', got '%s'", config.Telemetry.ServiceName)
	}
	if !config.Remote.Enabled {
		t.Error("Expected remote to be enabled")
	}
	if config.Remote.NATSURL != "nats://test:4222" {
		t.Errorf("Expected remote NATS URL 'nats://test:4222', got '%s'", config.Remote.NATSURL)
	}
	if config.Remote.CredsPath != "/tmp/creds.jwt" {
		t.Errorf("Expected remote creds path '/tmp/creds.jwt', got '%s'", config.Remote.CredsPath)
	}
}

// TestConfigWithEnvOverrides tests environment variable overrides
func TestConfigWithEnvOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("AMUX_CORE_REPO_ROOT", "/env/test")
	os.Setenv("AMUX_CORE_DEBUG", "true")
	os.Setenv("AMUX_SERVER_SOCKET_PATH", "/tmp/env.sock")
	os.Setenv("AMUX_SERVER_RPC_TIMEOUT", "90s")
	defer func() {
		// Clean up environment variables
		os.Unsetenv("AMUX_CORE_REPO_ROOT")
		os.Unsetenv("AMUX_CORE_DEBUG")
		os.Unsetenv("AMUX_SERVER_SOCKET_PATH")
		os.Unsetenv("AMUX_SERVER_RPC_TIMEOUT")
	}()

	// Load config with environment overrides
	config, err := LoadConfig("/nonexistent.toml") // Non-existent file to test defaults + env
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment overrides
	if config.Core.RepoRoot != "/env/test" {
		t.Errorf("Expected repo_root '/env/test' from env, got '%s'", config.Core.RepoRoot)
	}
	if !config.Core.Debug {
		t.Error("Expected debug to be true from env")
	}
	if config.Server.SocketPath != "/tmp/env.sock" {
		t.Errorf("Expected socket_path '/tmp/env.sock' from env, got '%s'", config.Server.SocketPath)
	}
	if config.Server.RPCTimeout != 90*time.Second {
		t.Errorf("Expected rpc_timeout 90s from env, got %v", config.Server.RPCTimeout)
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	config := &Config{}
	config.Core.RepoRoot = ""
	
	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for empty repo_root")
	}
	
	config.Core.RepoRoot = "/valid/path"
	err = config.Validate()
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

// TestExpandHomeDir tests expanding the home directory
func TestExpandHomeDir(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, "test/path")
	
	result := expandHomeDir("~/test/path")
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test with non-home path
	result = expandHomeDir("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("Expected '/absolute/path', got '%s'", result)
	}
}