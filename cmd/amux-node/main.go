// Command amux-node provides the daemon and manager functionality for amux.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/copilot-claude-sonnet-4/amux/internal/agent"
	"github.com/copilot-claude-sonnet-4/amux/internal/config"
	"github.com/copilot-claude-sonnet-4/amux/internal/paths"
	"github.com/copilot-claude-sonnet-4/amux/internal/rpc"
)

var (
	role       = flag.String("role", "director", "daemon role: director or manager")
	configFile = flag.String("config", "", "configuration file path")
	debug      = flag.Bool("debug", false, "enable debug logging")
)

func main() {
	flag.Parse()

	// Validate spec version (disabled for now)
	// if err := spec.ValidateVersion(); err != nil {
	//     log.Printf("Warning: Spec validation failed: %v", err)
	// }

	switch *role {
	case "director":
		runDirector()
	case "manager":
		runManager()
	default:
		log.Fatalf("Invalid role: %s (must be 'director' or 'manager')", *role)
	}
}

func runDirector() {
	log.Println("Starting amux director...")

	// Initialize configuration (simplified for now)
	_ = &config.Config{} // Suppress unused variable warning

	// Initialize path resolver
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	resolver, err := paths.NewResolver(cwd)
	if err != nil {
		log.Fatalf("Failed to create path resolver: %v", err)
	}

	// Initialize agent manager
	manager, err := agent.NewManager()
	if err != nil {
		log.Fatalf("Failed to create agent manager: %v", err)
	}

	// Populate agents from config (simplified for now)
	log.Println("Agent configuration loading not yet implemented")

	// TODO: Populate agents from config
	// for _, agentCfg := range cfg.Agents {
	//     // Implementation deferred
	// }

	// Initialize RPC server
	rpcServer := rpc.NewServer(manager, resolver)
	if err := rpcServer.Listen(); err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}

	log.Printf("Starting JSON-RPC server on %s", resolver.SocketPath())

	// Start RPC server in background
	go func() {
		if err := rpcServer.Serve(); err != nil {
			log.Printf("RPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Director started. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down director...")
	if err := rpcServer.Close(); err != nil {
		log.Printf("Failed to close RPC server: %v", err)
	}
}

func runManager() {
	log.Println("Starting amux manager...")
	
	// Manager role handles remote agent execution
	// TODO: Implement NATS connectivity and remote agent management
	
	log.Println("Manager mode not yet implemented")
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	log.Println("Manager started. Press Ctrl+C to stop.")
	<-sigChan
	
	log.Println("Shutting down manager...")
}