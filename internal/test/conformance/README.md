# package conformance

`import "github.com/agentflare-ai/amux/internal/test/conformance"`

- `type FlowResult`
- `type Result` — Result represents the outcome of a conformance run.
- `type Suite` — Suite represents the conformance test suite.

## type FlowResult

```go
type FlowResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pass, fail, skip
	Error  string `json:"error,omitempty"`
}
```

## type Result

```go
type Result struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Results     []FlowResult `json:"results"`
}
```

Result represents the outcome of a conformance run.
Spec §4.3 (Structured results)

### Functions returning Result

#### Run

```go
func Run(ctx context.Context) (*Result, error)
```

Run executes the conformance suite.
Phase 0: Returns a placeholder result.


## type Suite

```go
type Suite struct {
	Config *config.Config
}
```

Suite represents the conformance test suite.

