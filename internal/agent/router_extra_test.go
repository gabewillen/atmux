package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type stubDispatcher struct {
	publishErr error
	subscribeErr error
}

func (s stubDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return s.publishErr
}
func (s stubDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, s.subscribeErr
}
func (s stubDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (s stubDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (s stubDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (s stubDispatcher) MaxPayload() int { return 0 }
func (s stubDispatcher) JetStream() nats.JetStreamContext {
	return nil
}
func (s stubDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestEventRouterEmitErrors(t *testing.T) {
	agentMeta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "alpha",
		Adapter:  api.AdapterRef("stub"),
		RepoRoot: "/tmp/repo",
		Worktree: "/tmp/repo/work",
		Location: api.Location{Type: api.LocationLocal},
	}
	agt := &Agent{Agent: agentMeta}
	router := NewEventRouter(agt, nil)
	if err := router.EmitLifecycle(context.Background(), "start", nil); err == nil {
		t.Fatalf("expected emit error")
	}
	if err := router.EmitPresence(context.Background(), "", nil); err == nil {
		t.Fatalf("expected empty name error")
	}
}

func TestAgentEmitRecordsError(t *testing.T) {
	agentMeta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "alpha",
		Adapter:  api.AdapterRef("stub"),
		RepoRoot: "/tmp/repo",
		Worktree: "/tmp/repo/work",
		Location: api.Location{Type: api.LocationLocal},
	}
	dispatcher := stubDispatcher{publishErr: errors.New("fail")}
	agt := &Agent{Agent: agentMeta, dispatcher: dispatcher}
	agt.router = NewEventRouter(agt, dispatcher)
	if err := agt.EmitLifecycle(context.Background(), "start", nil); err == nil {
		t.Fatalf("expected emit error")
	}
	if agt.LastError() == nil {
		t.Fatalf("expected last error set")
	}
}

func TestAgentStartSubscribeError(t *testing.T) {
	dispatcher := stubDispatcher{subscribeErr: errors.New("subscribe fail")}
	agentMeta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "alpha",
		Adapter:  api.AdapterRef("stub"),
		RepoRoot: "/tmp/repo",
		Worktree: "/tmp/repo/work",
		Location: api.Location{Type: api.LocationLocal},
	}
	agt := &Agent{Agent: agentMeta, dispatcher: dispatcher}
	agt.router = NewEventRouter(agt, dispatcher)
	agt.Lifecycle, _ = NewLifecycle(agt, dispatcher)
	agt.Presence, _ = NewPresence(agt, dispatcher)
	agt.Start(context.Background())
	if agt.LastError() == nil {
		t.Fatalf("expected last error")
	}
}
