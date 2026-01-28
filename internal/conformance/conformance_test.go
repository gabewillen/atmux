package conformance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHarness(t *testing.T) {
	runConfig := &RunConfig{
		OutputPath: "test-results.json",
		Timeout:    5 * time.Second,
		Patterns:   []string{"agent-management"},
		Verbose:    true,
		CI:         false,
	}

	harness := NewHarness(runConfig)

	if harness == nil {
		t.Fatal("Expected non-nil harness")
	}

	if harness.config != runConfig {
		t.Error("Expected config to be set")
	}

	if harness.runResult == nil {
		t.Error("Expected non-nil run result")
	}

	if harness.runResult.RunID == "" {
		t.Error("Expected non-empty run ID")
	}

	if harness.runResult.SpecVersion != "v1.22" {
		t.Errorf("Expected spec version 'v1.22', got %q", harness.runResult.SpecVersion)
	}
}

func TestHarnessRun(t *testing.T) {
	// Create temporary output file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-results.json")

	runConfig := &RunConfig{
		OutputPath: outputPath,
		Timeout:    10 * time.Second,
		Patterns:   []string{}, // Run all
		Verbose:    false,
		CI:         false,
	}

	harness := NewHarness(runConfig)
	result, err := harness.Run()

	if err != nil {
		t.Fatalf("Unexpected error during run: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check result structure
	if result.RunID == "" {
		t.Error("Expected non-empty run ID")
	}

	if result.SpecVersion != "v1.22" {
		t.Errorf("Expected spec version 'v1.22', got %q", result.SpecVersion)
	}

	if result.FinishedAt == nil {
		t.Error("Expected finished time to be set")
	}

	if result.Summary == nil {
		t.Error("Expected summary to be set")
	}

	// Check that file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected results file to be created")
	}

	// Load and verify JSON content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read results file: %v", err)
	}

	var loadedResult RunResult
	if err := json.Unmarshal(data, &loadedResult); err != nil {
		t.Fatalf("Failed to unmarshal results: %v", err)
	}

	if loadedResult.RunID != result.RunID {
		t.Error("Loaded result doesn't match saved result")
	}
}

func TestFilterFlows(t *testing.T) {
	harness := NewHarness(&RunConfig{})

	allFlows := []string{
		"agent-management",
		"presence-status",
		"pty-monitoring",
		"process-tracking",
	}

	// Test pattern filter
	harness.config.Patterns = []string{"agent"}
	filtered := harness.filterFlows(allFlows)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered flow, got %d", len(filtered))
	}

	if filtered[0] != "agent-management" {
		t.Errorf("Expected 'agent-management', got %q", filtered[0])
	}

	// Test exact match
	harness.config.Patterns = []string{"presence-status"}
	filtered = harness.filterFlows(allFlows)

	if len(filtered) != 1 || filtered[0] != "presence-status" {
		t.Error("Failed to filter exact match")
	}

	// Test no pattern (should return all)
	harness.config.Patterns = []string{}
	filtered = harness.filterFlows(allFlows)

	if len(filtered) != len(allFlows) {
		t.Error("Should return all flows when no pattern specified")
	}
}

func TestCalculateSummary(t *testing.T) {
	harness := NewHarness(&RunConfig{})

	// Add some test flows
	harness.runResult.Flows = []*FlowResult{
		{Name: "flow1", Status: "pass"},
		{Name: "flow2", Status: "fail"},
		{Name: "flow3", Status: "skip"},
		{Name: "flow4", Status: "pass"},
	}

	summary := harness.calculateSummary()

	if summary.Total != 4 {
		t.Errorf("Expected total 4, got %d", summary.Total)
	}

	if summary.Passed != 2 {
		t.Errorf("Expected passed 2, got %d", summary.Passed)
	}

	if summary.Failed != 1 {
		t.Errorf("Expected failed 1, got %d", summary.Failed)
	}

	if summary.Skipped != 1 {
		t.Errorf("Expected skipped 1, got %d", summary.Skipped)
	}
}

func TestRunSuite(t *testing.T) {
	// Run the full suite
	if err := RunSuite(); err != nil {
		t.Errorf("Unexpected error running suite: %v", err)
	}

	// Check that default results file was created
	if _, err := os.Stat("conformance-results.json"); os.IsNotExist(err) {
		t.Error("Expected default results file to be created")
	}

	// Clean up
	os.Remove("conformance-results.json")
}

func TestValidateOutputPath(t *testing.T) {
	// Test valid path
	validPath := filepath.Join(t.TempDir(), "results.json")
	if err := ValidateOutputPath(validPath); err != nil {
		t.Errorf("Expected no error for valid path: %v", err)
	}

	// Test invalid path (e.g., non-existent directory)
	invalidPath := "/nonexistent/directory/results.json"
	if err := ValidateOutputPath(invalidPath); err == nil {
		t.Error("Expected error for invalid path")
	}
}
