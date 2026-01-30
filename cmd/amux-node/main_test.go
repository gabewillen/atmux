package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/rpc"
)

func TestApplyOverrides(t *testing.T) {
	cfg := config.Config{}
	applyOverrides(&cfg, "manager", "host", "nats://leaf", "/tmp/creds")
	if cfg.Node.Role != "manager" {
		t.Fatalf("expected role override")
	}
	if cfg.Remote.Manager.HostID != "host" {
		t.Fatalf("expected host id override")
	}
	if cfg.Remote.NATS.URL != "nats://leaf" {
		t.Fatalf("expected manager nats override")
	}
	if cfg.Remote.NATS.CredsPath != "/tmp/creds" {
		t.Fatalf("expected creds override")
	}
	applyOverrides(&cfg, "director", "", "nats://hub", "")
	if cfg.NATS.HubURL != "nats://hub" {
		t.Fatalf("expected hub override")
	}
}

func TestRunStatusAndStop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.status", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return map[string]any{"hub_connected": true}, nil
	})
	server.Register("daemon.stop", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return nil, nil
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	statusErr := runStatus(nil)
	_ = w.Close()
	os.Stdout = stdout
	if statusErr != nil {
		t.Fatalf("run status: %v", statusErr)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if got := buf.String(); got == "" {
		t.Fatalf("expected status output")
	}
	if err := runStop([]string{"--force"}); err != nil {
		t.Fatalf("run stop: %v", err)
	}
}

func TestRunStatusMissingField(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.status", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return map[string]any{"other": true}, nil
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	statusErr := runStatus(nil)
	_ = w.Close()
	os.Stdout = stdout
	if statusErr != nil {
		t.Fatalf("run status: %v", statusErr)
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if got := buf.String(); got == "" {
		t.Fatalf("expected status output")
	}
}

func TestRunStopParseError(t *testing.T) {
	if err := runStop([]string{"--bad-flag"}); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestRunStatusDialError(t *testing.T) {
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", filepath.Join(t.TempDir(), "missing.sock"))
	if err := runStatus(nil); err == nil {
		t.Fatalf("expected dial error")
	}
}

func TestApplyOverridesNilConfig(t *testing.T) {
	applyOverrides(nil, "manager", "host", "nats://leaf", "/tmp/creds")
}

func TestRunStopNoForce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.stop", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		if string(params) != `{"force":false}` {
			t.Fatalf("unexpected params: %s", string(params))
		}
		return nil, nil
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	if err := runStop(nil); err != nil {
		t.Fatalf("run stop: %v", err)
	}
}

func TestRunStatusRPCError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.status", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return nil, &rpc.Error{Code: rpc.CodeInternalError, Message: "boom"}
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	if err := runStatus(nil); err == nil {
		t.Fatalf("expected status error")
	}
}

func TestApplyOverridesDirectorNATS(t *testing.T) {
	cfg := config.Config{}
	applyOverrides(&cfg, "director", "", "nats://hub", "")
	if cfg.NATS.HubURL != "nats://hub" {
		t.Fatalf("expected hub override")
	}
}

func TestRunStatusParseError(t *testing.T) {
	if err := runStatus([]string{"--bad-flag"}); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestRunStopRPCError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.stop", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return nil, &rpc.Error{Code: rpc.CodeInternalError, Message: "boom"}
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	if err := runStop([]string{"--force"}); err == nil {
		t.Fatalf("expected stop error")
	}
}

func TestRunDaemonForegroundManagerRepoMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := runDaemon([]string{"--foreground", "--role", "manager"}); err == nil {
		t.Fatalf("expected repo root error")
	}
}

func TestRunDaemonForegroundDirectorRepoMissing(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := runDaemon([]string{"--foreground", "--role", "director"}); err == nil {
		t.Fatalf("expected repo root error")
	}
}


func TestDaemonizeSkipped(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "1")
	if err := daemonize(nil); err != nil {
		t.Fatalf("expected daemonize to skip: %v", err)
	}
}

func TestRunDaemonParseError(t *testing.T) {
	if err := runDaemon([]string{"--bad-flag"}); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestRunDaemonNoRepoDirector(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := runDaemon([]string{"--foreground", "--role", "director"}); err == nil {
		t.Fatalf("expected repo root error")
	}
}

func TestRunDaemonNoRepoManager(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := runDaemon([]string{"--foreground", "--role", "manager"}); err == nil {
		t.Fatalf("expected manager error")
	}
}

func TestMainVersion(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"amux-node", "version"}
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	main()
	_ = w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if buf.Len() == 0 {
		t.Fatalf("expected version output")
	}
}

func TestRunDaemonDaemonizeSkipped(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "1")
	if err := runDaemon(nil); err != nil {
		t.Fatalf("run daemon: %v", err)
	}
	if err := runDaemon([]string{"extra"}); err != nil {
		t.Fatalf("run daemon extra args: %v", err)
	}
}

func TestMainDefaultAndUnknownDaemonizeSkipped(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "1")
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"amux-node"}
	main()
	os.Args = []string{"amux-node", "unknown"}
	main()
}

func TestRunDaemonForegroundInferenceError(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "1")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := runDaemon([]string{"--foreground"}); err == nil {
		t.Fatalf("expected inference error")
	}
}

func TestMainStatusAndStop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "amuxd.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	server := rpc.NewServer(nil)
	server.Register("daemon.status", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return map[string]any{"hub_connected": true}, nil
	})
	server.Register("daemon.stop", func(ctx context.Context, params json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = params
		return nil, nil
	})
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", socketPath)
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"amux-node", "status"}
	main()
	os.Args = []string{"amux-node", "stop"}
	main()
}

func TestDaemonizeError(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "")
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"/nonexistent/amux-node"}
	if err := daemonize(nil); err == nil {
		t.Fatalf("expected daemonize error")
	}
}

func TestRunDaemonDaemonizeError(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "")
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })
	os.Args = []string{"/nonexistent/amux-node"}
	if err := runDaemon(nil); err == nil {
		t.Fatalf("expected runDaemon to return daemonize error")
	}
}

func TestRunDaemonForegroundSocketPathEmpty(t *testing.T) {
	t.Setenv("AMUX_DAEMONIZED", "1")
	t.Setenv("AMUX__DAEMON__SOCKET_PATH", "")
	if err := runDaemon([]string{"--foreground", "--role", "manager"}); err == nil {
		t.Fatalf("expected socket path error")
	}
}

func TestMainExitOnParseError(t *testing.T) {
	if os.Getenv("AMUX_NODE_HELPER") == "1" {
		os.Args = []string{"amux-node", "daemon", "--bad-flag"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitOnParseError")
	cmd.Env = append(os.Environ(), "AMUX_NODE_HELPER=1")
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected non-zero exit")
	}
}

func TestMainStatusExitOnError(t *testing.T) {
	if os.Getenv("AMUX_NODE_HELPER_STATUS") == "1" {
		os.Setenv("AMUX__DAEMON__SOCKET_PATH", filepath.Join(os.TempDir(), "missing.sock"))
		os.Args = []string{"amux-node", "status"}
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainStatusExitOnError")
	cmd.Env = append(os.Environ(), "AMUX_NODE_HELPER_STATUS=1")
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected non-zero exit")
	}
}
