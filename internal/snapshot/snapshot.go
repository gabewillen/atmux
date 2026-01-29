// Package snapshot implements the amux test snapshot functionality per spec §12.6.
package snapshot

import
(
    "bytes"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "syscall"
    "time"

    "github.com/pelletier/go-toml/v2"

    amuxerrors "github.com/stateforward/amux/internal/errors"
)

// SpecVersion is the version string recorded in snapshots.
const SpecVersion = "v1.22"

// Meta contains top-level metadata for a snapshot per spec §12.6.4.
type Meta struct {
    CreatedAt   time.Time `toml:"created_at"`
    ModuleRoot  string    `toml:"module_root"`
    SpecVersion string    `toml:"spec_version"`
}

// StepResult represents the outcome of a single verification step.
type StepResult struct {
    ExitCode       int   `toml:"exit_code"`
    DurationMillis int64 `toml:"duration_ms"`
}

// CoverageStep extends StepResult with total coverage percentage.
type CoverageStep struct {
    StepResult
    TotalPercent float64 `toml:"total_percent"`
}

// BenchmarkMetrics captures metrics for a single benchmark.
type BenchmarkMetrics struct {
    NsPerOp     float64  `toml:"ns_per_op"`
    BytesPerOp  *float64 `toml:"bytes_per_op,omitempty"`
    AllocsPerOp *float64 `toml:"allocs_per_op,omitempty"`
}

// BenchmarkStep extends StepResult with benchmark metrics keyed by pkg/name.
type BenchmarkStep struct {
    StepResult
    Benchmarks map[string]BenchmarkMetrics `toml:"benchmarks"`
}

// Steps groups all required verification steps per spec §12.6.2.
type Steps struct {
    GoModTidy   StepResult    `toml:"go_mod_tidy"`
    GoVet       StepResult    `toml:"go_vet"`
    GolangciLint StepResult   `toml:"golangci_lint"`
    TestsRace   StepResult    `toml:"tests_race"`
    Tests       StepResult    `toml:"tests"`
    Coverage    CoverageStep  `toml:"coverage"`
    Benchmarks  BenchmarkStep `toml:"benchmarks"`
}

// Snapshot represents a test snapshot per spec §12.6.
type Snapshot struct {
    Meta  Meta  `toml:"meta"`
    Steps Steps `toml:"steps"`
}

// Create creates a new snapshot by running the verification sequence per spec §12.6.
//
// The command sequence is:
//  1. go mod tidy
//  2. go vet ./...
//  3. golangci-lint run ./...
//  4. go test -race ./...
//  5. go test ./...
//  6. go test ./... -coverprofile=<coverprofile_path>
//  7. go test ./... -run=^$ -bench=. -benchmem ./...
//
// All commands are executed with working directory set to moduleRoot. Failures are
// recorded in the snapshot via exit codes; missing required executables (at minimum
// `go` and `golangci-lint`) are treated as fatal errors and returned to the caller.
func Create(moduleRoot string) (*Snapshot, error) {
    meta := Meta{
        CreatedAt:   time.Now().UTC(),
        ModuleRoot:  moduleRoot,
        SpecVersion: SpecVersion,
    }

    var steps Steps

    // 1. go mod tidy
    res, err := runStep(moduleRoot, "go", "mod", "tidy")
    if err != nil {
        return nil, err
    }
    steps.GoModTidy = res

    // 2. go vet ./...
    res, err = runStep(moduleRoot, "go", "vet", "./...")
    if err != nil {
        return nil, err
    }
    steps.GoVet = res

    // 3. golangci-lint run ./...
    res, err = runStep(moduleRoot, "golangci-lint", "run", "./...")
    if err != nil {
        return nil, err
    }
    steps.GolangciLint = res

    // 4. go test -race ./...
    res, err = runStep(moduleRoot, "go", "test", "-race", "./...")
    if err != nil {
        return nil, err
    }
    steps.TestsRace = res

    // 5. go test ./...
    res, err = runStep(moduleRoot, "go", "test", "./...")
    if err != nil {
        return nil, err
    }
    steps.Tests = res

    // 6. Coverage step: go test ./... -coverprofile=<coverprofile>
    covStep, err := runCoverageStep(moduleRoot)
    if err != nil {
        return nil, err
    }
    steps.Coverage = covStep

    // 7. Benchmarks: go test ./... -run=^$ -bench=. -benchmem
    benchStep, err := runBenchmarksStep(moduleRoot)
    if err != nil {
        return nil, err
    }
    steps.Benchmarks = benchStep

    snap := &Snapshot{
        Meta:  meta,
        Steps: steps,
    }

    return snap, nil
}

