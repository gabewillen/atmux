package manager

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestShutdownTargets(t *testing.T) {
	manager := &Manager{
		agents: map[api.AgentID]*agentState{
			api.NewAgentID(): {remote: true},
			api.NewAgentID(): nil,
			api.NewAgentID(): {remote: false, repoRoot: "/tmp", slug: "alpha"},
		},
	}
	targets := manager.shutdownTargets()
	if len(targets) != 1 {
		t.Fatalf("expected 1 shutdown target, got %d", len(targets))
	}
}
