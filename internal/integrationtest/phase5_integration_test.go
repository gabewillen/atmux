//go:build integration
// +build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/internal/session"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestIntegrationPhase5MonitorAndTUI(t *testing.T) {
	harness, err := NewHarness(t)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	ctx := harness.Context()
	natsContainer, err := harness.StartNATS(ctx, NATSContainerOptions{})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	dispatcher, err := protocol.NewNATSDispatcher(ctx, natsContainer.URL, protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("dispatcher: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(context.Background()); err != nil {
			t.Logf("close dispatcher: %v", err)
		}
	})
	repoRoot := initRepo(t)
	runner := git.NewRunner()
	worktree, err := runner.EnsureWorktree(ctx, repoRoot, "phase5")
	if err != nil {
		t.Fatalf("ensure worktree: %v", err)
	}
	meta, err := api.NewAgent("phase5", "", api.AdapterRef("stub"), repoRoot, worktree.Path, api.Location{Type: api.LocationLocal, RepoPath: repoRoot})
	if err != nil {
		t.Fatalf("agent meta: %v", err)
	}
	runtime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("agent runtime: %v", err)
	}
	sessMeta, err := api.NewSession(meta.ID, repoRoot, worktree.Path, meta.Location)
	if err != nil {
		t.Fatalf("session meta: %v", err)
	}
	cmd := session.Command{
		Argv: []string{os.Args[0], "-test.run=TestIntegrationPhase5Helper"},
		Env:  []string{"AMUX_PHASE5_HELPER=1"},
	}
	sess, err := session.NewLocalSession(sessMeta, runtime, cmd, worktree.Path, stubMatcher{}, dispatcher, session.Config{
		DrainTimeout: 2 * time.Second,
		IdleTimeout:  50 * time.Millisecond,
		StuckTimeout: 120 * time.Millisecond,
		TUIEnabled:   true,
		TUIRows:      5,
		TUICols:      20,
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := sess.Start(ctx); err != nil {
		t.Fatalf("start session: %v", err)
	}
	t.Cleanup(func() {
		if err := sess.Kill(context.Background()); err != nil {
			t.Logf("kill session: %v", err)
		}
	})
	events := make(chan protocol.Event, 16)
	sub, err := dispatcher.Subscribe(ctx, protocol.Subject("events", "pty"), func(event protocol.Event) {
		select {
		case events <- event:
		default:
		}
	})
	if err != nil {
		t.Fatalf("subscribe pty: %v", err)
	}
	t.Cleanup(func() {
		if err := sub.Unsubscribe(); err != nil {
			t.Logf("unsubscribe: %v", err)
		}
	})
	waitForTUIXML(t, sess, "HELLO")
	waitForPTYEvents(t, events, sessMeta.ID.String(), []string{
		agent.EventInactivityDetected,
		agent.EventStuckDetected,
	})
}

func TestIntegrationPhase5Helper(t *testing.T) {
	if os.Getenv("AMUX_PHASE5_HELPER") != "1" {
		return
	}
	_, _ = os.Stdout.Write([]byte("\x1b[?1049h"))
	_, _ = os.Stdout.Write([]byte("HELLO"))
	sigCh := make(chan os.Signal, 1)
	waitForSignal(sigCh)
}

func waitForTUIXML(t *testing.T, sess *session.LocalSession, want string) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			xml := sess.TUIXML()
			if strings.Contains(xml, want) {
				return
			}
		case <-deadline.C:
			t.Fatalf("timeout waiting for TUI XML containing %q", want)
		}
	}
}

func waitForPTYEvents(t *testing.T, events <-chan protocol.Event, sessionID string, names []string) {
	t.Helper()
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	for len(want) > 0 {
		select {
		case event := <-events:
			if _, ok := want[event.Name]; !ok {
				continue
			}
			payload := struct {
				SessionID string `json:"session_id"`
			}{}
			if data, err := json.Marshal(event.Payload); err == nil {
				_ = json.Unmarshal(data, &payload)
			}
			if payload.SessionID != sessionID {
				continue
			}
			delete(want, event.Name)
		case <-timeout.C:
			t.Fatalf("timeout waiting for pty events, remaining=%v", want)
		}
	}
}

func initRepo(t *testing.T) string {
	repoRoot := t.TempDir()
	runner := git.NewRunner()
	if _, err := runner.Exec(context.Background(), repoRoot, "init"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if _, err := runner.Exec(context.Background(), repoRoot, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	if _, err := runner.Exec(context.Background(), repoRoot, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config: %v", err)
	}
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := runner.Exec(context.Background(), repoRoot, "add", "README.md"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if _, err := runner.Exec(context.Background(), repoRoot, "commit", "-m", "init"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	return repoRoot
}

type stubMatcher struct{}

func (stubMatcher) Match(ctx context.Context, output []byte) ([]adapter.PatternMatch, error) {
	_ = ctx
	_ = output
	return nil, nil
}

func waitForSignal(sigCh chan os.Signal) {
	select {
	case <-sigCh:
		return
	case <-time.After(2 * time.Second):
		return
	}
}
