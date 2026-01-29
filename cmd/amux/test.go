// Package main implements the amux CLI client.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/spec"
	"github.com/pelletier/go-toml/v2"
)

// TestSnapshot represents a test snapshot.
type TestSnapshot struct {
	Timestamp time.Time
	Results   map[string]interface{}
}

// runTest implements the `amux test` command.
func runTest(args []string) error {
	var (
		noSnapshot   bool
		regression   bool
		snapshotPath string
	)

	// Parse flags (simplified for Phase 0)
	for i, arg := range args {
		switch arg {
		case "--no-snapshot":
			noSnapshot = true
		case "--regression":
			regression = true
		case "--snapshot":
			if i+1 < len(args) {
				snapshotPath = args[i+1]
			}
		}
	}

	// Find module root early for spec guard and snapshot path (plan Phase 0: guard test or startup check).
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find module root: %w", err)
	}
	if err := spec.CheckSpecVersion(moduleRoot); err != nil {
		return fmt.Errorf("spec version check failed: %w", err)
	}

	// Determine snapshot path
	if snapshotPath == "" {
		snapshotPath = filepath.Join(moduleRoot, "snapshots", fmt.Sprintf("amux-test-%s.toml", time.Now().Format("20060102-150405")))
	}

	// Run test sequence per spec §12.6
	results := make(map[string]interface{})

	// 1. go mod tidy
	if err := runCommand("go", "mod", "tidy"); err != nil {
		results["tidy"] = map[string]interface{}{"status": "failed", "error": err.Error()}
	} else {
		results["tidy"] = map[string]interface{}{"status": "passed"}
	}

	// 2. go vet
	if err := runCommand("go", "vet", "./..."); err != nil {
		results["vet"] = map[string]interface{}{"status": "failed", "error": err.Error()}
	} else {
		results["vet"] = map[string]interface{}{"status": "passed"}
	}

	// 3. go test -race
	if err := runCommand("go", "test", "-race", "./..."); err != nil {
		results["test_race"] = map[string]interface{}{"status": "failed", "error": err.Error()}
	} else {
		results["test_race"] = map[string]interface{}{"status": "passed"}
	}

	// 4. go test
	if err := runCommand("go", "test", "./..."); err != nil {
		results["test"] = map[string]interface{}{"status": "failed", "error": err.Error()}
	} else {
		results["test"] = map[string]interface{}{"status": "passed"}
	}

	// 5. Coverage (placeholder)
	results["coverage"] = map[string]interface{}{"status": "skipped"}

	// 6. Benchmarks (placeholder)
	results["bench"] = map[string]interface{}{"status": "skipped"}

	// Create snapshot
	snapshot := TestSnapshot{
		Timestamp: time.Now(),
		Results:   results,
	}

	// Write snapshot
	if noSnapshot {
		// Write to stdout
		data, err := toml.Marshal(snapshot)
		if err != nil {
			return fmt.Errorf("failed to marshal snapshot: %w", err)
		}
		os.Stdout.Write(data)
	} else {
		// Write to file
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
			return fmt.Errorf("failed to create snapshot directory: %w", err)
		}

		data, err := toml.Marshal(snapshot)
		if err != nil {
			return fmt.Errorf("failed to marshal snapshot: %w", err)
		}

		if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write snapshot file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Snapshot written to %s\n", snapshotPath)
	}

	// Regression check
	if regression {
		return checkRegression(snapshotPath)
	}

	return nil
}

// findModuleRoot finds the Go module root directory.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a Go module")
		}
		dir = parent
	}
}

// runCommand runs a command and returns an error if it fails.
// Per spec §12.6, commands should continue on failure but record the result.
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr // Write output to stderr per spec
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// checkRegression compares the current snapshot with the previous one.
func checkRegression(snapshotPath string) error {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find module root: %w", err)
	}

	snapshotsDir := filepath.Join(moduleRoot, "snapshots")
	
	// Find the most recent snapshot before the current one
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var previousSnapshot string
	var previousTime time.Time
	currentTime := time.Now()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "amux-test-") {
			continue
		}

		// Extract timestamp from filename
		if len(entry.Name()) < len("amux-test-YYYYMMDD-HHMMSS.toml") {
			continue
		}

		timeStr := entry.Name()[len("amux-test-"):len(entry.Name())-len(".toml")]
		t, err := time.Parse("20060102-150405", timeStr)
		if err != nil {
			continue
		}

		if t.Before(currentTime) && (previousSnapshot == "" || t.After(previousTime)) {
			previousSnapshot = entry.Name()
			previousTime = t
		}
	}

	if previousSnapshot == "" {
		fmt.Fprintf(os.Stderr, "No previous snapshot found for comparison\n")
		return nil
	}

	// Load both snapshots
	prevPath := filepath.Join(snapshotsDir, previousSnapshot)
	prevData, err := os.ReadFile(prevPath)
	if err != nil {
		return fmt.Errorf("failed to read previous snapshot: %w", err)
	}

	currData, err := os.ReadFile(snapshotPath)
	if err != nil {
		return fmt.Errorf("failed to read current snapshot: %w", err)
	}

	// Compare snapshots
	var prevSnapshot, currSnapshot TestSnapshot
	if err := toml.Unmarshal(prevData, &prevSnapshot); err != nil {
		return fmt.Errorf("failed to parse previous snapshot: %w", err)
	}

	if err := toml.Unmarshal(currData, &currSnapshot); err != nil {
		return fmt.Errorf("failed to parse current snapshot: %w", err)
	}

	// Check for regressions
	regressions := []string{}
	
	// Compare each test result
	for testName, prevResult := range prevSnapshot.Results {
		currResult, exists := currSnapshot.Results[testName]
		if !exists {
			regressions = append(regressions, fmt.Sprintf("Test %s removed", testName))
			continue
		}

		prevStatus := getStatus(prevResult)
		currStatus := getStatus(currResult)

		if prevStatus == "passed" && currStatus != "passed" {
			regressions = append(regressions, fmt.Sprintf("Test %s regressed: %s -> %s", testName, prevStatus, currStatus))
		}
	}

	if len(regressions) > 0 {
		fmt.Fprintf(os.Stderr, "Regression detected:\n")
		for _, reg := range regressions {
			fmt.Fprintf(os.Stderr, "  - %s\n", reg)
		}
		return fmt.Errorf("regression detected: %d test(s) regressed", len(regressions))
	}

	fmt.Fprintf(os.Stderr, "No regressions detected (compared to %s)\n", previousSnapshot)
	return nil
}

// getStatus extracts the status from a test result.
func getStatus(result interface{}) string {
	if m, ok := result.(map[string]interface{}); ok {
		if status, ok := m["status"].(string); ok {
			return status
		}
	}
	return "unknown"
}
