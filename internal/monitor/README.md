# package monitor

`import "github.com/agentflare-ai/amux/internal/monitor"`

Package monitor observes PTY output and detects adapter patterns.

- `ErrMatcherRequired` — ErrMatcherRequired is returned when a matcher is missing.
- `type Monitor` — Monitor scans PTY output with an adapter matcher.

### Variables

#### ErrMatcherRequired

```go
var ErrMatcherRequired = errors.New("matcher required")
```

ErrMatcherRequired is returned when a matcher is missing.


## type Monitor

```go
type Monitor struct {
	matcher adapter.PatternMatcher
}
```

Monitor scans PTY output with an adapter matcher.

### Functions returning Monitor

#### NewMonitor

```go
func NewMonitor(matcher adapter.PatternMatcher) (*Monitor, error)
```

NewMonitor constructs a monitor with the provided matcher.


### Methods

#### Monitor.Scan

```go
func () Scan(ctx context.Context, r io.Reader) ([]adapter.PatternMatch, error)
```

Scan reads from r and emits pattern matches.


