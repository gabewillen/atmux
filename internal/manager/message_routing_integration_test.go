package manager

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestStartMessageRoutingAndOutboundFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: filepath.Join(t.TempDir(), "jetstream"),
	})
	if err != nil {
		t.Fatalf("start hub: %v", err)
	}
	t.Cleanup(func() { server.Shutdown() })
	dispatcher, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() { _ = dispatcher.Close(context.Background()) })

	hostID := api.MustParseHostID("host")
	director := &remote.Director{}
	setUnexportedField(director, "hostID", hostID)

	mgr := &Manager{
		dispatcher:     dispatcher,
		cfg:            config.Config{Remote: config.RemoteConfig{NATS: config.RemoteNATSConfig{SubjectPrefix: "amux"}}},
		remoteDirector: director,
		agents:          map[api.AgentID]*agentState{},
	}
	sender := api.NewAgentID()
	mgr.agents[sender] = &agentState{slug: "alpha", remote: false}

	if err := mgr.startMessageRouting(ctx); err != nil {
		t.Fatalf("start routing: %v", err)
	}

	msgCh := make(chan protocol.Message, 1)
	subject := remote.BroadcastCommSubject("amux")
	sub, err := dispatcher.SubscribeRaw(ctx, subject, func(msg protocol.Message) {
		select {
		case msgCh <- msg:
		default:
		}
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = sub.Unsubscribe() })

	event := protocol.Event{
		Name:    "message.outbound",
		Payload: api.OutboundMessage{AgentID: &sender, ToSlug: "broadcast", Content: "hi"},
		OccurredAt: time.Now().UTC(),
	}
	if err := dispatcher.Publish(ctx, protocol.Subject("events", "message"), event); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case <-msgCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected broadcast message")
	}
}