// runStep executes a command in moduleRoot and returns a StepResult. It treats
// missing executables as fatal errors (returned to the caller) and records
// non-zero exit codes in the StepResult.
func runStep(moduleRoot, name string, args ...string) (StepResult, error) {
    start := time.Now()

    cmd := exec.Command(name, args...)
    cmd.Dir = moduleRoot
    cmd.Stdout = os.Stderr
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    duration := time.Since(start)

    exitCode := 0
    if err != nil {
        // Missing executable is fatal per spec §12.6.2.
        var execErr *exec.Error
        if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
            return StepResult{}, fmt.Errorf("missing required executable %q: %w", name, err)
        }

        // For non-zero exits, capture exit code but do not treat as fatal.
        if ee, ok := err.(*exec.ExitError); ok {
            if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
                exitCode = ws.ExitStatus()
            } else {
                exitCode = 1
            }
        } else {
            exitCode = 1
        }
    }

    return StepResult{
        ExitCode:       exitCode,
        DurationMillis: duration.Milliseconds(),
    }, nil
}

// runCoverageStep runs the coverage command and computes total_percent via
// `go tool cover -func`. It returns a CoverageStep and only treats missing
// executables as fatal.
func runCoverageStep(moduleRoot string) (CoverageStep, error) {
    // Use a temporary cover profile inside moduleRoot.
    coverPath := filepath.Join(moduleRoot, "cover.out")

    start := time.Now()

    cmd := exec.Command("go", "test", "./...", "-coverprofile", coverPath)
    cmd.Dir = moduleRoot
    cmd.Stdout = os.Stderr
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    duration := time.Since(start)

    exitCode := 0
    if err != nil {
        var execErr *exec.Error
        if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
            return CoverageStep{}, fmt.Errorf("missing required executable %q: %w", "go", err)
        }
        if ee, ok := err.(*exec.ExitError); ok {
            if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
                exitCode = ws.ExitStatus()
            } else {
                exitCode = 1
            }
        } else {
            exitCode = 1
        }
    }

    total := 0.0
    if exitCode == 0 {
        // Compute total_percent using `go tool cover -func`.
        toolCmd := exec.Command("go", "tool", "cover", "-func", coverPath)
        toolCmd.Dir = moduleRoot
        var buf bytes.Buffer
        toolCmd.Stdout = &buf
        toolCmd.Stderr = os.Stderr
        if err := toolCmd.Run(); err == nil {
            re := regexp.MustCompile(`total:\s*\(statements\)\s*([0-9.]+)%`)
            for _, line := range strings.Split(buf.String(), "\n") {
                m := re.FindStringSubmatch(line)
                if len(m) == 2 {
                    if v, err := strconv.ParseFloat(m[1], 64); err == nil {
                        total = v
                    }
                    break
                }
            }
        }
    }

    return CoverageStep{
        StepResult: StepResult{
            ExitCode:       exitCode,
            DurationMillis: duration.Milliseconds(),
        },
        TotalPercent: total,
    }, nil
}

