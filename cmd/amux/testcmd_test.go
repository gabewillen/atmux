package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunTestWithStubbedGo(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "95.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if err := runTest([]string{"--no-snapshot"}); err != nil {
		t.Fatalf("runTest: %v", err)
	}
}

func TestRunTestCoverageFailure(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "79.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if err := runTest([]string{"--no-snapshot"}); err == nil || !strings.Contains(err.Error(), ErrCoverageBelowThreshold.Error()) {
		t.Fatalf("expected coverage error, got %v", err)
	}
}

func TestSnapshotEncodingAndParsing(t *testing.T) {
	steps := map[string]stepResult{
		"tests": {argv: []string{"go", "test"}, stdout: []byte("ok"), stderr: []byte(""), exitCode: 0, duration: time.Millisecond},
	}
	snap := buildSnapshot("/tmp", steps)
	encoded, err := encodeSnapshot(snap)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(encoded, "[meta]") {
		t.Fatalf("expected meta")
	}
	path := filepath.Join(t.TempDir(), "snapshots", "baseline.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	decoded, err := readSnapshot(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if decoded.Meta.ModuleRoot != "/tmp" {
		t.Fatalf("unexpected module root: %s", decoded.Meta.ModuleRoot)
	}
}

func TestParseCoverageTotal(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "92.5"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	value, err := parseCoverageTotal(filepath.Join(tmp, "coverage.out"))
	if err != nil {
		t.Fatalf("parse coverage: %v", err)
	}
	if value < 92.0 {
		t.Fatalf("unexpected coverage: %f", value)
	}
}

func TestParseCoverageTotalInvalidPercent(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "bad"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if _, err := parseCoverageTotal(filepath.Join(tmp, "coverage.out")); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestParseCoverageTotalMissingTotal(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGoOutput(binDir, "no total here", 0); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if _, err := parseCoverageTotal(filepath.Join(tmp, "coverage.out")); err == nil {
		t.Fatalf("expected missing total error")
	}
}

func TestParseBenchmarks(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 100 123 ns/op 10 B/op 1 allocs/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 1 {
		t.Fatalf("expected benchmark")
	}
	if benchmarks[0].Pkg != "github.com/example" {
		t.Fatalf("unexpected pkg: %s", benchmarks[0].Pkg)
	}
}

func TestParseBenchmarksSkipsInvalidLines(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 bad 123 ns/op\nBenchmarkBar-8 10 5 ns/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 1 {
		t.Fatalf("expected one benchmark")
	}
	if benchmarks[0].Name != "BenchmarkBar" {
		t.Fatalf("unexpected benchmark name: %s", benchmarks[0].Name)
	}
}

func TestParseBenchmarksInvalidFields(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 10 bad ns/op\nBenchmarkBar-8 1 2\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 0 {
		t.Fatalf("expected no benchmarks")
	}
}

func TestParseBenchmarksSkipsBadMetrics(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 10 5 ns/op bad B/op 2 allocs/op\nBenchmarkBar-8 10 5 ns/op 3 B/op bad allocs/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 2 {
		t.Fatalf("expected benchmarks")
	}
}

func TestParseBenchmarksNsPerOpSuffix(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 10 123ns/op 9 B/op 2 allocs/op\nBenchmarkBar-8 5 10 us/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 1 {
		t.Fatalf("expected one benchmark")
	}
	if benchmarks[0].BytesPerOp == nil || benchmarks[0].AllocsPerOp == nil {
		t.Fatalf("expected bytes/allocs")
	}
}

func TestParseBenchmarksWithUnits(t *testing.T) {
	output := []byte("pkg: github.com/example\nBenchmarkFoo-8 10 5 ns/op 12 B/op 2 allocs/op\nBenchmarkBar-8 5 7 ns/op 9 B/op 1 allocs/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 2 {
		t.Fatalf("expected benchmarks")
	}
	if benchmarks[0].BytesPerOp == nil || benchmarks[0].AllocsPerOp == nil {
		t.Fatalf("expected bytes/allocs")
	}
}

func TestRegressionDetection(t *testing.T) {
	base := snapshot{Steps: map[string]snapshotStep{
		"coverage": {ExitCode: 0, TotalPercent: floatPtr(95.0)},
		"tests":    {ExitCode: 0},
	}}
	curr := snapshot{Steps: map[string]snapshotStep{
		"coverage": {ExitCode: 0, TotalPercent: floatPtr(90.0)},
		"tests":    {ExitCode: 1},
	}}
	regress := regressions(base, curr)
	if len(regress) == 0 {
		t.Fatalf("expected regressions")
	}
}

func TestRegressionBenchmarks(t *testing.T) {
	base := snapshot{Benchmarks: []benchmark{{
		Name:        "BenchmarkFoo",
		Pkg:         "pkg",
		NsPerOp:     10,
		BytesPerOp:  intPtr(5),
		AllocsPerOp: intPtr(1),
	}}}
	curr := snapshot{Benchmarks: []benchmark{{
		Name:        "BenchmarkFoo",
		Pkg:         "pkg",
		NsPerOp:     20,
		BytesPerOp:  intPtr(10),
		AllocsPerOp: intPtr(2),
	}}}
	regress := regressions(base, curr)
	if len(regress) < 3 {
		t.Fatalf("expected benchmark regressions")
	}
}

func TestHelpers(t *testing.T) {
	if quoteString("a") != "\"a\"" {
		t.Fatalf("unexpected quote")
	}
	if hashBytes([]byte("a")) == "" {
		t.Fatalf("expected hash")
	}
	if formatFloat(1.25) != "1.25" {
		t.Fatalf("unexpected format")
	}
	array := []string{"a", "b"}
	formatted := formatStringArray(array)
	if formatted == "" {
		t.Fatalf("expected formatted array")
	}
	parsed, ok := parseStringArray([]any{"a", "b"})
	if !ok || len(parsed) != 2 {
		t.Fatalf("unexpected parse: %v", parsed)
	}
	if toInt(float64(2)) != 2 {
		t.Fatalf("unexpected toInt")
	}
	if toInt(4) != 4 {
		t.Fatalf("unexpected toInt int")
	}
	if toInt(int64(3)) != 3 {
		t.Fatalf("unexpected toInt int64")
	}
	if toInt("bad") != 0 {
		t.Fatalf("expected default toInt")
	}
}

func TestCheckRegressionNoBaseline(t *testing.T) {
	tmp := t.TempDir()
	current := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}}
	_, err := checkRegression(tmp, current, "")
	if err == nil {
		t.Fatalf("expected regression error")
	}
}

