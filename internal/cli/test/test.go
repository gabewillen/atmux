// Package test implements the 'amux test' subcommand.
//
// The amux test command runs a standardized Go verification suite for the
// current Go module and emits a TOML "test snapshot" capturing the results.
//
// See spec §12.6 for the full specification.
package test

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Run executes the amux test command.
func Run(ctx context.Context, args []string) error {
	// Parse flags
	flags, _ := parseFlags(args)

	noSnapshot := hasFlag(flags, "no-snapshot")
	regression := hasFlag(flags, "regression")

	// Find module root
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	moduleRoot, err := paths.FindModuleRoot(wd)
	if err != nil {
		return fmt.Errorf("no go.mod found: a Go module is required")
	}

	// Check required executables
	if err := checkRequiredExecutables(); err != nil {
		return err
	}

	// Create snapshot
	snapshot := NewSnapshot(moduleRoot)

	// Run all steps
	steps := []struct {
		name string
		key  string
		args []string
	}{
		{"go mod tidy", "go_mod_tidy", []string{"go", "mod", "tidy"}},
		{"go vet", "go_vet", []string{"go", "vet", "./..."}},
		{"golangci-lint", "golangci_lint", []string{"golangci-lint", "run", "./..."}},
		{"go test -race", "tests_race", []string{"go", "test", "-race", "./..."}},
		{"go test", "tests", []string{"go", "test", "./..."}},
		{"go test -cover", "coverage", nil}, // coverprofile handled specially
		{"go test -bench", "benchmarks", []string{"go", "test", "-run=^$", "-bench=.", "-benchmem", "./..."}},
	}

	// Output based on mode
	var output io.Writer = os.Stderr
	if noSnapshot {
		output = os.Stderr
	}

	for _, step := range steps {
		fmt.Fprintf(output, "Running: %s\n", step.name)

		var result *StepResult
		if step.key == "coverage" {
			// Handle coverage specially
			coverFile := filepath.Join(os.TempDir(), fmt.Sprintf("amux-coverage-%d.out", os.Getpid()))
			defer os.Remove(coverFile)
			result = runCommand(ctx, moduleRoot, []string{"go", "test", "./...", "-coverprofile=" + coverFile})

			// Parse coverage if successful
			if result.ExitCode == 0 {
				if percent, err := parseCoverageProfile(coverFile); err == nil {
					result.TotalPercent = &percent
				}
			}
		} else {
			result = runCommand(ctx, moduleRoot, step.args)
		}

		snapshot.SetStep(step.key, result)

		if result.ExitCode != 0 {
			fmt.Fprintf(output, "  FAILED (exit %d)\n", result.ExitCode)
		} else {
			fmt.Fprintf(output, "  OK\n")
		}
	}

	// Parse benchmarks from the benchmarks step
	// Note: A full implementation would capture and parse benchmark output here
	_ = snapshot.Steps.Benchmarks // Mark as used

	// Handle regression checking
	if regression {
		if err := checkRegression(moduleRoot, snapshot, output); err != nil {
			// Still write the snapshot
			if !noSnapshot {
				if err := writeSnapshot(moduleRoot, snapshot); err != nil {
					fmt.Fprintf(output, "Warning: failed to write snapshot: %v\n", err)
				}
			} else {
				writeSnapshotToStdout(snapshot)
			}
			return err
		}
	}

	// Write snapshot
	if noSnapshot {
		writeSnapshotToStdout(snapshot)
	} else {
		if err := writeSnapshot(moduleRoot, snapshot); err != nil {
			return fmt.Errorf("write snapshot: %w", err)
		}
	}

	// Check for any failures
	if snapshot.HasFailures() {
		return fmt.Errorf("one or more steps failed")
	}

	return nil
}

// Snapshot represents the test snapshot TOML structure.
type Snapshot struct {
	Meta       MetaInfo   `toml:"meta"`
	Steps      Steps      `toml:"steps"`
	Benchmarks []Benchmark `toml:"benchmarks,omitempty"`
}

