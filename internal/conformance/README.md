# package conformance

`import "github.com/agentflare-ai/amux/internal/conformance"`

Package conformance provides the conformance harness and test runner for amux.

- `func RunSuite() error` — RunSuite runs the conformance suite with default configuration.
- `func ValidateOutputPath(path string) error` — ValidateOutputPath checks if the output path is valid.
- `func generateRunID() string` — generateRunID generates a unique run ID.
- `type Artifact` — Artifact represents an output artifact from a flow.
- `type FlowResult` — FlowResult represents the result of a single conformance flow.
- `type Harness` — Harness manages conformance test execution.
- `type RunConfig` — RunConfig contains configuration for conformance runs.
- `type RunResult` — RunResult represents the result of a conformance run.
- `type RunSummary` — RunSummary provides a summary of the conformance run.

### Functions

#### RunSuite

```go
func RunSuite() error
```

RunSuite runs the conformance suite with default configuration.

#### ValidateOutputPath

```go
func ValidateOutputPath(path string) error
```

ValidateOutputPath checks if the output path is valid.

#### generateRunID

```go
func generateRunID() string
```

generateRunID generates a unique run ID.


## type Artifact

```go
type Artifact struct {
	Type string `json:"type"` // "log", "screenshot", "config", etc.
	Path string `json:"path"`
	Size int64  `json:"size,omitempty"`
}
```

Artifact represents an output artifact from a flow.

## type FlowResult

```go
type FlowResult struct {
	Name       string     `json:"name"`
	Status     string     `json:"status"` // "pass", "fail", "skip"
	Error      string     `json:"error,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Artifacts  []Artifact `json:"artifacts,omitempty"`
}
```

FlowResult represents the result of a single conformance flow.

## type Harness

```go
type Harness struct {
	config    *RunConfig
	runResult *RunResult
	ctx       context.Context
	cancel    context.CancelFunc
}
```

Harness manages conformance test execution.

### Functions returning Harness

#### NewHarness

```go
func NewHarness(config *RunConfig) *Harness
```

NewHarness creates a new conformance harness.


### Methods

#### Harness.Run

```go
func () Run() (*RunResult, error)
```

Run executes the conformance suite.

#### Harness.calculateSummary

```go
func () calculateSummary() *RunSummary
```

calculateSummary computes run summary from flow results.

#### Harness.filterFlows

```go
func () filterFlows(flows []string) []string
```

filterFlows filters flows by pattern matching.

#### Harness.matchesPattern

```go
func () matchesPattern(flow, pattern string) bool
```

matchesPattern checks if a flow name matches a pattern.

#### Harness.runFlow

```go
func () runFlow(flowName string) error
```

runFlow executes a single conformance flow.

#### Harness.runFlows

```go
func () runFlows() error
```

runFlows executes all conformance flows.

#### Harness.saveResults

```go
func () saveResults() error
```

saveResults writes the run results to the output file.


## type RunConfig

```go
type RunConfig struct {
	// Output path for results
	OutputPath string

	// Timeout for the entire run
	Timeout time.Duration

	// Test patterns to run
	Patterns []string

	// Verbose output
	Verbose bool

	// Run in CI mode
	CI bool
}
```

RunConfig contains configuration for conformance runs.

## type RunResult

```go
type RunResult struct {
	RunID       string        `json:"run_id"`
	SpecVersion string        `json:"spec_version"`
	StartedAt   time.Time     `json:"started_at"`
	FinishedAt  *time.Time    `json:"finished_at,omitempty"`
	Flows       []*FlowResult `json:"flows"`
	Summary     *RunSummary   `json:"summary"`
}
```

RunResult represents the result of a conformance run.

## type RunSummary

```go
type RunSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}
```

RunSummary provides a summary of the conformance run.

