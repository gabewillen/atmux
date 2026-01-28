// Package conformance implements tests for the conformance harness
package conformance

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestNewHarness tests creating a new conformance harness
func TestNewHarness(t *testing.T) {
	harness := NewHarness("v1.0.0")

	if harness == nil {
		t.Fatal("Expected harness to be created")
	}

	if harness.specVersion != "v1.0.0" {
		t.Errorf("Expected spec version 'v1.0.0', got '%s'", harness.specVersion)
	}

	if harness.runID == "" {
		t.Error("Expected run ID to be generated")
	}

	if harness.results == nil {
		t.Error("Expected results slice to be initialized")
	}

	if len(harness.results) != 0 {
		t.Error("Expected empty results initially")
	}
}

// TestRunFlow tests running a single conformance flow
func TestRunFlow(t *testing.T) {
	harness := NewHarness("v1.0.0")

	ctx := context.Background()

	// Test successful flow
	harness.RunFlow(ctx, "success-flow", func(ctx context.Context) error {
		return nil
	})

	if len(harness.results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(harness.results))
	}

	result := harness.results[0]
	if result.Name != "success-flow" {
		t.Errorf("Expected flow name 'success-flow', got '%s'", result.Name)
	}

	if result.Status != "pass" {
		t.Errorf("Expected status 'pass', got '%s'", result.Status)
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got '%s'", result.Error)
	}

	if result.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be set")
	}

	if result.EndedAt.IsZero() {
		t.Error("Expected EndedAt to be set")
	}

	if result.EndedAt.Before(result.StartedAt) {
		t.Error("Expected EndedAt to be after StartedAt")
	}
}

// TestRunFlowWithError tests running a flow that returns an error
func TestRunFlowWithError(t *testing.T) {
	harness := NewHarness("v1.0.0")

	ctx := context.Background()

	// Test failing flow
	harness.RunFlow(ctx, "failure-flow", func(ctx context.Context) error {
		return &FlowError{"test error"}
	})

	if len(harness.results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(harness.results))
	}

	result := harness.results[0]
	if result.Name != "failure-flow" {
		t.Errorf("Expected flow name 'failure-flow', got '%s'", result.Name)
	}

	if result.Status != "fail" {
		t.Errorf("Expected status 'fail', got '%s'", result.Status)
	}

	if result.Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", result.Error)
	}
}

// TestRun tests running multiple flows
func TestRun(t *testing.T) {
	harness := NewHarness("v1.0.0")

	flows := map[string]func(context.Context) error{
		"flow-1": func(ctx context.Context) error { return nil },
		"flow-2": func(ctx context.Context) error { return &FlowError{"error"} },
		"flow-3": func(ctx context.Context) error { return nil },
	}

	ctx := context.Background()
	harness.Run(ctx, flows)

	if len(harness.results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(harness.results))
	}

	// Check individual results
	nameToResult := make(map[string]FlowResult)
	for _, r := range harness.results {
		nameToResult[r.Name] = r
	}

	r1 := nameToResult["flow-1"]
	if r1.Status != "pass" {
		t.Errorf("Expected flow-1 to pass, got %s", r1.Status)
	}

	r2 := nameToResult["flow-2"]
	if r2.Status != "fail" {
		t.Errorf("Expected flow-2 to fail, got %s", r2.Status)
	}

	r3 := nameToResult["flow-3"]
	if r3.Status != "pass" {
		t.Errorf("Expected flow-3 to pass, got %s", r3.Status)
	}
}

// TestWriteResults tests writing results to a file
func TestWriteResults(t *testing.T) {
	harness := NewHarness("v1.0.0")

	ctx := context.Background()
	harness.RunFlow(ctx, "test-flow", func(ctx context.Context) error {
		return nil
	})

	tempDir := t.TempDir()
	resultPath := filepath.Join(tempDir, "results.json")

	err := harness.WriteResults(resultPath)
	if err != nil {
		t.Fatalf("Unexpected error writing results: %v", err)
	}

	// Check that file exists and has content
	content, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("Error reading results file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty results file")
	}

	// Verify the content is valid JSON with expected structure
	var runResult RunResult
	err = json.Unmarshal(content, &runResult)
	if err != nil {
		t.Fatalf("Invalid JSON in results file: %v", err)
	}

	if runResult.RunID != harness.runID {
		t.Errorf("Expected RunID '%s', got '%s'", harness.runID, runResult.RunID)
	}

	if runResult.SpecVersion != "v1.0.0" {
		t.Errorf("Expected spec version 'v1.0.0', got '%s'", runResult.SpecVersion)
	}

	if len(runResult.Results) != 1 {
		t.Errorf("Expected 1 result in RunResult, got %d", len(runResult.Results))
	}

	if runResult.Results[0].Name != "test-flow" {
		t.Errorf("Expected result name 'test-flow', got '%s'", runResult.Results[0].Name)
	}
}

// TestResultCounts tests the counting methods
func TestResultCounts(t *testing.T) {
	harness := NewHarness("v1.0.0")

	ctx := context.Background()

	// Add a passing flow
	harness.RunFlow(ctx, "passing", func(ctx context.Context) error {
		return nil
	})

	// Add a failing flow
	harness.RunFlow(ctx, "failing", func(ctx context.Context) error {
		return &FlowError{"error"}
	})

	// Add another passing flow
	harness.RunFlow(ctx, "passing2", func(ctx context.Context) error {
		return nil
	})

	// Add a skipped flow (we'll simulate this by manually adding it)
	harness.results = append(harness.results, FlowResult{
		Name:   "skipped",
		Status: "skip",
	})

	if harness.CountPasses() != 2 {
		t.Errorf("Expected 2 passes, got %d", harness.CountPasses())
	}

	if harness.CountFailures() != 1 {
		t.Errorf("Expected 1 failure, got %d", harness.CountFailures())
	}

	if harness.CountSkipped() != 1 {
		t.Errorf("Expected 1 skipped, got %d", harness.CountSkipped())
	}

	if harness.TotalFlows() != 4 {
		t.Errorf("Expected 4 total flows, got %d", harness.TotalFlows())
	}
}

// FlowError is a simple error type for testing
type FlowError struct {
	msg string
}

func (e *FlowError) Error() string {
	return e.msg
}