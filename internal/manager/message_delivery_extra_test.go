package manager

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestHandleCommMessageDeliversToTarget(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{
		slug:      "alpha",
		session:   &session.LocalSession{},
		formatter: stubFormatter{},
	}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	payload := api.AgentMessage{
		ID:        api.NewRuntimeID(),
		From:      api.NewRuntimeID(),
		To:        api.TargetIDFromRuntime(id.RuntimeID),
		Content:   "hello",
		Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleCommMessage(protocol.Message{Subject: "amux.comm.agent.host." + id.String(), Data: data})
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected message event")
	}
}

func TestHandleCommMessageBroadcastDeliver(t *testing.T) {
	dispatcher := &recordDispatcher{}
	id := api.NewAgentID()
	state := &agentState{session: &session.LocalSession{}}
	mgr := &Manager{
		dispatcher: dispatcher,
		agents:     map[api.AgentID]*agentState{id: state},
	}
	payload := api.AgentMessage{
		ID:        api.NewRuntimeID(),
		From:      api.NewRuntimeID(),
		To:        api.TargetID{},
		Content:   "broadcast",
		Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleCommMessage(protocol.Message{Subject: "amux.comm.broadcast", Data: data})
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected broadcast event")
	}
}

type errorFormatter struct{}

func (errorFormatter) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	_ = input
	return "", errors.New("format failed")
}

func TestMirrorMessageToStateFormatterError(t *testing.T) {
	state := &agentState{
		session:   &session.LocalSession{},
		formatter: errorFormatter{},
	}
	mgr := &Manager{}
	mgr.mirrorMessageToState("subject", api.AgentMessage{Content: "hello"}, state)
}
