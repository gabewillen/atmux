package remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type noopMatcher struct{}

func (noopMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return nil, nil
}

type noopDispatcher struct{}

func (noopDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	return nil
}
func (noopDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (noopDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	return nil
}
func (noopDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return nil, nil
}
func (noopDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (noopDispatcher) MaxPayload() int { return 1024 }
func (noopDispatcher) JetStream() nats.JetStreamContext { return nil }
func (noopDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestHandleKillSuccess(t *testing.T) {
	repoRoot := t.TempDir()
	worktree := filepath.Join(repoRoot, "worktree")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	agentID := api.NewAgentID()
	meta, err := api.NewSession(agentID, repoRoot, worktree, api.Location{Type: api.LocationLocal})
	if err != nil {
		t.Fatalf("session meta: %v", err)
	}
	sess, err := session.NewLocalSession(meta, nil, session.Command{Argv: []string{"sh", "-c", "cat"}}, worktree, noopMatcher{}, noopDispatcher{}, session.Config{})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	if err := sess.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		_ = sess.Kill(context.Background())
	})
	dispatcher := &recordRawDispatcher{}
	sessionID := meta.ID
	manager := &HostManager{
		dispatcher: dispatcher,
		sessions:   map[api.SessionID]*remoteSession{sessionID: {agentID: agentID, sessionID: sessionID, runtime: sess}},
		agentIndex: map[api.AgentID]*remoteSession{agentID: {agentID: agentID, sessionID: sessionID, runtime: sess}},
	}
	control, err := EncodePayload("kill", KillRequest{SessionID: sessionID.String()})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	manager.handleKill("reply", control)
	msg, err := DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg.Type != "kill" {
		t.Fatalf("expected kill response")
	}
	var resp KillResponse
	if err := DecodePayload(msg, &resp); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !resp.Killed {
		t.Fatalf("expected killed")
	}
}
