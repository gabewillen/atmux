package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/remote/director"
	"github.com/agentflare-ai/amux/internal/remote/manager"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/stateforward/hsm-go/muid"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	type Runner interface {
		Start(context.Context) error
		Stop()
	}

	var runner Runner

	role := cfg.Node.Role
	if role == "" {
		role = "manager"
	}

	switch role {
	case "director":
		runner = director.New(director.Options{
			Config: *cfg,
		})
	case "manager":
		wtMgr, err := worktree.NewManager()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init worktree manager: %v\n", err)
			os.Exit(1)
		}

		// TODO: Load persistent HostID
		hostID := muid.Make().String()
		fmt.Printf("Starting Manager with HostID: %s\n", hostID)

		runner = manager.New(manager.Options{
			Config:   *cfg,
			HostID:   hostID,
			Worktree: wtMgr,
		})
	default:
		fmt.Fprintf(os.Stderr, "Unknown role: %s\n", role)
		os.Exit(1)
	}

	fmt.Printf("amux-node starting (role=%s)...\n", role)
	if err := runner.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start: %v\n", err)
		os.Exit(1)
	}

	// Wait for context cancellation (signal)
	<-ctx.Done()

	fmt.Println("Shutting down...")
	runner.Stop()
	fmt.Println("amux-node stopped")
}
