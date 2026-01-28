# package monitor

`import "github.com/stateforward/amux/internal/monitor"`

Package monitor implements PTY output monitoring (delegates to adapters)

- `ErrMonitorFailure` — ErrMonitorFailure is returned when monitoring fails

### Variables

#### ErrMonitorFailure

```go
var ErrMonitorFailure = errors.New("monitor operation failed")
```

ErrMonitorFailure is returned when monitoring fails


