# package harness

`import "github.com/stateforward/amux/conformance/harness"`

Package harness provides the conformance testing harness per spec §4.3.1.

- `func WriteResults(result *RunResult, path string) error` — WriteResults writes conformance results to a file.
- `type FlowResult` — FlowResult represents the result of a single conformance flow.
- `type RunResult` — RunResult represents the result of a conformance run.

### Functions

#### WriteResults

```go
func WriteResults(result *RunResult, path string) error
```

WriteResults writes conformance results to a file.


## type FlowResult

```go
type FlowResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // pass, fail, skip
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}
```

FlowResult represents the result of a single conformance flow.

## type RunResult

```go
type RunResult struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Flows       []FlowResult `json:"flows"`
}
```

RunResult represents the result of a conformance run.

### Functions returning RunResult

#### Run

```go
func Run(ctx context.Context) (*RunResult, error)
```

Run executes a minimal conformance suite against the local amux binaries.

Phase 0–3: This boots the daemon stub (amux-node), runs basic CLI flows, and
invokes `amux test` to exercise the verification entrypoint, then records
structured JSON results per the output contract.


