Package main implements the amux CLI client.
agent.go implements amux agent add (spec §5.2).

Command amux is the main CLI client for amux.

Package main implements the amux CLI client.
test.go implements amux test per spec §12.6 (snapshot schema, regression, required sequence).

- `benchLineRe` — benchmark line: "BenchmarkName-8   1000000  1234 ns/op  56 B/op  2 allocs/op" (tabs or spaces)
- `func checkRegression(moduleRoot, currentPath string, current TestSnapshot, noSnapshot bool) error` — checkRegression: baseline = lexicographically greatest amux-test-*.toml excluding current (§12.6.5).
- `func checkRegressionRules(prev, curr TestSnapshot) []string`
- `func findModuleRoot() (string, error)`
- `func main()`
- `func run() error`
- `func runAgent(args []string) error`
- `func runAgentAdd(args []string) error`
- `func runCoverFunc(moduleRoot, coverPath string) *float64`
- `func runStepSequence(moduleRoot string) (StepsSnapshot, *float64, []BenchmarkEntry)`
- `func runTest(args []string) error`
- `func snapshotPathFor(moduleRoot string) string` — snapshotPathFor returns path for new snapshot: amux-test-<UTC>.toml with -1, -2 if exists (§12.6.3).
- `type BenchmarkEntry`
- `type MetaSnapshot`
- `type StepResult`
- `type StepsSnapshot`
- `type TestSnapshot` — Snapshot schema per spec §12.6.4: [meta], [steps.*], [[benchmarks]].

### Variables

#### benchLineRe

```go
var benchLineRe = regexp.MustCompile(`^Benchmark(\S+)\s+(\d+)\s+(\d+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)
```

benchmark line: "BenchmarkName-8   1000000  1234 ns/op  56 B/op  2 allocs/op" (tabs or spaces)


### Functions

#### checkRegression

```go
func checkRegression(moduleRoot, currentPath string, current TestSnapshot, noSnapshot bool) error
```

checkRegression: baseline = lexicographically greatest amux-test-*.toml excluding current (§12.6.5).

#### checkRegressionRules

```go
func checkRegressionRules(prev, curr TestSnapshot) []string
```

#### findModuleRoot

```go
func findModuleRoot() (string, error)
```

#### main

```go
func main()
```

#### run

```go
func run() error
```

#### runAgent

```go
func runAgent(args []string) error
```

#### runAgentAdd

```go
func runAgentAdd(args []string) error
```

#### runCoverFunc

```go
func runCoverFunc(moduleRoot, coverPath string) *float64
```

#### runStepSequence

```go
func runStepSequence(moduleRoot string) (StepsSnapshot, *float64, []BenchmarkEntry)
```

#### runTest

```go
func runTest(args []string) error
```

#### snapshotPathFor

```go
func snapshotPathFor(moduleRoot string) string
```

snapshotPathFor returns path for new snapshot: amux-test-<UTC>.toml with -1, -2 if exists (§12.6.3).


## type BenchmarkEntry

```go
type BenchmarkEntry struct {
	Name        string  `toml:"name"`
	Pkg         string  `toml:"pkg"`
	NsPerOp     float64 `toml:"ns_per_op"`
	Iterations  int     `toml:"iterations"`
	BytesPerOp  *int    `toml:"bytes_per_op,omitempty"`
	AllocsPerOp *int    `toml:"allocs_per_op,omitempty"`
}
```

### Functions returning BenchmarkEntry

#### parseBenchmarkOutput

```go
func parseBenchmarkOutput(stdout []byte) []BenchmarkEntry
```


## type MetaSnapshot

```go
type MetaSnapshot struct {
	CreatedAt   string `toml:"created_at"`
	ModuleRoot  string `toml:"module_root"`
	SpecVersion string `toml:"spec_version"`
}
```

## type StepResult

```go
type StepResult struct {
	Argv         []string `toml:"argv"`
	ExitCode     int      `toml:"exit_code"`
	DurationMs   int      `toml:"duration_ms"`
	StdoutSha256 string   `toml:"stdout_sha256"`
	StderrSha256 string   `toml:"stderr_sha256"`
	StdoutBytes  int      `toml:"stdout_bytes"`
	StderrBytes  int      `toml:"stderr_bytes"`
	TotalPercent *float64 `toml:"total_percent,omitempty"`
}
```

### Functions returning StepResult

#### runStep

```go
func runStep(moduleRoot string, name string, args ...string) StepResult
```

#### runStepCapture

```go
func runStepCapture(moduleRoot string, name string, args ...string) (stdout []byte, result StepResult)
```


## type StepsSnapshot

```go
type StepsSnapshot struct {
	GoModTidy    StepResult `toml:"go_mod_tidy"`
	GoVet        StepResult `toml:"go_vet"`
	GolangciLint StepResult `toml:"golangci_lint"`
	TestsRace    StepResult `toml:"tests_race"`
	Tests        StepResult `toml:"tests"`
	Coverage     StepResult `toml:"coverage"`
	Benchmarks   StepResult `toml:"benchmarks"`
}
```

## type TestSnapshot

```go
type TestSnapshot struct {
	Meta       MetaSnapshot     `toml:"meta"`
	Steps      StepsSnapshot    `toml:"steps"`
	Benchmarks []BenchmarkEntry `toml:"benchmarks"`
}
```

Snapshot schema per spec §12.6.4: [meta], [steps.*], [[benchmarks]].

