// Package main implements the amux-node daemon per spec §12.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/paths"
	"github.com/stateforward/amux/internal/telemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry
	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: init telemetry: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: shutdown telemetry: %v\n", err)
		}
	}()

	// Initialize config actor
	userConfigPath, err := paths.UserConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: determine user config path: %v\n", err)
		os.Exit(1)
	}

	loader := func() (*config.Config, error) {
		return config.Load(userConfigPath)
	}

	cfgActor, err := config.NewActor(ctx, loader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: init config actor: %v\n", err)
		os.Exit(1)
	}

	_ = cfgActor // Will use this for daemon lifecycle in later phases

	// Phase 2: Basic daemon stub with OTel + config actor initialized
	fmt.Println("amux-node: daemon started (Phase 2)")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("amux-node: shutting down")
}