func TestDecodeBenchmarks(t *testing.T) {
	rows := []map[string]any{{
		"name":          "BenchmarkFoo",
		"pkg":           "pkg",
		"ns_per_op":     12.5,
		"iterations":    int64(10),
		"bytes_per_op":  float64(4),
		"allocs_per_op": float64(1),
	}}
	raw := []any{rows[0]}
	decoded := decodeBenchmarks(raw)
	if len(decoded) != 1 || decoded[0].Name != "BenchmarkFoo" {
		t.Fatalf("unexpected decoded")
	}
}

func TestDecodeBenchmarksSkipsInvalid(t *testing.T) {
	raw := []any{"bad", map[string]any{"name": "BenchmarkFoo"}}
	decoded := decodeBenchmarks(raw)
	if len(decoded) != 1 {
		t.Fatalf("expected one decoded entry")
	}
}

func TestWriteSnapshotFile(t *testing.T) {
	tmp := t.TempDir()
	created := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	snap := snapshot{Meta: snapshotMeta{CreatedAt: created}}
	prePath := filepath.Join(tmp, "snapshots", "amux-test-"+created.Format("20060102T150405Z")+".toml")
	if err := os.MkdirAll(filepath.Dir(prePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(prePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write pre snapshot: %v", err)
	}
	path, err := writeSnapshotFile(tmp, snap)
	if err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	if path == "" {
		t.Fatalf("expected snapshot path")
	}
	if !strings.Contains(filepath.Base(path), "-1.toml") {
		t.Fatalf("expected snapshot suffix, got %s", path)
	}
}

func TestWriteSnapshotFileInvalidDir(t *testing.T) {
	tmp := t.TempDir()
	snap := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}}
	blocker := filepath.Join(tmp, "snapshots")
	if err := os.WriteFile(blocker, []byte("nope"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	if _, err := writeSnapshotFile(tmp, snap); err == nil {
		t.Fatalf("expected snapshot dir error")
	}
}

func writeStubGo(binDir string, coverage string) error {
	script := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "tool" ] && [ "$2" = "cover" ]; then
  echo "total: (statements) %s%%"
  exit 0
fi
for arg in "$@"; do
  case "$arg" in
    -coverprofile=*)
      path="${arg#-coverprofile=}"
      mkdir -p "$(dirname "$path")"
      echo "mode: set" > "$path"
      ;;
  esac
done
exit 0
`, coverage)
	path := filepath.Join(binDir, "go")
	return os.WriteFile(path, []byte(script), 0o755)
}

func writeStubGoOutput(binDir string, output string, exitCode int) error {
	script := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "tool" ] && [ "$2" = "cover" ]; then
  echo "%s"
  exit %d
fi
exit 0
`, output, exitCode)
	path := filepath.Join(binDir, "go")
	return os.WriteFile(path, []byte(script), 0o755)
}

