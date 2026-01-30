package remote

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type recordDispatcher struct {
	events []protocol.Event
}

func (r *recordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	r.events = append(r.events, event)
	return nil
}
func (r *recordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *recordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	return nil
}
func (r *recordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *recordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (r *recordDispatcher) MaxPayload() int { return 0 }
func (r *recordDispatcher) JetStream() nats.JetStreamContext {
	return nil
}
func (r *recordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestPresenceTransitionEvents(t *testing.T) {
	if presenceTransitionEvents("online", "online") != nil {
		t.Fatalf("expected nil transitions")
	}
	events := presenceTransitionEvents("offline", "busy")
	if len(events) == 0 {
		t.Fatalf("expected transitions")
	}
}

func TestHandleActionUpdatePresence(t *testing.T) {
	manager := &HostManager{}
	session := &remoteSession{agentRuntime: &agent.Agent{Agent: api.Agent{ID: api.NewAgentID()}}}
	payload, _ := json.Marshal(actionUpdatePresence{Presence: "busy"})
	manager.handleActionUpdatePresence(context.Background(), session, payload)
}

func TestHandleActionEmitEvent(t *testing.T) {
	dispatcher := &recordDispatcher{}
	manager := &HostManager{dispatcher: dispatcher}
	payload, _ := json.Marshal(actionEmitEvent{Event: adapter.Event{Type: "test"}})
	manager.handleActionEmitEvent(context.Background(), payload)
	if len(dispatcher.events) != 1 {
		t.Fatalf("expected event")
	}
	manager.handleActionEmitEvent(context.Background(), json.RawMessage("{"))
}
