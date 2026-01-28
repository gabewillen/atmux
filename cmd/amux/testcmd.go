package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/config"
)

const specVersion = "v1.22"

func runTest(args []string) error {
	flags := flag.NewFlagSet("amux test", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var regression bool
	var noSnapshot bool
	flags.BoolVar(&regression, "regression", false, "compare against previous snapshot")
	flags.BoolVar(&noSnapshot, "no-snapshot", false, "write snapshot to stdout")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return err
	}
	if err := ensureExecutables([]string{"go", "golangci-lint"}); err != nil {
		return err
	}
	steps, err := runSteps(moduleRoot)
	if err != nil {
		return err
	}
	snapshot := buildSnapshot(moduleRoot, steps)
	benchmarks := parseBenchmarks(steps["benchmarks"].stdout)
	snapshot.Benchmarks = benchmarks
	coverageStep := steps["coverage"]
	if coverageStep.exitCode == 0 {
		percent, err := parseCoverageTotal(coverageStep.coverageProfile)
		if err != nil {
			return fmt.Errorf("coverage parse: %w", err)
		}
		step := snapshot.Steps["coverage"]
		step.TotalPercent = &percent
		snapshot.Steps["coverage"] = step
	}
	var logOut io.Writer = os.Stderr
	var snapshotPath string
	if !noSnapshot {
		snapshotPath, err = writeSnapshotFile(moduleRoot, snapshot)
		if err != nil {
			return err
		}
	} else {
		encoded, err := encodeSnapshot(snapshot)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(os.Stdout, encoded); err != nil {
			return fmt.Errorf("write snapshot: %w", err)
		}
	}
	var regressErr error
	if regression {
		report, err := checkRegression(moduleRoot, snapshot, snapshotPath)
		if err != nil {
			reportErr := fmt.Errorf("regression: %w", err)
			fmt.Fprintln(logOut, reportErr.Error())
			regressErr = reportErr
		} else if len(report) > 0 {
			fmt.Fprintln(logOut, "regressions detected:")
			for _, entry := range report {
				fmt.Fprintf(logOut, "- %s baseline=%s new=%s\n", entry.Metric, entry.Baseline, entry.Current)
			}
			regressErr = fmt.Errorf("regressions detected")
		}
	}
	if regressErr != nil {
		return regressErr
	}
	return summarizeSteps(logOut, steps)
}

type stepResult struct {
	argv            []string
	stdout          []byte
	stderr          []byte
	exitCode        int
	duration        time.Duration
	coverageProfile string
}

type snapshot struct {
	Meta       snapshotMeta
	Steps      map[string]snapshotStep
	Benchmarks []benchmark
}

type snapshotMeta struct {
	CreatedAt  time.Time
	ModuleRoot string
	SpecVersion string
}

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

type benchmark struct {
	Name        string
	Pkg         string
	NsPerOp     float64
	Iterations  int
	BytesPerOp  *int
	AllocsPerOp *int
}

type regressionEntry struct {
	Metric   string
	Baseline string
	Current  string
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("module root: %w", err)
	}
	current := wd
	for {
		path := filepath.Join(current, "go.mod")
		if _, err := os.Stat(path); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("module root: go.mod not found")
		}
		current = parent
	}
}

func ensureExecutables(names []string) error {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("missing executable: %s", name)
		}
	}
	return nil
}

func runSteps(moduleRoot string) (map[string]stepResult, error) {
	steps := make(map[string]stepResult)
	steps["go_mod_tidy"] = runStep(moduleRoot, []string{"go", "mod", "tidy"})
	steps["go_vet"] = runStep(moduleRoot, []string{"go", "vet", "./..."})
	steps["golangci_lint"] = runStep(moduleRoot, []string{"golangci-lint", "run", "./..."})
	steps["tests_race"] = runStep(moduleRoot, []string{"go", "test", "-race", "./..."})
	steps["tests"] = runStep(moduleRoot, []string{"go", "test", "./..."})
	coveragePath := filepath.Join(moduleRoot, ".amux", "coverage", "coverage.out")
	if err := os.MkdirAll(filepath.Dir(coveragePath), 0o755); err != nil {
		return steps, fmt.Errorf("coverage dir: %w", err)
	}
	coverageStep := runStep(moduleRoot, []string{"go", "test", "./...", "-coverprofile=" + coveragePath})
	coverageStep.coverageProfile = coveragePath
	steps["coverage"] = coverageStep
	steps["benchmarks"] = runStep(moduleRoot, []string{"go", "test", "-run=^$", "-bench=.", "-benchmem", "./..."})
	return steps, nil
}

func runStep(moduleRoot string, argv []string) stepResult {
	start := time.Now()
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = moduleRoot
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return stepResult{
		argv:     argv,
		stdout:   stdout.Bytes(),
		stderr:   stderr.Bytes(),
		exitCode: exitCode,
		duration: time.Since(start),
	}
}

