# package test

`import "github.com/agentflare-ai/amux/internal/cli/test"`

Package test implements the 'amux test' subcommand.

The amux test command runs a standardized Go verification suite for the
current Go module and emits a TOML "test snapshot" capturing the results.

See spec §12.6 for the full specification.

- `BenchmarkRegex` — BenchmarkRegex matches benchmark output lines.
- `func Run(ctx context.Context, args []string) error` — Run executes the amux test command.
- `func checkRegression(moduleRoot string, newSnapshot *Snapshot, output io.Writer) error`
- `func checkRequiredExecutables() error`
- `func checkStepRegression(name string, baseline, current *StepResult) []string`
- `func fileExists(path string) bool`
- `func hasFlag(flags map[string]string, name string) bool`
- `func parseCoverageProfile(path string) (float64, error)`
- `func parseFlags(args []string) (map[string]string, []string)`
- `func sha256Hex(data []byte) string`
- `func writeSnapshot(moduleRoot string, snapshot *Snapshot) error`
- `func writeSnapshotToStdout(snapshot *Snapshot)`
- `pkgLineRegex` — pkgLineRegex matches "pkg: <package>" lines in go test -bench output.
- `type Benchmark` — Benchmark contains benchmark results.
- `type MetaInfo` — MetaInfo contains snapshot metadata.
- `type Snapshot` — Snapshot represents the test snapshot TOML structure.
- `type StepResult` — StepResult contains the result of a single step.
- `type Steps` — Steps contains results for each step.

### Variables

#### BenchmarkRegex

```go
var BenchmarkRegex = regexp.MustCompile(`^Benchmark(\w+)(?:-\d+)?\s+(\d+)\s+([\d.]+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)
```

BenchmarkRegex matches benchmark output lines.

#### pkgLineRegex

```go
var pkgLineRegex = regexp.MustCompile(`^pkg:\s+(\S+)`)
```

pkgLineRegex matches "pkg: <package>" lines in go test -bench output.


### Functions

#### Run

```go
func Run(ctx context.Context, args []string) error
```

Run executes the amux test command.

#### checkRegression

```go
func checkRegression(moduleRoot string, newSnapshot *Snapshot, output io.Writer) error
```

#### checkRequiredExecutables

```go
func checkRequiredExecutables() error
```

#### checkStepRegression

```go
func checkStepRegression(name string, baseline, current *StepResult) []string
```

#### fileExists

```go
func fileExists(path string) bool
```

#### hasFlag

```go
func hasFlag(flags map[string]string, name string) bool
```

#### parseCoverageProfile

```go
func parseCoverageProfile(path string) (float64, error)
```

#### parseFlags

```go
func parseFlags(args []string) (map[string]string, []string)
```

#### sha256Hex

```go
func sha256Hex(data []byte) string
```

#### writeSnapshot

```go
func writeSnapshot(moduleRoot string, snapshot *Snapshot) error
```

#### writeSnapshotToStdout

```go
func writeSnapshotToStdout(snapshot *Snapshot)
```


## type Benchmark

```go
type Benchmark struct {
	Name        string  `toml:"name"`
	Pkg         string  `toml:"pkg"`
	NsPerOp     float64 `toml:"ns_per_op"`
	Iterations  int64   `toml:"iterations"`
	BytesPerOp  *int64  `toml:"bytes_per_op,omitempty"`
	AllocsPerOp *int64  `toml:"allocs_per_op,omitempty"`
}
```

Benchmark contains benchmark results.

### Functions returning Benchmark

#### ParseBenchmarkOutput

```go
func ParseBenchmarkOutput(output string, pkg string) []Benchmark
```

ParseBenchmarkOutput parses go test -bench output into Benchmark entries.

#### ParseBenchmarksMultiPkg

```go
func ParseBenchmarksMultiPkg(output string) []Benchmark
```

ParseBenchmarksMultiPkg parses go test -bench output that may contain
results from multiple packages (as produced by "go test -bench=. ./...").
Package context is tracked via "pkg:" lines emitted by the test runner.


## type MetaInfo

```go
type MetaInfo struct {
	CreatedAt   string `toml:"created_at"`
	ModuleRoot  string `toml:"module_root"`
	SpecVersion string `toml:"spec_version"`
}
```

MetaInfo contains snapshot metadata.

## type Snapshot

```go
type Snapshot struct {
	Meta       MetaInfo    `toml:"meta"`
	Steps      Steps       `toml:"steps"`
	Benchmarks []Benchmark `toml:"benchmarks,omitempty"`
}
```

Snapshot represents the test snapshot TOML structure.

### Functions returning Snapshot

#### NewSnapshot

```go
func NewSnapshot(moduleRoot string) *Snapshot
```

NewSnapshot creates a new snapshot with metadata initialized.


### Methods

#### Snapshot.HasFailures

```go
func () HasFailures() bool
```

HasFailures returns true if any step failed.

#### Snapshot.SetStep

```go
func () SetStep(key string, result *StepResult)
```

SetStep sets the result for a step.


## type StepResult

```go
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
```

StepResult contains the result of a single step.

### Functions returning StepResult

#### runCommand

```go
func runCommand(ctx context.Context, dir string, args []string) (*StepResult, []byte)
```


## type Steps

```go
type Steps struct {
	GoModTidy    *StepResult `toml:"go_mod_tidy"`
	GoVet        *StepResult `toml:"go_vet"`
	GolangciLint *StepResult `toml:"golangci_lint"`
	Staticcheck  *StepResult `toml:"staticcheck"`
	TestsRace    *StepResult `toml:"tests_race"`
	Tests        *StepResult `toml:"tests"`
	Coverage     *StepResult `toml:"coverage"`
	Benchmarks   *StepResult `toml:"benchmarks"`
}
```

Steps contains results for each step.

