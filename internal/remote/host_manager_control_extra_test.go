package remote

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type stubDispatcher struct {
	lastSubject string
	lastPayload []byte
}

func (s *stubDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	_ = event
	return nil
}

func (s *stubDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (s *stubDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = reply
	s.lastSubject = subject
	s.lastPayload = append([]byte(nil), payload...)
	return nil
}

func (s *stubDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (s *stubDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (s *stubDispatcher) MaxPayload() int { return 0 }

func (s *stubDispatcher) JetStream() nats.JetStreamContext { return nil }

func (s *stubDispatcher) Closed() <-chan struct{} { return make(chan struct{}) }

func TestHandlePingKillReplay(t *testing.T) {
	dispatcher := &stubDispatcher{}
	manager := &HostManager{dispatcher: dispatcher}
	pingMsg, err := EncodePayload("ping", PingPayload{UnixMS: 123})
	if err != nil {
		t.Fatalf("encode ping: %v", err)
	}
	manager.handlePing("reply.ping", pingMsg)
	control, err := DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode control: %v", err)
	}
	if control.Type != "pong" {
		t.Fatalf("expected pong response")
	}
	killMsg, err := EncodePayload("kill", KillRequest{SessionID: api.NewSessionID().String()})
	if err != nil {
		t.Fatalf("encode kill: %v", err)
	}
	manager.handleKill("reply.kill", killMsg)
	control, err = DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode kill: %v", err)
	}
	if control.Type != "kill" {
		t.Fatalf("expected kill response")
	}
	replayMsg, err := EncodePayload("replay", ReplayRequest{SessionID: api.NewSessionID().String()})
	if err != nil {
		t.Fatalf("encode replay: %v", err)
	}
	manager.handleReplay("reply.replay", replayMsg)
	control, err = DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode replay: %v", err)
	}
	if control.Type != "replay" {
		t.Fatalf("expected replay response")
	}
}

func TestHandleSpawnErrors(t *testing.T) {
	dispatcher := &stubDispatcher{}
	manager := &HostManager{
		dispatcher: dispatcher,
		agentIndex: map[api.AgentID]*remoteSession{},
	}
	badControl := ControlMessage{Type: "spawn", Payload: []byte("bad")}
	manager.handleSpawn("reply.bad", badControl)
	if dispatcher.lastSubject == "" {
		t.Fatalf("expected error reply for bad payload")
	}
	req := SpawnRequest{AgentID: "bad", AgentSlug: "slug", RepoPath: "/tmp", Adapter: "stub", Command: []string{"cmd"}}
	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	manager.handleSpawn("reply.agent", ControlMessage{Type: "spawn", Payload: payload})
	req = SpawnRequest{AgentID: api.NewAgentID().String(), AgentSlug: "", RepoPath: "/tmp", Adapter: "stub", Command: []string{"cmd"}}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.slug", ControlMessage{Type: "spawn", Payload: payload})
	req = SpawnRequest{AgentID: api.NewAgentID().String(), AgentSlug: "slug", RepoPath: "/tmp", Adapter: "", Command: []string{"cmd"}}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.adapter", ControlMessage{Type: "spawn", Payload: payload})
	req = SpawnRequest{AgentID: api.NewAgentID().String(), AgentSlug: "slug", RepoPath: "", Adapter: "stub", Command: []string{"cmd"}}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.repo", ControlMessage{Type: "spawn", Payload: payload})
	req = SpawnRequest{AgentID: api.NewAgentID().String(), AgentSlug: "slug", RepoPath: "/tmp/missing", Adapter: "stub", Command: []string{"cmd"}}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.missing", ControlMessage{Type: "spawn", Payload: payload})
	agentID := api.NewAgentID()
	manager.agentIndex[agentID] = &remoteSession{agentID: agentID, slug: "other", repoPath: "/tmp", adapter: "stub"}
	req = SpawnRequest{AgentID: agentID.String(), AgentSlug: "slug", RepoPath: "/tmp", Adapter: "stub", Command: []string{"cmd"}}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.conflict", ControlMessage{Type: "spawn", Payload: payload})
	req = SpawnRequest{AgentID: api.NewAgentID().String(), AgentSlug: "slug", RepoPath: "/tmp", Adapter: "stub", Command: nil}
	payload, _ = json.Marshal(req)
	manager.handleSpawn("reply.command", ControlMessage{Type: "spawn", Payload: payload})
}

func TestHandleControlErrors(t *testing.T) {
	dispatcher := &stubDispatcher{}
	manager := &HostManager{dispatcher: dispatcher}
	manager.handleControl(protocol.Message{Reply: "reply.notready", Data: []byte("bad")})
	control, err := DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode not ready: %v", err)
	}
	if control.Type != "error" {
		t.Fatalf("expected error response")
	}
	manager.ready = true
	manager.handleControl(protocol.Message{Reply: "reply.invalid", Data: []byte("bad")})
	control, err = DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode invalid: %v", err)
	}
	if control.Type != "error" {
		t.Fatalf("expected error response")
	}
	msg, err := EncodePayload("noop", map[string]string{"ok": "yes"})
	if err != nil {
		t.Fatalf("encode noop: %v", err)
	}
	manager.handleControl(protocol.Message{Reply: "reply.unknown", Data: mustEncodeControl(t, msg)})
	control, err = DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode unknown: %v", err)
	}
	if control.Type != "error" {
		t.Fatalf("expected error response")
	}
}

func mustEncodeControl(t *testing.T, msg ControlMessage) []byte {
	t.Helper()
	data, err := EncodeControlMessage(msg)
	if err != nil {
		t.Fatalf("encode control: %v", err)
	}
	return data
}
