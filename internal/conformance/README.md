# package conformance

`import "github.com/agentflare-ai/amux/internal/conformance"`

Package conformance provides the conformance harness skeleton.

- `func writeResults(path string, results Results) error`
- `type CLIFixture` — CLIFixture boots a CLI client for conformance runs.
- `type DaemonFixture` — DaemonFixture boots a daemon for conformance runs.
- `type FlowResult` — FlowResult describes a single flow outcome.
- `type NoopFixture` — NoopFixture is a placeholder fixture.
- `type Results` — Results describes a conformance run.
- `type Runner` — Runner executes conformance flows and writes results.

### Functions

#### writeResults

```go
func writeResults(path string, results Results) error
```


## type CLIFixture

```go
type CLIFixture interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

CLIFixture boots a CLI client for conformance runs.

## type DaemonFixture

```go
type DaemonFixture interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

DaemonFixture boots a daemon for conformance runs.

## type FlowResult

```go
type FlowResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
```

FlowResult describes a single flow outcome.

## type NoopFixture

```go
type NoopFixture struct{}
```

NoopFixture is a placeholder fixture.

### Methods

#### NoopFixture.Start

```go
func () Start(ctx context.Context) error
```

Start is a no-op.

#### NoopFixture.Stop

```go
func () Stop(ctx context.Context) error
```

Stop is a no-op.


## type Results

```go
type Results struct {
	RunID       string       `json:"run_id"`
	SpecVersion string       `json:"spec_version"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
	Flows       []FlowResult `json:"flows"`
}
```

Results describes a conformance run.

## type Runner

```go
type Runner struct {
	SpecVersion string
	OutputPath  string
	Daemon      DaemonFixture
	CLI         CLIFixture
	Clock       func() time.Time
}
```

Runner executes conformance flows and writes results.

### Methods

#### Runner.Run

```go
func () Run(ctx context.Context) (Results, error)
```

Run executes the conformance suite and writes structured JSON results.

#### Runner.runFlow

```go
func () runFlow(ctx context.Context) FlowResult
```


