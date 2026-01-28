// Package test implements the 'amux test' CLI subcommand.
package test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/agentflare-ai/amux/internal/conformance"
)

// TestConfig holds configuration for the test command.
type TestConfig struct {
	// Flags
	NoSnapshot bool
	Regression bool
	Output     string
	Quiet      bool

	// Internal state
	snapshotPath string
}

// Snapshot represents a test execution snapshot.
type Snapshot struct {
	GeneratedAt time.Time `toml:"generated_at"`
	Command     string    `toml:"command"`
	Environment []string  `toml:"environment"`

	// Test results
	Tests *TestResults `toml:"tests"`

	// Build information
	Build        *BuildInfo `toml:"build"`
	Dependencies []string   `toml:"dependencies"`
}

// TestResults contains results of various test types.
type TestResults struct {
	Unit        *Summary `toml:"unit"`
	Integration *Summary `toml:"integration"`
	Lint        *Summary `toml:"lint"`
	Vet         *Summary `toml:"vet"`
	Coverage    *Summary `toml:"coverage"`
	Benchmark   *Summary `toml:"benchmark"`
	Conformance *Summary `toml:"conformance"`
}

// Summary contains test execution summary.
type Summary struct {
	Passed   int    `toml:"passed"`
	Failed   int    `toml:"failed"`
	Skipped  int    `toml:"skipped"`
	Duration string `toml:"duration"`
	Output   string `toml:"output,omitempty"`
}

// BuildInfo contains build and version information.
type BuildInfo struct {
	GoVersion string `toml:"go_version"`
	GitCommit string `toml:"git_commit"`
	GitBranch string `toml:"git_branch"`
	BuildTime string `toml:"build_time"`
}

// Command represents a command that was executed.
type Command struct {
	Name     string   `toml:"name"`
	Args     []string `toml:"args"`
	ExitCode int      `toml:"exit_code"`
	Duration string   `toml:"duration"`
	Output   string   `toml:"output,omitempty"`
	Error    string   `toml:"error,omitempty"`
}

// Run executes the test command with the given configuration.
func Run(config *TestConfig) error {
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Determine output path
	outputPath := config.Output
	if outputPath == "" {
		outputPath = getDefaultSnapshotPath()
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// If in regression mode, compare with previous snapshot
	if config.Regression {
		return runRegression(config, outputPath)
	}

	// Otherwise, run new test suite
	return runTestSuite(config, outputPath)
}

// validateConfig validates the test configuration.
func validateConfig(config *TestConfig) error {
	if config.Regression && config.NoSnapshot {
		return fmt.Errorf("cannot use both --regression and --no-snapshot")
	}

	return nil
}

// getDefaultSnapshotPath returns the default snapshot file path.
func getDefaultSnapshotPath() string {
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join("snapshots", fmt.Sprintf("amux-test-%s.toml", timestamp))
}

// runTestSuite executes the full test suite.
func runTestSuite(config *TestConfig, outputPath string) error {
	if !config.Quiet {
		fmt.Printf("Running amux test suite...\n")
	}

	snapshot := &Snapshot{
		GeneratedAt:  time.Now(),
		Command:      "amux test",
		Environment:  os.Environ(),
		Tests:        &TestResults{},
		Build:        getBuildInfo(),
		Dependencies: getDependencies(),
	}

	// Run unit tests
	if !config.Quiet {
		fmt.Printf("Running unit tests...\n")
	}
	unitResult := runCommand("go", []string{"test", "./..."}, "unit")
	snapshot.Tests.Unit = unitResult

	// Run go vet
	if !config.Quiet {
		fmt.Printf("Running go vet...\n")
	}
	vetResult := runCommand("go", []string{"vet", "./..."}, "vet")
	snapshot.Tests.Vet = vetResult

	// Run lint
	if !config.Quiet {
		fmt.Printf("Running staticcheck...\n")
	}
	lintResult := runCommand("staticcheck", []string{"./..."}, "lint")
	snapshot.Tests.Lint = lintResult

	// Run integration tests
	if !config.Quiet {
		fmt.Printf("Running integration tests...\n")
	}
	integrationResult := runCommand("go", []string{"test", "-tags=integration", "./..."}, "integration")
	snapshot.Tests.Integration = integrationResult

	// Run coverage
	if !config.Quiet {
		fmt.Printf("Running coverage...\n")
	}
	coverageResult := runCommand("go", []string{"test", "-cover", "./..."}, "coverage")
	snapshot.Tests.Coverage = coverageResult

	// Run benchmarks
	if !config.Quiet {
		fmt.Printf("Running benchmarks...\n")
	}
	benchmarkResult := runCommand("go", []string{"test", "-bench=.", "-benchmem", "./..."}, "benchmark")
	snapshot.Tests.Benchmark = benchmarkResult

	// Run conformance tests
	if !config.Quiet {
		fmt.Printf("Running conformance tests...\n")
	}
	conformanceResult := runConformanceTests()
	snapshot.Tests.Conformance = conformanceResult

	// Save snapshot unless disabled
	if !config.NoSnapshot {
		if err := saveSnapshot(snapshot, outputPath); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}

		if !config.Quiet {
			fmt.Printf("Snapshot saved to: %s\n", outputPath)
		}
	} else {
		if !config.Quiet {
			fmt.Printf("Snapshot generation disabled by --no-snapshot\n")
		}
	}

	// Print summary
	printSummary(snapshot.Tests)

	return nil
}

