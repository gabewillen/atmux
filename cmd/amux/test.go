// Package main implements the amux CLI client.
// test.go implements amux test per spec §12.6 (snapshot schema, regression, required sequence).
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/spec"
	"github.com/pelletier/go-toml/v2"
)

// Snapshot schema per spec §12.6.4: [meta], [steps.*], [[benchmarks]].
type TestSnapshot struct {
	Meta       MetaSnapshot    `toml:"meta"`
	Steps      StepsSnapshot  `toml:"steps"`
	Benchmarks []BenchmarkEntry `toml:"benchmarks"`
}

type MetaSnapshot struct {
	CreatedAt   string `toml:"created_at"`
	ModuleRoot  string `toml:"module_root"`
	SpecVersion string `toml:"spec_version"`
}

type StepsSnapshot struct {
	GoModTidy    StepResult `toml:"go_mod_tidy"`
	GoVet        StepResult `toml:"go_vet"`
	GolangciLint StepResult `toml:"golangci_lint"`
	TestsRace    StepResult `toml:"tests_race"`
	Tests        StepResult `toml:"tests"`
	Coverage     StepResult `toml:"coverage"`
	Benchmarks   StepResult `toml:"benchmarks"`
}

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

type BenchmarkEntry struct {
	Name        string `toml:"name"`
	Pkg         string `toml:"pkg"`
	NsPerOp     float64 `toml:"ns_per_op"`
	Iterations  int    `toml:"iterations"`
	BytesPerOp  *int   `toml:"bytes_per_op,omitempty"`
	AllocsPerOp *int   `toml:"allocs_per_op,omitempty"`
}

func runTest(args []string) error {
	var noSnapshot, regression bool
	var snapshotPath string
	for i, arg := range args {
		switch arg {
		case "--no-snapshot":
			noSnapshot = true
		case "--regression":
			regression = true
		case "--snapshot":
			if i+1 < len(args) {
				snapshotPath = args[i+1]
			}
		}
	}

	moduleRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find module root: %w", err)
	}
	if err := spec.CheckSpecVersion(moduleRoot); err != nil {
		return fmt.Errorf("spec version check failed: %w", err)
	}

	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintf(os.Stderr, "amux test: required executable not found: go\n")
		return fmt.Errorf("required executable not found: go")
	}
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		fmt.Fprintf(os.Stderr, "amux test: required executable not found: golangci-lint (install: https://golangci-lint.run/usage/install/)\n")
		return fmt.Errorf("required executable not found: golangci-lint")
	}

	steps, coveragePct, benchmarks := runStepSequence(moduleRoot)
	if steps.Coverage.ExitCode == 0 && coveragePct != nil {
		steps.Coverage.TotalPercent = coveragePct
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	snap := TestSnapshot{
		Meta: MetaSnapshot{
			CreatedAt:   createdAt,
			ModuleRoot: moduleRoot,
			SpecVersion: spec.ExpectedSpecVersion,
		},
		Steps:      steps,
		Benchmarks: benchmarks,
	}

	if snapshotPath == "" && !noSnapshot {
		snapshotPath = snapshotPathFor(moduleRoot)
	}

	if noSnapshot {
		data, err := toml.Marshal(snap)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}
		_, _ = os.Stdout.Write(data)
	} else if snapshotPath != "" {
		if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
			return fmt.Errorf("create snapshot dir: %w", err)
		}
		data, err := toml.Marshal(snap)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}
		if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
			return fmt.Errorf("write snapshot: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Snapshot written to %s\n", snapshotPath)
	}

	if regression {
		return checkRegression(moduleRoot, snapshotPath, snap, noSnapshot)
	}
	return nil
}

// snapshotPathFor returns path for new snapshot: amux-test-<UTC>.toml with -1, -2 if exists (§12.6.3).
func snapshotPathFor(moduleRoot string) string {
	dir := filepath.Join(moduleRoot, "snapshots")
	ts := time.Now().UTC().Format("20060102T150405Z")
	for i := 0; ; i++ {
		var name string
		if i == 0 {
			name = "amux-test-" + ts + ".toml"
		} else {
			name = fmt.Sprintf("amux-test-%s-%d.toml", ts, i)
		}
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
	}
}

