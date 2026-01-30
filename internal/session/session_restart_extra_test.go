package session

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/protocol"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestLocalSessionRestartAndKill(t *testing.T) {
	repoRoot := initRepo(t)
	runner := git.NewRunner()
	ctx := context.Background()
	worktree, err := runner.EnsureWorktree(ctx, repoRoot, "restart")
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
	meta, err := api.NewAgent("restart", "", api.AdapterRef("test"), repoRoot, worktree.Path, api.Location{Type: api.LocationLocal, RepoPath: repoRoot})
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
	sess, err := NewLocalSession(sessMeta, runtime, cmd, worktree.Path, stubMatcher{}, dispatcher, Config{DrainTimeout: 2 * time.Second})
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
	readLineWithTimeout(t, conn, 2*time.Second)
	_ = conn.Close()
	if err := sess.Restart(ctx); err != nil {
		t.Fatalf("restart session: %v", err)
	}
	conn, err = sess.Attach()
	if err != nil {
		t.Fatalf("attach after restart: %v", err)
	}
	readLineWithTimeout(t, conn, 2*time.Second)
	_ = conn.Close()
	if err := sess.Kill(ctx); err != nil {
		t.Fatalf("kill session: %v", err)
	}
	select {
	case <-sess.Done():
	case <-time.After(2 * time.Second):
		t.Fatalf("expected session to stop after kill")
	}
}

func readLineWithTimeout(t *testing.T, conn net.Conn, timeout time.Duration) string {
	t.Helper()
	reader := bufio.NewReader(conn)
	lineCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		line, err := reader.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		lineCh <- line
	}()
	select {
	case line := <-lineCh:
		return line
	case err := <-errCh:
		t.Fatalf("read output: %v", err)
	case <-time.After(timeout):
		t.Fatalf("timeout reading output")
	}
	return ""
}
