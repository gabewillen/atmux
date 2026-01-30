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

type recordSub struct {
	unsubscribed bool
}

func (r *recordSub) Unsubscribe() error {
	r.unsubscribed = true
	return nil
}

type sequenceDispatcher struct {
	calls      int
	firstSub   *recordSub
	secondErr  error
	publishErr error
}

func (s *sequenceDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return s.publishErr
}
func (s *sequenceDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	s.calls++
	if s.calls == 1 {
		sub := &recordSub{}
		s.firstSub = sub
		return sub, nil
	}
	return nil, s.secondErr
}
func (s *sequenceDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (s *sequenceDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (s *sequenceDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (s *sequenceDispatcher) MaxPayload() int { return 0 }
func (s *sequenceDispatcher) JetStream() nats.JetStreamContext { return nil }
func (s *sequenceDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestEventRouterStartUnsubscribesOnFailure(t *testing.T) {
	dispatcher := &sequenceDispatcher{secondErr: errors.New("subscribe fail")}
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
	if err := agt.router.Start(context.Background()); err == nil {
		t.Fatalf("expected start error")
	}
	if dispatcher.firstSub == nil || !dispatcher.firstSub.unsubscribed {
		t.Fatalf("expected lifecycle subscription to unsubscribe")
	}
}

func TestEventRouterStartNoopWhenStarted(t *testing.T) {
	dispatcher := &sequenceDispatcher{}
	agentMeta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "alpha",
		Adapter:  api.AdapterRef("stub"),
		RepoRoot: "/tmp/repo",
		Worktree: "/tmp/repo/work",
		Location: api.Location{Type: api.LocationLocal},
	}
	agt := &Agent{Agent: agentMeta, dispatcher: dispatcher}
	router := NewEventRouter(agt, dispatcher)
	router.started = true
	if err := router.Start(context.Background()); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if dispatcher.calls != 0 {
		t.Fatalf("expected no subscriptions when already started")
	}
}
