// Package conformance provides the conformance test harness.
// This package implements the conformance suite skeleton that boots a daemon
// + CLI client fixture and records structured JSON results.
package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Common sentinel errors for conformance operations.
var (
	// ErrHarnessNotReady indicates the test harness is not ready.
	ErrHarnessNotReady = errors.New("harness not ready")

	// ErrTestFailed indicates a conformance test failed.
	ErrTestFailed = errors.New("test failed")

	// ErrFixtureSetupFailed indicates test fixture setup failed.
	ErrFixtureSetupFailed = errors.New("fixture setup failed")
)

// TestResult represents a conformance test result.
type TestResult struct {
	// TestName is the name of the test.
	TestName string `json:"test_name"`

	// Status is the test status ("pass", "fail", "skip").
	Status string `json:"status"`

	// Duration is how long the test took.
	Duration time.Duration `json:"duration"`

	// Error is the error message if the test failed.
	Error string `json:"error,omitempty"`

	// Metadata contains test-specific metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Suite represents a conformance test suite.
type Suite struct {
	results []TestResult
	ready   bool
}

// NewSuite creates a new conformance test suite.
func NewSuite() *Suite {
	return &Suite{
		results: make([]TestResult, 0),
		ready:   false,
	}
}

// Setup initializes the test fixtures (daemon + CLI client).
// Phase 0: Placeholder implementation.
func (s *Suite) Setup() error {
	// Phase 0: Fixture setup not yet implemented
	return fmt.Errorf("conformance fixtures not implemented: %w", ErrFixtureSetupFailed)
}

// RunTest executes a single conformance test.
// Phase 0: Placeholder implementation.
func (s *Suite) RunTest(testName string) error {
	if !s.ready {
		return fmt.Errorf("test harness not ready: %w", ErrHarnessNotReady)
	}

	start := time.Now()
	result := TestResult{
		TestName: testName,
		Status:   "skip",
		Duration: time.Since(start),
		Error:    "Phase 0: conformance tests not implemented",
	}

	s.results = append(s.results, result)
	return nil
}

// GetResults returns the test results as JSON.
func (s *Suite) GetResults() ([]byte, error) {
	return json.MarshalIndent(s.results, "", "  ")
}

// Cleanup tears down test fixtures.
func (s *Suite) Cleanup() error {
	s.ready = false
	return nil
}