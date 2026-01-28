// Command amux is the main CLI client for amux.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		fmt.Println("amux CLI (Phase 0)")
		fmt.Println("Usage: amux <command> [args...]")
		return nil
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "test":
		return runTest(args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
