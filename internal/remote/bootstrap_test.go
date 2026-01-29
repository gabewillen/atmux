package remote

import (
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestBootstrapRemote_Simulation(t *testing.T) {
	// Setup temp home for keys
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	// Generate keys via ConfigureEmbeddedHub
	dummyCfg := config.DefaultConfig()
	dummyCfg.NATS.Mode = "embedded"
	if err := ConfigureEmbeddedHub(&dummyCfg); err != nil {
		t.Fatalf("Failed to setup NATS keys: %v", err)
	}

	// Skip actual SSH
	t.Setenv("AMUX_TEST_SKIP_SSH", "1")

	cfg := config.AgentConfig{
		Location: config.LocationConfig{
			Host: "example.com",
			User: "testuser",
		},
	}
	hostID := api.HostID("test-host")

	err := BootstrapRemote(cfg, hostID)
	if err != nil {
		t.Errorf("BootstrapRemote failed: %v", err)
	}
}
