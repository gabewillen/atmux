# package conformance

`import "github.com/copilot-claude-sonnet-4/amux/internal/conformance"`

Package conformance provides the conformance test harness.
This package implements the conformance suite skeleton that boots a daemon
+ CLI client fixture and records structured JSON results.

- `ErrHarnessNotReady, ErrTestFailed, ErrFixtureSetupFailed` — Common sentinel errors for conformance operations.
- `type Suite` — Suite represents a conformance test suite.
- `type TestResult` — TestResult represents a conformance test result.

### Variables

#### ErrHarnessNotReady, ErrTestFailed, ErrFixtureSetupFailed

```go
var (
	// ErrHarnessNotReady indicates the test harness is not ready.
	ErrHarnessNotReady = errors.New("harness not ready")

	// ErrTestFailed indicates a conformance test failed.
	ErrTestFailed = errors.New("test failed")

	// ErrFixtureSetupFailed indicates test fixture setup failed.
	ErrFixtureSetupFailed = errors.New("fixture setup failed")
)
```

Common sentinel errors for conformance operations.


## type Suite

```go
type Suite struct {
	results []TestResult
	ready   bool
}
```

Suite represents a conformance test suite.

### Functions returning Suite

#### NewSuite

```go
func NewSuite() *Suite
```

NewSuite creates a new conformance test suite.


### Methods

#### Suite.Cleanup

```go
func () Cleanup() error
```

Cleanup tears down test fixtures.

#### Suite.GetResults

```go
func () GetResults() ([]byte, error)
```

GetResults returns the test results as JSON.

#### Suite.RunTest

```go
func () RunTest(testName string) error
```

RunTest executes a single conformance test.
Phase 0: Placeholder implementation.

#### Suite.Setup

```go
func () Setup() error
```

Setup initializes the test fixtures (daemon + CLI client).
Phase 0: Placeholder implementation.


## type TestResult

```go
type TestResult struct {
	// TestName is the name of the test.
	TestName string `json:"test_name"`

	// Status is the test status ("pass", "fail", "skip").
	Status string `json:"status"`

	// Duration is how long the test took.
	Duration time.Duration `json:"duration"`

	// Error is the error message if the test failed.
	Error string `json:"error,omitempty"`

	// Metadata contains test-specific metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

TestResult represents a conformance test result.