func runStepSequence(moduleRoot string) (StepsSnapshot, *float64, []BenchmarkEntry) {
	var steps StepsSnapshot
	var coveragePct *float64
	var benchmarks []BenchmarkEntry

	// 1. go mod tidy
	steps.GoModTidy = runStep(moduleRoot, "go", "mod", "tidy")

	// 2. go vet ./...
	steps.GoVet = runStep(moduleRoot, "go", "vet", "./...")

	// 3. golangci-lint run ./...
	steps.GolangciLint = runStep(moduleRoot, "golangci-lint", "run", "./...")

	// 4. go test -race ./...
	steps.TestsRace = runStep(moduleRoot, "go", "test", "-race", "./...")

	// 5. go test ./...
	steps.Tests = runStep(moduleRoot, "go", "test", "./...")

	// 6. go test ./... -coverprofile=<path>
	coverFile := filepath.Join(moduleRoot, ".amux-test-cover.out")
	steps.Coverage = runStep(moduleRoot, "go", "test", "./...", "-coverprofile="+coverFile)
	if steps.Coverage.ExitCode == 0 {
		if pct := runCoverFunc(moduleRoot, coverFile); pct != nil {
			coveragePct = pct
		}
	}
	_ = os.Remove(coverFile)

	// 7. go test -run=^$ -bench=. -benchmem ./...
	out, res := runStepCapture(moduleRoot, "go", "test", "-run=^$", "-bench=.", "-benchmem", "./...")
	steps.Benchmarks = res
	benchmarks = parseBenchmarkOutput(out)
	return steps, coveragePct, benchmarks
}

func runStep(moduleRoot string, name string, args ...string) StepResult {
	_, r := runStepCapture(moduleRoot, name, args...)
	return r
}

func runStepCapture(moduleRoot string, name string, args ...string) (stdout []byte, result StepResult) {
	result.Argv = append([]string{name}, args...)
	cmd := exec.Command(name, args...)
	cmd.Dir = moduleRoot
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)
	result.DurationMs = int(elapsed.Milliseconds())
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}
	outB := outBuf.Bytes()
	errB := errBuf.Bytes()
	result.StdoutBytes = len(outB)
	result.StderrBytes = len(errB)
	h := sha256.Sum256(outB)
	result.StdoutSha256 = hex.EncodeToString(h[:])
	h = sha256.Sum256(errB)
	result.StderrSha256 = hex.EncodeToString(h[:])
	return outB, result
}

func runCoverFunc(moduleRoot, coverPath string) *float64 {
	cmd := exec.Command("go", "tool", "cover", "-func="+coverPath)
	cmd.Dir = moduleRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	// Last line: "total: (statements) 45.2%"
	scanner := bufio.NewScanner(bytes.NewReader(out))
	var last string
	for scanner.Scan() {
		last = scanner.Text()
	}
	if last == "" {
		return nil
	}
	if !strings.Contains(last, "total:") {
		return nil
	}
	// Line is "total:\t\t(statements)\t18.4%" — extract the number before "%".
	if i := strings.Index(last, "%"); i > 0 {
		numStr := last[:i]
		// Last token before "%" is the percentage.
		fields := strings.Fields(numStr)
		if len(fields) == 0 {
			return nil
		}
		pct, err := strconv.ParseFloat(fields[len(fields)-1], 64)
		if err != nil {
			return nil
		}
		return &pct
	}
	return nil
}

