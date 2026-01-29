package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/rpc"
)

func connectDaemon(ctx context.Context) (*rpc.Client, *paths.Resolver, config.Config, error) {
	resolver, err := paths.NewResolver(".")
	if err != nil {
		return nil, nil, config.Config{}, fmt.Errorf("connect daemon: %w", err)
	}
	cfg, err := config.Load(config.LoadOptions{Resolver: resolver, AdapterDefaults: adapter.NewDefaultsProvider(resolver)})
	if err != nil {
		return nil, nil, config.Config{}, fmt.Errorf("connect daemon: %w", err)
	}
	client, err := rpc.Dial(ctx, cfg.Daemon.SocketPath)
	if err == nil {
		return client, resolver, cfg, nil
	}
	if !cfg.Daemon.Autostart {
		return nil, nil, config.Config{}, fmt.Errorf("connect daemon: %w", err)
	}
	if err := startDaemon(ctx, resolver, cfg.Daemon.SocketPath); err != nil {
		return nil, nil, config.Config{}, fmt.Errorf("connect daemon: %w", err)
	}
	client, err = rpc.Dial(ctx, cfg.Daemon.SocketPath)
	if err != nil {
		return nil, nil, config.Config{}, fmt.Errorf("connect daemon: %w", err)
	}
	return client, resolver, cfg, nil
}

func startDaemon(ctx context.Context, resolver *paths.Resolver, socketPath string) error {
	cmd := exec.CommandContext(ctx, "amux-node")
	cmd.Dir = resolver.RepoRoot()
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("autostart daemon: %w", err)
	}
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("autostart daemon: %w", err)
	}
	return waitForSocket(ctx, socketPath)
}

func waitForSocket(ctx context.Context, socketPath string) error {
	deadline := time.Now().Add(5 * time.Second)
	if dl, ok := ctx.Deadline(); ok {
		deadline = dl
	}
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("wait for daemon: timeout")
		}
		conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		_ = err
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for daemon: %w", ctx.Err())
		case <-time.After(200 * time.Millisecond):
		}
	}
}
