# package conformance

`import "github.com/stateforward/amux/internal/conformance"`

Package conformance implements the conformance harness for amux

- `type FlowResult` — FlowResult represents the result of a single conformance flow
- `type Harness` — Harness manages the execution of conformance tests
- `type RunResult` — RunResult represents the overall conformance run result

## type FlowResult

```go
type FlowResult struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "pass", "fail", "skip"
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}
```

FlowResult represents the result of a single conformance flow

## type Harness

```go
type Harness struct {
	specVersion string
	runID       string
	results     []FlowResult
}
```

Harness manages the execution of conformance tests

### Functions returning Harness

#### NewHarness

```go
func NewHarness(specVersion string) *Harness
```

NewHarness creates a new conformance harness


### Methods

#### Harness.CountFailures

```go
func () CountFailures() int
```

CountFailures returns the number of failed flows

#### Harness.CountPasses

```go
func () CountPasses() int
```

CountPasses returns the number of passed flows

#### Harness.CountSkipped

```go
func () CountSkipped() int
```

CountSkipped returns the number of skipped flows

#### Harness.Run

```go
func () Run(ctx context.Context, flows map[string]func(context.Context) error)
```

Run executes all registered conformance flows

#### Harness.RunFlow

```go
func () RunFlow(ctx context.Context, name string, flowFn func(context.Context) error)
```

RunFlow executes a single conformance flow

#### Harness.TotalFlows

```go
func () TotalFlows() int
```

TotalFlows returns the total number of flows

#### Harness.WriteResults

```go
func () WriteResults(path string) error
```

WriteResults writes the conformance results to the specified path


## type RunResult

```go
type RunResult struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Results     []FlowResult `json:"results"`
}
```

RunResult represents the overall conformance run result

