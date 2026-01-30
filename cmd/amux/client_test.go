package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

func TestConnectDaemonNoRepo(t *testing.T) {
	tmp := t.TempDir()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if _, _, _, err := connectDaemon(context.Background()); err == nil {
		t.Fatalf("expected connect error")
	}
}

func TestConnectDaemonNoAutostart(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath := filepath.Join(repo, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repo, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = false }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	if _, _, _, err := connectDaemon(context.Background()); err == nil {
		t.Fatalf("expected connect error")
	}
}

func TestWaitForSocket(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "daemon.sock")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ready := make(chan struct{})
	go func() {
		timer := time.NewTimer(200 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			close(ready)
			return
		}
		ln, err := net.Listen("unix", socketPath)
		if err != nil {
			return
		}
		close(ready)
		<-ctx.Done()
		_ = ln.Close()
	}()
	<-ready
	if err := waitForSocket(ctx, socketPath); err != nil {
		t.Fatalf("waitForSocket: %v", err)
	}
}

func TestWaitForSocketTimeout(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "missing.sock")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := waitForSocket(ctx, socketPath); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestStartDaemonMissingBinary(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	resolver, err := paths.NewResolver(repo)
	if err != nil {
		t.Fatalf("resolver: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := startDaemon(ctx, resolver, filepath.Join(repo, ".amux", "amuxd.sock")); err == nil {
		t.Fatalf("expected start daemon error")
	}
}

func TestConnectDaemonSuccess(t *testing.T) {
	repoRoot, _, cleanup := setupDaemonSocket(t)
	defer cleanup()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	client, _, _, err := connectDaemon(context.Background())
	if err != nil {
		t.Fatalf("connect daemon: %v", err)
	}
	_ = client.Close()
}

func TestConnectDaemonAutostart(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath := filepath.Join(repo, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repo, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = true }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	binDir := filepath.Join(repo, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	stub := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "amux-node"), []byte(stub), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		timer := time.NewTimer(200 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return
		}
		ln, err := net.Listen("unix", socketPath)
		if err != nil {
			return
		}
		defer ln.Close()
		<-ctx.Done()
	}()
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		t.Fatalf("connect daemon: %v", err)
	}
	_ = client.Close()
}

func TestConnectDaemonAutostartFailure(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}
	socketPath := filepath.Join(repo, ".amux", "amuxd.sock")
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o755); err != nil {
		t.Fatalf("mkdir socket dir: %v", err)
	}
	configPath := filepath.Join(repo, ".amux", "config.toml")
	configData := "daemon = { socket_path = \"" + socketPath + "\", autostart = true }\n"
	if err := os.WriteFile(configPath, []byte(configData), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("HOME", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, _, _, err := connectDaemon(ctx); err == nil {
		t.Fatalf("expected autostart error")
	}
}
