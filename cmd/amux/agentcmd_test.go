package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentflare-ai/amux/internal/rpc"
	"github.com/agentflare-ai/amux/pkg/api"
)

func setupDaemonSocket(t *testing.T) (repoRoot string, socketPath string, cleanup func()) {
	t.Helper()
	repoRoot = t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath = filepath.Join(repoRoot, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repoRoot, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = false }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := rpc.NewServer(nil)
	server.Register("agent.add", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"agent_id": "1"}, nil
	})
	server.Register("agent.list", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"roster": []api.RosterEntry{{Kind: api.RosterAgent, RuntimeID: api.NewRuntimeID(), Name: "alpha", Presence: "online"}}}, nil
	})
	server.Register("agent.remove", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	server.Register("agent.start", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	server.Register("agent.stop", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	server.Register("agent.kill", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	server.Register("agent.restart", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	server.Register("agent.attach", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"socket_path": filepath.Join(repoRoot, "missing.sock")}, nil
	})
	server.Register("git.merge", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"ok": true}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	cleanup = func() {
		cancel()
		_ = listener.Close()
	}
	return repoRoot, socketPath, cleanup
}

func captureStdout(t *testing.T) func() string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	return func() string {
		_ = w.Close()
		os.Stdout = orig
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		return buf.String()
	}
}

func TestRunAgentCommands(t *testing.T) {
	repoRoot, _, cleanup := setupDaemonSocket(t)
	defer cleanup()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	out := captureStdout(t)
	if err := runAgentAdd([]string{"--name", "alpha", "--adapter", "stub"}); err != nil {
		t.Fatalf("runAgentAdd: %v", err)
	}
	if !strings.Contains(out(), "1") {
		t.Fatalf("expected agent id output")
	}
	out = captureStdout(t)
	if err := runAgentList(nil); err != nil {
		t.Fatalf("runAgentList: %v", err)
	}
	if !strings.Contains(out(), "alpha") {
		t.Fatalf("expected roster output")
	}
	if err := runAgentRemove([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentRemove: %v", err)
	}
	if err := runAgentStart([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentStart: %v", err)
	}
	if err := runAgentStop([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentStop: %v", err)
	}
	if err := runAgentKill([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentKill: %v", err)
	}
	if err := runAgentRestart([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentRestart: %v", err)
	}
	if err := runGitMerge([]string{"--id", "1"}); err != nil {
		t.Fatalf("runGitMerge: %v", err)
	}
	if err := runAgentAttach([]string{"--id", "1"}); err == nil {
		t.Fatalf("expected attach error")
	}
}

func TestRunAgentListAgentID(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath := filepath.Join(repoRoot, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repoRoot, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = false }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := rpc.NewServer(nil)
	agentID := api.NewAgentID()
	server.Register("agent.list", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"roster": []api.RosterEntry{{Kind: api.RosterAgent, RuntimeID: api.NewRuntimeID(), AgentID: &agentID, Name: "alpha", Presence: "online"}}}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = server.Serve(ctx, listener)
	}()
	defer func() {
		cancel()
		_ = listener.Close()
	}()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	out := captureStdout(t)
	if err := runAgentList(nil); err != nil {
		t.Fatalf("runAgentList: %v", err)
	}
	if !strings.Contains(out(), agentID.String()) {
		t.Fatalf("expected agent id output")
	}
}

func TestRunAgentDispatch(t *testing.T) {
	repoRoot, _, cleanup := setupDaemonSocket(t)
	defer cleanup()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	if err := runAgent([]string{"add", "--name", "alpha", "--adapter", "stub"}); err != nil {
		t.Fatalf("runAgent add: %v", err)
	}
	if err := runAgent([]string{"list"}); err != nil {
		t.Fatalf("runAgent list: %v", err)
	}
	if err := runAgent([]string{"remove", "--id", "1"}); err != nil {
		t.Fatalf("runAgent remove: %v", err)
	}
	if err := runAgent([]string{"start", "--id", "1"}); err != nil {
		t.Fatalf("runAgent start: %v", err)
	}
	if err := runAgent([]string{"stop", "--id", "1"}); err != nil {
		t.Fatalf("runAgent stop: %v", err)
	}
	if err := runAgent([]string{"kill", "--id", "1"}); err != nil {
		t.Fatalf("runAgent kill: %v", err)
	}
	if err := runAgent([]string{"restart", "--id", "1"}); err != nil {
		t.Fatalf("runAgent restart: %v", err)
	}
}

func TestRunAgentAttachSuccess(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath := filepath.Join(repoRoot, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repoRoot, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = false }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	attachSocket := filepath.Join(repoRoot, "attach.sock")
	ln, err := net.Listen("unix", attachSocket)
	if err != nil {
		t.Fatalf("listen attach: %v", err)
	}
	defer ln.Close()
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := rpc.NewServer(nil)
	server.Register("agent.attach", func(ctx context.Context, raw json.RawMessage) (any, *rpc.Error) {
		_ = ctx
		_ = raw
		return map[string]any{"socket_path": attachSocket}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = server.Serve(ctx, listener) }()
	defer func() { _ = listener.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_, _ = conn.Write([]byte("hello"))
		_ = conn.Close()
	}()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = w.Close() }()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()
	out := captureStdout(t)
	if err := runAgentAttach([]string{"--id", "1"}); err != nil {
		t.Fatalf("runAgentAttach: %v", err)
	}
	if !strings.Contains(out(), "hello") {
		t.Fatalf("expected attach output")
	}
}

func TestParseAgentRefFlags(t *testing.T) {
	if _, err := parseAgentRefFlags("agent", nil); err == nil {
		t.Fatalf("expected missing id error")
	}
	params, err := parseAgentRefFlags("agent", []string{"--id", "1", "--name", "alpha"})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if params["agent_id"] != "1" || params["name"] != "alpha" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestRunAgentRefCommandMissingRef(t *testing.T) {
	if err := runAgentRefCommand(nil, "agent.start"); err == nil {
		t.Fatalf("expected missing ref error")
	}
}

func TestRunAgentAddInvalidFlag(t *testing.T) {
	if err := runAgentAdd([]string{"--bad"}); err == nil {
		t.Fatalf("expected flag error")
	}
}

func TestRunAgentInvalid(t *testing.T) {
	if err := runAgent(nil); err == nil {
		t.Fatalf("expected usage error")
	}
	if err := runAgent([]string{"unknown"}); err == nil {
		t.Fatalf("expected unknown command error")
	}
}
