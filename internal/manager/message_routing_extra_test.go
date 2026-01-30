package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRouteOutboundMessagePublishes(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "nats")
	server, err := protocol.StartHubServer(context.Background(), protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: cfg.NATS.JetStreamDir,
	})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() { _ = dispatcher.Close(context.Background()) })
	mgr, err := NewManager(context.Background(), resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	record, err := mgr.AddAgent(context.Background(), AddRequest{
		Name:     "alpha",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:      repoRoot,
	})
	if err != nil {
		t.Fatalf("add agent: %v", err)
	}
	hostID := mgr.localHostID()
	subject := protocol.Subject("amux", "comm", "agent", hostID.String(), record.AgentID.String())
	received := make(chan struct{}, 1)
	_, err = dispatcher.SubscribeRaw(context.Background(), subject, func(msg protocol.Message) {
		received <- struct{}{}
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	payload := api.OutboundMessage{AgentID: record.AgentID, ToSlug: "alpha", Content: "hi"}
	mgr.routeOutboundMessage(context.Background(), payload)
	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected outbound message")
	}
	// Exercise unknown recipient path.
	mgr.mu.Lock()
	state := mgr.agents[*record.AgentID]
	state.session = &session.LocalSession{}
	mgr.mu.Unlock()
	mgr.routeOutboundMessage(context.Background(), api.OutboundMessage{AgentID: record.AgentID, ToSlug: "missing", Content: "hi"})
}

