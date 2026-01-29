package manager

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

type stubRegistry struct {
	cmd []string
}

type stubAdapter struct {
	name string
	cmd  []string
}

func (s *stubRegistry) Load(ctx context.Context, name string) (adapter.Adapter, error) {
	_ = ctx
	return &stubAdapter{name: name, cmd: s.cmd}, nil
}

func (s *stubAdapter) Name() string {
	return s.name
}

func (s *stubAdapter) Manifest() adapter.Manifest {
	return adapter.Manifest{
		Name: s.name,
		Commands: adapter.AdapterCommands{
			Start: s.cmd,
		},
	}
}

func (s *stubAdapter) Matcher() adapter.PatternMatcher {
	return &adapter.NoopMatcher{}
}

func (s *stubAdapter) Formatter() adapter.ActionFormatter {
	return &adapter.NoopFormatter{}
}

func TestAddAgentWritesConfig(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	server, err := protocol.StartEmbeddedServer(context.Background(), "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	mgr, err := NewLocalManager(context.Background(), resolver, cfg, dispatcher)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	record, err := mgr.AddAgent(context.Background(), AddRequest{
		Name:    "alpha",
		About:   "",
		Adapter: "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:     repoRoot,
	})
	if err != nil {
		t.Fatalf("add agent: %v", err)
	}
	if record.ID.IsZero() {
		t.Fatalf("expected agent id")
	}
	configPath := resolver.ProjectConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "[[agents]]") {
		t.Fatalf("expected agents table")
	}
	if err := mgr.RemoveAgent(context.Background(), RemoveRequest{Name: "alpha"}); err != nil {
		t.Fatalf("remove agent: %v", err)
	}
}

func TestAddAgentRequiresRepo(t *testing.T) {
	repoRoot := initRepo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	server, err := protocol.StartEmbeddedServer(context.Background(), "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	mgr, err := NewLocalManager(context.Background(), resolver, cfg, dispatcher)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	nonRepo := t.TempDir()
	_, err = mgr.AddAgent(context.Background(), AddRequest{
		Name:    "alpha",
		Adapter: "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:     nonRepo,
	})
	if err == nil {
		t.Fatalf("expected error for non-repo cwd")
	}
}

func TestManagerHelperProcess(t *testing.T) {
	if os.Getenv("AMUX_HELPER") != "1" {
		return
	}
	writer := bufio.NewWriter(os.Stdout)
	_, _ = writer.WriteString("ready\n")
	_ = writer.Flush()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
}

func initRepo(t *testing.T) string {
	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.email", "test@example.com")
	runGit(t, repoRoot, "config", "user.name", "Test")
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoRoot, "add", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
	return repoRoot
}

func runGit(t *testing.T, dir string, args ...string) {
	runner := git.NewRunner()
	result, err := runner.Exec(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}
