package session

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestLocalSessionStartStop(t *testing.T) {
	repoRoot := initRepo(t)
	runner := git.NewRunner()
	ctx := context.Background()
	worktree, err := runner.EnsureWorktree(ctx, repoRoot, "alpha")
	if err != nil {
		t.Fatalf("ensure worktree: %v", err)
	}
	jetstreamDir := filepath.Join(t.TempDir(), "nats")
	server, err := protocol.StartHubServer(ctx, protocol.HubServerConfig{
		Listen:       "127.0.0.1:-1",
		JetStreamDir: jetstreamDir,
	})
	if err != nil {
		t.Fatalf("start nats: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("close nats: %v", err)
		}
	})
	dispatcher, err := protocol.NewNATSDispatcher(ctx, server.URL(), protocol.NATSOptions{})
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	t.Cleanup(func() {
		if err := dispatcher.Close(ctx); err != nil {
			t.Errorf("close dispatcher: %v", err)
		}
	})
	meta, err := api.NewAgent("alpha", "", api.AdapterRef("test"), repoRoot, worktree.Path, api.Location{Type: api.LocationLocal, RepoPath: repoRoot})
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	runtime, err := agent.NewAgent(meta, dispatcher)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	sessMeta, err := api.NewSession(meta.ID, repoRoot, worktree.Path, meta.Location)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	cmd := Command{
		Argv: []string{os.Args[0], "-test.run=TestHelperProcess"},
		Env:  []string{"AMUX_HELPER=1"},
	}
	sess, err := NewLocalSession(sessMeta, runtime, cmd, worktree.Path, &adapter.NoopMatcher{}, dispatcher, Config{DrainTimeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := sess.Start(ctx); err != nil {
		t.Fatalf("start session: %v", err)
	}
	conn, err := sess.Attach()
	if err != nil {
		t.Fatalf("attach session: %v", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(line, worktree.Path) {
		t.Fatalf("unexpected output: %s", line)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close pty: %v", err)
	}
	if err := sess.Stop(ctx); err != nil {
		t.Fatalf("stop session: %v", err)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("AMUX_HELPER") != "1" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stdout, "cwd=unknown\n")
	} else {
		fmt.Fprintf(os.Stdout, "cwd=%s\n", cwd)
	}
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
	result, err := execGit(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}

func execGit(ctx context.Context, dir string, args ...string) (git.ExecResult, error) {
	runner := git.NewRunner()
	return runner.Exec(ctx, dir, args...)
}
