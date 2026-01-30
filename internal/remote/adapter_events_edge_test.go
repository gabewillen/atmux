package remote

import (
	"context"
	"encoding/json"
	"os"
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

type publishRecorder struct {
	subject string
	event   protocol.Event
}

func (p *publishRecorder) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	p.subject = subject
	p.event = event
	return nil
}
func (p *publishRecorder) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (p *publishRecorder) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}
func (p *publishRecorder) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}
func (p *publishRecorder) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}
func (p *publishRecorder) MaxPayload() int { return 0 }
func (p *publishRecorder) JetStream() nats.JetStreamContext { return nil }
func (p *publishRecorder) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestHandleActionSendInputErrors(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	runtime := &session.LocalSession{}
	setUnexportedField(runtime, "ptyPair", &pty.Pair{Master: writer})
	sess := &remoteSession{runtime: runtime}
	manager := &HostManager{}
	manager.handleActionSendInput(sess, json.RawMessage(`bad`))
	manager.handleActionSendInput(sess, json.RawMessage(`{"data_b64":""}`))
	manager.handleActionSendInput(sess, json.RawMessage(`{"data_b64":"@@@"}`))
	_ = writer.Close()
	buf := make([]byte, 16)
	n, _ := reader.Read(buf)
	if n != 0 {
		t.Fatalf("expected no data written")
	}
}

func TestHandleActionUpdatePresenceInvalidAndNoop(t *testing.T) {
	dispatcher := &adapterEventDispatcher{}
	agentMeta := api.Agent{
		ID:       api.NewAgentID(),
		Name:     "agent",
		Adapter:  "adapter",
		RepoRoot: "/tmp/repo",
		Worktree: "/tmp/repo/work",
		Location: api.Location{Type: api.LocationLocal},
	}
	agentRuntime, err := agent.NewAgent(agentMeta, dispatcher)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	sess := &remoteSession{agentRuntime: agentRuntime, presence: agent.PresenceBusy}
	manager := &HostManager{}
	manager.handleActionUpdatePresence(context.Background(), sess, json.RawMessage(`bad`))
	manager.handleActionUpdatePresence(context.Background(), sess, json.RawMessage(`{"presence":"busy"}`))
	if len(dispatcher.events) != 0 {
		t.Fatalf("expected no presence events")
	}
}

func TestHandleActionEmitEventEdgeCases(t *testing.T) {
	dispatcher := &publishRecorder{}
	manager := &HostManager{dispatcher: dispatcher}
	manager.handleActionEmitEvent(context.Background(), json.RawMessage(`bad`))
	manager.handleActionEmitEvent(context.Background(), json.RawMessage(`{"event":{"type":""}}`))
	payload := json.RawMessage(`{"x":1}`)
	req := actionEmitEvent{Event: adapter.Event{Type: "adapter.notice", Payload: payload}}
	encoded, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager.handleActionEmitEvent(context.Background(), encoded)
	if dispatcher.subject != protocol.Subject("events", "adapter") || dispatcher.event.Name != "adapter.notice" {
		t.Fatalf("unexpected publish")
	}
	if _, ok := dispatcher.event.Payload.(json.RawMessage); !ok {
		t.Fatalf("expected raw payload")
	}
}

func TestPresenceTransitionEventsEdgeCases(t *testing.T) {
	if got := presenceTransitionEvents(agent.PresenceOnline, agent.PresenceBusy); len(got) != 1 || got[0] != agent.EventTaskAssigned {
		t.Fatalf("unexpected online->busy events: %#v", got)
	}
	if got := presenceTransitionEvents(agent.PresenceOffline, agent.PresenceBusy); len(got) != 2 {
		t.Fatalf("unexpected offline->busy events: %#v", got)
	}
	if got := presenceTransitionEvents(agent.PresenceAway, agent.PresenceOnline); len(got) != 1 || got[0] != agent.EventActivity {
		t.Fatalf("unexpected away->online events: %#v", got)
	}
	if got := presenceTransitionEvents(agent.PresenceOnline, agent.PresenceOffline); len(got) != 1 || got[0] != agent.EventRateLimit {
		t.Fatalf("unexpected online->offline events: %#v", got)
	}
	if got := presenceTransitionEvents(agent.PresenceOnline, agent.PresenceAway); len(got) != 1 || got[0] != agent.EventStuckDetected {
		t.Fatalf("unexpected online->away events: %#v", got)
	}
}
