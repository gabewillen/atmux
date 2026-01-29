// Package main provides the unified daemon binary.
// This binary serves as both amuxd (daemon) and amux-manager roles,
// with the active role determined by configuration and/or flags.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/copilot-claude-sonnet-4/amux/internal/config"
	"github.com/copilot-claude-sonnet-4/amux/internal/remote"
)

const version = "v1.22.0-phase3"

// Command line flags
var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	roleFlag    = flag.String("role", "", "Role to run as (director|manager)")
	hostIDFlag  = flag.String("host-id", "", "Host identifier")
	natsURLFlag = flag.String("nats-url", "", "NATS server URL")
	natsCredsFlag = flag.String("nats-creds", "", "NATS credentials file")
)

func main() {
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  daemon           Start daemon process\n")
		fmt.Fprintf(os.Stderr, "  status           Check daemon status\n")
		fmt.Fprintf(os.Stderr, "  stop             Stop daemon gracefully\n")
		fmt.Fprintf(os.Stderr, "  version          Show version\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	if *showVersion {
		fmt.Printf("amux-node %s\n", version)
		return
	}
	
	// Handle command
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	
	command := args[0]
	switch command {
	case "version":
		fmt.Printf("amux-node %s\n", version)
	case "daemon":
		handleDaemon()
	case "status":
		handleStatus()
	case "stop":
		handleStop()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func handleDaemon() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Override config with command line flags
	role := *roleFlag
	if role == "" {
		if cfg.Remote.Enabled {
			role = "manager" // Default role for remote hosts
		} else {
			role = "director" // Default role for local operation
		}
	}
	
	hostID := *hostIDFlag
	if hostID == "" {
		hostID = remote.GenerateHostID()
	}
	
	natsURL := *natsURLFlag
	if natsURL == "" {
		natsURL = cfg.Remote.NATS.URL
	}
	
	natsCreds := *natsCredsFlag
	if natsCreds == "" {
		natsCreds = cfg.Remote.NATS.CredsPath
	}
	
	log.Printf("Starting amux-node %s in %s role", version, role)
	log.Printf("Host ID: %s", hostID)
	
	// Parse timeout
	timeout, err := time.ParseDuration(cfg.Remote.RequestTimeout)
	if err != nil {
		timeout = 30 * time.Second
	}
	
	// Create remote manager configuration
	remoteConfig := &remote.RemoteConfig{
		Role:           role,
		HostID:         hostID,
		NATSURL:        natsURL,
		CredsPath:      natsCreds,
		SubjectPrefix:  cfg.Remote.NATS.SubjectPrefix,
		KVBucket:       cfg.Remote.NATS.KVBucket,
		RequestTimeout: timeout,
		BufferSize:     cfg.Remote.BufferSize,
	}
	
	// Create remote manager
	rm, err := remote.NewRemoteManager(remoteConfig)
	if err != nil {
		log.Fatalf("Failed to create remote manager: %v", err)
	}
	
	// Start the daemon
	if err := rm.Start(); err != nil {
		log.Fatalf("Failed to start remote manager: %v", err)
	}
	
	log.Printf("Daemon started successfully in %s role", role)
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("Received shutdown signal, stopping daemon...")
	
	// Graceful shutdown
	if err := rm.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	
	log.Println("Daemon stopped")
}

func handleStatus() {
	// TODO: Implement status check
	// For now, just check if process is running
	fmt.Println("Daemon status check not implemented yet")
	os.Exit(1)
}

func handleStop() {
	// TODO: Implement graceful stop
	// For now, just indicate that stop is not implemented
	fmt.Println("Graceful stop not implemented yet")
	os.Exit(1)
}