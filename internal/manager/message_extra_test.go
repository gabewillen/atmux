package manager

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type rawRecordDispatcher struct {
	mu        sync.Mutex
	rawCalls  []protocol.Message
	events    []protocol.Event
}

func (r *rawRecordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
	return nil
}

func (r *rawRecordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *rawRecordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	r.mu.Lock()
	r.rawCalls = append(r.rawCalls, protocol.Message{Subject: subject, Reply: reply, Data: append([]byte(nil), payload...)})
	r.mu.Unlock()
	return nil
}

func (r *rawRecordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *rawRecordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (r *rawRecordDispatcher) MaxPayload() int {
	return 1024 * 1024
}

func (r *rawRecordDispatcher) JetStream() nats.JetStreamContext {
	return nil
}

func (r *rawRecordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestResolveSenderAndTarget(t *testing.T) {
	t.Parallel()
	id := api.NewAgentID()
	state := &agentState{slug: "alpha"}
	mgr := &Manager{agents: map[api.AgentID]*agentState{id: state}}
	payload := api.OutboundMessage{AgentID: &id}
	gotID, gotState, ok := mgr.resolveSender(payload)
	if !ok || gotState != state || gotID != id {
		t.Fatalf("resolve sender by id failed")
	}
	payload = api.OutboundMessage{From: id.RuntimeID.String()}
	gotID, gotState, ok = mgr.resolveSender(payload)
	if !ok || gotState != state || gotID != id {
		t.Fatalf("resolve sender by runtime id failed")
	}
	target, ok := mgr.resolveToID("alpha")
	if !ok || target.Value() != id.RuntimeID.Value() {
		t.Fatalf("resolve to id failed")
	}
}

func TestBuildAndPublishMessage(t *testing.T) {
	t.Parallel()
	dispatcher := &rawRecordDispatcher{}
	id := api.NewAgentID()
	mgr := &Manager{
		dispatcher: dispatcher,
		cfg:        config.Config{},
		agents:     map[api.AgentID]*agentState{id: &agentState{slug: "alpha"}},
	}
	msg, err := mgr.buildAgentMessage(id, api.OutboundMessage{ToSlug: "alpha", Content: "hi"})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	mgr.publishAgentMessage(id, msg)
	dispatcher.mu.Lock()
	count := len(dispatcher.rawCalls)
	dispatcher.mu.Unlock()
	if count == 0 {
		t.Fatalf("expected publish calls")
	}
}

func TestBuildAgentMessageUnknownRecipient(t *testing.T) {
	t.Parallel()
	mgr := &Manager{}
	_, err := mgr.buildAgentMessage(api.NewAgentID(), api.OutboundMessage{ToSlug: "missing", Content: "hi"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestHandleCommMessageBroadcast(t *testing.T) {
	t.Parallel()
	dispatcher := &rawRecordDispatcher{}
	mgr := &Manager{dispatcher: dispatcher}
	payload := api.AgentMessage{
		ID:        api.NewRuntimeID(),
		From:      api.NewRuntimeID(),
		To:        api.TargetID{},
		Content:   "hello",
		Timestamp: time.Now().UTC(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mgr.handleCommMessage(protocol.Message{Subject: "amux.comm.broadcast", Data: data})
	dispatcher.mu.Lock()
	count := len(dispatcher.events)
	dispatcher.mu.Unlock()
	if count == 0 {
		t.Fatalf("expected broadcast event")
	}
}
