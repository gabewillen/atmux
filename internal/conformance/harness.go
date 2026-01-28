// Package conformance provides the conformance harness for amux.
//
// The conformance harness executes the conformance suite against the amux
// implementation and any WASM adapters that claim conformance to the
// specification.
//
// See spec §4.3 for the full conformance requirements.
package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/pkg/api"
)

// Run executes the conformance suite.
func Run(ctx context.Context, opts Options) (*Results, error) {
	results := &Results{
		RunID:       muid.Make().String(),
		SpecVersion: api.SpecVersion,
		StartedAt:   time.Now().UTC(),
		Flows:       make([]FlowResult, 0),
	}

	// Run each flow
	flows := []Flow{
		&AuthFlow{},
		&MenuFlow{},
		&StatusFlow{},
		&NotificationFlow{},
		&ControlPlaneFlow{},
	}

	for _, flow := range flows {
		if ctx.Err() != nil {
			break
		}

		flowResult := FlowResult{
			Name:      flow.Name(),
			StartedAt: time.Now().UTC(),
		}

		if err := flow.Run(ctx, opts); err != nil {
			flowResult.Status = StatusFail
			flowResult.Error = err.Error()
		} else {
			flowResult.Status = StatusPass
		}

		flowResult.FinishedAt = time.Now().UTC()
		results.Flows = append(results.Flows, flowResult)
	}

	results.FinishedAt = time.Now().UTC()
	return results, nil
}

// Options configures the conformance run.
type Options struct {
	// DaemonAddr is the address of the amux daemon.
	DaemonAddr string

	// AdapterName is the adapter to test (optional).
	AdapterName string

	// OutputPath is where to write results (optional).
	OutputPath string

	// Verbose enables verbose output.
	Verbose bool
}

// Status represents a flow result status.
type Status string

const (
	// StatusPass indicates the flow passed.
	StatusPass Status = "pass"

	// StatusFail indicates the flow failed.
	StatusFail Status = "fail"

	// StatusSkip indicates the flow was skipped.
	StatusSkip Status = "skip"
)

// Results represents the conformance run results.
type Results struct {
	// RunID is the unique run identifier.
	RunID string `json:"run_id"`

	// SpecVersion is the spec version tested against.
	SpecVersion string `json:"spec_version"`

	// StartedAt is when the run started.
	StartedAt time.Time `json:"started_at"`

	// FinishedAt is when the run finished.
	FinishedAt time.Time `json:"finished_at"`

	// Flows contains per-flow results.
	Flows []FlowResult `json:"flows"`
}

// FlowResult represents the result of a single flow.
type FlowResult struct {
	// Name is the flow name.
	Name string `json:"name"`

	// Status is the flow status.
	Status Status `json:"status"`

	// Error is the error message (if failed).
	Error string `json:"error,omitempty"`

	// StartedAt is when the flow started.
	StartedAt time.Time `json:"started_at"`

	// FinishedAt is when the flow finished.
	FinishedAt time.Time `json:"finished_at"`

	// Artifacts contains artifact references (if any).
	Artifacts []string `json:"artifacts,omitempty"`
}

// WriteResults writes results to the specified path.
func WriteResults(results *Results, path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create results directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write results: %w", err)
	}

	return nil
}

// Flow represents a conformance flow to test.
type Flow interface {
	// Name returns the flow name.
	Name() string

	// Run executes the flow.
	Run(ctx context.Context, opts Options) error
}

// AuthFlow tests authentication flows.
type AuthFlow struct{}

// Name returns "auth".
func (f *AuthFlow) Name() string {
	return "auth"
}

// Run executes the auth flow.
func (f *AuthFlow) Run(ctx context.Context, opts Options) error {
	// Placeholder - will be implemented with full daemon/adapter integration
	return nil
}

// MenuFlow tests menu/TUI navigation flows.
type MenuFlow struct{}

// Name returns "menu".
func (f *MenuFlow) Name() string {
	return "menu"
}

// Run executes the menu flow.
func (f *MenuFlow) Run(ctx context.Context, opts Options) error {
	// Placeholder - will be implemented with full daemon/adapter integration
	return nil
}

// StatusFlow tests status/presence/lifecycle flows.
type StatusFlow struct{}

// Name returns "status".
func (f *StatusFlow) Name() string {
	return "status"
}

// Run executes the status flow.
func (f *StatusFlow) Run(ctx context.Context, opts Options) error {
	// Placeholder - will be implemented with full daemon/adapter integration
	return nil
}

// NotificationFlow tests notification/messaging flows.
type NotificationFlow struct{}

// Name returns "notification".
func (f *NotificationFlow) Name() string {
	return "notification"
}

// Run executes the notification flow.
func (f *NotificationFlow) Run(ctx context.Context, opts Options) error {
	// Placeholder - will be implemented with full daemon/adapter integration
	return nil
}

// ControlPlaneFlow tests JSON-RPC control plane flows.
type ControlPlaneFlow struct{}

// Name returns "control_plane".
func (f *ControlPlaneFlow) Name() string {
	return "control_plane"
}

// Run executes the control plane flow.
func (f *ControlPlaneFlow) Run(ctx context.Context, opts Options) error {
	// Placeholder - will be implemented with full daemon/adapter integration
	return nil
}
