package manager

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestUpdateRemotePresenceEmits(t *testing.T) {
	dispatcher := &recordDispatcher{}
	hostID := api.MustParseHostID("host")
	id := api.NewAgentID()
	state := &agentState{remote: true, remoteHost: hostID, presence: agent.PresenceAway}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	mgr.updateRemotePresence(context.Background(), hostID, agent.PresenceOnline, false)
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected presence events")
	}
}

func TestHandleRemoteEventConnectionLost(t *testing.T) {
	dispatcher := &recordDispatcher{}
	hostID := api.MustParseHostID("host")
	id := api.NewAgentID()
	state := &agentState{remote: true, remoteHost: hostID, presence: agent.PresenceOnline}
	mgr := &Manager{
		dispatcher: dispatcher,
		cfg:        config.Config{},
		agents:     map[api.AgentID]*agentState{id: state},
	}
	msg, err := remote.EncodeEventMessageJSON(remote.EventMessage{
		Type: remote.MsgBroadcast,
		Event: remote.WireEvent{
			Name: "connection.lost",
			Data: json.RawMessage(`{"peer_id":"1","timestamp":"2020-01-01T00:00:00Z"}`),
		},
	})
	if err != nil {
		t.Fatalf("encode event: %v", err)
	}
	subject := remote.EventsSubject("", hostID)
	mgr.handleRemoteEvent(protocol.Message{Subject: subject, Data: msg})
	if state.presence != agent.PresenceAway {
		t.Fatalf("expected presence away")
	}
}
