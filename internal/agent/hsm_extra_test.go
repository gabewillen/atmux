package agent

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/stateforward/hsm-go"
)

func TestLifecycleErrorTransitionEmits(t *testing.T) {
	t.Parallel()
	dispatcher := &recordDispatcher{}
	meta := api.Agent{ID: api.NewAgentID(), Name: "agent", Adapter: "adapter", RepoRoot: "/tmp/repo", Worktree: "/tmp/repo/work", Location: api.Location{Type: api.LocationLocal}}
	agent := &Agent{Agent: meta, dispatcher: dispatcher}
	lifecycle, err := NewLifecycle(agent, dispatcher)
	if err != nil {
		t.Fatalf("new lifecycle: %v", err)
	}
	started := hsm.Started(context.Background(), lifecycle, &LifecycleModel)
	<-hsm.Dispatch(started.Context(), started, hsm.Event{Name: EventError, Data: "boom"})
	if started.State() != "/agent.lifecycle/errored" {
		t.Fatalf("unexpected state: %s", started.State())
	}
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected lifecycle events")
	}
}

