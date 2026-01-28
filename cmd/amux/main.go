// Package amux is the Agent Multiplexer implementation.
package main

import (
	"os"

	"github.com/agentflare-ai/amux/internal/test"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "test":
		handleTestCommand()
	default:
		printUsage()
		os.Exit(1)
	}
}

// handleTestCommand processes the 'amux test' command.
func handleTestCommand() {
	config, err := test.ParseFlags(os.Args[2:])
	if err != nil {
		printUsage()
		os.Exit(1)
	}

	if err := test.Run(config); err != nil {
		// Error already printed by test.Run
		os.Exit(1)
	}
}

// printUsage prints command usage information.
func printUsage() {
	println("amux - Agent Multiplexer")
	println("")
	println("Usage:")
	println("  amux test [flags]  Run test suite and create snapshot")
	println("")
	println("Test flags:")
	println("  --no-snapshot      Skip writing snapshot file")
	println("  --regression      Compare with previous snapshot and fail on regressions")
	println("  --output PATH     Output file path (default: snapshots/amux-test-YYYYMMDD-HHMMSS.toml)")
	println("  --quiet           Suppress verbose output")
}
