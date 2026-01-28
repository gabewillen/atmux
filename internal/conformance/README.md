# package conformance

`import "github.com/agentflare-ai/amux/internal/conformance"`

Package conformance provides the conformance harness for amux.

The conformance harness executes the conformance suite against the amux
implementation and any WASM adapters that claim conformance to the
specification.

See spec §4.3 for the full conformance requirements.

- `func WriteResults(results *Results, path string) error` — WriteResults writes results to the specified path.
- `type AuthFlow` — AuthFlow tests authentication flows.
- `type ControlPlaneFlow` — ControlPlaneFlow tests JSON-RPC control plane flows.
- `type FlowResult` — FlowResult represents the result of a single flow.
- `type Flow` — Flow represents a conformance flow to test.
- `type MenuFlow` — MenuFlow tests menu/TUI navigation flows.
- `type NotificationFlow` — NotificationFlow tests notification/messaging flows.
- `type Options` — Options configures the conformance run.
- `type Results` — Results represents the conformance run results.
- `type StatusFlow` — StatusFlow tests status/presence/lifecycle flows.
- `type Status` — Status represents a flow result status.

### Functions

#### WriteResults

```go
func WriteResults(results *Results, path string) error
```

WriteResults writes results to the specified path.


## type AuthFlow

```go
type AuthFlow struct{}
```

AuthFlow tests authentication flows.

### Methods

#### AuthFlow.Name

```go
func () Name() string
```

Name returns "auth".

#### AuthFlow.Run

```go
func () Run(ctx context.Context, opts Options) error
```

Run executes the auth flow.


## type ControlPlaneFlow

```go
type ControlPlaneFlow struct{}
```

ControlPlaneFlow tests JSON-RPC control plane flows.

### Methods

#### ControlPlaneFlow.Name

```go
func () Name() string
```

Name returns "control_plane".

#### ControlPlaneFlow.Run

```go
func () Run(ctx context.Context, opts Options) error
```

Run executes the control plane flow.


## type Flow

```go
type Flow interface {
	// Name returns the flow name.
	Name() string

	// Run executes the flow.
	Run(ctx context.Context, opts Options) error
}
```

Flow represents a conformance flow to test.

## type FlowResult

```go
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
```

FlowResult represents the result of a single flow.

## type MenuFlow

```go
type MenuFlow struct{}
```

MenuFlow tests menu/TUI navigation flows.

### Methods

#### MenuFlow.Name

```go
func () Name() string
```

Name returns "menu".

#### MenuFlow.Run

```go
func () Run(ctx context.Context, opts Options) error
```

Run executes the menu flow.


## type NotificationFlow

```go
type NotificationFlow struct{}
```

NotificationFlow tests notification/messaging flows.

### Methods

#### NotificationFlow.Name

```go
func () Name() string
```

Name returns "notification".

#### NotificationFlow.Run

```go
func () Run(ctx context.Context, opts Options) error
```

Run executes the notification flow.


## type Options

```go
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
```

Options configures the conformance run.

## type Results

```go
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
```

Results represents the conformance run results.

### Functions returning Results

#### Run

```go
func Run(ctx context.Context, opts Options) (*Results, error)
```

Run executes the conformance suite.


## type Status

```go
type Status string
```

Status represents a flow result status.

### Constants

#### StatusPass, StatusFail, StatusSkip

```go
const (
	// StatusPass indicates the flow passed.
	StatusPass Status = "pass"

	// StatusFail indicates the flow failed.
	StatusFail Status = "fail"

	// StatusSkip indicates the flow was skipped.
	StatusSkip Status = "skip"
)
```


## type StatusFlow

```go
type StatusFlow struct{}
```

StatusFlow tests status/presence/lifecycle flows.

### Methods

#### StatusFlow.Name

```go
func () Name() string
```

Name returns "status".

#### StatusFlow.Run

```go
func () Run(ctx context.Context, opts Options) error
```

Run executes the status flow.


