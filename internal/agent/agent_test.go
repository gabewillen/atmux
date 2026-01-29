package agent

import (
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestNewAgentRequiresDispatcher(t *testing.T) {
	meta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "alpha",
		About:    "desc",
		Adapter:  api.AdapterRef("adapter"),
		RepoRoot: "/repo",
		Worktree: "/repo/.amux/worktrees/alpha",
		Location: api.Location{Type: api.LocationLocal},
	}
	if _, err := NewAgent(meta, nil); err == nil {
		t.Fatalf("expected error for nil dispatcher")
	}
}

func TestNewLifecycleRequiresDispatcher(t *testing.T) {
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}}
	if _, err := NewLifecycle(agent, nil); err == nil {
		t.Fatalf("expected error for nil dispatcher")
	}
}

func TestNewPresenceRequiresDispatcher(t *testing.T) {
	agent := &Agent{Agent: api.Agent{ID: api.NewAgentID()}}
	if _, err := NewPresence(agent, nil); err == nil {
		t.Fatalf("expected error for nil dispatcher")
	}
}
