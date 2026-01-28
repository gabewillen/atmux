// Package conformance implements the conformance harness for amux
package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// FlowResult represents the result of a single conformance flow
type FlowResult struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "pass", "fail", "skip"
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// RunResult represents the overall conformance run result
type RunResult struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Results     []FlowResult `json:"results"`
}

// Harness manages the execution of conformance tests
type Harness struct {
	specVersion string
	runID       string
	results     []FlowResult
}

// NewHarness creates a new conformance harness
func NewHarness(specVersion string) *Harness {
	return &Harness{
		specVersion: specVersion,
		runID:       fmt.Sprintf("run-%d", time.Now().UnixNano()),
		results:     make([]FlowResult, 0),
	}
}

// RunFlow executes a single conformance flow
func (h *Harness) RunFlow(ctx context.Context, name string, flowFn func(context.Context) error) {
	startTime := time.Now()
	
	result := FlowResult{
		Name:      name,
		Status:    "pass",
		StartedAt: startTime,
	}
	
	err := flowFn(ctx)
	if err != nil {
		result.Status = "fail"
		result.Error = err.Error()
	}
	
	result.EndedAt = time.Now()
	h.results = append(h.results, result)
}

// Run executes all registered conformance flows
func (h *Harness) Run(ctx context.Context, flows map[string]func(context.Context) error) {
	for name, flow := range flows {
		h.RunFlow(ctx, name, flow)
	}
}

// WriteResults writes the conformance results to the specified path
func (h *Harness) WriteResults(path string) error {
	result := RunResult{
		RunID:       h.runID,
		SpecVersion: h.specVersion,
		StartedAt:   time.Time{}, // Will be computed as min of all flow start times
		FinishedAt:  time.Time{}, // Will be computed as max of all flow end times
		Results:     h.results,
	}
	
	// Calculate overall start/end times
	if len(h.results) > 0 {
		result.StartedAt = h.results[0].StartedAt
		result.FinishedAt = h.results[0].EndedAt
		
		for _, r := range h.results {
			if r.StartedAt.Before(result.StartedAt) {
				result.StartedAt = r.StartedAt
			}
			if r.EndedAt.After(result.FinishedAt) {
				result.FinishedAt = r.EndedAt
			}
		}
	}
	
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// CountPasses returns the number of passed flows
func (h *Harness) CountPasses() int {
	count := 0
	for _, r := range h.results {
		if r.Status == "pass" {
			count++
		}
	}
	return count
}

// CountFailures returns the number of failed flows
func (h *Harness) CountFailures() int {
	count := 0
	for _, r := range h.results {
		if r.Status == "fail" {
			count++
		}
	}
	return count
}

// CountSkipped returns the number of skipped flows
func (h *Harness) CountSkipped() int {
	count := 0
	for _, r := range h.results {
		if r.Status == "skip" {
			count++
		}
	}
	return count
}

// TotalFlows returns the total number of flows
func (h *Harness) TotalFlows() int {
	return len(h.results)
}