package manager

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

type stubRegistry struct {
	cmd []string
}

type stubAdapter struct {
	name string
	cmd  []string
}

type recordDispatcher struct {
	mu     sync.Mutex
	events []protocol.Event
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
	return stubMatcher{}
}

func (s *stubAdapter) Formatter() adapter.ActionFormatter {
	return stubFormatter{}
}

func (s *stubAdapter) OnEvent(ctx context.Context, event adapter.Event) ([]adapter.Action, error) {
	_ = ctx
	_ = event
	return nil, nil
}

type stubMatcher struct{}

func (stubMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return nil, nil
}

type stubFormatter struct{}

func (stubFormatter) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	return input, nil
}

func (r *recordDispatcher) Publish(ctx context.Context, subject string, event protocol.Event) error {
	_ = ctx
	_ = subject
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

func (r *recordDispatcher) Subscribe(ctx context.Context, subject string, handler func(protocol.Event)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *recordDispatcher) PublishRaw(ctx context.Context, subject string, payload []byte, reply string) error {
	_ = ctx
	_ = subject
	_ = payload
	_ = reply
	return nil
}

func (r *recordDispatcher) SubscribeRaw(ctx context.Context, subject string, handler func(protocol.Message)) (protocol.Subscription, error) {
	_ = ctx
	_ = subject
	_ = handler
	return nil, nil
}

func (r *recordDispatcher) Request(ctx context.Context, subject string, payload []byte, timeout time.Duration) (protocol.Message, error) {
	_ = ctx
	_ = subject
	_ = payload
	_ = timeout
	return protocol.Message{}, nil
}

func (r *recordDispatcher) MaxPayload() int {
	return 1024 * 1024
}

func (r *recordDispatcher) JetStream() nats.JetStreamContext {
	return nil
}

func (r *recordDispatcher) Closed() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func TestAddAgentWritesConfig(t *testing.T) {
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
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
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
		About:    "",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:      repoRoot,
	})
	if err != nil {
		t.Fatalf("add agent: %v", err)
	}
	if record.AgentID == nil || record.AgentID.IsZero() {
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
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "nats")
	server, err := protocol.StartHubServer(context.Background(), protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: cfg.NATS.JetStreamDir,
	})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	mgr, err := NewManager(context.Background(), resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	nonRepo := t.TempDir()
	_, err = mgr.AddAgent(context.Background(), AddRequest{
		Name:     "alpha",
		Adapter:  "stub",
		Location: api.Location{Type: api.LocationLocal},
		Cwd:      nonRepo,
	})
	if err == nil {
		t.Fatalf("expected error for non-repo cwd")
	}
}

func TestAddAgentRequiresRepoPathForMultiRepo(t *testing.T) {
	repoRoot := initRepo(t)
	otherRoot := initRepo(t)
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
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	mgr, err := NewManager(context.Background(), resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	alpha, err := mgr.AddAgent(context.Background(), AddRequest{
		Name:    "alpha",
		Adapter: "stub",
		Location: api.Location{
			Type: api.LocationLocal,
		},
		Cwd: repoRoot,
	})
	if err != nil {
		t.Fatalf("add agent alpha: %v", err)
	}
	t.Cleanup(func() {
		if err := mgr.RemoveAgent(context.Background(), RemoveRequest{AgentID: *alpha.AgentID}); err != nil {
			t.Errorf("remove agent alpha: %v", err)
		}
	})
	if _, err := mgr.AddAgent(context.Background(), AddRequest{
		Name:    "beta",
		Adapter: "stub",
		Location: api.Location{
			Type: api.LocationLocal,
		},
		Cwd: otherRoot,
	}); err == nil {
		t.Fatalf("expected repo_path error for multi-repo add")
	}
	beta, err := mgr.AddAgent(context.Background(), AddRequest{
		Name:    "beta",
		Adapter: "stub",
		Location: api.Location{
			Type:     api.LocationLocal,
			RepoPath: otherRoot,
		},
		Cwd: otherRoot,
	})
	if err != nil {
		t.Fatalf("add agent beta: %v", err)
	}
	t.Cleanup(func() {
		if err := mgr.RemoveAgent(context.Background(), RemoveRequest{AgentID: *beta.AgentID}); err != nil {
			t.Errorf("remove agent beta: %v", err)
		}
	})
}

func TestShutdownEmitsEvents(t *testing.T) {
	dispatcher := &recordDispatcher{}
	mgr := &Manager{
		dispatcher: dispatcher,
		cfg: config.Config{
			Shutdown: config.ShutdownConfig{
				DrainTimeout: 10 * time.Millisecond,
			},
		},
	}
	if err := mgr.Shutdown(context.Background(), false); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	names := map[string]struct{}{}
	dispatcher.mu.Lock()
	for _, event := range dispatcher.events {
		names[event.Name] = struct{}{}
	}
	dispatcher.mu.Unlock()
	if _, ok := names[shutdownEventRequest]; !ok {
		t.Fatalf("expected %s event", shutdownEventRequest)
	}
	if _, ok := names[shutdownEventDrainComplete]; !ok {
		t.Fatalf("expected %s event", shutdownEventDrainComplete)
	}
}

func TestRosterUpdatedEvent(t *testing.T) {
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
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(context.Background(), server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	mgr, err := NewManager(context.Background(), resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestManagerHelperProcess"}}, nil
	})
	events := make(chan protocol.Event, 4)
	sub, err := dispatcher.Subscribe(context.Background(), protocol.Subject("events", "presence"), func(event protocol.Event) {
		if event.Name == "roster.updated" {
			events <- event
		}
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() {
		if err := sub.Unsubscribe(); err != nil {
			t.Errorf("unsubscribe: %v", err)
		}
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
	if record.AgentID == nil {
		t.Fatalf("expected agent id")
	}
	t.Cleanup(func() {
		if err := mgr.RemoveAgent(context.Background(), RemoveRequest{AgentID: *record.AgentID}); err != nil {
			t.Errorf("remove agent: %v", err)
		}
	})
	select {
	case <-events:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected roster.updated event")
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
