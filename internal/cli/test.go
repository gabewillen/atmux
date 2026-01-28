package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/snapshot"
)

func RunTest(ctx context.Context, regression bool) error {
	fmt.Println("Running amux verification...")

	// Capture snapshot
	snap, err := snapshot.Capture()
	if err != nil {
		return fmt.Errorf("failed to capture snapshot: %w", err)
	}

	// Determine snapshot path
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	snapshotDir := filepath.Join(cwd, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("amux-test-%s.toml", timestamp)
	path := filepath.Join(snapshotDir, filename)

	// Save snapshot
	if err := snapshot.Save(path, snap); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}
	fmt.Printf("Snapshot saved to %s\n", path)

	if regression {
		return runRegression(snapshotDir, snap)
	}

	return nil
}

func runRegression(dir string, current *snapshot.Snapshot) error {
	// Find previous snapshot
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var snapshots []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "amux-test-") && strings.HasSuffix(e.Name(), ".toml") {
			snapshots = append(snapshots, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(snapshots)

	if len(snapshots) < 2 {
		fmt.Println("No previous snapshot to compare against.")
		return nil
	}
	
	// The last one is the current one we just saved. The one before is the baseline.
	baselinePath := snapshots[len(snapshots)-2]
	fmt.Printf("Comparing against baseline: %s\n", baselinePath)
	
baseline, err := snapshot.Load(baselinePath)
	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	if err := snapshot.Compare(baseline, current); err != nil {
		return fmt.Errorf("regression detected: %w", err)
	}
	
	fmt.Println("Regression check passed.")
	return nil
}
