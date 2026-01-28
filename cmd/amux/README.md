- `func benchKey(item benchmark) string`
- `func benchmarkIndex(items []benchmark) map[string]benchmark`
- `func encodeSnapshot(s snapshot) (string, error)`
- `func ensureExecutables(names []string) error`
- `func findBaselineSnapshot(moduleRoot string, currentPath string) (string, error)`
- `func findModuleRoot() (string, error)`
- `func formatFloat(value float64) string`
- `func formatStringArray(values []string) string`
- `func hashBytes(data []byte) string`
- `func main()`
- `func parseCoverageTotal(profilePath string) (float64, error)`
- `func parseStringArray(value any) ([]string, bool)`
- `func quoteString(value string) string`
- `func runSteps(moduleRoot string) (map[string]stepResult, error)`
- `func runTest(args []string) error`
- `func summarizeSteps(w io.Writer, steps map[string]stepResult) error`
- `func toInt(value any) int`
- `func writeSnapshotFile(moduleRoot string, snap snapshot) (string, error)`
- `specVersion`
- `type benchmark`
- `type regressionEntry`
- `type snapshotMeta`
- `type snapshotStep`
- `type snapshot`
- `type stepResult`

### Constants

#### specVersion

```go
const specVersion = "v1.22"
```


### Functions

#### benchKey

```go
func benchKey(item benchmark) string
```

#### benchmarkIndex

```go
func benchmarkIndex(items []benchmark) map[string]benchmark
```

#### encodeSnapshot

```go
func encodeSnapshot(s snapshot) (string, error)
```

#### ensureExecutables

```go
func ensureExecutables(names []string) error
```

#### findBaselineSnapshot

```go
func findBaselineSnapshot(moduleRoot string, currentPath string) (string, error)
```

#### findModuleRoot

```go
func findModuleRoot() (string, error)
```

#### formatFloat

```go
func formatFloat(value float64) string
```

#### formatStringArray

```go
func formatStringArray(values []string) string
```

#### hashBytes

```go
func hashBytes(data []byte) string
```

#### main

```go
func main()
```

#### parseCoverageTotal

```go
func parseCoverageTotal(profilePath string) (float64, error)
```

#### parseStringArray

```go
func parseStringArray(value any) ([]string, bool)
```

#### quoteString

```go
func quoteString(value string) string
```

#### runSteps

```go
func runSteps(moduleRoot string) (map[string]stepResult, error)
```

#### runTest

```go
func runTest(args []string) error
```

#### summarizeSteps

```go
func summarizeSteps(w io.Writer, steps map[string]stepResult) error
```

#### toInt

```go
func toInt(value any) int
```

#### writeSnapshotFile

```go
func writeSnapshotFile(moduleRoot string, snap snapshot) (string, error)
```


## type benchmark

```go
type benchmark struct {
	Name        string
	Pkg         string
	NsPerOp     float64
	Iterations  int
	BytesPerOp  *int
	AllocsPerOp *int
}
```

### Functions returning benchmark

#### decodeBenchmarks

```go
func decodeBenchmarks(raw any) []benchmark
```

#### parseBenchmarks

```go
func parseBenchmarks(output []byte) []benchmark
```


## type regressionEntry

```go
type regressionEntry struct {
	Metric   string
	Baseline string
	Current  string
}
```

### Functions returning regressionEntry

#### checkRegression

```go
func checkRegression(moduleRoot string, current snapshot, currentPath string) ([]regressionEntry, error)
```

#### regressions

```go
func regressions(baseline snapshot, current snapshot) []regressionEntry
```


## type snapshot

```go
type snapshot struct {
	Meta       snapshotMeta
	Steps      map[string]snapshotStep
	Benchmarks []benchmark
}
```

### Functions returning snapshot

#### buildSnapshot

```go
func buildSnapshot(moduleRoot string, steps map[string]stepResult) snapshot
```

#### decodeSnapshot

```go
func decodeSnapshot(raw map[string]any) (snapshot, error)
```

#### readSnapshot

```go
func readSnapshot(path string) (snapshot, error)
```


## type snapshotMeta

```go
type snapshotMeta struct {
	CreatedAt   time.Time
	ModuleRoot  string
	SpecVersion string
}
```

## type snapshotStep

```go
type snapshotStep struct {
	Argv         []string
	ExitCode     int
	DurationMS   int64
	StdoutSHA256 string
	StderrSHA256 string
	StdoutBytes  int
	StderrBytes  int
	TotalPercent *float64
}
```

## type stepResult

```go
type stepResult struct {
	argv            []string
	stdout          []byte
	stderr          []byte
	exitCode        int
	duration        time.Duration
	coverageProfile string
}
```

### Functions returning stepResult

#### runStep

```go
func runStep(moduleRoot string, argv []string) stepResult
```


