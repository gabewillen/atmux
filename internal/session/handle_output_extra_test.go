package session

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/monitor"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type matchMatcher struct {
	matches []adapter.PatternMatch
}

func (m matchMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return m.matches, nil
}

type recordDispatcher struct {
	subjects []string
	events   []protocol.Event
}

func (r *recordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	r.subjects = append(r.subjects, subject)
	r.events = append(r.events, event)
	return nil
}
func (r *recordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *recordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (r *recordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (r *recordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (r *recordDispatcher) MaxPayload() int { return 0 }
func (r *recordDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r *recordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestHandleOutputPatterns(t *testing.T) {
	repo := t.TempDir()
	worktree := filepath.Join(repo, "wt")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	meta, err := api.NewAgent("alpha", "about", "stub", repo, worktree, api.Location{Type: api.LocationLocal})
	if err != nil {
		t.Fatalf("agent meta: %v", err)
	}
	dispatcher := &recordDispatcher{}
	runtime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	outbound := api.OutboundMessage{ToSlug: "beta", Content: "hello"}
	encoded, err := json.Marshal(outbound)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	matcher := matchMatcher{matches: []adapter.PatternMatch{
		{Pattern: "prompt"},
		{Pattern: "rate_limit"},
		{Pattern: "completion"},
		{Pattern: "message", Text: string(encoded)},
	}}
	mon, err := monitor.NewMonitor(matcher, monitor.Options{})
	if err != nil {
		t.Fatalf("monitor: %v", err)
	}
	sess := &LocalSession{
		agent:      runtime,
		monitor:    mon,
		dispatcher: dispatcher,
	}
	sess.handleOutput(context.Background(), []byte("data"))
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected events")
	}
	foundOutbound := false
	for i, subject := range dispatcher.subjects {
		if subject == protocol.Subject("events", "message") {
			if dispatcher.events[i].Name == "message.outbound" {
				foundOutbound = true
			}
		}
	}
	if !foundOutbound {
		t.Fatalf("expected outbound message event")
	}
}