func writeStubLint(binDir string) error {
	script := "#!/bin/sh\nexit 0\n"
	path := filepath.Join(binDir, "golangci-lint")
	return os.WriteFile(path, []byte(script), 0o755)
}

func floatPtr(value float64) *float64 {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func TestRunStepsStubbed(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "93.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	steps, err := runSteps(tmp)
	if err != nil {
		t.Fatalf("run steps: %v", err)
	}
	if _, ok := steps["coverage"]; !ok {
		t.Fatalf("expected coverage step")
	}
}

func TestParseStringArrayErrors(t *testing.T) {
	if _, ok := parseStringArray([]any{"a", 1}); ok {
		t.Fatalf("expected parse error")
	}
}

func TestSummarizeSteps(t *testing.T) {
	steps := map[string]stepResult{
		"tests": {argv: []string{"go", "test"}, stdout: []byte("ok"), stderr: []byte(""), exitCode: 0, duration: time.Millisecond},
	}
	if err := summarizeSteps(&bytes.Buffer{}, steps); err != nil {
		t.Fatalf("summarize: %v", err)
	}
	steps["tests"] = stepResult{argv: []string{"go", "test"}, stdout: []byte("fail"), stderr: []byte(""), exitCode: 1, duration: time.Millisecond}
	if err := summarizeSteps(&bytes.Buffer{}, steps); err == nil {
		t.Fatalf("expected summarize error")
	}
}

func TestRunTestInvalidModuleRoot(t *testing.T) {
	wd, _ := os.Getwd()
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := runTest([]string{}); err == nil {
		t.Fatalf("expected module root error")
	}
}

func TestRunTestWritesSnapshot(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "95.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := runTest([]string{}); err != nil {
		t.Fatalf("runTest: %v", err)
	}
	snapDir := filepath.Join(tmp, "snapshots")
	entries, err := os.ReadDir(snapDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected snapshot files")
	}
}

func TestRunTestRegressionReport(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "95.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	snapDir := filepath.Join(tmp, "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshots: %v", err)
	}
	base := snapshot{
		Meta: snapshotMeta{CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), ModuleRoot: tmp, SpecVersion: specVersion},
		Steps: map[string]snapshotStep{
			"coverage": {ExitCode: 0, TotalPercent: floatPtr(96.0)},
		},
	}
	encoded, err := encodeSnapshot(base)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	basePath := filepath.Join(snapDir, "amux-test-20200101T000000Z.toml")
	if err := os.WriteFile(basePath, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := runTest([]string{"--no-snapshot", "--regression"}); err == nil {
		t.Fatalf("expected regression error")
	}
}