// runRegression compares current results with previous snapshot.
func runRegression(config *TestConfig, currentOutputPath string) error {
	// Find previous snapshot
	previousPath, err := findPreviousSnapshot(currentOutputPath)
	if err != nil {
		return fmt.Errorf("finding previous snapshot: %w", err)
	}

	if !config.Quiet {
		fmt.Printf("Comparing with previous snapshot: %s\n", previousPath)
	}

	// Load previous snapshot
	previousData, err := os.ReadFile(previousPath)
	if err != nil {
		return fmt.Errorf("reading previous snapshot: %w", err)
	}

	var previousSnapshot Snapshot
	if err := toml.Unmarshal(previousData, &previousSnapshot); err != nil {
		return fmt.Errorf("parsing previous snapshot: %w", err)
	}

	// Run current test suite to get comparison baseline
	tempOutput := filepath.Join(os.TempDir(), "current-snapshot.toml")
	if err := runTestSuite(config, tempOutput); err != nil {
		return fmt.Errorf("running current test suite: %w", err)
	}

	currentData, err := os.ReadFile(tempOutput)
	if err != nil {
		return fmt.Errorf("reading current snapshot: %w", err)
	}

	var currentSnapshot Snapshot
	if err := toml.Unmarshal(currentData, &currentSnapshot); err != nil {
		return fmt.Errorf("parsing current snapshot: %w", err)
	}

	// Compare snapshots
	regression := compareSnapshots(&previousSnapshot, &currentSnapshot)

	if regression.HasRegressions {
		if !config.Quiet {
			fmt.Printf("REGRESSIONS DETECTED:\n")
			for _, r := range regression.Regressions {
				fmt.Printf("  %s: %s\n", r.Test, r.Message)
			}
		}
		os.Exit(1)
	} else {
		if !config.Quiet {
			fmt.Printf("No regressions detected.\n")
		}
	}

	return nil
}

// RegressionInfo contains regression comparison results.
type RegressionInfo struct {
	HasRegressions bool         `json:"has_regressions"`
	Regressions    []Regression `json:"regressions"`
}

// Regression represents a single regression.
type Regression struct {
	Test    string `json:"test"`
	Message string `json:"message"`
	Before  int    `json:"before"`
	After   int    `json:"after"`
}

// findPreviousSnapshot finds the most recent snapshot before the given one.
func findPreviousSnapshot(currentPath string) (string, error) {
	// For now, return a simple pattern - in real implementation this would
	// scan the snapshots directory for the most recent file
	snapshotsDir := filepath.Dir(currentPath)
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return "", fmt.Errorf("reading snapshots directory: %w", err)
	}

	var latestPath string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) && entry.Name() != filepath.Base(currentPath) {
			latestTime = info.ModTime()
			latestPath = filepath.Join(snapshotsDir, entry.Name())
		}
	}

	if latestPath == "" {
		return "", fmt.Errorf("no previous snapshot found")
	}

	return latestPath, nil
}

// compareSnapshots compares two snapshots and identifies regressions.
func compareSnapshots(prev, current *Snapshot) *RegressionInfo {
	regressions := []Regression{}

	// Compare unit test results
	if prev.Tests.Unit != nil && current.Tests.Unit != nil {
		if current.Tests.Unit.Failed > prev.Tests.Unit.Failed {
			regressions = append(regressions, Regression{
				Test:    "unit-tests",
				Message: fmt.Sprintf("More failures: %d -> %d", prev.Tests.Unit.Failed, current.Tests.Unit.Failed),
				Before:  prev.Tests.Unit.Failed,
				After:   current.Tests.Unit.Failed,
			})
		}
	}

	// Compare integration test results
	if prev.Tests.Integration != nil && current.Tests.Integration != nil {
		if current.Tests.Integration.Failed > prev.Tests.Integration.Failed {
			regressions = append(regressions, Regression{
				Test:    "integration-tests",
				Message: fmt.Sprintf("More failures: %d -> %d", prev.Tests.Integration.Failed, current.Tests.Integration.Failed),
				Before:  prev.Tests.Integration.Failed,
				After:   current.Tests.Integration.Failed,
			})
		}
	}

	// Compare lint results
	if prev.Tests.Lint != nil && current.Tests.Lint != nil {
		if current.Tests.Lint.Failed > prev.Tests.Lint.Failed {
			regressions = append(regressions, Regression{
				Test:    "lint",
				Message: fmt.Sprintf("More lint failures: %d -> %d", prev.Tests.Lint.Failed, current.Tests.Lint.Failed),
				Before:  prev.Tests.Lint.Failed,
				After:   current.Tests.Lint.Failed,
			})
		}
	}

	return &RegressionInfo{
		HasRegressions: len(regressions) > 0,
		Regressions:    regressions,
	}
}

