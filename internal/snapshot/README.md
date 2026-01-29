# package snapshot

`import "github.com/stateforward/amux/internal/snapshot"`

Package snapshot implements the amux test snapshot functionality per spec §12.6.

- `SpecVersion` — SpecVersion is the version string recorded in snapshots.
- `func Compare(baseline, current *Snapshot) (bool, string)` — Compare compares two snapshots and returns a regression report per spec §12.6.5.
- `func FindLatestSnapshot(moduleRoot string) (string, error)` — FindLatestSnapshot finds the most recent snapshot file in the snapshots directory using lexicographic ordering of file names matching amux-test-*.toml per spec §12.6.5.
- `func GenerateSnapshotPath(moduleRoot string) string` — GenerateSnapshotPath generates a snapshot file path under <module_root>/snapshots with the name amux-test-<timestamp>.toml where timestamp is formatted as YYYYMMDDThhmmssZ per spec §12.6.3.
- `func Write(snapshot *Snapshot, path string) error` — Write writes a snapshot to a TOML file.
- `type BenchmarkMetrics` — BenchmarkMetrics captures metrics for a single benchmark.
- `type BenchmarkStep` — BenchmarkStep extends StepResult with benchmark metrics keyed by pkg/name.
- `type CoverageStep` — CoverageStep extends StepResult with total coverage percentage.
- `type Meta` — Meta contains top-level metadata for a snapshot per spec §12.6.4.
- `type Snapshot` — Snapshot represents a test snapshot per spec §12.6.
- `type StepResult` — StepResult represents the outcome of a single verification step.
- `type Steps` — Steps groups all required verification steps per spec §12.6.2.

### Constants

#### SpecVersion

```go
const SpecVersion = "v1.22"
```

SpecVersion is the version string recorded in snapshots.


### Functions

#### Compare

```go
func Compare(baseline, current *Snapshot) (bool, string)
```

Compare compares two snapshots and returns a regression report per spec §12.6.5.

Regression rules:
  - For any step present in both snapshots, if the baseline exit_code is 0 and the
    new exit_code is non-zero, this is a regression.
  - If both snapshots report coverage.exit_code = 0 and the new total_percent is
    less than the baseline total_percent, this is a coverage regression.
  - For any benchmark present in both snapshots, a regression occurs when any of:
      * ns_per_op increases
      * bytes_per_op increases (when present in both)
      * allocs_per_op increases (when present in both)

#### FindLatestSnapshot

```go
func FindLatestSnapshot(moduleRoot string) (string, error)
```

FindLatestSnapshot finds the most recent snapshot file in the snapshots
directory using lexicographic ordering of file names matching
amux-test-*.toml per spec §12.6.5.

#### GenerateSnapshotPath

```go
func GenerateSnapshotPath(moduleRoot string) string
```

GenerateSnapshotPath generates a snapshot file path under <module_root>/snapshots
with the name amux-test-<timestamp>.toml where timestamp is formatted as
YYYYMMDDThhmmssZ per spec §12.6.3. If a file with that name already exists,
a numeric suffix -1, -2, ... is appended before the .toml extension.

#### Write

```go
func Write(snapshot *Snapshot, path string) error
```

Write writes a snapshot to a TOML file.


## type BenchmarkMetrics

```go
type BenchmarkMetrics struct {
	NsPerOp     float64  `toml:"ns_per_op"`
	BytesPerOp  *float64 `toml:"bytes_per_op,omitempty"`
	AllocsPerOp *float64 `toml:"allocs_per_op,omitempty"`
}
```

BenchmarkMetrics captures metrics for a single benchmark.

## type BenchmarkStep

```go
type BenchmarkStep struct {
	StepResult
	Benchmarks map[string]BenchmarkMetrics `toml:"benchmarks"`
}
```

BenchmarkStep extends StepResult with benchmark metrics keyed by pkg/name.

### Functions returning BenchmarkStep

#### runBenchmarksStep

```go
func runBenchmarksStep(moduleRoot string) (BenchmarkStep, error)
```

runBenchmarksStep runs go test benchmarks and parses benchmark metrics into
a BenchmarkStep. Missing executables are treated as fatal errors; benchmark
failures do not affect other step results beyond the exit code.


## type CoverageStep

```go
type CoverageStep struct {
	StepResult
	TotalPercent float64 `toml:"total_percent"`
}
```

CoverageStep extends StepResult with total coverage percentage.

### Functions returning CoverageStep

#### runCoverageStep

```go
func runCoverageStep(moduleRoot string) (CoverageStep, error)
```

runCoverageStep runs the coverage command and computes total_percent via
`go tool cover -func`. It returns a CoverageStep and only treats missing
executables as fatal.


## type Meta

```go
type Meta struct {
	CreatedAt   time.Time `toml:"created_at"`
	ModuleRoot  string    `toml:"module_root"`
	SpecVersion string    `toml:"spec_version"`
}
```

Meta contains top-level metadata for a snapshot per spec §12.6.4.

## type Snapshot

```go
type Snapshot struct {
	Meta  Meta  `toml:"meta"`
	Steps Steps `toml:"steps"`
}
```

Snapshot represents a test snapshot per spec §12.6.

### Functions returning Snapshot

#### Create

```go
func Create(moduleRoot string) (*Snapshot, error)
```

Create creates a new snapshot by running the verification sequence per spec §12.6.

The command sequence is:
 1. go mod tidy
 2. go vet ./...
 3. golangci-lint run ./...
 4. go test -race ./...
 5. go test ./...
 6. go test ./... -coverprofile=<coverprofile_path>
 7. go test ./... -run=^$ -bench=. -benchmem ./...

All commands are executed with working directory set to moduleRoot. Failures are
recorded in the snapshot via exit codes; missing required executables (at minimum
`go` and `golangci-lint`) are treated as fatal errors and returned to the caller.

#### Read

```go
func Read(path string) (*Snapshot, error)
```

Read reads a snapshot from a TOML file.


## type StepResult

```go
type StepResult struct {
	ExitCode       int   `toml:"exit_code"`
	DurationMillis int64 `toml:"duration_ms"`
}
```

StepResult represents the outcome of a single verification step.

### Functions returning StepResult

#### runStep

```go
func runStep(moduleRoot, name string, args ...string) (StepResult, error)
```

runStep executes a command in moduleRoot and returns a StepResult. It treats
missing executables as fatal errors (returned to the caller) and records
non-zero exit codes in the StepResult.


## type Steps

```go
type Steps struct {
	GoModTidy    StepResult    `toml:"go_mod_tidy"`
	GoVet        StepResult    `toml:"go_vet"`
	GolangciLint StepResult    `toml:"golangci_lint"`
	TestsRace    StepResult    `toml:"tests_race"`
	Tests        StepResult    `toml:"tests"`
	Coverage     CoverageStep  `toml:"coverage"`
	Benchmarks   BenchmarkStep `toml:"benchmarks"`
}
```

Steps groups all required verification steps per spec §12.6.2.

