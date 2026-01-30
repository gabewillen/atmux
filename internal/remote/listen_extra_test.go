package remote

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/pty"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type listenSub struct {
	unsubscribed bool
}

func (s *listenSub) Unsubscribe() error {
	s.unsubscribed = true
	return nil
}

type listenDispatcher struct {
	subs        map[string]func(protocol.Message)
	subObjs     map[string]*listenSub
	subscribeErr map[string]error
}

func (d *listenDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	return nil
}
func (d *listenDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	return nil, nil
}
func (d *listenDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	return nil
}
func (d *listenDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	if d.subscribeErr != nil {
		if err := d.subscribeErr[subject]; err != nil {
			return nil, err
		}
	}
	if d.subs == nil {
		d.subs = make(map[string]func(protocol.Message))
	}
	if d.subObjs == nil {
		d.subObjs = make(map[string]*listenSub)
	}
	d.subs[subject] = handler
	sub := &listenSub{}
	d.subObjs[subject] = sub
	return sub, nil
}
func (d *listenDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	return protocol.Message{}, nil
}
func (d *listenDispatcher) MaxPayload() int { return 1024 }
func (d *listenDispatcher) JetStream() nats.JetStreamContext { return nil }
func (d *listenDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestListenSubjectForTarget(t *testing.T) {
	manager := &HostManager{
		cfg: config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID:     api.MustParseHostID("host"),
		agentIndex: map[api.AgentID]*remoteSession{},
	}
	agentID := api.NewAgentID()
	manager.agentIndex[agentID] = &remoteSession{slug: "Alpha"}
	if subject, ok := manager.listenSubjectForTarget("subject: custom.topic"); !ok || subject != "custom.topic" {
		t.Fatalf("expected subject prefix passthrough")
	}
	if subject, ok := manager.listenSubjectForTarget("amux.comm.director"); !ok || subject != "amux.comm.director" {
		t.Fatalf("expected full subject passthrough")
	}
	if subject, ok := manager.listenSubjectForTarget("broadcast"); !ok || subject != BroadcastCommSubject("amux") {
		t.Fatalf("expected broadcast subject")
	}
	if subject, ok := manager.listenSubjectForTarget("manager"); !ok || subject != ManagerCommSubject("amux", manager.hostID) {
		t.Fatalf("expected manager subject")
	}
	if subject, ok := manager.listenSubjectForTarget("manager@host"); !ok || subject != ManagerCommSubject("amux", manager.hostID) {
		t.Fatalf("expected manager@host subject")
	}
	if subject, ok := manager.listenSubjectForTarget("alpha"); !ok || subject != AgentCommSubject("amux", manager.hostID, agentID) {
		t.Fatalf("expected agent subject")
	}
	if _, ok := manager.listenSubjectForTarget("manager@"); ok {
		t.Fatalf("expected invalid host parse")
	}
}

func TestResolveListenSubjectsDedup(t *testing.T) {
	manager := &HostManager{
		cfg:         config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID:      api.MustParseHostID("host"),
		agentIndex:  map[api.AgentID]*remoteSession{},
	}
	targets := []string{"broadcast", "broadcast", "  ", "subject:foo.bar"}
	subjects := manager.resolveListenSubjects(targets)
	if len(subjects) != 2 {
		t.Fatalf("expected 2 subjects, got %d", len(subjects))
	}
}

func TestUpdateListenTargetsSubscribesAndUnsubscribes(t *testing.T) {
	dispatcher := &listenDispatcher{}
	agentID := api.NewAgentID()
	manager := &HostManager{
		cfg:         config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID:      api.MustParseHostID("host"),
		dispatcher:  dispatcher,
		listenSubs:  make(map[string]*listenSubscription),
		listenTargets: make(map[string]map[api.AgentID]struct{}),
		agentIndex: map[api.AgentID]*remoteSession{
			agentID: {agentID: agentID, listenSubjects: []string{"custom.one"}},
		},
	}
	oldSub := &listenSub{}
	manager.listenTargets["custom.one"] = map[api.AgentID]struct{}{agentID: {}}
	manager.listenSubs["custom.one"] = &listenSubscription{subject: "custom.one", sub: oldSub}
	manager.updateListenTargets(context.Background(), agentID, []string{"custom.two"})
	if !oldSub.unsubscribed {
		t.Fatalf("expected old subscription to unsubscribe")
	}
	if _, ok := manager.listenTargets["custom.two"]; !ok {
		t.Fatalf("expected new listen target")
	}
	if _, ok := manager.listenSubs["custom.two"]; !ok {
		t.Fatalf("expected new subscription")
	}
}

func TestShouldSubscribeListenSubject(t *testing.T) {
	manager := &HostManager{
		cfg:    config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID: api.MustParseHostID("host"),
	}
	if manager.shouldSubscribeListenSubject(ManagerCommSubject("amux", manager.hostID)) {
		t.Fatalf("expected manager subject not to subscribe")
	}
	if manager.shouldSubscribeListenSubject(BroadcastCommSubject("amux")) {
		t.Fatalf("expected broadcast subject not to subscribe")
	}
	agentSubject := AgentCommSubject("amux", manager.hostID, api.NewAgentID())
	if manager.shouldSubscribeListenSubject(agentSubject) {
		t.Fatalf("expected agent subject not to subscribe")
	}
	if !manager.shouldSubscribeListenSubject("custom.subject") {
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
	sessionID := api.NewSessionID()
	agentID := api.NewAgentID()
	manager := &HostManager{
		listenTargets: map[string]map[api.AgentID]struct{}{
			"custom.subject": {agentID: {}},
		},
		agentIndex: map[api.AgentID]*remoteSession{
			agentID: {agentID: agentID, sessionID: sessionID, runtime: runtime},
		},
	}
	manager.mirrorListenedMessage("custom.subject", api.AgentMessage{Content: "hello"})
	buf := make([]byte, 64)
	n, readErr := read.Read(buf)
	if readErr != nil {
		t.Fatalf("read: %v", readErr)
	}
	if n == 0 {
		t.Fatalf("expected output")
	}
}

func TestUpdateListenTargetsSubscribeError(t *testing.T) {
	dispatcher := &listenDispatcher{
		subscribeErr: map[string]error{"custom.subject": errors.New("fail")},
	}
	agentID := api.NewAgentID()
	manager := &HostManager{
		cfg:          config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID:       api.MustParseHostID("host"),
		dispatcher:   dispatcher,
		listenSubs:   make(map[string]*listenSubscription),
		listenTargets: make(map[string]map[api.AgentID]struct{}),
		agentIndex:   map[api.AgentID]*remoteSession{agentID: {agentID: agentID}},
	}
	manager.updateListenTargets(context.Background(), agentID, []string{"custom.subject"})
	if len(manager.listenSubs) != 0 {
		t.Fatalf("expected no subscriptions on error")
	}
}

func TestClearListenRemovesTargets(t *testing.T) {
	dispatcher := &listenDispatcher{}
	agentID := api.NewAgentID()
	session := &remoteSession{agentID: agentID, listenSubjects: []string{"custom.subject"}}
	manager := &HostManager{
		cfg:          config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		hostID:       api.MustParseHostID("host"),
		dispatcher:   dispatcher,
		listenSubs:   make(map[string]*listenSubscription),
		listenTargets: map[string]map[api.AgentID]struct{}{
			"custom.subject": {agentID: {}},
		},
		agentIndex: map[api.AgentID]*remoteSession{agentID: session},
	}
	manager.clearListen(session)
	if len(manager.listenTargets) != 0 {
		t.Fatalf("expected listen targets cleared")
	}
}
