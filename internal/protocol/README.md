# package protocol

`import "github.com/agentflare-ai/amux/internal/protocol"`

Package protocol defines the remote communication protocol.

- `type Dispatcher` — Dispatcher handles event routing.
- `type Event` — Event represents a system event.

## type Dispatcher

```go
type Dispatcher interface {
	Dispatch(event Event) error
	Subscribe(pattern string) (<-chan Event, func())
}
```

Dispatcher handles event routing.

## type Event

```go
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
	Payload   interface{} `json:"payload"`
}
```

Event represents a system event.

