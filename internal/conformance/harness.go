// Package conformance provides the conformance test harness for amux.
package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Result represents a conformance test result.
type Result struct {
	RunID       string    `json:"run_id"`
	SpecVersion string    `json:"spec_version"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Flows       []FlowResult `json:"flows"`
}

// FlowResult represents the result of a single conformance flow.
type FlowResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "fail", "skip"
	Error  string `json:"error,omitempty"`
}

// Harness runs the conformance suite.
type Harness struct {
	outputPath string
}

// NewHarness creates a new conformance harness.
func NewHarness(outputPath string) *Harness {
	return &Harness{
		outputPath: outputPath,
	}
}

// Run runs the conformance suite and writes results.
func (h *Harness) Run(ctx context.Context) error {
	result := Result{
		RunID:       fmt.Sprintf("run-%d", time.Now().Unix()),
		SpecVersion: "v1.22",
		StartedAt:   time.Now(),
		Flows:       []FlowResult{},
	}

	// Phase 0: Placeholder flows
	flows := []string{
		"auth",
		"menu",
		"status",
		"notification",
		"control_plane",
	}

	for _, flowName := range flows {
		flowResult := FlowResult{
			Name:   flowName,
			Status: "skip", // Phase 0: All flows skipped
		}
		result.Flows = append(result.Flows, flowResult)
	}

	result.FinishedAt = time.Now()

	// Write results
	if h.outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(h.outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}

		if err := os.WriteFile(h.outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}
	}

	return nil
}
