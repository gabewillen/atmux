package remote

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
)

func TestConfigureEmbeddedHub(t *testing.T) {
	// Mock config dir by setting HOME
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	jsDir := filepath.Join(tmpHome, "nats-data")

	cfg := &config.Config{
		NATS: config.NATSConfig{
			Mode:         "embedded",
			JetStreamDir: jsDir,
			Listen:       "127.0.0.1:4222",
		},
	}

	if err := ConfigureEmbeddedHub(cfg); err != nil {
		t.Fatalf("ConfigureEmbeddedHub failed: %v", err)
	}

	// Verify JS dir
	if _, err := os.Stat(jsDir); os.IsNotExist(err) {
		t.Error("JetStream dir not created")
	}

	// Verify config file
	confDir, _ := paths.DefaultConfigDir()
	confPath := filepath.Join(confDir, "nats-hub.conf")
	content, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("Failed to read nats config: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("NATS config is empty")
	}
}

func TestConfigureLeaf(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "test.creds")
	
	// Should fail if missing
	if err := ConfigureLeaf(nil, credsPath); err == nil {
		t.Error("Expected error for missing creds")
	}
	
	// Create creds
	os.WriteFile(credsPath, []byte("dummy"), 0600)
	
	// Should pass
	if err := ConfigureLeaf(nil, credsPath); err != nil {
		t.Errorf("ConfigureLeaf failed: %v", err)
	}
}