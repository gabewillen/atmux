# package monitor

`import "github.com/agentflare-ai/amux/internal/monitor"`

Package monitor provides PTY output monitoring for amux.

The monitor observes PTY output to detect activity, inactivity, and
state changes. Pattern matching is delegated to adapters via the
PatternMatcher interface.

See spec §7 for PTY monitoring requirements.

- `type Config` — Config holds monitor configuration.
- `type Monitor` — Monitor observes PTY output and emits events.

## type Config

```go
type Config struct {
	// AgentID is the agent being monitored.
	AgentID muid.MUID

	// Reader is the PTY output reader.
	Reader io.Reader

	// Matcher is the pattern matcher (may be noop).
	Matcher adapter.PatternMatcher

	// Dispatcher is the event dispatcher.
	Dispatcher event.Dispatcher

	// IdleTimeout is the idle detection timeout.
	IdleTimeout time.Duration

	// StuckTimeout is the stuck detection timeout.
	StuckTimeout time.Duration
}
```

Config holds monitor configuration.

## type Monitor

```go
type Monitor struct {
	mu           sync.Mutex
	agentID      muid.MUID
	reader       io.Reader
	matcher      adapter.PatternMatcher
	dispatcher   event.Dispatcher
	idleTimeout  time.Duration
	stuckTimeout time.Duration
	running      bool
	cancel       context.CancelFunc
}
```

Monitor observes PTY output and emits events.

### Functions returning Monitor

#### New

```go
func New(cfg Config) *Monitor
```

New creates a new monitor.


### Methods

#### Monitor.Start

```go
func () Start(ctx context.Context) error
```

Start begins monitoring.

#### Monitor.Stop

```go
func () Stop()
```

Stop stops monitoring.

#### Monitor.run

```go
func () run(ctx context.Context)
```