// MetaInfo contains snapshot metadata.
type MetaInfo struct {
	CreatedAt   string `toml:"created_at"`
	ModuleRoot  string `toml:"module_root"`
	SpecVersion string `toml:"spec_version"`
}

// Steps contains results for each step.
type Steps struct {
	GoModTidy    *StepResult `toml:"go_mod_tidy"`
	GoVet        *StepResult `toml:"go_vet"`
	GolangciLint *StepResult `toml:"golangci_lint"`
	TestsRace    *StepResult `toml:"tests_race"`
	Tests        *StepResult `toml:"tests"`
	Coverage     *StepResult `toml:"coverage"`
	Benchmarks   *StepResult `toml:"benchmarks"`
}

// StepResult contains the result of a single step.
type StepResult struct {
	Argv         []string `toml:"argv"`
	ExitCode     int      `toml:"exit_code"`
	DurationMs   int64    `toml:"duration_ms"`
	StdoutSha256 string   `toml:"stdout_sha256"`
	StderrSha256 string   `toml:"stderr_sha256"`
	StdoutBytes  int      `toml:"stdout_bytes"`
	StderrBytes  int      `toml:"stderr_bytes"`
	TotalPercent *float64 `toml:"total_percent,omitempty"`
}

// Benchmark contains benchmark results.
type Benchmark struct {
	Name        string  `toml:"name"`
	Pkg         string  `toml:"pkg"`
	NsPerOp     float64 `toml:"ns_per_op"`
	Iterations  int64   `toml:"iterations"`
	BytesPerOp  *int64  `toml:"bytes_per_op,omitempty"`
	AllocsPerOp *int64  `toml:"allocs_per_op,omitempty"`
}

// NewSnapshot creates a new snapshot with metadata initialized.
func NewSnapshot(moduleRoot string) *Snapshot {
	return &Snapshot{
		Meta: MetaInfo{
			CreatedAt:   time.Now().UTC().Format(time.RFC3339),
			ModuleRoot:  moduleRoot,
			SpecVersion: api.SpecVersion,
		},
	}
}

// SetStep sets the result for a step.
func (s *Snapshot) SetStep(key string, result *StepResult) {
	switch key {
	case "go_mod_tidy":
		s.Steps.GoModTidy = result
	case "go_vet":
		s.Steps.GoVet = result
	case "golangci_lint":
		s.Steps.GolangciLint = result
	case "tests_race":
		s.Steps.TestsRace = result
	case "tests":
		s.Steps.Tests = result
	case "coverage":
		s.Steps.Coverage = result
	case "benchmarks":
		s.Steps.Benchmarks = result
	}
}

// HasFailures returns true if any step failed.
func (s *Snapshot) HasFailures() bool {
	steps := []*StepResult{
		s.Steps.GoModTidy,
		s.Steps.GoVet,
		s.Steps.GolangciLint,
		s.Steps.TestsRace,
		s.Steps.Tests,
		s.Steps.Coverage,
		s.Steps.Benchmarks,
	}

	for _, step := range steps {
		if step != nil && step.ExitCode != 0 {
			return true
		}
	}

	return false
}

func runCommand(ctx context.Context, dir string, args []string) *StepResult {
	if len(args) == 0 {
		return &StepResult{ExitCode: 1}
	}

	start := time.Now()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return &StepResult{
			Argv:     args,
			ExitCode: 1,
		}
	}

	// Read stdout
	stdoutData, _ := io.ReadAll(stdout)
	stderrData, _ := io.ReadAll(stderr)

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	duration := time.Since(start)

	return &StepResult{
		Argv:         args,
		ExitCode:     exitCode,
		DurationMs:   duration.Milliseconds(),
		StdoutSha256: sha256Hex(stdoutData),
		StderrSha256: sha256Hex(stderrData),
		StdoutBytes:  len(stdoutData),
		StderrBytes:  len(stderrData),
	}
}

