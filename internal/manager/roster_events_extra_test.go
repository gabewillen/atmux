package manager

import (
	"context"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestEmitRosterUpdated(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{config: config.AgentConfig{Name: "alpha", Adapter: "stub"}}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	mgr.emitRosterUpdated(context.Background())
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected roster events")
	}
}

func TestHandlePresenceEvent(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{presence: agent.PresenceOnline}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	payload := agent.PresenceEvent{AgentID: id, Presence: agent.PresenceBusy}
	event := protocol.Event{Name: agent.EventPresenceChanged, Payload: payload}
	mgr.handlePresenceEvent(event)
	if state.presence != agent.PresenceBusy {
		t.Fatalf("expected presence update")
	}
}
