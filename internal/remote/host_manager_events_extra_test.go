package remote

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestPublishHostEventEnqueuesWhenDisconnected(t *testing.T) {
	manager := &HostManager{
		subjectPrefix: "amux",
		hostID:        api.MustParseHostID("host"),
		outbox:        NewOutbox(1024),
		connected:     false,
	}
	manager.publishHostEvent(context.Background(), "connection.lost", ConnectionLostPayload{PeerID: "peer", Timestamp: "now"})
	entries := manager.outbox.Drain()
	if len(entries) != 1 {
		t.Fatalf("expected outbox entry")
	}
}

func TestMarkDisconnectedSetsReplayGate(t *testing.T) {
	session := &remoteSession{}
	manager := &HostManager{
		hostID:    api.MustParseHostID("host"),
		peerID:    api.NewPeerID(),
		outbox:    NewOutbox(1024),
		connected: true,
		ready:     true,
		sessions:  map[api.SessionID]*remoteSession{api.NewSessionID(): session},
	}
	manager.markDisconnected("io_error")
	if manager.connected || manager.ready {
		t.Fatalf("expected disconnected")
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if !session.replayGate {
		t.Fatalf("expected replay gate set")
	}
	if len(manager.outbox.Drain()) == 0 {
		t.Fatalf("expected outbox event")
	}
}

func TestReplyErrorPublishes(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	manager := &HostManager{dispatcher: dispatcher}
	if err := manager.replyError("reply", "ping", "bad", "oops"); err != nil {
		t.Fatalf("reply error: %v", err)
	}
	if dispatcher.lastSubject != "reply" {
		t.Fatalf("expected reply subject")
	}
	msg, err := DecodeControlMessage(dispatcher.lastPayload)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg.Type != "error" {
		t.Fatalf("expected error message")
	}
}

func TestReconnectDelayBackoff(t *testing.T) {
	cfg := config.Config{Remote: config.RemoteConfig{
		ReconnectBackoffBase: 100 * time.Millisecond,
		ReconnectBackoffMax:  500 * time.Millisecond,
	}}
	if delay := reconnectDelay(cfg, 1); delay != 100*time.Millisecond {
		t.Fatalf("unexpected delay: %v", delay)
	}
	if delay := reconnectDelay(cfg, 3); delay != 400*time.Millisecond {
		t.Fatalf("unexpected delay: %v", delay)
	}
	if delay := reconnectDelay(cfg, 4); delay != 500*time.Millisecond {
		t.Fatalf("unexpected delay: %v", delay)
	}
}

func TestWriteHeartbeatNilKV(t *testing.T) {
	manager := &HostManager{hostID: api.MustParseHostID("host")}
	if err := manager.writeHeartbeat(context.Background()); err == nil {
		t.Fatalf("expected heartbeat error")
	}
}

func TestHandleKillInvalidPayload(t *testing.T) {
	dispatcher := &recordRawDispatcher{}
	manager := &HostManager{dispatcher: dispatcher}
	manager.handleKill("reply", ControlMessage{Type: "kill"})
	if dispatcher.lastSubject != "reply" {
		t.Fatalf("expected reply on invalid payload")
	}
}