func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func checkRequiredExecutables() error {
	required := []string{"go", "golangci-lint"}

	for _, exe := range required {
		if _, err := exec.LookPath(exe); err != nil {
			return fmt.Errorf("required executable not found: %s", exe)
		}
	}

	return nil
}

func parseCoverageProfile(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var totalStatements, coveredStatements int64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		// Format: name.go:line.column,line.column statements count
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		statements, _ := strconv.ParseInt(parts[1], 10, 64)
		count, _ := strconv.ParseInt(parts[2], 10, 64)

		totalStatements += statements
		if count > 0 {
			coveredStatements += statements
		}
	}

	if totalStatements == 0 {
		return 0, nil
	}

	return float64(coveredStatements) / float64(totalStatements) * 100, nil
}

func writeSnapshot(moduleRoot string, snapshot *Snapshot) error {
	snapshotsDir := filepath.Join(moduleRoot, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return err
	}

	// Generate filename
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("amux-test-%s.toml", timestamp)
	path := filepath.Join(snapshotsDir, filename)

	// Handle collision
	for i := 1; fileExists(path); i++ {
		filename = fmt.Sprintf("amux-test-%s-%d.toml", timestamp, i)
		path = filepath.Join(snapshotsDir, filename)
	}

	data, err := toml.Marshal(snapshot)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Snapshot written to: %s\n", path)
	return nil
}

func writeSnapshotToStdout(snapshot *Snapshot) {
	data, err := toml.Marshal(snapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling snapshot: %v\n", err)
		return
	}
	os.Stdout.Write(data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkRegression(moduleRoot string, newSnapshot *Snapshot, output io.Writer) error {
	snapshotsDir := filepath.Join(moduleRoot, "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no baseline snapshot exists")
		}
		return err
	}

	// Find the lexicographically greatest snapshot
	var baselinePath string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "amux-test-") && strings.HasSuffix(entry.Name(), ".toml") {
			candidate := filepath.Join(snapshotsDir, entry.Name())
			if candidate > baselinePath {
				baselinePath = candidate
			}
		}
	}

	if baselinePath == "" {
		return fmt.Errorf("no baseline snapshot exists")
	}

	// Load baseline
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return fmt.Errorf("read baseline: %w", err)
	}

	var baseline Snapshot
	if err := toml.Unmarshal(data, &baseline); err != nil {
		return fmt.Errorf("parse baseline: %w", err)
	}

	// Check for regressions
	var regressions []string

	// Step exit regressions
	regressions = append(regressions, checkStepRegression("go_mod_tidy", baseline.Steps.GoModTidy, newSnapshot.Steps.GoModTidy)...)
	regressions = append(regressions, checkStepRegression("go_vet", baseline.Steps.GoVet, newSnapshot.Steps.GoVet)...)
	regressions = append(regressions, checkStepRegression("golangci_lint", baseline.Steps.GolangciLint, newSnapshot.Steps.GolangciLint)...)
	regressions = append(regressions, checkStepRegression("tests_race", baseline.Steps.TestsRace, newSnapshot.Steps.TestsRace)...)
	regressions = append(regressions, checkStepRegression("tests", baseline.Steps.Tests, newSnapshot.Steps.Tests)...)
	regressions = append(regressions, checkStepRegression("coverage", baseline.Steps.Coverage, newSnapshot.Steps.Coverage)...)
	regressions = append(regressions, checkStepRegression("benchmarks", baseline.Steps.Benchmarks, newSnapshot.Steps.Benchmarks)...)

	// Coverage regression
	if baseline.Steps.Coverage != nil && newSnapshot.Steps.Coverage != nil {
		if baseline.Steps.Coverage.ExitCode == 0 && newSnapshot.Steps.Coverage.ExitCode == 0 {
			if baseline.Steps.Coverage.TotalPercent != nil && newSnapshot.Steps.Coverage.TotalPercent != nil {
				if *newSnapshot.Steps.Coverage.TotalPercent < *baseline.Steps.Coverage.TotalPercent {
					regressions = append(regressions, fmt.Sprintf(
						"coverage regression: %.2f%% -> %.2f%%",
						*baseline.Steps.Coverage.TotalPercent,
						*newSnapshot.Steps.Coverage.TotalPercent,
					))
				}
			}
		}
	}

	// Benchmark regressions (compare by pkg+name)
	baselineBenchmarks := make(map[string]Benchmark)
	for _, b := range baseline.Benchmarks {
		key := b.Pkg + ":" + b.Name
		baselineBenchmarks[key] = b
	}

	for _, b := range newSnapshot.Benchmarks {
		key := b.Pkg + ":" + b.Name
		if baseB, ok := baselineBenchmarks[key]; ok {
			if b.NsPerOp > baseB.NsPerOp {
				regressions = append(regressions, fmt.Sprintf(
					"benchmark %s ns_per_op regression: %.2f -> %.2f",
					key, baseB.NsPerOp, b.NsPerOp,
				))
			}
			if b.BytesPerOp != nil && baseB.BytesPerOp != nil && *b.BytesPerOp > *baseB.BytesPerOp {
				regressions = append(regressions, fmt.Sprintf(
					"benchmark %s bytes_per_op regression: %d -> %d",
					key, *baseB.BytesPerOp, *b.BytesPerOp,
				))
			}
			if b.AllocsPerOp != nil && baseB.AllocsPerOp != nil && *b.AllocsPerOp > *baseB.AllocsPerOp {
				regressions = append(regressions, fmt.Sprintf(
					"benchmark %s allocs_per_op regression: %d -> %d",
					key, *baseB.AllocsPerOp, *b.AllocsPerOp,
				))
			}
		}
	}

	if len(regressions) > 0 {
		fmt.Fprintln(output, "Regressions detected:")
		for _, r := range regressions {
			fmt.Fprintf(output, "  - %s\n", r)
		}
		return fmt.Errorf("%d regressions detected", len(regressions))
	}

	fmt.Fprintln(output, "No regressions detected")
	return nil
}

