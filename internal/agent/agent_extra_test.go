package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type errorDispatcher struct{}

func (e *errorDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	return errors.New("publish error")
}

func (e *errorDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (e *errorDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (e *errorDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (e *errorDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (e *errorDispatcher) MaxPayload() int {
	return 1024
}

func (e *errorDispatcher) JetStream() nats.JetStreamContext {
	return nil
}

func (e *errorDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestAgentEmitPresenceError(t *testing.T) {
	t.Parallel()
	dispatcher := &errorDispatcher{}
	repoRoot := t.TempDir()
	worktree := filepath.Join(repoRoot, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	meta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "a",
		Adapter:  "b",
		RepoRoot: repoRoot,
		Worktree: worktree,
		Location: api.Location{Type: api.LocationLocal},
	}
	agent, err := NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	if err := agent.EmitPresence(context.Background(), "presence", nil); err == nil {
		t.Fatalf("expected error")
	}
	if agent.LastError() == nil {
		t.Fatalf("expected last error to be recorded")
	}
}
