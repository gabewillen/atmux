package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type stubDispatcher struct{}

func (stubDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}
func (stubDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (stubDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (stubDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (stubDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (stubDispatcher) MaxPayload() int { return 0 }
func (stubDispatcher) JetStream() nats.JetStreamContext {
	return nil
}
func (stubDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestNewLocalSessionErrors(t *testing.T) {
	meta := api.Session{}
	if _, err := NewLocalSession(meta, nil, Command{}, "", stubMatcher{}, stubDispatcher{}, Config{}); err == nil {
		t.Fatalf("expected invalid argv error")
	}
	cmd := Command{Argv: []string{"echo"}}
	if _, err := NewLocalSession(meta, nil, cmd, "", stubMatcher{}, stubDispatcher{}, Config{}); err == nil {
		t.Fatalf("expected invalid worktree error")
	}
	if _, err := NewLocalSession(meta, nil, cmd, "/tmp", nil, stubDispatcher{}, Config{}); err == nil {
		t.Fatalf("expected matcher error")
	}
	if _, err := NewLocalSession(meta, nil, cmd, "/tmp", stubMatcher{}, nil, Config{}); err == nil {
		t.Fatalf("expected dispatcher error")
	}
}

func TestSessionNotRunningErrors(t *testing.T) {
	repo := t.TempDir()
	worktree := filepath.Join(repo, "wt")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	meta, err := api.NewSession(api.NewAgentID(), repo, worktree, api.Location{Type: api.LocationLocal})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	sess, err := NewLocalSession(meta, nil, Command{Argv: []string{"echo"}}, worktree, stubMatcher{}, stubDispatcher{}, Config{DrainTimeout: time.Millisecond})
	if err != nil {
		t.Fatalf("new local session: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sess.Start(ctx); err == nil {
		t.Fatalf("expected start context error")
	}
	if _, err := sess.Attach(); err == nil {
		t.Fatalf("expected attach error")
	}
	if err := sess.Stop(context.Background()); err == nil {
		t.Fatalf("expected stop error")
	}
	if err := sess.Kill(context.Background()); err == nil {
		t.Fatalf("expected kill error")
	}
	if err := sess.Send([]byte("hi")); err == nil {
		t.Fatalf("expected send error")
	}
	if err := sess.Restart(context.Background()); err == nil {
		t.Fatalf("expected restart error")
	}
}
