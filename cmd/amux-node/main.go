package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/agentflare-ai/amux/internal/adapter"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/daemon"
	"github.com/agentflare-ai/amux/internal/inference"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/rpc"
)

func main() {
	if len(os.Args) < 2 {
		if err := runDaemon(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	}
	switch os.Args[1] {
	case "daemon":
		if err := runDaemon(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "status":
		if err := runStatus(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "stop":
		if err := runStop(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "version":
		fmt.Printf("amux-manager %s\n", daemon.AmuxVersion)
	default:
		if err := runDaemon(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func runDaemon(args []string) error {
	flags := flag.NewFlagSet("amux-node daemon", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var role string
	var hostID string
	var natsURL string
	var natsCreds string
	var foreground bool
	flags.StringVar(&role, "role", "", "node role (director|manager)")
	flags.StringVar(&hostID, "host-id", "", "host id for manager role")
	flags.StringVar(&natsURL, "nats-url", "", "hub NATS URL")
	flags.StringVar(&natsCreds, "nats-creds", "", "path to NATS creds file")
	flags.BoolVar(&foreground, "foreground", false, "run in foreground")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if !foreground && runtime.GOOS != "windows" {
		if err := daemonize(args); err != nil {
			return fmt.Errorf("daemonize: %w", err)
		}
		return nil
	}
	resolver, err := paths.NewResolverOptionalRepo(".")
	if err != nil {
		return err
	}
	logger := log.New(os.Stderr, "amux-node ", log.LstdFlags)
	cfg, err := config.Load(config.LoadOptions{Resolver: resolver, AdapterDefaults: adapter.NewDefaultsProvider(resolver)})
	if err != nil {
		return err
	}
	applyOverrides(&cfg, role, hostID, natsURL, natsCreds)
	if resolver.RepoRoot() == "" && strings.TrimSpace(cfg.Node.Role) != "manager" {
		return fmt.Errorf("run daemon: %w", paths.ErrRepoRootNotFound)
	}
	if _, err := inference.NewDefaultEngine(resolver.RepoRoot(), logger); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	node, err := daemon.New(ctx, resolver, cfg, logger)
	if err != nil {
		return err
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
		_ = node.Close(context.Background(), false)
		return err
	}
	if err := node.Close(context.Background(), false); err != nil {
		return err
	}
	return nil
}

func runStatus(args []string) error {
	flags := flag.NewFlagSet("amux-node status", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := rpc.Dial(ctx, cfg.Daemon.SocketPath)
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	defer client.Close()
	var result map[string]any
	if err := client.Call(ctx, "daemon.status", nil, &result); err != nil {
		return fmt.Errorf("status: %w", err)
	}
	if hub, ok := result["hub_connected"]; ok {
		fmt.Printf("hub_connected=%v\n", hub)
		return nil
	}
	fmt.Println("hub_connected=false")
	return nil
}

func runStop(args []string) error {
	flags := flag.NewFlagSet("amux-node stop", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	var force bool
	flags.BoolVar(&force, "force", false, "force stop")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := rpc.Dial(ctx, cfg.Daemon.SocketPath)
	if err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	defer client.Close()
	params := map[string]any{"force": force}
	if err := client.Call(ctx, "daemon.stop", params, nil); err != nil {
		return fmt.Errorf("stop: %w", err)
	}
	return nil
}

func loadConfig() (config.Config, error) {
	resolver, err := paths.NewResolverOptionalRepo(".")
	if err != nil {
		return config.Config{}, err
	}
	return config.Load(config.LoadOptions{Resolver: resolver, AdapterDefaults: adapter.NewDefaultsProvider(resolver)})
}

func applyOverrides(cfg *config.Config, role, hostID, natsURL, natsCreds string) {
	if cfg == nil {
		return
	}
	if role != "" {
		cfg.Node.Role = role
	}
	effectiveRole := cfg.Node.Role
	if hostID != "" {
		cfg.Remote.Manager.HostID = hostID
	}
	if natsURL != "" {
		if effectiveRole == "manager" {
			cfg.Remote.NATS.URL = natsURL
		} else {
			cfg.NATS.HubURL = natsURL
		}
	}
	if natsCreds != "" {
		cfg.Remote.NATS.CredsPath = natsCreds
	}
}

func daemonize(args []string) error {
	if os.Getenv("AMUX_DAEMONIZED") == "1" {
		return nil
	}
	childArgs := append([]string{"daemon", "--foreground"}, args...)
	cmd := exec.Command(os.Args[0], childArgs...)
	cmd.Env = append(os.Environ(), "AMUX_DAEMONIZED=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()
	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
