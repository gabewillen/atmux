package remote

import (
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestBootstrapRemote_Simulation(t *testing.T) {
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
