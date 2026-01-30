package manager

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/pkg/api"
)

func TestResolveSenderFallbacks(t *testing.T) {
	mgr := &Manager{agents: map[api.AgentID]*agentState{}}
	missing := api.NewAgentID()
	payload := api.OutboundMessage{AgentID: &missing, ToSlug: "broadcast", Content: "hi"}
	if _, _, ok := mgr.resolveSender(payload); ok {
		t.Fatalf("expected missing sender")
	}
	payload = api.OutboundMessage{From: "not-a-number", ToSlug: "broadcast", Content: "hi"}
	if _, _, ok := mgr.resolveSender(payload); ok {
		t.Fatalf("expected invalid runtime id")
	}
	sender := api.NewAgentID()
	mgr.agents[sender] = &agentState{slug: "alpha"}
	payload = api.OutboundMessage{From: sender.String(), ToSlug: "broadcast", Content: "hi"}
	if _, _, ok := mgr.resolveSender(payload); !ok {
		t.Fatalf("expected sender resolved")
	}
}

func TestBuildAgentMessageUnknownTarget(t *testing.T) {
	mgr := &Manager{}
	_, err := mgr.buildAgentMessage(api.NewAgentID(), api.OutboundMessage{ToSlug: "missing", Content: "hi"})
	if err == nil {
		t.Fatalf("expected target unknown error")
	}
	_, err = mgr.buildAgentMessage(api.NewAgentID(), api.OutboundMessage{ToSlug: "broadcast", Content: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRouteOutboundMessageSkipsEmpty(t *testing.T) {
	mgr := &Manager{}
	mgr.routeOutboundMessage(context.Background(), api.OutboundMessage{ToSlug: "", Content: ""})
}
