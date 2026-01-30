package manager

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestStatePresence(t *testing.T) {
	if got := statePresence(nil); got == "" {
		t.Fatalf("expected presence for nil state")
	}
	state := &agentState{presence: "  Busy "}
	if got := statePresence(state); got != "busy" {
		t.Fatalf("unexpected presence: %s", got)
	}
	state = &agentState{}
	if got := statePresence(state); got == "" {
		t.Fatalf("expected default presence")
	}
}

func TestMergeAgentErrors(t *testing.T) {
	manager := &Manager{agents: map[api.AgentID]*agentState{}}
	if _, err := manager.MergeAgent(context.Background(), api.NewAgentID(), git.MergeStrategy(""), ""); err == nil {
		t.Fatalf("expected missing agent error")
	}
	agentID := api.NewAgentID()
	manager.agents[agentID] = &agentState{remote: true}
	if _, err := manager.MergeAgent(context.Background(), agentID, git.MergeStrategy(""), ""); err == nil {
		t.Fatalf("expected remote agent error")
	}
	manager.agents[agentID] = &agentState{remote: false, repoRoot: "/missing"}
	manager.cfg = config.Config{}
	if _, err := manager.MergeAgent(context.Background(), agentID, git.MergeStrategy(""), ""); err == nil {
		t.Fatalf("expected base branch error")
	}
}
