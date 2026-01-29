package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/agentflare-ai/amux/pkg/api"
)

// DaemonFixture boots a daemon for conformance runs.
type DaemonFixture interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// CLIFixture boots a CLI client for conformance runs.
type CLIFixture interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Runner executes conformance flows and writes results.
type Runner struct {
	SpecVersion string
	OutputPath  string
	Daemon      DaemonFixture
	CLI         CLIFixture
	Clock       func() time.Time
}

// Run executes the conformance suite and writes structured JSON results.
func (r *Runner) Run(ctx context.Context) (Results, error) {
	clock := r.Clock
	if clock == nil {
		clock = time.Now
	}
	results := Results{
		RunID:       api.NewRuntimeID().String(),
		SpecVersion: r.SpecVersion,
		StartedAt:   clock().UTC(),
	}
	flow := r.runFlow(ctx)
	results.Flows = append(results.Flows, flow)
	results.FinishedAt = clock().UTC()
	if err := writeResults(r.OutputPath, results); err != nil {
		return results, err
	}
	return results, nil
}

func (r *Runner) runFlow(ctx context.Context) FlowResult {
	flow := FlowResult{Name: "bootstrap", Status: "pass"}
	if r.Daemon != nil {
		if err := r.Daemon.Start(ctx); err != nil {
			flow.Status = "fail"
			flow.Error = err.Error()
			return flow
		}
		if err := r.Daemon.Stop(ctx); err != nil {
			flow.Status = "fail"
			flow.Error = err.Error()
			return flow
		}
	}
	if r.CLI != nil {
		if err := r.CLI.Start(ctx); err != nil {
			flow.Status = "fail"
			flow.Error = err.Error()
			return flow
		}
		if err := r.CLI.Stop(ctx); err != nil {
			flow.Status = "fail"
			flow.Error = err.Error()
			return flow
		}
	}
	return flow
}

func writeResults(path string, results Results) error {
	if path == "" {
		return fmt.Errorf("results path is required")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create results dir: %w", err)
	}
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("encode results: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write results: %w", err)
	}
	return nil
}