func TestCheckRegressionWithBaseline(t *testing.T) {
	tmp := t.TempDir()
	base := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}, Steps: map[string]snapshotStep{"tests": {ExitCode: 0}}}
	encoded, _ := encodeSnapshot(base)
	path := filepath.Join(tmp, "snapshots", "amux-test-20200101T000000Z.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	current := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}, Steps: map[string]snapshotStep{"tests": {ExitCode: 0}}}
	report, err := checkRegression(tmp, current, "")
	if err != nil {
		t.Fatalf("check regression: %v", err)
	}
	if len(report) != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestReadSnapshotErrors(t *testing.T) {
	if _, err := readSnapshot("nope"); err == nil {
		t.Fatalf("expected read error")
	}
	if _, err := decodeSnapshot(map[string]any{}); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestDecodeSnapshotInvalidTimestamp(t *testing.T) {
	raw := map[string]any{
		"meta": map[string]any{"created_at": "not-a-time"},
		"steps": map[string]any{
			"tests": map[string]any{"exit_code": 0},
		},
	}
	if _, err := decodeSnapshot(raw); err == nil {
		t.Fatalf("expected invalid timestamp error")
	}
}

func TestDecodeSnapshotTotalPercentInt(t *testing.T) {
	tmp := t.TempDir()
	steps := map[string]snapshotStep{
		"coverage": {ExitCode: 0, TotalPercent: floatPtr(96.0)},
	}
	snap := snapshot{
		Meta:  snapshotMeta{CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), ModuleRoot: tmp, SpecVersion: specVersion},
		Steps: steps,
	}
	encoded, err := encodeSnapshot(snap)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	path := filepath.Join(tmp, "snapshots", "baseline.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	decoded, err := readSnapshot(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if decoded.Steps["coverage"].TotalPercent == nil || *decoded.Steps["coverage"].TotalPercent != 96 {
		t.Fatalf("expected total percent decoded")
	}
}

func TestFindBaselineSnapshotSelectsLatest(t *testing.T) {
	tmp := t.TempDir()
	snapDir := filepath.Join(tmp, "snapshots")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	first := filepath.Join(snapDir, "amux-test-20200101T000000Z.toml")
	second := filepath.Join(snapDir, "amux-test-20200102T000000Z.toml")
	if err := os.WriteFile(first, []byte("data"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(second, []byte("data"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	found, err := findBaselineSnapshot(tmp, second)
	if err != nil {
		t.Fatalf("find baseline: %v", err)
	}
	if found != first {
		t.Fatalf("expected baseline %s, got %s", first, found)
	}
}

func TestEnsureExecutables(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "90.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if err := ensureExecutables([]string{"go"}); err != nil {
		t.Fatalf("ensure executables: %v", err)
	}
	if err := ensureExecutables([]string{"missing-exec"}); err == nil {
		t.Fatalf("expected missing executable")
	}
}

func TestFindModuleRoot(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	if root != tmp {
		t.Fatalf("unexpected root: %s", root)
	}
}

func TestCheckRegressionOutput(t *testing.T) {
	tmp := t.TempDir()
	base := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}, Steps: map[string]snapshotStep{"tests": {ExitCode: 0}}}
	encoded, _ := encodeSnapshot(base)
	path := filepath.Join(tmp, "snapshots", "amux-test-20200101T000000Z.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(encoded), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	current := snapshot{Meta: snapshotMeta{CreatedAt: time.Now().UTC()}, Steps: map[string]snapshotStep{"tests": {ExitCode: 1}}}
	report, err := checkRegression(tmp, current, "")
	if err != nil {
		t.Fatalf("check regression: %v", err)
	}
	if len(report) == 0 {
		t.Fatalf("expected regression report")
	}
}

func TestRunTestRegressionFlag(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGo(binDir, "95.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLint(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := runTest([]string{"--no-snapshot", "--regression"}); err == nil {
		t.Fatalf("expected regression error without baseline")
	}
}

func TestParseCoverageTotalMissingLine(t *testing.T) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	script := "#!/bin/sh\necho \"no total\"; exit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "go"), []byte(script), 0o755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	pathEnv := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+pathEnv)
	t.Cleanup(func() { _ = os.Setenv("PATH", pathEnv) })
	if _, err := parseCoverageTotal(filepath.Join(tmp, "coverage.out")); err == nil {
		t.Fatalf("expected coverage total error")
	}
}

func TestRunStepExitCode(t *testing.T) {
	tmp := t.TempDir()
	script := "#!/bin/sh\nexit 2\n"
	path := filepath.Join(tmp, "fail")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	result := runStep(tmp, []string{path})
	if result.exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", result.exitCode)
	}
}

func TestFormatStringArrayEmpty(t *testing.T) {
	if formatStringArray(nil) != "[]" {
		t.Fatalf("expected empty array")
	}
	if _, ok := parseStringArray([]any{}); !ok {
		t.Fatalf("parse empty failed")
	}
}

func TestWriteJSONSnapshotRoundTrip(t *testing.T) {
	snap := snapshot{
		Meta: snapshotMeta{CreatedAt: time.Now().UTC(), ModuleRoot: "root", SpecVersion: specVersion},
		Steps: map[string]snapshotStep{
			"coverage": {TotalPercent: floatPtr(95.5)},
		},
		Benchmarks: []benchmark{{
			Name:       "BenchmarkFoo",
			Pkg:        "pkg",
			NsPerOp:    1,
			Iterations: 1,
			BytesPerOp: intPtr(4),
			AllocsPerOp: intPtr(2),
		}},
	}
	encoded, err := encodeSnapshot(snap)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(encoded, "spec_version") {
		t.Fatalf("expected spec_version")
	}
}

func TestBenchmarkIndex(t *testing.T) {
	items := []benchmark{{Name: "BenchmarkFoo", Pkg: "pkg", NsPerOp: 1}}
	idx := benchmarkIndex(items)
	if idx[benchKey(items[0])].Name != "BenchmarkFoo" {
		t.Fatalf("unexpected benchmark")
	}
}

func TestRunTestWithBadFlags(t *testing.T) {
	if err := runTest([]string{"--unknown"}); err == nil {
		t.Fatalf("expected flag error")
	}
}

func TestParseBenchmarksSkipsInvalid(t *testing.T) {
	output := []byte("BenchmarkBad\nBenchmarkFoo-8 bad 123 ns/op\n")
	benchmarks := parseBenchmarks(output)
	if len(benchmarks) != 0 {
		t.Fatalf("expected no benchmarks")
	}
}

func TestDecodeSnapshotErrors(t *testing.T) {
	if _, err := decodeSnapshot(map[string]any{}); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestRunTestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = ctx
}