// runBenchmarksStep runs go test benchmarks and parses benchmark metrics into
// a BenchmarkStep. Missing executables are treated as fatal errors; benchmark
// failures do not affect other step results beyond the exit code.
func runBenchmarksStep(moduleRoot string) (BenchmarkStep, error) {
    start := time.Now()

    cmd := exec.Command("go", "test", "./...", "-run=^$", "-bench=.", "-benchmem")
    cmd.Dir = moduleRoot
    var buf bytes.Buffer
    cmd.Stdout = &buf
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    duration := time.Since(start)

    exitCode := 0
    if err != nil {
        var execErr *exec.Error
        if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
            return BenchmarkStep{}, fmt.Errorf("missing required executable %q: %w", "go", err)
        }
        if ee, ok := err.(*exec.ExitError); ok {
            if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
                exitCode = ws.ExitStatus()
            } else {
                exitCode = 1
            }
        } else {
            exitCode = 1
        }
    }

    metrics := make(map[string]BenchmarkMetrics)

    // Parse go test -bench output. Typical lines look like:
    // pkg/path  BenchmarkName-8   12345 ns/op   67 B/op   1 allocs/op
    re := regexp.MustCompile(`^(\S+)\s+Benchmark(\S+)-\d+\s+([0-9]+)\s+ns/op(?:\s+([0-9]+)\s+B/op)?(?:\s+([0-9]+)\s+allocs/op)?`)

    for _, line := range strings.Split(buf.String(), "\n") {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        m := re.FindStringSubmatch(line)
        if len(m) == 0 {
            continue
        }

        pkg := m[1]
        name := m[2]
        key := fmt.Sprintf("%s/%s", pkg, name)

        ns, _ := strconv.ParseFloat(m[3], 64)

        var bytesPerOp *float64
        if m[4] != "" {
            if v, err := strconv.ParseFloat(m[4], 64); err == nil {
                bytesPerOp = &v
            }
        }

        var allocsPerOp *float64
        if m[5] != "" {
            if v, err := strconv.ParseFloat(m[5], 64); err == nil {
                allocsPerOp = &v
            }
        }

        metrics[key] = BenchmarkMetrics{
            NsPerOp:     ns,
            BytesPerOp:  bytesPerOp,
            AllocsPerOp: allocsPerOp,
        }
    }

    return BenchmarkStep{
        StepResult: StepResult{
            ExitCode:       exitCode,
            DurationMillis: duration.Milliseconds(),
        },
        Benchmarks: metrics,
    }, nil
}

// Write writes a snapshot to a TOML file.
func Write(snapshot *Snapshot, path string) error {
    // Ensure directory exists
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return amuxerrors.Wrapf(err, "create snapshot directory: %s", dir)
    }

    data, err := toml.Marshal(snapshot)
    if err != nil {
        return amuxerrors.Wrap(err, "marshal snapshot")
    }

    if err := os.WriteFile(path, data, 0o644); err != nil {
        return amuxerrors.Wrapf(err, "write snapshot: %s", path)
    }

    return nil
}

// Read reads a snapshot from a TOML file.
func Read(path string) (*Snapshot, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, amuxerrors.Wrapf(err, "read snapshot: %s", path)
    }

    var snapshot Snapshot
    if err := toml.Unmarshal(data, &snapshot); err != nil {
        return nil, amuxerrors.Wrapf(err, "unmarshal snapshot: %s", path)
    }

    return &snapshot, nil
}

