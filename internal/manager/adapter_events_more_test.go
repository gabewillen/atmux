package manager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type recordingAdapter struct {
	events []adapter.Event
}

func (r *recordingAdapter) Name() string { return "recording" }
func (r *recordingAdapter) Manifest() adapter.Manifest { return adapter.Manifest{Name: "recording"} }
func (r *recordingAdapter) Matcher() adapter.PatternMatcher { return nil }
func (r *recordingAdapter) Formatter() adapter.ActionFormatter { return nil }
func (r *recordingAdapter) OnEvent(ctx context.Context, event adapter.Event) ([]adapter.Action, error) {
	_ = ctx
	r.events = append(r.events, event)
	return nil, nil
}

func TestDispatchRosterToAdapters(t *testing.T) {
	adapterA := &recordingAdapter{}
	adapterB := &recordingAdapter{}
	mgr := &Manager{
		agents: map[api.AgentID]*agentState{
			api.NewAgentID(): {adapter: adapterA, remote: false},
			api.NewAgentID(): {adapter: adapterB, remote: true},
		},
	}
	roster := []api.RosterEntry{{Name: "alpha", RuntimeID: api.NewRuntimeID(), Kind: api.RosterAgent, Presence: agent.PresenceOnline}}
	mgr.dispatchRosterToAdapters(context.Background(), roster)
	if len(adapterA.events) != 1 {
		t.Fatalf("expected roster event for local adapter")
	}
	if len(adapterB.events) != 0 {
		t.Fatalf("expected no roster event for remote adapter")
	}
}

func TestExecuteAdapterActionsAllBranches(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = read.Close()
		_ = write.Close()
	})
	runtime := &session.LocalSession{}
	setUnexportedField(runtime, "ptyPair", &pty.Pair{Master: write})
	dispatcher := &adapterRecordDispatcher{}
	repoRoot := t.TempDir()
	worktree := filepath.Join(repoRoot, ".amux", "worktrees", "agent")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}
	meta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "agent",
		Adapter:  "adapter",
		RepoRoot: repoRoot,
		Worktree: worktree,
		Location: api.Location{Type: api.LocationLocal},
	}
	agentRuntime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	state := &agentState{
		session: runtime,
		runtime: agentRuntime,
		presence: agent.PresenceOnline,
	}
	sendPayload, err := json.Marshal(actionSendInput{DataB64: base64.StdEncoding.EncodeToString([]byte("hi"))})
	if err != nil {
		t.Fatalf("marshal send: %v", err)
	}
	emitPayload, err := json.Marshal(actionEmitEvent{Event: adapter.Event{Type: "custom", Payload: json.RawMessage(`{"ok":true}`)}})
	if err != nil {
		t.Fatalf("marshal emit: %v", err)
	}
	updatePayload, err := json.Marshal(actionUpdatePresence{Presence: agent.PresenceBusy})
	if err != nil {
		t.Fatalf("marshal update: %v", err)
	}
	mgr := &Manager{dispatcher: dispatcher}
	actions := []adapter.Action{
		{Type: "send.input", Payload: sendPayload},
		{Type: "emit.event", Payload: emitPayload},
		{Type: "update.presence", Payload: updatePayload},
	}
	mgr.executeAdapterActions(context.Background(), state, actions)
	buf := make([]byte, 16)
	if _, err := read.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(dispatcher.events) == 0 {
		t.Fatalf("expected adapter event published")
	}
}

func TestHandleActionSendInputErrors(t *testing.T) {
	mgr := &Manager{}
	state := &agentState{session: &session.LocalSession{}}
	if payload, err := json.Marshal(actionSendInput{DataB64: "!!bad"}); err == nil {
		mgr.handleActionSendInput(context.Background(), state, payload)
	}
	if payload, err := json.Marshal(actionSendInput{}); err == nil {
		mgr.handleActionSendInput(context.Background(), state, payload)
	}
}

func TestPresenceTransitionCoverage(t *testing.T) {
	cases := []struct {
		current string
		target  string
	}{
		{agent.PresenceBusy, agent.PresenceOnline},
		{agent.PresenceOffline, agent.PresenceBusy},
		{agent.PresenceAway, agent.PresenceOnline},
		{agent.PresenceOnline, agent.PresenceOffline},
		{agent.PresenceAway, agent.PresenceBusy},
		{agent.PresenceAway, agent.PresenceAway},
	}
	for _, tc := range cases {
		_ = presenceTransitionEvents(tc.current, tc.target)
	}
}

type adapterRecordDispatcher struct {
	events []protocol.Event
}

func (r *adapterRecordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	r.events = append(r.events, event)
	return nil
}
func (r *adapterRecordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *adapterRecordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	return nil
}
func (r *adapterRecordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *adapterRecordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (r *adapterRecordDispatcher) MaxPayload() int { return 1024 }
func (r *adapterRecordDispatcher) JetStream() nats.JetStreamContext { return nil }
func (r *adapterRecordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
