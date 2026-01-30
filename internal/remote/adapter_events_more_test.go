package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type adapterEventStub struct {
	mu    sync.Mutex
	calls int
}

func (s *adapterEventStub) Name() string                       { return "stub" }
func (s *adapterEventStub) Manifest() adapter.Manifest         { return adapter.Manifest{} }
func (s *adapterEventStub) Matcher() adapter.PatternMatcher    { return nil }
func (s *adapterEventStub) Formatter() adapter.ActionFormatter { return nil }
func (s *adapterEventStub) OnEvent(ctx context.Context, event adapter.Event) ([]adapter.Action, error) {
	_ = ctx
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	return []adapter.Action{
		{Type: "emit.event", Payload: json.RawMessage(`{"event":{"type":"adapter.notice"}}`)},
	}, nil
}

type adapterEventDispatcher struct {
	mu     sync.Mutex
	events []protocol.Event
}

func (r *adapterEventDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

func (r *adapterEventDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *adapterEventDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (r *adapterEventDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *adapterEventDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (r *adapterEventDispatcher) MaxPayload() int { return 1024 }
func (r *adapterEventDispatcher) JetStream() nats.JetStreamContext {
	return nil
}
func (r *adapterEventDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestHandlePresenceEventRoster(t *testing.T) {
	dispatcher := &adapterEventDispatcher{}
	stub := &adapterEventStub{}
	sess := &remoteSession{adapterRef: stub}
	manager := &HostManager{
		dispatcher: dispatcher,
		sessions:   map[api.SessionID]*remoteSession{api.NewSessionID(): sess},
	}
	roster := []api.RosterEntry{{Kind: api.RosterAgent, Name: "agent", RuntimeID: api.NewRuntimeID()}}
	event := protocol.Event{Name: agent.EventPresenceChanged, Payload: roster}
	manager.handlePresenceEvent(event)
	stub.mu.Lock()
	calls := stub.calls
	stub.mu.Unlock()
	if calls == 0 {
		t.Fatalf("expected adapter OnEvent to be called")
	}
	dispatcher.mu.Lock()
	events := len(dispatcher.events)
	dispatcher.mu.Unlock()
	if events == 0 {
		t.Fatalf("expected adapter event publish")
	}
}

func TestHandlePresenceEventUpdateSession(t *testing.T) {
	id := api.NewAgentID()
	session := &remoteSession{presence: agent.PresenceOffline}
	manager := &HostManager{agentIndex: map[api.AgentID]*remoteSession{id: session}}
	event := protocol.Event{Name: agent.EventPresenceChanged, Payload: agent.PresenceEvent{AgentID: id, Presence: agent.PresenceBusy}}
	manager.handlePresenceEvent(event)
	if session.presence != agent.PresenceBusy {
		t.Fatalf("expected presence to update")
	}
}

func TestHandleActionSendInputAndUpdatePresence(t *testing.T) {
	dispatcher := &adapterEventDispatcher{}
	repoRoot := t.TempDir()
	worktree := repoRoot + "/work"
	agentMeta := api.Agent{ID: api.NewAgentID(), Name: "agent", Adapter: "adapter", RepoRoot: repoRoot, Worktree: worktree, Location: api.Location{Type: api.LocationLocal}}
	agentRuntime, err := agent.NewAgent(agentMeta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	sess := &remoteSession{
		agentRuntime: agentRuntime,
		runtime:      &session.LocalSession{},
		presence:     agent.PresenceOnline,
	}
	manager := &HostManager{dispatcher: dispatcher}
	inputPayload, err := json.Marshal(actionSendInput{DataB64: base64.StdEncoding.EncodeToString([]byte("hi"))})
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	manager.handleActionSendInput(sess, inputPayload)
	updatePayload, err := json.Marshal(actionUpdatePresence{Presence: agent.PresenceBusy})
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	manager.handleActionUpdatePresence(context.Background(), sess, updatePayload)
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected presence events")
	}
}
