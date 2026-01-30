package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleOutboundMessage(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{slug: "alpha"}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	payload := api.OutboundMessage{AgentID: &id, ToSlug: "alpha", Content: "hi"}
	event := protocol.Event{Name: "message.outbound", Payload: payload}
	mgr.handleOutboundMessage(event)
}

func TestShutdownForcePath(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{session: &session.LocalSession{}, runtime: nil}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	controller := mgr.ensureShutdownController()
	controller.signal(context.Background(), shutdownEventRequest, map[string]any{})
	controller.signal(context.Background(), shutdownEventForce, map[string]any{})
	if err := controller.wait(context.Background()); err != nil {
		t.Fatalf("shutdown wait: %v", err)
	}
	controller.recordError(errors.New("boom"))
}
