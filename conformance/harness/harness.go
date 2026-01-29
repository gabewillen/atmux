// Package harness provides the conformance testing harness per spec §4.3.1.
package harness

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
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

// Run executes a minimal conformance suite against the local amux binaries.
//
// Phase 0–3: This boots the daemon stub (amux-node), runs basic CLI flows, and
// invokes `amux test` to exercise the verification entrypoint, then records
// structured JSON results per the output contract.
func Run(ctx context.Context) (*RunResult, error) {
	start := time.Now()
	result := &RunResult{
		RunID:       "phase0-3-local",
		SpecVersion: "v1.22",
		StartedAt:   start,
	}

	var flows []FlowResult

	// Flow 1: Boot the daemon stub (amux-node) with default behavior.
	{
		flowStart := time.Now()
		cmd := exec.CommandContext(ctx, "amux-node")
		if err := cmd.Run(); err != nil {
			flows = append(flows, FlowResult{
				Name:     "daemon.stub",
				Status:   "fail",
				Error:    err.Error(),
				Duration: time.Since(flowStart).String(),
			})
		} else {
			flows = append(flows, FlowResult{
				Name:     "daemon.stub",
				Status:   "pass",
				Duration: time.Since(flowStart).String(),
			})
		}
	}

	// Flow 2: Run a simple CLI command (amux version).
	{
		flowStart := time.Now()
		cmd := exec.CommandContext(ctx, "amux", "version")
		if err := cmd.Run(); err != nil {
			flows = append(flows, FlowResult{
				Name:     "cli.version",
				Status:   "fail",
				Error:    err.Error(),
				Duration: time.Since(flowStart).String(),
			})
		} else {
			flows = append(flows, FlowResult{
				Name:     "cli.version",
				Status:   "pass",
				Duration: time.Since(flowStart).String(),
			})
		}
	}

	// Flow 3: Run `amux test --no-snapshot` to exercise the verification
	// snapshot pipeline without writing snapshot files.
	{
		flowStart := time.Now()
		cmd := exec.CommandContext(ctx, "amux", "test", "--no-snapshot")
		if err := cmd.Run(); err != nil {
			flows = append(flows, FlowResult{
				Name:     "cli.amux_test_no_snapshot",
				Status:   "fail",
				Error:    err.Error(),
				Duration: time.Since(flowStart).String(),
			})
		} else {
			flows = append(flows, FlowResult{
				Name:     "cli.amux_test_no_snapshot",
				Status:   "pass",
				Duration: time.Since(flowStart).String(),
			})
		}
	}

	result.Flows = flows
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
