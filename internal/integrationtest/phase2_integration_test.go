//go:build integration
// +build integration

package integrationtest

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/manager"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

type phase2Registry struct {
	cmd []string
}

type phase2Adapter struct {
	name string
	cmd  []string
}

type phase2Matcher struct{}

type phase2Formatter struct{}

func (r *phase2Registry) Load(ctx context.Context, name string) (adapter.Adapter, error) {
	_ = ctx
	return &phase2Adapter{name: name, cmd: r.cmd}, nil
}

func (a *phase2Adapter) Name() string {
	return a.name
}

func (a *phase2Adapter) Manifest() adapter.Manifest {
	return adapter.Manifest{
		Name: a.name,
		Commands: adapter.AdapterCommands{
			Start: a.cmd,
		},
	}
}

func (a *phase2Adapter) Matcher() adapter.PatternMatcher {
	return phase2Matcher{}
}

func (a *phase2Adapter) Formatter() adapter.ActionFormatter {
	return phase2Formatter{}
}

func (a *phase2Adapter) OnEvent(ctx context.Context, event adapter.Event) ([]adapter.Action, error) {
	_ = ctx
	_ = event
	return nil, nil
}

func (phase2Matcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return nil, nil
}

func (phase2Formatter) Format(ctx context.Context, input string) (string, error) {
	_ = ctx
	return input, nil
}

func TestIntegrationPhase2LocalLifecycle(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx, NATSContainerOptions{})
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	repoRoot := initPhase2Repo(t)
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "nats")
	dispatcher, err := protocol.NewNATSDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Errorf("dispatcher close: %v", err)
		}
	})
	mgr, err := manager.NewManager(ctx, resolver, cfg, dispatcher, "test")
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	mgr.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &phase2Registry{
			cmd: []string{"env", "AMUX_HELPER=1", os.Args[0], "-test.run=TestIntegrationPhase2Helper"},
		}, nil
	})
	record, err := mgr.AddAgent(ctx, manager.AddRequest{
		Name:     "alpha",
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
	if record.Worktree == "" {
		t.Fatalf("expected worktree path")
	}
	if _, err := os.Stat(record.Worktree); err != nil {
		t.Fatalf("worktree missing: %v", err)
	}
	configPath := resolver.ProjectConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "[[agents]]") {
		t.Fatalf("expected agents table")
	}
	conn, err := mgr.AttachAgent(*record.AgentID)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("deadline: %v", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read attach: %v", err)
	}
	if !strings.Contains(line, "ready") {
		t.Fatalf("expected ready output, got %q", line)
	}
	if err := mgr.StopAgent(ctx, *record.AgentID); err != nil {
		t.Fatalf("stop agent: %v", err)
	}
	if err := mgr.RemoveAgent(ctx, manager.RemoveRequest{AgentID: *record.AgentID}); err != nil {
		t.Fatalf("remove agent: %v", err)
	}
	if _, err := os.Stat(record.Worktree); err == nil {
		t.Fatalf("expected worktree removed")
	}
}

func TestIntegrationPhase2Helper(t *testing.T) {
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

func initPhase2Repo(t *testing.T) string {
	repoRoot := t.TempDir()
	runPhase2Git(t, repoRoot, "init")
	runPhase2Git(t, repoRoot, "config", "user.email", "test@example.com")
	runPhase2Git(t, repoRoot, "config", "user.name", "Test")
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runPhase2Git(t, repoRoot, "add", "README.md")
	runPhase2Git(t, repoRoot, "commit", "-m", "init")
	return repoRoot
}

func runPhase2Git(t *testing.T, dir string, args ...string) {
	runner := git.NewRunner()
	result, err := runner.Exec(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}
