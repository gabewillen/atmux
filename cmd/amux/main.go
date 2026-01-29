package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/agentflare-ai/amux/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// Verify spec presence
	cwd, _ := os.Getwd()
	if err := cli.CheckSpec(cwd); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "test":
		testCmd := flag.NewFlagSet("test", flag.ExitOnError)
		regression := testCmd.Bool("regression", false, "Compare against previous snapshot")
		_ = testCmd.Parse(args)
		
		if err := cli.RunTest(context.Background(), *regression); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "daemon":
		fmt.Println("Use 'amux-node' to run the daemon.")
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: amux <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  test       Run verification snapshot")
	fmt.Println("  daemon     (Hint: use amux-node binary)")
}