func buildSnapshot(moduleRoot string, steps map[string]stepResult) snapshot {
	meta := snapshotMeta{
		CreatedAt:  time.Now().UTC(),
		ModuleRoot: moduleRoot,
		SpecVersion: specVersion,
	}
	stepSnapshots := make(map[string]snapshotStep)
	for name, step := range steps {
		stepSnapshots[name] = snapshotStep{
			Argv:         step.argv,
			ExitCode:     step.exitCode,
			DurationMS:   step.duration.Milliseconds(),
			StdoutSHA256: hashBytes(step.stdout),
			StderrSHA256: hashBytes(step.stderr),
			StdoutBytes:  len(step.stdout),
			StderrBytes:  len(step.stderr),
		}
	}
	return snapshot{Meta: meta, Steps: stepSnapshots}
}

func encodeSnapshot(s snapshot) (string, error) {
	var b strings.Builder
	b.WriteString("[meta]\n")
	b.WriteString(fmt.Sprintf("created_at = %s\n", quoteString(s.Meta.CreatedAt.Format(time.RFC3339))))
	b.WriteString(fmt.Sprintf("module_root = %s\n", quoteString(s.Meta.ModuleRoot)))
	b.WriteString(fmt.Sprintf("spec_version = %s\n\n", quoteString(s.Meta.SpecVersion)))
	stepNames := make([]string, 0, len(s.Steps))
	for name := range s.Steps {
		stepNames = append(stepNames, name)
	}
	sort.Strings(stepNames)
	for _, name := range stepNames {
		step := s.Steps[name]
		b.WriteString("[steps." + name + "]\n")
		b.WriteString("argv = " + formatStringArray(step.Argv) + "\n")
		b.WriteString(fmt.Sprintf("exit_code = %d\n", step.ExitCode))
		b.WriteString(fmt.Sprintf("duration_ms = %d\n", step.DurationMS))
		b.WriteString(fmt.Sprintf("stdout_sha256 = %s\n", quoteString(step.StdoutSHA256)))
		b.WriteString(fmt.Sprintf("stderr_sha256 = %s\n", quoteString(step.StderrSHA256)))
		b.WriteString(fmt.Sprintf("stdout_bytes = %d\n", step.StdoutBytes))
		b.WriteString(fmt.Sprintf("stderr_bytes = %d\n", step.StderrBytes))
		if step.TotalPercent != nil {
			b.WriteString(fmt.Sprintf("total_percent = %s\n", formatFloat(*step.TotalPercent)))
		}
		b.WriteString("\n")
	}
	for _, bench := range s.Benchmarks {
		b.WriteString("[[benchmarks]]\n")
		b.WriteString(fmt.Sprintf("name = %s\n", quoteString(bench.Name)))
		b.WriteString(fmt.Sprintf("pkg = %s\n", quoteString(bench.Pkg)))
		b.WriteString(fmt.Sprintf("ns_per_op = %s\n", formatFloat(bench.NsPerOp)))
		b.WriteString(fmt.Sprintf("iterations = %d\n", bench.Iterations))
		if bench.BytesPerOp != nil {
			b.WriteString(fmt.Sprintf("bytes_per_op = %d\n", *bench.BytesPerOp))
		}
		if bench.AllocsPerOp != nil {
			b.WriteString(fmt.Sprintf("allocs_per_op = %d\n", *bench.AllocsPerOp))
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func writeSnapshotFile(moduleRoot string, snap snapshot) (string, error) {
	encoded, err := encodeSnapshot(snap)
	if err != nil {
		return "", err
	}
	snapDir := filepath.Join(moduleRoot, "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return "", fmt.Errorf("snapshot dir: %w", err)
	}
	ts := snap.Meta.CreatedAt.UTC().Format("20060102T150405Z")
	name := "amux-test-" + ts + ".toml"
	path := filepath.Join(snapDir, name)
	counter := 1
	for {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			break
		}
		name = fmt.Sprintf("amux-test-%s-%d.toml", ts, counter)
		path = filepath.Join(snapDir, name)
		counter++
	}
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		return "", fmt.Errorf("write snapshot: %w", err)
	}
	return path, nil
}

func parseCoverageTotal(profilePath string) (float64, error) {
	cmd := exec.Command("go", "tool", "cover", "-func="+profilePath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("go tool cover: %w", err)
	}
	lines := strings.Split(stdout.String(), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				return 0, fmt.Errorf("invalid total line")
			}
			percent := strings.TrimSuffix(parts[len(parts)-1], "%")
			value, err := strconv.ParseFloat(percent, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid coverage percent")
			}
			return value, nil
		}
	}
	return 0, fmt.Errorf("coverage total not found")
}

