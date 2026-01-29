package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	cfg, err := config.Load(config.LoadOptions{Resolver: resolver})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if _, err := inference.NewDefaultEngine(resolver.RepoRoot(), logger); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	node, err := daemon.New(ctx, resolver, cfg, logger)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := node.Serve(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		_ = node.Close(ctx)
		os.Exit(1)
	}
	if err := node.Close(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
