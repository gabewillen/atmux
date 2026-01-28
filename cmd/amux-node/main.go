// Package main is the entry point for the amux unified node binary.
//
// This binary serves as both the daemon (amuxd) and manager (amux-manager)
// depending on configuration. The role is determined by the node.role
// configuration option:
//   - director: Runs the amux director with hub-mode NATS
//   - manager: Runs as a host manager with leaf-mode NATS
//
// See spec §3.44-§3.46 and §12 for the full specification.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agentflare-ai/amux/internal/daemon"
	"github.com/agentflare-ai/amux/pkg/api"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "Received shutdown signal")
		cancel()
	}()

	if err := daemon.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// Version returns the amux-node version string.
func Version() string {
	return "0.1.0-dev"
}

// SpecVersion returns the spec version this implementation targets.
func SpecVersion() string {
	return api.SpecVersion
}
