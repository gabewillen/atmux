// Package protocol implements remote communication protocol (transports events)
package protocol

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stateforward/amux/internal/config"
)

// NATSServer manages NATS server configuration for both hub (director) and leaf (manager) modes
type NATSServer struct {
	cfg *config.NATSConfig
	server *server.Server
	nc *nats.Conn
}

// NewNATSServer creates a new NATS server instance
func NewNATSServer(cfg *config.NATSConfig) *NATSServer {
	return &NATSServer{
		cfg: cfg,
	}
}

// StartHubServer starts a NATS server in hub mode (for director role)
func (ns *NATSServer) StartHubServer(ctx context.Context) error {
	opts := &server.Options{
		Host:      "0.0.0.0", // Listen on all interfaces
		Port:      -1,        // Use random available port
		JetStream: true,      // Enable JetStream
		StoreDir:  expandHomeDir("~/.amux/nats"), // Store JetStream data
	}

	// Start the NATS server
	s, err := server.NewServer(opts)
	if err != nil {
		return fmt.Errorf("failed to create NATS server: %w", err)
	}

	ns.server = s
	server.ConfigureLogger(s, server.StandAloneMode, false, false)

	go s.Start()

	// Wait for server to be ready
	if !s.ReadyForConnections(10 * time.Second) {
		return fmt.Errorf("NATS server failed to start in time")
	}

	log.Printf("NATS Hub server started on %s", s.ClientURL())

	// Connect to the server
	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server: %w", err)
	}
	ns.nc = nc

	return nil
}

// StartLeafServer starts a NATS server in leaf mode (for manager role)
func (ns *NATSServer) StartLeafServer(ctx context.Context, hubURL, credsPath string) error {
	// Connect to the hub as a leaf node
	leafOpts := fmt.Sprintf(`
		port: -1
		leafnodes {
			remotes: [
				{
					url: "%s"
					credentials: "%s"
				}
			]
		}
		jetstream: enabled
		store_dir: "%s"
	`, hubURL, credsPath, expandHomeDir("~/.amux/nats"))

	configFile, err := os.CreateTemp("", "nats-leaf-config-*.conf")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	defer os.Remove(configFile.Name())

	if _, err := configFile.WriteString(leafOpts); err != nil {
		configFile.Close()
		return fmt.Errorf("failed to write leaf config: %w", err)
	}
	configFile.Close()

	opts := &server.Options{}
	if err := opts.ProcessConfigFile(configFile.Name()); err != nil {
		return fmt.Errorf("failed to process leaf config: %w", err)
	}

	s, err := server.NewServer(opts)
	if err != nil {
		return fmt.Errorf("failed to create NATS leaf server: %w", err)
	}

	ns.server = s
	server.ConfigureLogger(s, server.StandAloneMode, false, false)

	go s.Start()

	// Wait for server to be ready
	if !s.ReadyForConnections(10 * time.Second) {
		return fmt.Errorf("NATS leaf server failed to start in time")
	}

	log.Printf("NATS Leaf server started, connected to hub at %s", hubURL)

	// Connect to the local server
	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		return fmt.Errorf("failed to connect to NATS leaf server: %w", err)
	}
	ns.nc = nc

	return nil
}

// Stop stops the NATS server
func (ns *NATSServer) Stop() {
	if ns.nc != nil {
		ns.nc.Close()
	}
	if ns.server != nil {
		ns.server.Shutdown()
	}
}

// GetClient returns the NATS connection
func (ns *NATSServer) GetClient() *nats.Conn {
	return ns.nc
}

// expandHomeDir expands the ~ symbol to the user's home directory
func expandHomeDir(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}

// WaitForShutdown waits for a signal to shut down the server
func (ns *NATSServer) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down NATS server...")
	ns.Stop()
}