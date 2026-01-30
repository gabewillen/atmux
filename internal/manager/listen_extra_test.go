package manager

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type listenDispatcher struct {
	subs []string
}

func (l *listenDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}

func (l *listenDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (l *listenDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (l *listenDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = handler
	l.subs = append(l.subs, subject)
	return &listenSub{}, nil
}

func (l *listenDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (l *listenDispatcher) MaxPayload() int { return 1024 }
func (l *listenDispatcher) JetStream() nats.JetStreamContext { return nil }
func (l *listenDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

type listenSub struct {
	unsubscribed bool
}

func (s *listenSub) Unsubscribe() error {
	s.unsubscribed = true
	return nil
}

func TestListenSubjectForTarget(t *testing.T) {
	mgr := &Manager{cfg: config.Config{}}
	if subject, ok := mgr.listenSubjectForTarget("subject:foo.bar"); !ok || subject != "foo.bar" {
		t.Fatalf("expected subject: prefix")
	}
	if subject, ok := mgr.listenSubjectForTarget("amux.comm.broadcast"); !ok || subject == "" {
		t.Fatalf("expected prefixed subject")
	}
	if subject, ok := mgr.listenSubjectForTarget("manager@host"); !ok || subject == "" {
		t.Fatalf("expected manager@host subject")
	}
	if _, ok := mgr.listenSubjectForTarget(""); ok {
		t.Fatalf("expected empty target to fail")
	}
}

func TestUpdateListenTargetsSubscribeAndUnsubscribe(t *testing.T) {
	dispatcher := &listenDispatcher{}
	id := api.NewAgentID()
	subject := "amux.comm.broadcast"
	state := &agentState{listenSubjects: []string{subject}}
	mgr := &Manager{
		dispatcher:    dispatcher,
		agents:        map[api.AgentID]*agentState{id: state},
		listenSubs:    map[string]*listenSubscription{subject: &listenSubscription{subject: subject, sub: &listenSub{}}},
		listenTargets: map[string]map[api.AgentID]struct{}{subject: {id: {}}},
		cfg:           config.Config{},
	}
	mgr.updateListenTargets(context.Background(), id, []string{"amux.comm.director"})
	if len(dispatcher.subs) == 0 {
		t.Fatalf("expected subscription")
	}
	if len(state.listenSubjects) != 1 {
		t.Fatalf("expected updated listen subjects")
	}
}

func TestResolveListenSubjectsDedup(t *testing.T) {
	mgr := &Manager{cfg: config.Config{}}
	subjects := mgr.resolveListenSubjects([]string{"broadcast", "broadcast", "subject:foo.bar"})
	if len(subjects) != 2 {
		t.Fatalf("expected deduped subjects")
	}
}

func TestShouldSubscribeListenSubject(t *testing.T) {
	host := api.MustParseHostID("host")
	director := &remote.Director{}
	setUnexportedField(director, "hostID", host)
	mgr := &Manager{
		cfg: config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		remoteDirector: director,
	}
	if mgr.shouldSubscribeListenSubject(remote.ManagerCommSubject("amux", host)) {
		t.Fatalf("expected manager subject not to subscribe")
	}
	if mgr.shouldSubscribeListenSubject(remote.BroadcastCommSubject("amux")) {
		t.Fatalf("expected broadcast subject not to subscribe")
	}
	if !mgr.shouldSubscribeListenSubject("custom.subject") {
		t.Fatalf("expected custom subject to subscribe")
	}
}

func TestMirrorListenedMessage(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() {
		_ = read.Close()
		_ = write.Close()
	}()
	runtime := &session.LocalSession{}
	setUnexportedField(runtime, "ptyPair", &pty.Pair{Master: write})
	agentID := api.NewAgentID()
	state := &agentState{
		session:   runtime,
		formatter: stubFormatter{},
	}
	mgr := &Manager{
		agents: map[api.AgentID]*agentState{agentID: state},
		listenTargets: map[string]map[api.AgentID]struct{}{
			"amux.comm.broadcast": {agentID: {}},
		},
	}
	payload := api.AgentMessage{Content: "hello"}
	mgr.mirrorListenedMessage("amux.comm.broadcast", payload)
	buf := make([]byte, 128)
	if _, err := read.Read(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
}