// benchmark line: "BenchmarkName-8   1000000  1234 ns/op  56 B/op  2 allocs/op" (tabs or spaces)
var benchLineRe = regexp.MustCompile(`^Benchmark(\S+)\s+(\d+)\s+(\d+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)

func parseBenchmarkOutput(stdout []byte) []BenchmarkEntry {
	var entries []BenchmarkEntry
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	var pkg string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "pkg: ") {
			pkg = strings.TrimSpace(strings.TrimPrefix(line, "pkg:"))
			continue
		}
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		matches := benchLineRe.FindStringSubmatch(line)
		if len(matches) < 4 {
			continue
		}
		// matches[1]=name with -N suffix; strip -N for storage per spec
		name := matches[1]
		if idx := strings.LastIndex(name, "-"); idx > 0 {
			if _, err := strconv.Atoi(name[idx+1:]); err == nil {
				name = name[:idx]
			}
		}
		iters, _ := strconv.Atoi(matches[2])
		nsPerOp, _ := strconv.ParseFloat(matches[3], 64)
		e := BenchmarkEntry{Name: name, Pkg: pkg, NsPerOp: nsPerOp, Iterations: iters}
		if len(matches) >= 5 && matches[4] != "" {
			b, _ := strconv.Atoi(matches[4])
			e.BytesPerOp = &b
		}
		if len(matches) >= 6 && matches[5] != "" {
			a, _ := strconv.Atoi(matches[5])
			e.AllocsPerOp = &a
		}
		entries = append(entries, e)
	}
	return entries
}

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a Go module")
		}
		dir = parent
	}
}

// checkRegression: baseline = lexicographically greatest amux-test-*.toml excluding current (§12.6.5).
func checkRegression(moduleRoot, currentPath string, current TestSnapshot, noSnapshot bool) error {
	snapDir := filepath.Join(moduleRoot, "snapshots")
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		return fmt.Errorf("read snapshots dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "amux-test-") || !strings.HasSuffix(e.Name(), ".toml") {
			continue
		}
		names = append(names, e.Name())
	}
	// §12.6.5: baseline = lexicographically greatest amux-test-*.toml excluding current file.
	currentBase := ""
	if currentPath != "" {
		currentBase = filepath.Base(currentPath)
	}
	var filtered []string
	for _, n := range names {
		if n != currentBase {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		fmt.Fprintf(os.Stderr, "amux test --regression: no baseline snapshot found (no other amux-test-*.toml in snapshots/)\n")
		return fmt.Errorf("no baseline snapshot found for regression comparison")
	}
	sort.Strings(filtered)
	baselineName := filtered[len(filtered)-1]

	prevPath := filepath.Join(snapDir, baselineName)
	prevData, err := os.ReadFile(prevPath)
	if err != nil {
		return fmt.Errorf("read baseline snapshot: %w", err)
	}
	var prev TestSnapshot
	if err := toml.Unmarshal(prevData, &prev); err != nil {
		return fmt.Errorf("parse baseline snapshot: %w", err)
	}

	var currentData []byte
	if noSnapshot {
		var err error
		currentData, err = toml.Marshal(current)
		if err != nil {
			return err
		}
	} else {
		currentData, err = os.ReadFile(currentPath)
		if err != nil {
			return fmt.Errorf("read current snapshot: %w", err)
		}
	}
	var curr TestSnapshot
	if err := toml.Unmarshal(currentData, &curr); err != nil {
		return fmt.Errorf("parse current snapshot: %w", err)
	}

	regressions := checkRegressionRules(prev, curr)
	if len(regressions) > 0 {
		fmt.Fprintf(os.Stderr, "Regression report (baseline: %s):\n", baselineName)
		for _, r := range regressions {
			fmt.Fprintf(os.Stderr, "  - %s\n", r)
		}
		return fmt.Errorf("regression detected: %d issue(s)", len(regressions))
	}
	fmt.Fprintf(os.Stderr, "No regressions detected (compared to %s)\n", baselineName)
	return nil
}

func checkRegressionRules(prev, curr TestSnapshot) []string {
	var out []string
	stepsPrev := []struct {
		name string
		s    StepResult
	}{
		{"go_mod_tidy", prev.Steps.GoModTidy},
		{"go_vet", prev.Steps.GoVet},
		{"golangci_lint", prev.Steps.GolangciLint},
		{"tests_race", prev.Steps.TestsRace},
		{"tests", prev.Steps.Tests},
		{"coverage", prev.Steps.Coverage},
		{"benchmarks", prev.Steps.Benchmarks},
	}
	stepsCurr := []struct {
		name string
		s    StepResult
	}{
		{"go_mod_tidy", curr.Steps.GoModTidy},
		{"go_vet", curr.Steps.GoVet},
		{"golangci_lint", curr.Steps.GolangciLint},
		{"tests_race", curr.Steps.TestsRace},
		{"tests", curr.Steps.Tests},
		{"coverage", curr.Steps.Coverage},
		{"benchmarks", curr.Steps.Benchmarks},
	}
	for i := range stepsPrev {
		if stepsPrev[i].s.ExitCode == 0 && stepsCurr[i].s.ExitCode != 0 {
			out = append(out, fmt.Sprintf("step %s: exit regression (baseline 0, current %d)", stepsPrev[i].name, stepsCurr[i].s.ExitCode))
		}
	}
	if prev.Steps.Coverage.ExitCode == 0 && curr.Steps.Coverage.ExitCode == 0 {
		pPrev := 0.0
		if prev.Steps.Coverage.TotalPercent != nil {
			pPrev = *prev.Steps.Coverage.TotalPercent
		}
		pCurr := 0.0
		if curr.Steps.Coverage.TotalPercent != nil {
			pCurr = *curr.Steps.Coverage.TotalPercent
		}
		if pCurr < pPrev {
			out = append(out, fmt.Sprintf("coverage: total_percent regression (baseline %.2f%%, current %.2f%%)", pPrev, pCurr))
		}
	}
	benchPrev := make(map[string]BenchmarkEntry)
	for _, b := range prev.Benchmarks {
		benchPrev[b.Pkg+"\t"+b.Name] = b
	}
	for _, b := range curr.Benchmarks {
		key := b.Pkg + "\t" + b.Name
		p, ok := benchPrev[key]
		if !ok {
			continue
		}
		if b.NsPerOp > p.NsPerOp {
			out = append(out, fmt.Sprintf("benchmark %s %s: ns_per_op regression (baseline %.0f, current %.0f)", b.Pkg, b.Name, p.NsPerOp, b.NsPerOp))
		}
		if b.BytesPerOp != nil && p.BytesPerOp != nil && *b.BytesPerOp > *p.BytesPerOp {
			out = append(out, fmt.Sprintf("benchmark %s %s: bytes_per_op regression (baseline %d, current %d)", b.Pkg, b.Name, *p.BytesPerOp, *b.BytesPerOp))
		}
		if b.AllocsPerOp != nil && p.AllocsPerOp != nil && *b.AllocsPerOp > *p.AllocsPerOp {
			out = append(out, fmt.Sprintf("benchmark %s %s: allocs_per_op regression (baseline %d, current %d)", b.Pkg, b.Name, *p.AllocsPerOp, *b.AllocsPerOp))
		}
	}
	return out
}
