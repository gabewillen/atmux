# package conformance

`import "github.com/agentflare-ai/amux/internal/conformance"`

Package conformance provides the conformance test harness for amux.

- `type FlowResult` — FlowResult represents the result of a single conformance flow.
- `type Harness` — Harness runs the conformance suite.
- `type Result` — Result represents a conformance test result.

## type FlowResult

```go
type FlowResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "pass", "fail", "skip"
	Error  string `json:"error,omitempty"`
}
```

FlowResult represents the result of a single conformance flow.

## type Harness

```go
type Harness struct {
	outputPath string
}
```

Harness runs the conformance suite.

### Functions returning Harness

#### NewHarness

```go
func NewHarness(outputPath string) *Harness
```

NewHarness creates a new conformance harness.


### Methods

#### Harness.Run

```go
func () Run(ctx context.Context) error
```

Run runs the conformance suite and writes results.


## type Result

```go
type Result struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Flows       []FlowResult `json:"flows"`
}
```

Result represents a conformance test result.

