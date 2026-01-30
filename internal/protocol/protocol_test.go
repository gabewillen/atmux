package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestSubjectJoin(t *testing.T) {
	subject := Subject("events", "", "presence", "agent")
	if subject != "events.presence.agent" {
		t.Fatalf("unexpected subject: %s", subject)
	}
}

func TestDispatcherPublishSubscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server, err := StartHubServer(ctx, HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	defer server.Shutdown()
	dispatcher, err := NewNATSDispatcher(ctx, server.URL(), NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	if dispatcher.JetStream() == nil {
		t.Fatalf("expected jetstream context")
	}
	defer func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Fatalf("close: %v", err)
		}
	}()
	ch := make(chan Event, 1)
	sub, err := dispatcher.Subscribe(ctx, Subject("events", "presence"), func(event Event) {
		ch <- event
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() {
		if err := sub.Unsubscribe(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
			t.Errorf("unsubscribe: %v", err)
		}
	})
	payload := map[string]any{"ok": true}
	if err := dispatcher.Publish(ctx, Subject("events", "presence"), Event{Name: "presence.changed", Payload: payload, OccurredAt: time.Now().UTC()}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case event := <-ch:
		if event.Name != "presence.changed" {
			t.Fatalf("unexpected event: %s", event.Name)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for event")
	}
}

func TestDispatcherRawAndRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server, err := StartHubServer(ctx, HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	defer server.Shutdown()
	dispatcher, err := NewNATSDispatcher(ctx, server.URL(), NATSOptions{AllowNoJetStream: true})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	subject := Subject("request", "ping")
	_, err = dispatcher.SubscribeRaw(ctx, subject, func(msg Message) {
		if msg.Reply == "" {
			return
		}
		_ = dispatcher.PublishRaw(ctx, msg.Reply, []byte("pong"), "")
	})
	if err != nil {
		t.Fatalf("subscribe raw: %v", err)
	}
	reply, err := dispatcher.Request(ctx, subject, []byte("ping"), 2*time.Second)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if string(reply.Data) != "pong" {
		t.Fatalf("unexpected reply: %s", string(reply.Data))
	}
	if dispatcher.MaxPayload() == 0 {
		t.Fatalf("expected max payload")
	}
	select {
	case <-dispatcher.Closed():
		t.Fatalf("dispatcher should not be closed yet")
	default:
	}
	if err := dispatcher.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
	select {
	case <-dispatcher.Closed():
	case <-time.After(time.Second):
		t.Fatalf("closed channel not closed")
	}
}

func TestNatsServerHelpers(t *testing.T) {
	host, port, err := splitHostPort("127.0.0.1:0")
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if host == "" {
		t.Fatalf("unexpected host")
	}
	if _, _, err := splitHostPort("invalid"); err == nil {
		t.Fatalf("expected split error")
	}
	if _, err := parsePort("not-a-port"); err == nil {
		t.Fatalf("expected parse error")
	}
	leafHost, leafPort, err := resolveLeafListen(HubServerConfig{LeafListen: "127.0.0.1:0"}, host, port)
	if err != nil {
		t.Fatalf("resolve leaf: %v", err)
	}
	if leafHost == "" || leafPort == 0 {
		t.Fatalf("unexpected leaf: %s:%d", leafHost, leafPort)
	}
	if _, err := parsePort(""); err == nil {
		t.Fatalf("expected parse port error")
	}
	if _, err := allocatePort("invalid-addr"); err == nil {
		t.Fatalf("expected allocate port error")
	}
	if url := buildLeafURL("127.0.0.1", 4223, "", false); url == "" {
		t.Fatalf("expected leaf url")
	}
}

func TestStartLeafServerErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := StartLeafServer(ctx, LeafServerConfig{Listen: "bad"}); err == nil {
		t.Fatalf("expected invalid listen error")
	}
	if _, err := StartLeafServer(ctx, LeafServerConfig{Listen: "127.0.0.1:-1", HubURL: "://bad"}); err == nil {
		t.Fatalf("expected bad url error")
	}
}

func TestEventJSONRoundTrip(t *testing.T) {
	event := Event{Name: "test", Payload: map[string]any{"ok": true}, OccurredAt: time.Now().UTC()}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Name != event.Name {
		t.Fatalf("unexpected name: %s", decoded.Name)
	}
}

func TestStartHubServerErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := StartHubServer(ctx, HubServerConfig{Listen: "bad"}); err == nil {
		t.Fatalf("expected invalid listen error")
	}
	if _, err := StartHubServer(ctx, HubServerConfig{Listen: "127.0.0.1:-1"}); err == nil {
		t.Fatalf("expected missing jetstream dir error")
	}
}

func TestNATSServerCloseNil(t *testing.T) {
	var server *NATSServer
	if server.URL() != "" {
		t.Fatalf("expected empty url")
	}
	if server.LeafURL() != "" {
		t.Fatalf("expected empty leaf url")
	}
	if server.LeafCount() != 0 {
		t.Fatalf("expected zero leaf count")
	}
	server.Shutdown()
	if err := server.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestParseHostPortAllocation(t *testing.T) {
	port, err := allocatePort("127.0.0.1")
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_ = conn.Close()
	if port == 0 {
		t.Fatalf("expected port")
	}
}
