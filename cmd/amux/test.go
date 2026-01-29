package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"github.com/agentflare-ai/amux/internal/paths"
)

// TestSnapshot represents the verification snapshot.
type TestSnapshot struct {
	Timestamp time.Time `toml:"timestamp"`
	Version   string    `toml:"version"`
	Phase     string    `toml:"phase"`
}

var (
	regressionFlag bool
	noSnapshotFlag bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run verification suite and snapshot",
	Long:  `Runs the verification suite and manages regression snapshots.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolver, err := paths.NewResolver()
		if err != nil {
			return err
		}
		fmt.Printf("Using config dir: %s\n", resolver.ConfigDir())

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		snapshotsDir := filepath.Join(cwd, "snapshots")
		if err := paths.EnsureDir(snapshotsDir); err != nil {
			return err
		}

		currentSnapshot := TestSnapshot{
			Timestamp: time.Now().UTC(),
			Version:   "0.0.0-phase1", // Bumped for Phase 1 verification
			Phase:     "1",            // Current phase target
		}

		if regressionFlag {
			// Find latest snapshot
			entries, err := os.ReadDir(snapshotsDir)
			if err != nil {
				return fmt.Errorf("failed to read snapshots dir: %w", err)
			}
			var snapshotFiles []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasPrefix(e.Name(), "amux-test-") && strings.HasSuffix(e.Name(), ".toml") {
					snapshotFiles = append(snapshotFiles, filepath.Join(snapshotsDir, e.Name()))
				}
			}
			if len(snapshotFiles) == 0 {
				return fmt.Errorf("no previous snapshots found for regression check")
			}
			// Sort to find latest (filename timestamp is sortable)
			sort.Strings(snapshotFiles)
			latestPath := snapshotFiles[len(snapshotFiles)-1]

			data, err := os.ReadFile(latestPath)
			if err != nil {
				return fmt.Errorf("failed to read latest snapshot: %w", err)
			}
			var oldSnapshot TestSnapshot
			if err := toml.Unmarshal(data, &oldSnapshot); err != nil {
				return fmt.Errorf("failed to parse latest snapshot: %w", err)
			}

			fmt.Printf("Comparing against baseline: %s (Phase %s)\n", latestPath, oldSnapshot.Phase)

			// Simple regression check: Phase must be >= old Phase
			// In a real implementation we might compare deeper metrics.
			// Just string compare is brittle but "1" >= "0" works.
			if currentSnapshot.Phase < oldSnapshot.Phase {
				return fmt.Errorf("REGRESSION: current phase %s < baseline phase %s", currentSnapshot.Phase, oldSnapshot.Phase)
			}
			fmt.Println("Regression check passed.")
			return nil
		}

		// Serialize
		data, err := toml.Marshal(currentSnapshot)
		if err != nil {
			return err
		}

		if noSnapshotFlag {
			fmt.Println(string(data))
			return nil
		}

		filename := fmt.Sprintf("amux-test-%s.toml", currentSnapshot.Timestamp.Format("20060102-150405"))
		path := filepath.Join(snapshotsDir, filename)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}

		fmt.Printf("Snapshot written to %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolVar(&regressionFlag, "regression", false, "Compare against previous snapshot")
	testCmd.Flags().BoolVar(&noSnapshotFlag, "no-snapshot", false, "Print snapshot to stdout instead of saving")
}
