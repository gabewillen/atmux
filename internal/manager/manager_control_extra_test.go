package manager

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestRestartAndKillAgent(t *testing.T) {
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
		t.Fatalf("connect nats: %v", err)
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
	if err := mgr.StartAgent(context.Background(), *record.AgentID); err != nil {
		t.Fatalf("start agent: %v", err)
	}
	if err := mgr.RestartAgent(context.Background(), *record.AgentID); err != nil {
		t.Fatalf("restart agent: %v", err)
	}
	conn, err := mgr.AttachAgent(*record.AgentID)
	if err != nil {
		t.Fatalf("attach agent: %v", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(line, "ready") {
		t.Fatalf("unexpected output: %s", line)
	}
	_ = conn.Close()
	if err := mgr.KillAgent(context.Background(), *record.AgentID); err != nil {
		t.Fatalf("kill agent: %v", err)
	}
}
