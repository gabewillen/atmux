# package process

`import "github.com/agentflare-ai/amux/internal/process"`

Package process tracks child processes launched within agent sessions.

- `type Tracker` — Tracker observes process lifecycle events.

## type Tracker

```go
type Tracker struct{}
```

Tracker observes process lifecycle events.

### Methods

#### Tracker.Start

```go
func () Start(ctx context.Context, agentID muid.MUID) error
```

Start begins tracking processes for the given agent.


