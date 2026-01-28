// Package daemon implements the amux daemon (amuxd) and manager (amux-manager).
//
// The daemon serves a JSON-RPC 2.0 control plane over a Unix socket,
// manages agent lifecycles, and coordinates with remote hosts.
//
// The role (director vs manager) is determined by the node.role configuration:
//   - director: Runs the amux director with hub-mode NATS
//   - manager: Runs as a host manager with leaf-mode NATS
//
// See spec §12 for the full daemon specification.
package daemon

import (
	"context"
	"fmt"
	"os"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Version is the daemon version string.
const Version = "0.1.0-dev"

// Run starts the daemon with the given arguments.
func Run(ctx context.Context, args []string) error {
	// Parse arguments
	if len(args) > 0 {
		switch args[0] {
		case "help", "--help", "-h":
			return showHelp()
		case "version", "--version", "-v":
			return showVersion()
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Determine role
	role := cfg.Node.Role
	if role == "" {
		role = "director"
	}

	fmt.Fprintf(os.Stderr, "amux-node starting (role=%s, spec=%s)\n", role, api.SpecVersion)

	// Start the appropriate role
	switch role {
	case "director":
		return runDirector(ctx, cfg)
	case "manager":
		return runManager(ctx, cfg)
	default:
		return fmt.Errorf("unknown node role: %s", role)
	}
}

func showHelp() error {
	help := `amux-node - Agent Multiplexer Daemon

Usage:
  amux-node [options]

The daemon role is determined by the 'node.role' configuration:
  - director: Run as the amux director with hub-mode NATS
  - manager:  Run as a host manager with leaf-mode NATS

Options:
  --help, -h      Show this help
  --version, -v   Show version information

Configuration:
  User config:    ~/.config/amux/config.toml
  Project config: .amux/config.toml
  Environment:    AMUX__* variables

Spec version: ` + api.SpecVersion

	fmt.Println(help)
	return nil
}

func showVersion() error {
	fmt.Printf("amux-node version %s\n", Version)
	fmt.Printf("spec version %s\n", api.SpecVersion)
	return nil
}

func runDirector(ctx context.Context, cfg *config.Config) error {
	fmt.Fprintln(os.Stderr, "Director mode: not yet implemented")

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "Director shutting down")
	return nil
}

func runManager(ctx context.Context, cfg *config.Config) error {
	fmt.Fprintln(os.Stderr, "Manager mode: not yet implemented")

	// Wait for context cancellation
	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "Manager shutting down")
	return nil
}