func parseBenchmarks(output []byte) []benchmark {
	var benchmarks []benchmark
	lines := strings.Split(string(output), "\n")
	currentPkg := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "pkg:") {
			currentPkg = strings.TrimSpace(strings.TrimPrefix(line, "pkg:"))
			continue
		}
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		name := strings.Split(fields[0], "-")[0]
		iterations, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		valueField := fields[2]
		unitField := ""
		if strings.HasSuffix(valueField, "ns/op") {
			unitField = "ns/op"
			valueField = strings.TrimSuffix(valueField, "ns/op")
		} else if len(fields) > 3 {
			unitField = fields[3]
		}
		if unitField != "ns/op" {
			continue
		}
		nsPerOp, err := strconv.ParseFloat(valueField, 64)
		if err != nil {
			continue
		}
		bench := benchmark{
			Name:       name,
			Pkg:        currentPkg,
			Iterations: iterations,
			NsPerOp:    nsPerOp,
		}
		for i := 2; i < len(fields); i++ {
			field := fields[i]
			if strings.HasSuffix(field, "B/op") {
				value, err := strconv.Atoi(strings.TrimSuffix(field, "B/op"))
				if err == nil {
					bench.BytesPerOp = &value
				}
				continue
			}
			if strings.HasSuffix(field, "allocs/op") {
				value, err := strconv.Atoi(strings.TrimSuffix(field, "allocs/op"))
				if err == nil {
					bench.AllocsPerOp = &value
				}
				continue
			}
			if i+1 < len(fields) && (fields[i+1] == "B/op" || fields[i+1] == "allocs/op") {
				value, err := strconv.Atoi(field)
				if err != nil {
					continue
				}
				if fields[i+1] == "B/op" {
					bench.BytesPerOp = &value
				}
				if fields[i+1] == "allocs/op" {
					bench.AllocsPerOp = &value
				}
				i++
			}
		}
		benchmarks = append(benchmarks, bench)
	}
	return benchmarks
}
func checkRegression(moduleRoot string, current snapshot, currentPath string) ([]regressionEntry, error) {
	baselinePath, err := findBaselineSnapshot(moduleRoot, currentPath)
	if err != nil {
		return nil, err
	}
	baseline, err := readSnapshot(baselinePath)
	if err != nil {
		return nil, err
	}
	return regressions(baseline, current), nil
}

func findBaselineSnapshot(moduleRoot string, currentPath string) (string, error) {
	snapDir := filepath.Join(moduleRoot, "snapshots")
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		return "", fmt.Errorf("baseline snapshot: %w", err)
	}
	var candidates []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "amux-test-") || !strings.HasSuffix(name, ".toml") {
			continue
		}
		path := filepath.Join(snapDir, name)
		if currentPath != "" && path == currentPath {
			continue
		}
		candidates = append(candidates, path)
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("baseline snapshot: none found")
	}
	sort.Strings(candidates)
	return candidates[len(candidates)-1], nil
}

func readSnapshot(path string) (snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return snapshot{}, fmt.Errorf("read snapshot: %w", err)
	}
	parsed, err := config.ParseTOML(data)
	if err != nil {
		return snapshot{}, fmt.Errorf("parse snapshot: %w", err)
	}
	return decodeSnapshot(parsed)
}

func decodeSnapshot(raw map[string]any) (snapshot, error) {
	metaRaw, ok := raw["meta"].(map[string]any)
	if !ok {
		return snapshot{}, fmt.Errorf("snapshot meta missing")
	}
	createdRaw, _ := metaRaw["created_at"].(string)
	createdAt, err := time.Parse(time.RFC3339, createdRaw)
	if err != nil {
		return snapshot{}, fmt.Errorf("snapshot created_at: %w", err)
	}
	moduleRoot, _ := metaRaw["module_root"].(string)
	specVer, _ := metaRaw["spec_version"].(string)
	stepsRaw, ok := raw["steps"].(map[string]any)
	if !ok {
		return snapshot{}, fmt.Errorf("snapshot steps missing")
	}
	steps := make(map[string]snapshotStep)
	for name, value := range stepsRaw {
		stepMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		step := snapshotStep{}
		if argv, ok := parseStringArray(stepMap["argv"]); ok {
			step.Argv = argv
		}
		step.ExitCode = toInt(stepMap["exit_code"])
		step.DurationMS = int64(toInt(stepMap["duration_ms"]))
		step.StdoutSHA256, _ = stepMap["stdout_sha256"].(string)
		step.StderrSHA256, _ = stepMap["stderr_sha256"].(string)
		step.StdoutBytes = toInt(stepMap["stdout_bytes"])
		step.StderrBytes = toInt(stepMap["stderr_bytes"])
		if total, ok := stepMap["total_percent"].(float64); ok {
			step.TotalPercent = &total
		}
		steps[name] = step
	}
	benchmarks := decodeBenchmarks(raw["benchmarks"])
	return snapshot{
		Meta: snapshotMeta{CreatedAt: createdAt, ModuleRoot: moduleRoot, SpecVersion: specVer},
		Steps: steps,
		Benchmarks: benchmarks,
	}, nil
}