// Compare compares two snapshots and returns a regression report per spec §12.6.5.
//
// Regression rules:
//   - For any step present in both snapshots, if the baseline exit_code is 0 and the
//     new exit_code is non-zero, this is a regression.
//   - If both snapshots report coverage.exit_code = 0 and the new total_percent is
//     less than the baseline total_percent, this is a coverage regression.
//   - For any benchmark present in both snapshots, a regression occurs when any of:
//       * ns_per_op increases
//       * bytes_per_op increases (when present in both)
//       * allocs_per_op increases (when present in both)
func Compare(baseline, current *Snapshot) (bool, string) {
    var b strings.Builder
    regressions := false

    // Helper to compare simple steps.
    compareStep := func(label string, base, cur StepResult) {
        if base.ExitCode == 0 && cur.ExitCode != 0 {
            regressions = true
            fmt.Fprintf(&b, "%s exit_code regression: baseline=0, current=%d\n", label, cur.ExitCode)
        }
    }

    compareStep("go_mod_tidy", baseline.Steps.GoModTidy, current.Steps.GoModTidy)
    compareStep("go_vet", baseline.Steps.GoVet, current.Steps.GoVet)
    compareStep("golangci_lint", baseline.Steps.GolangciLint, current.Steps.GolangciLint)
    compareStep("tests_race", baseline.Steps.TestsRace, current.Steps.TestsRace)
    compareStep("tests", baseline.Steps.Tests, current.Steps.Tests)
    compareStep("coverage", baseline.Steps.Coverage.StepResult, current.Steps.Coverage.StepResult)
    compareStep("benchmarks", baseline.Steps.Benchmarks.StepResult, current.Steps.Benchmarks.StepResult)

    // Coverage regression.
    if baseline.Steps.Coverage.ExitCode == 0 && current.Steps.Coverage.ExitCode == 0 {
        if current.Steps.Coverage.TotalPercent < baseline.Steps.Coverage.TotalPercent {
            regressions = true
            fmt.Fprintf(&b, "coverage total_percent regression: baseline=%.2f%%, current=%.2f%%\n",
                baseline.Steps.Coverage.TotalPercent, current.Steps.Coverage.TotalPercent)
        }
    }

    // Benchmark regressions.
    for key, baseMetrics := range baseline.Steps.Benchmarks.Benchmarks {
        curMetrics, ok := current.Steps.Benchmarks.Benchmarks[key]
        if !ok {
            continue
        }

        if curMetrics.NsPerOp > baseMetrics.NsPerOp {
            regressions = true
            fmt.Fprintf(&b, "benchmark %s ns_per_op regression: baseline=%.2f, current=%.2f\n",
                key, baseMetrics.NsPerOp, curMetrics.NsPerOp)
        }

        if baseMetrics.BytesPerOp != nil && curMetrics.BytesPerOp != nil && *curMetrics.BytesPerOp > *baseMetrics.BytesPerOp {
            regressions = true
            fmt.Fprintf(&b, "benchmark %s bytes_per_op regression: baseline=%.2f, current=%.2f\n",
                key, *baseMetrics.BytesPerOp, *curMetrics.BytesPerOp)
        }

        if baseMetrics.AllocsPerOp != nil && curMetrics.AllocsPerOp != nil && *curMetrics.AllocsPerOp > *baseMetrics.AllocsPerOp {
            regressions = true
            fmt.Fprintf(&b, "benchmark %s allocs_per_op regression: baseline=%.2f, current=%.2f\n",
                key, *baseMetrics.AllocsPerOp, *curMetrics.AllocsPerOp)
        }
    }

    if !regressions {
        return true, "No regressions detected"
    }

    return false, strings.TrimSpace(b.String())
}

// GenerateSnapshotPath generates a snapshot file path under <module_root>/snapshots
// with the name amux-test-<timestamp>.toml where timestamp is formatted as
// YYYYMMDDThhmmssZ per spec §12.6.3. If a file with that name already exists,
// a numeric suffix -1, -2, ... is appended before the .toml extension.
func GenerateSnapshotPath(moduleRoot string) string {
    ts := time.Now().UTC().Format("20060102T150405Z")
    base := fmt.Sprintf("amux-test-%s.toml", ts)
    dir := filepath.Join(moduleRoot, "snapshots")

    path := filepath.Join(dir, base)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return path
    }

    // Append numeric suffix until we find a free name.
    for i := 1; ; i++ {
        candidate := fmt.Sprintf("amux-test-%s-%d.toml", ts, i)
        full := filepath.Join(dir, candidate)
        if _, err := os.Stat(full); os.IsNotExist(err) {
            return full
        }
    }
}

// FindLatestSnapshot finds the most recent snapshot file in the snapshots
// directory using lexicographic ordering of file names matching
// amux-test-*.toml per spec §12.6.5.
func FindLatestSnapshot(moduleRoot string) (string, error) {
    snapshotsDir := filepath.Join(moduleRoot, "snapshots")

    entries, err := os.ReadDir(snapshotsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return "", amuxerrors.Wrap(amuxerrors.ErrNotFound, "no snapshots directory")
        }
        return "", amuxerrors.Wrap(err, "read snapshots directory")
    }

    var names []string
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        name := entry.Name()
        if strings.HasPrefix(name, "amux-test-") && strings.HasSuffix(name, ".toml") {
            names = append(names, name)
        }
    }

    if len(names) == 0 {
        return "", amuxerrors.Wrap(amuxerrors.ErrNotFound, "no snapshot files found")
    }

    sort.Strings(names)
    latest := names[len(names)-1]
    return filepath.Join(snapshotsDir, latest), nil
}
