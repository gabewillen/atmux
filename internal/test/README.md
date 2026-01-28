# package test

`import "github.com/agentflare-ai/amux/internal/test"`

Package test implements the 'amux test' CLI subcommand.

- `func Run(config *TestConfig) error` — Run executes the test command with the given configuration.
- `func findPreviousSnapshot(currentPath string) (string, error)` — findPreviousSnapshot finds the most recent snapshot before the given one.
- `func getDefaultSnapshotPath() string` — getDefaultSnapshotPath returns the default snapshot file path.
- `func getDependencies() []string` — getDependencies returns a list of key dependencies.
- `func printSummary(tests *TestResults)` — printSummary prints a test summary.
- `func runRegression(config *TestConfig, currentOutputPath string) error` — runRegression compares current results with previous snapshot.
- `func runTestSuite(config *TestConfig, outputPath string) error` — runTestSuite executes the full test suite.
- `func saveSnapshot(snapshot *Snapshot, path string) error` — saveSnapshot writes a snapshot to a TOML file.
- `func validateConfig(config *TestConfig) error` — validateConfig validates the test configuration.
- `type BuildInfo` — BuildInfo contains build and version information.
- `type Command` — Command represents a command that was executed.
- `type RegressionInfo` — RegressionInfo contains regression comparison results.
- `type Regression` — Regression represents a single regression.
- `type Snapshot` — Snapshot represents a test execution snapshot.
- `type Summary` — Summary contains test execution summary.
- `type TestConfig` — TestConfig holds configuration for the test command.
- `type TestResults` — TestResults contains results of various test types.

### Functions

#### Run

```go
func Run(config *TestConfig) error
```

Run executes the test command with the given configuration.

#### findPreviousSnapshot

```go
func findPreviousSnapshot(currentPath string) (string, error)
```

findPreviousSnapshot finds the most recent snapshot before the given one.

#### getDefaultSnapshotPath

```go
func getDefaultSnapshotPath() string
```

getDefaultSnapshotPath returns the default snapshot file path.

#### getDependencies

```go
func getDependencies() []string
```

getDependencies returns a list of key dependencies.

#### printSummary

```go
func printSummary(tests *TestResults)
```

printSummary prints a test summary.

#### runRegression

```go
func runRegression(config *TestConfig, currentOutputPath string) error
```

runRegression compares current results with previous snapshot.

#### runTestSuite

```go
func runTestSuite(config *TestConfig, outputPath string) error
```

runTestSuite executes the full test suite.

#### saveSnapshot

```go
func saveSnapshot(snapshot *Snapshot, path string) error
```

saveSnapshot writes a snapshot to a TOML file.

#### validateConfig

```go
func validateConfig(config *TestConfig) error
```

validateConfig validates the test configuration.


## type BuildInfo

```go
type BuildInfo struct {
	GoVersion string `toml:"go_version"`
	GitCommit string `toml:"git_commit"`
	GitBranch string `toml:"git_branch"`
	BuildTime string `toml:"build_time"`
}
```

BuildInfo contains build and version information.

### Functions returning BuildInfo

#### getBuildInfo

```go
func getBuildInfo() *BuildInfo
```

getBuildInfo returns build information.


## type Command

```go
type Command struct {
	Name     string   `toml:"name"`
	Args     []string `toml:"args"`
	ExitCode int      `toml:"exit_code"`
	Duration string   `toml:"duration"`
	Output   string   `toml:"output,omitempty"`
	Error    string   `toml:"error,omitempty"`
}
```

Command represents a command that was executed.

## type Regression

```go
type Regression struct {
	Test    string `json:"test"`
	Message string `json:"message"`
	Before  int    `json:"before"`
	After   int    `json:"after"`
}
```

Regression represents a single regression.

## type RegressionInfo

```go
type RegressionInfo struct {
	HasRegressions bool         `json:"has_regressions"`
	Regressions    []Regression `json:"regressions"`
}
```

RegressionInfo contains regression comparison results.

### Functions returning RegressionInfo

#### compareSnapshots

```go
func compareSnapshots(prev, current *Snapshot) *RegressionInfo
```

compareSnapshots compares two snapshots and identifies regressions.


## type Snapshot

```go
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
```

Snapshot represents a test execution snapshot.

## type Summary

```go
type Summary struct {
	Passed   int    `toml:"passed"`
	Failed   int    `toml:"failed"`
	Skipped  int    `toml:"skipped"`
	Duration string `toml:"duration"`
	Output   string `toml:"output,omitempty"`
}
```

Summary contains test execution summary.

### Functions returning Summary

#### runCommand

```go
func runCommand(name string, args []string, testType string) *Summary
```

runCommand executes a command and returns a summary.

#### runConformanceTests

```go
func runConformanceTests() *Summary
```

runConformanceTests runs the conformance test suite.


## type TestConfig

```go
type TestConfig struct {
	// Flags
	NoSnapshot bool
	Regression bool
	Output     string
	Quiet      bool

	// Internal state
	snapshotPath string
}
```

TestConfig holds configuration for the test command.

### Functions returning TestConfig

#### ParseFlags

```go
func ParseFlags(args []string) (*TestConfig, error)
```

ParseFlags parses command line flags for the test command.


## type TestResults

```go
type TestResults struct {
	Unit        *Summary `toml:"unit"`
	Integration *Summary `toml:"integration"`
	Lint        *Summary `toml:"lint"`
	Vet         *Summary `toml:"vet"`
	Coverage    *Summary `toml:"coverage"`
	Benchmark   *Summary `toml:"benchmark"`
	Conformance *Summary `toml:"conformance"`
}
```

TestResults contains results of various test types.

