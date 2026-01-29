// Package main implements the amux CLI client
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stateforward/amux/internal/snapshot"
)

func main() {
	args := os.Args[1:]
	
	if len(args) == 0 {
		fmt.Println("amux CLI client - Agent Multiplexer")
		fmt.Println("Usage: amux [command]")
		fmt.Println("Commands: test, conformance, agent, chat")
		return
	}
	
	command := args[0]
	switch command {
	case "test":
		testCmd(args[1:])
	case "conformance":
		conformanceCmd(args[1:])
	case "agent":
		agentCmd(args[1:])
	case "chat":
		chatCmd(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: test, conformance, agent, chat")
		os.Exit(1)
	}
}

func testCmd(args []string) {
	var regression bool
	var noSnapshot bool
	
	for _, arg := range args {
		switch arg {
		case "--regression":
			regression = true
		case "--no-snapshot":
			noSnapshot = true
		default:
			fmt.Printf("Unknown flag: %s\n", arg)
			fmt.Println("Usage: amux test [--regression] [--no-snapshot]")
			os.Exit(1)
		}
	}
	
	if regression {
		// Run regression test
		if err := runRegressionTest(noSnapshot); err != nil {
			fmt.Fprintf(os.Stderr, "Regression test failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Run snapshot test
		if err := runSnapshotTest(noSnapshot); err != nil {
			fmt.Fprintf(os.Stderr, "Snapshot test failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func runSnapshotTest(noSnapshot bool) error {
	moduleRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	snap, err := snapshot.Create(moduleRoot)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}
	
	if noSnapshot {
		// Print to stdout instead of saving to file
		fmt.Println("Snapshot created (not saved due to --no-snapshot):")
		fmt.Printf("Timestamp: %v\n", snap.Timestamp)
		fmt.Printf("Version: %v\n", snap.Version)
		fmt.Printf("Module: %v\n", snap.Module)
		fmt.Printf("Number of files: %d\n", len(snap.Files))
	} else {
		// Create snapshots directory if it doesn't exist
		snapshotsDir := filepath.Join(moduleRoot, "snapshots")
		if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
			return fmt.Errorf("failed to create snapshots directory: %w", err)
		}
		
		// Generate filename with timestamp
		filename := fmt.Sprintf("amux-test-%d.toml", snap.Timestamp.Unix())
		snapshotPath := filepath.Join(snapshotsDir, filename)
		
		if err := snap.Save(snapshotPath); err != nil {
			return fmt.Errorf("failed to save snapshot: %w", err)
		}
		
		fmt.Printf("Snapshot saved to: %s\n", snapshotPath)
	}
	
	return nil
}

func runRegressionTest(noSnapshot bool) error {
	moduleRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	
	// Load the most recent snapshot
	snapshotsDir := filepath.Join(moduleRoot, "snapshots")
	
	// For simplicity, we'll just look for the first snapshot file
	// In a real implementation, we'd find the most recent one
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return fmt.Errorf("failed to read snapshots directory: %w", err)
	}
	
	var prevSnapshotPath string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".toml" {
			prevSnapshotPath = filepath.Join(snapshotsDir, entry.Name())
			break
		}
	}
	
	if prevSnapshotPath == "" {
		return fmt.Errorf("no previous snapshot found in %s", snapshotsDir)
	}
	
	// Load previous snapshot
	prevSnap, err := snapshot.Load(prevSnapshotPath)
	if err != nil {
		return fmt.Errorf("failed to load previous snapshot: %w", err)
	}
	
	// Create new snapshot
	newSnap, err := snapshot.Create(moduleRoot)
	if err != nil {
		return fmt.Errorf("failed to create new snapshot: %w", err)
	}
	
	// Compare snapshots
	comparison := prevSnap.Compare(newSnap)
	
	hasChanges := len(comparison.Added) > 0 || len(comparison.Removed) > 0 || len(comparison.Modified) > 0
	
	if hasChanges {
		fmt.Println("REGRESSION DETECTED:")
		
		if len(comparison.Added) > 0 {
			fmt.Printf("Added files (%d):\n", len(comparison.Added))
			for path := range comparison.Added {
				fmt.Printf("  + %s\n", path)
			}
		}
		
		if len(comparison.Removed) > 0 {
			fmt.Printf("Removed files (%d):\n", len(comparison.Removed))
			for path := range comparison.Removed {
				fmt.Printf("  - %s\n", path)
			}
		}
		
		if len(comparison.Modified) > 0 {
			fmt.Printf("Modified files (%d):\n", len(comparison.Modified))
			for path := range comparison.Modified {
				fmt.Printf("  ~ %s\n", path)
			}
		}
		
		return fmt.Errorf("regression detected: snapshots differ")
	} else {
		fmt.Println("No regressions detected: snapshots match")
	}
	
	// If not noSnapshot, save the new snapshot
	if !noSnapshot {
		// Generate filename with timestamp
		filename := fmt.Sprintf("amux-test-%d.toml", newSnap.Timestamp.Unix())
		snapshotPath := filepath.Join(snapshotsDir, filename)
		
		if err := newSnap.Save(snapshotPath); err != nil {
			return fmt.Errorf("failed to save new snapshot: %w", err)
		}
		
		fmt.Printf("New snapshot saved to: %s\n", snapshotPath)
	}
	
	return nil
}

func conformanceCmd(args []string) {
	fmt.Println("Running conformance suite...")
	// Implementation would go here
}

func agentCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: amux agent [add|list|remove]")
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		agentAddCmd(args[1:])
	case "list":
		agentListCmd(args[1:])
	case "remove":
		agentRemoveCmd(args[1:])
	default:
		fmt.Printf("Unknown agent subcommand: %s\n", subcommand)
		fmt.Println("Available subcommands: add, list, remove")
		os.Exit(1)
	}
}

func agentAddCmd(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: amux agent add <name> <adapter>")
		fmt.Println("Example: amux agent add my-agent claude-code")
		os.Exit(1)
	}

	name := args[0]
	adapter := args[1]

	// For now, just print what would happen
	fmt.Printf("Would add agent: name=%s, adapter=%s\n", name, adapter)

	// In a real implementation, this would call the daemon via JSON-RPC
	// to add the agent with the specified parameters
}

func agentListCmd(args []string) {
	// For now, just print what would happen
	fmt.Println("Listing agents...")

	// In a real implementation, this would call the daemon via JSON-RPC
	// to list all agents
}

func agentRemoveCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: amux agent remove <agent-id>")
		os.Exit(1)
	}

	agentID := args[0]
	fmt.Printf("Would remove agent: %s\n", agentID)

	// In a real implementation, this would call the daemon via JSON-RPC
	// to remove the specified agent
}

func chatCmd(args []string) {
	fmt.Println("Chat functionality...")
	// Implementation would go here
}