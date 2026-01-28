package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/agentflare-ai/amux/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "amux",
	Short: "Agent Multiplexer CLI",
	Long:  `amux is an agent-agnostic orchestrator for coding agents.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config to verify it works (Phase 0 check)
		// In real CLI, we might delay until needed or rely on defaults
		_, err := config.Load("") // Empty root means valid only if inside repo or just user config
		if err != nil {
			// Don't fail hard on config load for help/version, but warn?
			// For now, just ignore or log debug.
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands here
	// testCmd is defined in test.go (in same package main)
}