func decodeBenchmarks(raw any) []benchmark {
	entries, ok := raw.([]any)
	if !ok {
		return nil
	}
	var results []benchmark
	for _, entry := range entries {
		row, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		bench := benchmark{}
		bench.Name, _ = row["name"].(string)
		bench.Pkg, _ = row["pkg"].(string)
		if value, ok := row["ns_per_op"].(float64); ok {
			bench.NsPerOp = value
		}
		bench.Iterations = toInt(row["iterations"])
		if value, ok := row["bytes_per_op"].(int64); ok {
			val := int(value)
			bench.BytesPerOp = &val
		}
		if value, ok := row["bytes_per_op"].(float64); ok {
			val := int(value)
			bench.BytesPerOp = &val
		}
		if value, ok := row["allocs_per_op"].(int64); ok {
			val := int(value)
			bench.AllocsPerOp = &val
		}
		if value, ok := row["allocs_per_op"].(float64); ok {
			val := int(value)
			bench.AllocsPerOp = &val
		}
		results = append(results, bench)
	}
	return results
}

func regressions(baseline snapshot, current snapshot) []regressionEntry {
	var regress []regressionEntry
	for name, baseStep := range baseline.Steps {
		currentStep, ok := current.Steps[name]
		if !ok {
			continue
		}
		if baseStep.ExitCode == 0 && currentStep.ExitCode != 0 {
			regress = append(regress, regressionEntry{
				Metric:   "step." + name + ".exit_code",
				Baseline: "0",
				Current:  strconv.Itoa(currentStep.ExitCode),
			})
		}
	}
	baseCoverage := baseline.Steps["coverage"]
	currCoverage := current.Steps["coverage"]
	if baseCoverage.ExitCode == 0 && currCoverage.ExitCode == 0 && baseCoverage.TotalPercent != nil && currCoverage.TotalPercent != nil {
		if *currCoverage.TotalPercent < *baseCoverage.TotalPercent {
			regress = append(regress, regressionEntry{
				Metric:   "coverage.total_percent",
				Baseline: formatFloat(*baseCoverage.TotalPercent),
				Current:  formatFloat(*currCoverage.TotalPercent),
			})
		}
	}
	baseBench := benchmarkIndex(baseline.Benchmarks)
	for _, bench := range current.Benchmarks {
		key := benchKey(bench)
		base, ok := baseBench[key]
		if !ok {
			continue
		}
		if bench.NsPerOp > base.NsPerOp {
			regress = append(regress, regressionEntry{
				Metric:   key + ".ns_per_op",
				Baseline: formatFloat(base.NsPerOp),
				Current:  formatFloat(bench.NsPerOp),
			})
		}
		if bench.BytesPerOp != nil && base.BytesPerOp != nil && *bench.BytesPerOp > *base.BytesPerOp {
			regress = append(regress, regressionEntry{
				Metric:   key + ".bytes_per_op",
				Baseline: strconv.Itoa(*base.BytesPerOp),
				Current:  strconv.Itoa(*bench.BytesPerOp),
			})
		}
		if bench.AllocsPerOp != nil && base.AllocsPerOp != nil && *bench.AllocsPerOp > *base.AllocsPerOp {
			regress = append(regress, regressionEntry{
				Metric:   key + ".allocs_per_op",
				Baseline: strconv.Itoa(*base.AllocsPerOp),
				Current:  strconv.Itoa(*bench.AllocsPerOp),
			})
		}
	}
	return regress
}

func benchmarkIndex(items []benchmark) map[string]benchmark {
	idx := make(map[string]benchmark)
	for _, item := range items {
		idx[benchKey(item)] = item
	}
	return idx
}

func benchKey(item benchmark) string {
	return item.Pkg + ":" + item.Name
}

func summarizeSteps(w io.Writer, steps map[string]stepResult) error {
	var failed bool
	for name, step := range steps {
		if step.exitCode != 0 {
			failed = true
			fmt.Fprintf(w, "step %s failed (exit %d)\n", name, step.exitCode)
		}
	}
	if failed {
		return fmt.Errorf("one or more steps failed")
	}
	return nil
}

func hashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

func quoteString(value string) string {
	return strconv.Quote(value)
}

func formatStringArray(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, quoteString(value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func parseStringArray(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		str, ok := item.(string)
		if !ok {
			return nil, false
		}
		result = append(result, str)
	}
	return result, true
}

func toInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}
