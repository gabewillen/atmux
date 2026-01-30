package manager

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestShutdownForce(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	manager := &Manager{
		cfg:        config.DefaultConfig(resolver),
		agents:     map[api.AgentID]*agentState{},
		resolver:   resolver,
		registries: map[string]adapter.Registry{},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	if err := manager.Shutdown(ctx, true); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}
