// Package main is the entry point for the amux CLI client.
//
// The amux CLI communicates with the amux daemon (amuxd) over JSON-RPC 2.0
// via a Unix socket. It provides commands for managing agents, plugins,
// and running the test suite.
//
// See spec §12 for the full CLI specification.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/agentflare-ai/amux/internal/cli"
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
		cancel()
	}()

	if err := cli.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// Version returns the amux version string.
func Version() string {
	return "0.1.0-dev"
}

// SpecVersion returns the spec version this implementation targets.
func SpecVersion() string {
	return api.SpecVersion
}
