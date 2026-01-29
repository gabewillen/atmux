package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/daemon"
	"github.com/agentflare-ai/amux/internal/remote"
	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/nats.go"
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file (currently ignored, uses standard paths)")
	flag.Parse()

	if *configPath != "" {
		log.Printf("Loading config from %s (override not fully implemented)", *configPath)
	}

	// 1. Load Config
	repoRoot, _ := os.Getwd() // Default to cwd as repo root for now
	cfg, err := config.Load(repoRoot)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Setup Signal Handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 3. Start JSON-RPC Server (The Daemon)
	// This listens on the unix socket for CLI commands.
	srv := daemon.NewServer(cfg.Daemon)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Failed to start daemon server: %v", err)
	}
	log.Printf("Daemon listening on %s", cfg.Daemon.SocketPath)

	// 4. Remote Orchestration (Phase 3)
	// Check Role
	role := cfg.Node.Role
	log.Printf("Starting node in role: %s", role)

	var nc *nats.Conn

	if role == "director" {
		// Director Mode
		// A. Configure Embedded NATS (if applicable)
		if cfg.NATS.Mode == "embedded" {
			if err := remote.ConfigureEmbeddedHub(cfg); err != nil {
				log.Printf("Warning: Failed to configure embedded NATS hub: %v", err)
			} else {
				log.Println("Embedded NATS hub configuration generated.")
			}
		}

		// B. Connect to NATS
		nc, err = nats.Connect(cfg.Remote.NATS.URL, nats.Name("amux-director"))
		if err != nil {
			log.Printf("Director failed to connect to NATS at %s: %v", cfg.Remote.NATS.URL, err)
		} else {
			log.Printf("Director connected to NATS at %s", cfg.Remote.NATS.URL)
			
			// C. Start Director Protocol
			dp := remote.NewDirectorProtocol(nc, cfg.Remote.NATS.SubjectPrefix)
			if err := dp.Start(ctx); err != nil {
				log.Fatalf("Failed to start Director Protocol: %v", err)
			}
			log.Println("Director Protocol started")
		}

	} else if role == "manager" {
		// Manager Mode
		// Connect to NATS (Director's Hub)
		opts := []nats.Option{
			nats.Name("amux-manager-" + hostname()),
			nats.ReconnectWait(2 * time.Second),
			nats.MaxReconnects(cfg.Remote.ReconnectMaxAttempts),
		}
		
		if cfg.Remote.NATS.CredsPath != "" {
			opts = append(opts, nats.UserCredentials(cfg.Remote.NATS.CredsPath))
		}

		nc, err = nats.Connect(cfg.Remote.NATS.URL, opts...)
		if err != nil {
			log.Printf("Manager failed to connect to NATS: %v", err)
		} else {
			log.Printf("Manager connected to NATS at %s", cfg.Remote.NATS.URL)
			
			// Start Manager
			hostID := api.HostID(hostname()) // Use hostname as HostID for now
			mgr := remote.NewManager(cfg, hostID)
			
			// Manager.Start is blocking, run in goroutine
			go func() {
				if err := mgr.Start(ctx, nc); err != nil {
					log.Printf("Manager error: %v", err)
					cancel() // Stop daemon on manager failure?
				}
			}()
		}
	}

	// Wait for shutdown
	<-sigChan
	log.Println("Shutting down...")
	
	if nc != nil {
		nc.Close()
	}
	srv.Stop()
}

func hostname() string {
	h, _ := os.Hostname()
	if h == "" {
		return "unknown-host"
	}
	return h
}
