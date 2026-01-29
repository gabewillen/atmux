# package monitor

`import "github.com/stateforward/amux/internal/monitor"`

Package monitor implements PTY output monitoring (delegates to adapters)

- `func contains(s, substr string) bool` — Helper function to check if a string contains a substring
- `func find(s, substr string) bool` — Helper function to find a substring
- `type Monitor` — Monitor monitors PTY output and detects various conditions

### Functions

#### contains

```go
func contains(s, substr string) bool
```

Helper function to check if a string contains a substring

#### find

```go
func find(s, substr string) bool
```

Helper function to find a substring


## type Monitor

```go
type Monitor struct {
	ptyFile      *os.File
	outputChan   chan []byte
	stopChan     chan struct{}
	wg           sync.WaitGroup
	interval     time.Duration
	adapterIface adapteriface.Interface
	ctx          context.Context
	cancel       context.CancelFunc
}
```

Monitor monitors PTY output and detects various conditions

### Functions returning Monitor

#### NewMonitor

```go
func NewMonitor(ptyFile *os.File, adapterName string) *Monitor
```

NewMonitor creates a new PTY monitor


### Methods

#### Monitor.DetectActivity

```go
func () DetectActivity(timeout time.Duration) bool
```

DetectActivity detects if there's been recent activity in the PTY

#### Monitor.GetLastOutput

```go
func () GetLastOutput() []byte
```

GetLastOutput returns the last output from the PTY

#### Monitor.Start

```go
func () Start()
```

Start begins monitoring the PTY

#### Monitor.Stop

```go
func () Stop()
```

Stop stops monitoring the PTY

#### Monitor.WaitForPattern

```go
func () WaitForPattern(pattern string, timeout time.Duration) (bool, error)
```

WaitForPattern waits for a specific pattern to appear in the output

#### Monitor.processOutput

```go
func () processOutput()
```

processOutput processes the output from the PTY

#### Monitor.startReading

```go
func () startReading()
```

startReading starts the goroutine that reads from the PTY


