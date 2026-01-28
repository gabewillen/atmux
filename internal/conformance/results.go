package conformance

import "time"

// Results describes a conformance run.
type Results struct {
	RunID      string       `json:"run_id"`
	SpecVersion string      `json:"spec_version"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Flows      []FlowResult `json:"flows"`
}

// FlowResult describes a single flow outcome.
type FlowResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