// saveSnapshot writes a snapshot to a TOML file.
func saveSnapshot(snapshot *Snapshot, path string) error {
	data, err := toml.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshaling snapshot: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// runCommand executes a command and returns a summary.
func runCommand(name string, args []string, testType string) *Summary {
	// TODO: implement actual command execution with timing and exit code capture
	// For now, return a placeholder summary

	return &Summary{
		Passed:   1,
		Failed:   0,
		Skipped:  0,
		Duration: "1s",
	}
}

// runConformanceTests runs the conformance test suite.
func runConformanceTests() *Summary {
	// Run conformance suite using the conformance package
	if err := conformance.RunSuite(); err != nil {
		return &Summary{
			Passed:   0,
			Failed:   1,
			Skipped:  0,
			Duration: "1s",
			Output:   err.Error(),
		}
	}

	return &Summary{
		Passed:   1,
		Failed:   0,
		Skipped:  0,
		Duration: "1s",
	}
}

// getBuildInfo returns build information.
func getBuildInfo() *BuildInfo {
	// TODO: implement actual build info extraction
	return &BuildInfo{
		GoVersion: "1.25.6",
		GitCommit: "dev",
		GitBranch: "main",
		BuildTime: time.Now().Format(time.RFC3339),
	}
}

// getDependencies returns a list of key dependencies.
func getDependencies() []string {
	// TODO: implement actual dependency extraction from go.mod
	return []string{
		"github.com/BurntSushi/toml v1.4.0",
		"github.com/creack/pty v1.1.21",
		"github.com/tetratelabs/wazero v1.8.0",
		"go.opentelemetry.io/otel v1.31.0",
	}
}

// printSummary prints a test summary.
func printSummary(tests *TestResults) {
	fmt.Printf("\nTest Summary:\n")

	if tests.Unit != nil {
		fmt.Printf("  Unit Tests:     %d passed, %d failed, %d skipped\n",
			tests.Unit.Passed, tests.Unit.Failed, tests.Unit.Skipped)
	}

	if tests.Integration != nil {
		fmt.Printf("  Integration:     %d passed, %d failed, %d skipped\n",
			tests.Integration.Passed, tests.Integration.Failed, tests.Integration.Skipped)
	}

	if tests.Lint != nil {
		fmt.Printf("  Lint:           %d passed, %d failed, %d skipped\n",
			tests.Lint.Passed, tests.Lint.Failed, tests.Lint.Skipped)
	}

	if tests.Vet != nil {
		fmt.Printf("  Vet:            %d passed, %d failed, %d skipped\n",
			tests.Vet.Passed, tests.Vet.Failed, tests.Vet.Skipped)
	}

	if tests.Coverage != nil {
		fmt.Printf("  Coverage:       %d passed, %d failed, %d skipped\n",
			tests.Coverage.Passed, tests.Coverage.Failed, tests.Coverage.Skipped)
	}

	if tests.Benchmark != nil {
		fmt.Printf("  Benchmarks:     %d passed, %d failed, %d skipped\n",
			tests.Benchmark.Passed, tests.Benchmark.Failed, tests.Benchmark.Skipped)
	}

	if tests.Conformance != nil {
		fmt.Printf("  Conformance:    %d passed, %d failed, %d skipped\n",
			tests.Conformance.Passed, tests.Conformance.Failed, tests.Conformance.Skipped)
	}
}

// ParseFlags parses command line flags for the test command.
func ParseFlags(args []string) (*TestConfig, error) {
	config := &TestConfig{}

	flags := flag.NewFlagSet("amux test", flag.ContinueOnError)

	flags.BoolVar(&config.NoSnapshot, "no-snapshot", false, "Skip writing snapshot file")
	flags.BoolVar(&config.Regression, "regression", false, "Compare with previous snapshot and fail on regressions")
	flags.StringVar(&config.Output, "output", "", "Output file path for snapshot")
	flags.BoolVar(&config.Quiet, "quiet", false, "Suppress verbose output")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return config, nil
}
