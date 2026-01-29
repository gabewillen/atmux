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
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/remote/director"
	"github.com/agentflare-ai/amux/internal/remote/manager"
	"github.com/agentflare-ai/amux/internal/remote/natsconn"
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

	// Override role and host-id from arguments
	parseArgs(cfg, args)

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
  --role <role>   Set the daemon role (director|manager)
  --host-id <id>  Set the host identifier
  --nats-url <url>  Set the NATS server URL
  --nats-creds <path>  Set the NATS credentials file path

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

// parseArgs overrides config values from command-line arguments.
func parseArgs(cfg *config.Config, args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--role":
			if i+1 < len(args) {
				cfg.Node.Role = args[i+1]
				i++
			}
		case "--host-id":
			if i+1 < len(args) {
				// Store host-id in NATS config; it's not in a dedicated field
				// so we use an environment-based approach or pass it through
				os.Setenv("AMUX_HOST_ID", args[i+1])
				i++
			}
		case "--nats-url":
			if i+1 < len(args) {
				cfg.Remote.NATS.URL = args[i+1]
				i++
			}
		case "--nats-creds":
			if i+1 < len(args) {
				cfg.Remote.NATS.CredsPath = args[i+1]
				i++
			}
		}
	}
}

// getHostID returns the host identifier from env or generates one.
func getHostID() string {
	if id := os.Getenv("AMUX_HOST_ID"); id != "" {
		return id
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func runDirector(ctx context.Context, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	dispatcher := event.NewLocalDispatcher()
	hostID := getHostID()

	// Connect to NATS
	opts := natsconn.OptionsFromConfig(cfg, hostID)
	opts.DisconnectHandler = func(nc *nats.Conn, err error) {
		fmt.Fprintf(os.Stderr, "NATS disconnected: %v\n", err)
	}
	opts.ReconnectHandler = func(nc *nats.Conn) {
		fmt.Fprintln(os.Stderr, "NATS reconnected")
	}

	conn, err := natsconn.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("director nats connect: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(os.Stderr, "Director connected to NATS (host=%s)\n", hostID)

	// Create and start the director
	dir := director.New(conn, cfg, dispatcher)
	if err := dir.Start(ctx); err != nil {
		return fmt.Errorf("director start: %w", err)
	}
	defer func() { _ = dir.Stop() }()

	fmt.Fprintln(os.Stderr, "Director ready")

	// Wait for shutdown signal
	select {
	case sig := <-sigCh:
		fmt.Fprintf(os.Stderr, "Director received signal: %s, shutting down\n", sig)
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "Director context cancelled, shutting down")
	}

	return nil
}

func runManager(ctx context.Context, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	dispatcher := event.NewLocalDispatcher()
	hostID := getHostID()

	// Connect to NATS
	opts := natsconn.OptionsFromConfig(cfg, hostID)

	var mgr *manager.Manager

	opts.DisconnectHandler = func(nc *nats.Conn, err error) {
		fmt.Fprintf(os.Stderr, "NATS disconnected: %v\n", err)
		if mgr != nil {
			mgr.SetHubConnected(false)
		}
	}
	opts.ReconnectHandler = func(nc *nats.Conn) {
		fmt.Fprintln(os.Stderr, "NATS reconnected")
		if mgr != nil {
			mgr.SetHubConnected(true)
		}
	}

	conn, err := natsconn.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("manager nats connect: %w", err)
	}
	defer conn.Close()

	fmt.Fprintf(os.Stderr, "Manager connected to NATS (host=%s)\n", hostID)

	// Create and start the manager
	mgr = manager.New(conn, cfg, hostID, dispatcher)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager start: %w", err)
	}
	defer func() { _ = mgr.Stop() }()

	fmt.Fprintln(os.Stderr, "Manager ready")

	// Wait for shutdown signal
	select {
	case sig := <-sigCh:
		fmt.Fprintf(os.Stderr, "Manager received signal: %s, shutting down\n", sig)
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "Manager context cancelled, shutting down")
	}

	return nil
}
