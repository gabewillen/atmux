// Package snapshot implements the amux test snapshot functionality per spec §12.6.
package snapshot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stateforward/amux/internal/errors"
)

// Snapshot represents a test snapshot per spec §12.6.
type Snapshot struct {
	Timestamp    time.Time         `toml:"timestamp"`
	GoVersion    string            `toml:"go_version"`
	Module       string            `toml:"module"`
	TidyStatus   string            `toml:"tidy_status"`
	VetStatus    string            `toml:"vet_status"`
	LintStatus   string            `toml:"lint_status"`
	TestStatus   string            `toml:"test_status"`
	RaceStatus   string            `toml:"race_status"`
	Coverage     float64           `toml:"coverage"`
	BenchResults map[string]string `toml:"bench_results,omitempty"`
}

// Create creates a new snapshot by running the verification sequence per spec §12.6.
func Create(moduleRoot string) (*Snapshot, error) {
	snapshot := &Snapshot{
		Timestamp:    time.Now(),
		GoVersion:    "1.25.6", // Toolchain version per spec
		Module:       "github.com/stateforward/amux",
		TidyStatus:   "unknown",
		VetStatus:    "unknown",
		LintStatus:   "unknown",
		TestStatus:   "unknown",
		RaceStatus:   "unknown",
		Coverage:     0.0,
		BenchResults: make(map[string]string),
	}

	// Helper to run a command in the module root and return "pass" or "fail".
	run := func(name string, args ...string) string {
		cmd := exec.Command(name, args...)
		cmd.Dir = moduleRoot
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "fail"
		}
		return "pass"
	}

	// Run tidy → vet → lint → test -race → test per plan §12.6.
	snapshot.TidyStatus = run("go", "mod", "tidy")
	snapshot.VetStatus = run("go", "vet", "./...")

	// Lint via staticcheck when available; otherwise mark as "skip".
	if _, err := exec.LookPath("staticcheck"); err == nil {
		snapshot.LintStatus = run("staticcheck", "./...")
	} else {
		snapshot.LintStatus = "skip"
	}

	snapshot.RaceStatus = run("go", "test", "-race", "./...")
	snapshot.TestStatus = run("go", "test", "./...")

	// Coverage: run `go test -cover ./...` and parse the reported percentage.
	covCmd := exec.Command("go", "test", "-cover", "./...")
	covCmd.Dir = moduleRoot
	var buf bytes.Buffer
	covCmd.Stdout = &buf
	covCmd.Stderr = os.Stderr
	if err := covCmd.Run(); err == nil {
		lines := strings.Split(buf.String(), "\n")
		re := regexp.MustCompile(`coverage:\s+([0-9.]+)%`)
		for _, line := range lines {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 2 {
				if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
					snapshot.Coverage = v
					break
				}
			}
		}
	}

	// Benchmarks: best-effort; failures do not affect other statuses.
	benchCmd := exec.Command("go", "test", "-run=^$", "-bench=.", "./...")
	benchCmd.Dir = moduleRoot
	benchCmd.Stdout = os.Stderr
	benchCmd.Stderr = os.Stderr
	_ = benchCmd.Run()

	return snapshot, nil
}

// Write writes a snapshot to a TOML file.
func Write(snapshot *Snapshot, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrapf(err, "create snapshot directory: %s", dir)
	}
	
	data, err := toml.Marshal(snapshot)
	if err != nil {
		return errors.Wrap(err, "marshal snapshot")
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrapf(err, "write snapshot: %s", path)
	}
	
	return nil
}

// Read reads a snapshot from a TOML file.
func Read(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read snapshot: %s", path)
	}
	
	var snapshot Snapshot
	if err := toml.Unmarshal(data, &snapshot); err != nil {
		return nil, errors.Wrapf(err, "unmarshal snapshot: %s", path)
	}
	
	return &snapshot, nil
}

// Compare compares two snapshots and returns a regression report.
func Compare(baseline, current *Snapshot) (bool, string) {
	// Phase 0: Simple comparison
	if baseline.TestStatus == "pass" && current.TestStatus != "pass" {
		return false, "Tests regressed: baseline passed, current failed"
	}
	
	if baseline.VetStatus == "pass" && current.VetStatus != "pass" {
		return false, "Vet regressed: baseline passed, current failed"
	}
	
	if current.Coverage < baseline.Coverage {
		return false, fmt.Sprintf("Coverage regressed: %.2f%% -> %.2f%%", baseline.Coverage, current.Coverage)
	}
	
	return true, "No regressions detected"
}

// GenerateSnapshotPath generates a snapshot file path.
func GenerateSnapshotPath(moduleRoot string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("amux-test-%s.toml", timestamp)
	return filepath.Join(moduleRoot, "snapshots", filename)
}

// FindLatestSnapshot finds the most recent snapshot in the snapshots directory.
func FindLatestSnapshot(moduleRoot string) (string, error) {
	snapshotsDir := filepath.Join(moduleRoot, "snapshots")
	
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.Wrap(errors.ErrNotFound, "no snapshots directory")
		}
		return "", errors.Wrap(err, "read snapshots directory")
	}
	
	var latestPath string
	var latestTime time.Time
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestPath = filepath.Join(snapshotsDir, entry.Name())
		}
	}
	
	if latestPath == "" {
		return "", errors.Wrap(errors.ErrNotFound, "no snapshot files found")
	}
	
	return latestPath, nil
}
