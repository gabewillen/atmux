package main

import (
	"fmt"
	"os"

	"github.com/agentflare-ai/amux/internal/config"
)

func main() {
	// Phase 0: Just compile and load config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("amux-node starting (role=%s)...\n", cfg.Node.Role)
	// Daemon logic to follow in later phases
}
