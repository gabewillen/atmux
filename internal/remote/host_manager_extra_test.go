package remote

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type rawDispatcher struct {
	rawSubjects []string
}

func (r *rawDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	return nil
}
func (r *rawDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *rawDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = payload
	_ = reply
	r.rawSubjects = append(r.rawSubjects, subject)
	return nil
}
func (r *rawDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	return nil, nil
}
func (r *rawDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (r *rawDispatcher) MaxPayload() int { return 1024 }
func (r *rawDispatcher) JetStream() nats.JetStreamContext {
	return nil
}
func (r *rawDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestResolveToID(t *testing.T) {
	manager := &HostManager{
		hostID:    api.MustParseHostID("host"),
		peerID:    api.NewPeerID(),
		sessions:  map[api.SessionID]*remoteSession{},
		outbox:    NewOutbox(1024),
		connected: true,
	}
	if _, ok := manager.resolveToID("broadcast"); !ok {
		t.Fatalf("expected broadcast")
	}
	if _, ok := manager.resolveToID("manager"); !ok {
		t.Fatalf("expected manager")
	}
	if _, ok := manager.resolveToID("manager@other"); ok {
		t.Fatalf("expected manager mismatch")
	}
}

func TestBuildAgentMessage(t *testing.T) {
	manager := &HostManager{hostID: api.MustParseHostID("host"), peerID: api.NewPeerID(), sessions: map[api.SessionID]*remoteSession{}}
	session := &remoteSession{agentID: api.NewAgentID(), slug: "alpha"}
	_, err := manager.buildAgentMessage(nil, api.OutboundMessage{ToSlug: "manager", Content: "hi"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if _, err := manager.buildAgentMessage(session, api.OutboundMessage{ToSlug: "unknown", Content: "hi"}); err == nil {
		t.Fatalf("expected unknown recipient error")
	}
	msg, err := manager.buildAgentMessage(session, api.OutboundMessage{ToSlug: "manager", Content: "hi"})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	if msg.ToSlug != "manager" || msg.Content != "hi" {
		t.Fatalf("unexpected message")
	}
}

func TestPublishAgentMessage(t *testing.T) {
	dispatcher := &rawDispatcher{}
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		dispatcher:    dispatcher,
		outbox:        NewOutbox(1024),
		connected:     true,
	}
	session := &remoteSession{agentID: api.NewAgentID()}
	msg := api.AgentMessage{To: api.TargetIDFromRuntime(session.agentID.RuntimeID)}
	manager.publishAgentMessage(session, msg)
	manager.publishComm("amux.comm.broadcast", []byte("hi"))
	if len(dispatcher.rawSubjects) == 0 {
		t.Fatalf("expected publish")
	}
}

func TestPublishCommOutbox(t *testing.T) {
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		outbox:        NewOutbox(64),
		connected:     false,
	}
	manager.publishComm("amux.comm.broadcast", []byte("hi"))
	if len(manager.outbox.Drain()) == 0 {
		t.Fatalf("expected outbox entry")
	}
}

func TestCommSubjectForTarget(t *testing.T) {
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		peerID:        api.NewPeerID(),
		sessions:      map[api.SessionID]*remoteSession{},
	}
	agentID := api.NewAgentID()
	sessionID := api.NewSessionID()
	manager.sessions[sessionID] = &remoteSession{agentID: agentID}
	target := api.TargetIDFromRuntime(agentID.RuntimeID)
	if subject := manager.commSubjectForTarget(target); subject == "" {
		t.Fatalf("expected comm subject")
	}
}
