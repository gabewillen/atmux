package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/daemon"
	"github.com/agentflare-ai/amux/internal/inference"
	"github.com/agentflare-ai/amux/internal/paths"
)

func main() {
	resolver, err := paths.NewResolver(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	logger := log.New(os.Stderr, "amux-node ", log.LstdFlags)
	cfg, err := config.Load(config.LoadOptions{Resolver: resolver, AdapterDefaults: adapter.NewDefaultsProvider(resolver)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if _, err := inference.NewDefaultEngine(resolver.RepoRoot(), logger); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	node, err := daemon.New(ctx, resolver, cfg, logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	go func() {
		count := 0
		for range sigCh {
			count++
			if count == 1 {
				go func() {
					_ = node.Close(context.Background(), false)
					cancel()
				}()
				continue
			}
			forceCtx, forceCancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = node.Close(forceCtx, true)
			forceCancel()
			return
		}
	}()
	if err := node.Serve(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		_ = node.Close(context.Background(), false)
		os.Exit(1)
	}
	if err := node.Close(context.Background(), false); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