func checkStepRegression(name string, baseline, current *StepResult) []string {
	if baseline == nil || current == nil {
		return nil
	}

	if baseline.ExitCode == 0 && current.ExitCode != 0 {
		return []string{fmt.Sprintf("step %s exit regression: 0 -> %d", name, current.ExitCode)}
	}

	return nil
}

func parseFlags(args []string) (map[string]string, []string) {
	flags := make(map[string]string)
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			key := arg[2:]
			if idx := strings.Index(key, "="); idx >= 0 {
				flags[key[:idx]] = key[idx+1:]
			} else {
				flags[key] = "true"
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			flags[arg[1:]] = "true"
		} else {
			positional = append(positional, arg)
		}
	}

	return flags, positional
}

func hasFlag(flags map[string]string, name string) bool {
	_, ok := flags[name]
	return ok
}

// BenchmarkRegex matches benchmark output lines.
var BenchmarkRegex = regexp.MustCompile(`^Benchmark(\w+)(?:-\d+)?\s+(\d+)\s+([\d.]+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)

// ParseBenchmarkOutput parses go test -bench output into Benchmark entries.
func ParseBenchmarkOutput(output string, pkg string) []Benchmark {
	var benchmarks []Benchmark

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		matches := BenchmarkRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		name := matches[1]
		iterations, _ := strconv.ParseInt(matches[2], 10, 64)
		nsPerOp, _ := strconv.ParseFloat(matches[3], 64)

		b := Benchmark{
			Name:       name,
			Pkg:        pkg,
			NsPerOp:    nsPerOp,
			Iterations: iterations,
		}

		if matches[4] != "" {
			bytesPerOp, _ := strconv.ParseInt(matches[4], 10, 64)
			b.BytesPerOp = &bytesPerOp
		}

		if matches[5] != "" {
			allocsPerOp, _ := strconv.ParseInt(matches[5], 10, 64)
			b.AllocsPerOp = &allocsPerOp
		}

		benchmarks = append(benchmarks, b)
	}

	return benchmarks
}
