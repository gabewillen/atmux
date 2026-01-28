// Package main implements the amux CLI client per spec §12.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stateforward/amux/internal/snapshot"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "amux: command required")
		fmt.Fprintln(os.Stderr, "Usage: amux <command> [args...]")
		os.Exit(1)
	}

	// Phase 0: Minimal CLI stub
	command := os.Args[1]
	
	switch command {
	case "version":
		fmt.Println("amux v0.1.0-phase0")
	case "test":
		handleTestCommand()
	default:
		fmt.Fprintf(os.Stderr, "amux: unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleTestCommand() {
	// Parse flags
	testFlags := flag.NewFlagSet("test", flag.ExitOnError)
	regression := testFlags.Bool("regression", false, "Compare against previous snapshot")
	noSnapshot := testFlags.Bool("no-snapshot", false, "Write snapshot to stdout")
	testFlags.Parse(os.Args[2:])
	
	moduleRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: get working directory: %v\n", err)
		os.Exit(1)
	}
	
	// Create snapshot
	fmt.Fprintln(os.Stderr, "Running amux test suite...")
	snap, err := snapshot.Create(moduleRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: create snapshot: %v\n", err)
		os.Exit(1)
	}
	
	if *regression {
		// Regression mode: compare with latest snapshot
		latestPath, err := snapshot.FindLatestSnapshot(moduleRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: find baseline snapshot: %v\n", err)
			os.Exit(1)
		}
		
		baseline, err := snapshot.Read(latestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: read baseline snapshot: %v\n", err)
			os.Exit(1)
		}
		
		passed, report := snapshot.Compare(baseline, snap)
		fmt.Fprintln(os.Stderr, "\nRegression report:")
		fmt.Fprintln(os.Stderr, report)
		
		if !passed {
			fmt.Fprintln(os.Stderr, "\nREGRESSION DETECTED")
			os.Exit(1)
		}
		
		fmt.Fprintln(os.Stderr, "\nNo regressions detected")
		
		// Write new snapshot
		outPath := snapshot.GenerateSnapshotPath(moduleRoot)
		if err := snapshot.Write(snap, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write snapshot: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Snapshot written: %s\n", outPath)
	} else if *noSnapshot {
		// No-snapshot mode: write to stdout
		// (This is a stub for Phase 0)
		fmt.Fprintln(os.Stderr, "Phase 0: --no-snapshot writes to stdout")
	} else {
		// Normal mode: write snapshot
		outPath := snapshot.GenerateSnapshotPath(moduleRoot)
		if err := snapshot.Write(snap, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write snapshot: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Snapshot written: %s\n", outPath)
	}
}
