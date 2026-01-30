package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/rpc"
	"github.com/agentflare-ai/amux/pkg/api"
	"os/signal"
	"syscall"
)

func TestDaemonLifecycle(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "jetstream")
	cfg.NATS.Listen = "127.0.0.1:-1"
	cfg.NATS.AdvertiseURL = ""
	cfg.NATS.LeafAdvertiseURL = ""
	cfg.Daemon.SocketPath = filepath.Join(t.TempDir(), "amuxd.sock")
	logger := log.New(os.Stderr, "test ", log.LstdFlags)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	daemon, err := New(ctx, resolver, cfg, logger)
	if err != nil {
		t.Fatalf("new daemon: %v", err)
	}
	go func() {
		_ = daemon.Serve(ctx)
	}()
	if err := waitForSocket(ctx, cfg.Daemon.SocketPath); err != nil {
		t.Fatalf("wait for socket: %v", err)
	}
	client, err := rpc.Dial(context.Background(), cfg.Daemon.SocketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()
	var ping map[string]any
	if err := client.Call(context.Background(), "daemon.ping", nil, &ping); err != nil {
		t.Fatalf("ping: %v", err)
	}
	var status map[string]any
	if err := client.Call(context.Background(), "daemon.status", nil, &status); err != nil {
		t.Fatalf("status: %v", err)
	}
	if err := daemon.Close(context.Background(), true); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestDaemonServeErrors(t *testing.T) {
	var d *Daemon
	if err := d.Serve(context.Background()); err == nil {
		t.Fatalf("expected nil daemon error")
	}
	daemon := &Daemon{}
	if err := daemon.Serve(context.Background()); err == nil {
		t.Fatalf("expected socket path error")
	}
}

func TestDaemonHandlers(t *testing.T) {
	repo := t.TempDir()
	initRepo(t, repo)
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	cfg := config.DefaultConfig(resolver)
	cfg.NATS.JetStreamDir = filepath.Join(t.TempDir(), "jetstream")
	cfg.NATS.Listen = "127.0.0.1:-1"
	cfg.NATS.AdvertiseURL = ""
	cfg.NATS.LeafAdvertiseURL = ""
	cfg.Daemon.SocketPath = filepath.Join(t.TempDir(), "amuxd.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	daemon, err := New(ctx, resolver, cfg, log.New(os.Stderr, "test ", log.LstdFlags))
	if err != nil {
		t.Fatalf("new daemon: %v", err)
	}
	daemon.manager.SetRegistryFactory(func(resolver *paths.Resolver) (adapter.Registry, error) {
		_ = resolver
		return &stubRegistry{cmd: []string{"env", "AMUX_DAEMON_HELPER=1", os.Args[0], "-test.run=TestDaemonHelperProcess"}}, nil
	})
	raw, _ := json.Marshal(map[string]any{})
	if _, err := daemon.handlePing(context.Background(), raw); err != nil {
		t.Fatalf("ping handler: %v", err)
	}
	if _, err := daemon.handleVersion(context.Background(), raw); err != nil {
		t.Fatalf("version handler: %v", err)
	}
	addParams := agentAddParams{
		Name:    "alpha",
		Adapter: "stub",
		Location: locationParam{
			Type:     "local",
			RepoPath: repo,
		},
		Cwd: repo,
	}
	addRaw, _ := json.Marshal(addParams)
	addResp, rpcErr := daemon.handleAgentAdd(context.Background(), addRaw)
	if rpcErr != nil {
		t.Fatalf("agent add: %v", rpcErr)
	}
	addResult := addResp.(agentAddResult)
	refRaw, _ := json.Marshal(agentRefParams{AgentID: addResult.AgentID.String()})
	if _, rpcErr := daemon.handleAgentList(context.Background(), raw); rpcErr != nil {
		t.Fatalf("agent list: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentStart(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent start: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentAttach(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent attach: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentStop(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent stop: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentKill(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent kill: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentRestart(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent restart: %v", rpcErr)
	}
	if _, rpcErr := daemon.handleAgentRemove(context.Background(), refRaw); rpcErr != nil {
		t.Fatalf("agent remove: %v", rpcErr)
	}
}

func TestAttachProxy(t *testing.T) {
	d := &Daemon{}
	repo := t.TempDir()
	streamA, streamB := net.Pipe()
	defer streamA.Close()
	socketPath, err := d.startAttachProxy(context.Background(), repo, api.NewAgentID(), streamA)
	if err != nil {
		t.Fatalf("start attach proxy: %v", err)
	}
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	go func() {
		_, _ = streamB.Write([]byte("hello"))
	}()
	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("unexpected payload: %s", buf)
	}
}

func TestAttachProxyNilStream(t *testing.T) {
	d := &Daemon{}
	if _, err := d.startAttachProxy(context.Background(), t.TempDir(), api.NewAgentID(), nil); err == nil {
		t.Fatalf("expected nil stream error")
	}
}

func TestDaemonHelperProcess(t *testing.T) {
	if os.Getenv("AMUX_DAEMON_HELPER") != "1" {
		return
	}
	_, _ = io.WriteString(os.Stdout, "ready\n")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
}

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

func (s *stubAdapter) Name() string { return s.name }

func (s *stubAdapter) Manifest() adapter.Manifest {
	return adapter.Manifest{Name: s.name, Commands: adapter.AdapterCommands{Start: s.cmd}}
}

func (s *stubAdapter) Matcher() adapter.PatternMatcher { return stubMatcher{} }
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

func initRepo(t *testing.T, repoRoot string) {
	runGit(t, repoRoot, "init")
	runGit(t, repoRoot, "config", "user.email", "test@example.com")
	runGit(t, repoRoot, "config", "user.name", "Test")
	path := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoRoot, "add", "README.md")
	runGit(t, repoRoot, "commit", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	runner := git.NewRunner()
	result, err := runner.Exec(context.Background(), dir, args...)
	if err != nil {
		t.Fatalf("git %v: %v (output: %s)", args, err, string(result.Output))
	}
}

func TestDecodeParamsHelpers(t *testing.T) {
	if err := decodeParams(nil, &struct{}{}); err != nil {
		t.Fatalf("expected nil raw ok")
	}
	if err := decodeParams([]byte("null"), &struct{}{}); err != nil {
		t.Fatalf("expected null raw ok")
	}
	if err := decodeParams([]byte("{"), &struct{}{}); err == nil {
		t.Fatalf("expected decode error")
	}
	invalid := rpcInvalidParams(fmt.Errorf("bad"))
	if invalid.Code != rpc.CodeInvalidParams {
		t.Fatalf("unexpected invalid params code")
	}
	internal := rpcInternal(fmt.Errorf("bad"))
	if internal.Code != rpc.CodeInternalError {
		t.Fatalf("unexpected internal code")
	}
}

func waitForSocket(ctx context.Context, socketPath string) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			return context.DeadlineExceeded
		}
		conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}
