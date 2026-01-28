// Package harness provides the conformance testing harness per spec §4.3.1.
package harness

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/stateforward/amux/internal/errors"
)

// RunResult represents the result of a conformance run.
type RunResult struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Flows       []FlowResult `json:"flows"`
}

// FlowResult represents the result of a single conformance flow.
type FlowResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // pass, fail, skip
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

// Run executes the conformance suite.
// Phase 0: Placeholder implementation.
func Run(ctx context.Context) (*RunResult, error) {
	result := &RunResult{
		RunID:       "phase0-stub",
		SpecVersion: "v1.22",
		StartedAt:   time.Now(),
		Flows:       []FlowResult{},
	}
	
	// Phase 0: No actual tests yet
	result.FinishedAt = time.Now()
	
	return result, nil
}

// WriteResults writes conformance results to a file.
func WriteResults(result *RunResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshal results")
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrapf(err, "write results: %s", path)
	}
	
	return nil
}
