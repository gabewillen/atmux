# package hooks

`import "github.com/agentflare-ai/amux/hooks"`

Package hooks implements the exec hook library used for process tracking.

- `ErrHooksUnavailable` — ErrHooksUnavailable is returned when hooks are not built for the platform.
- `func Init() error` — Init initializes the exec hook library.
- `type FD` — FD is a placeholder file descriptor wrapper.
- `type MessageType` — MessageType identifies a hook message type.
- `type Message` — Message is a placeholder hook protocol envelope.

### Variables

#### ErrHooksUnavailable

```go
var ErrHooksUnavailable = errors.New("hooks unavailable")
```

ErrHooksUnavailable is returned when hooks are not built for the platform.


### Functions

#### Init

```go
func Init() error
```

Init initializes the exec hook library.


## type FD

```go
type FD struct {
	Value int
}
```

FD is a placeholder file descriptor wrapper.

## type Message

```go
type Message struct {
	Type    MessageType
	Payload []byte
}
```

Message is a placeholder hook protocol envelope.

## type MessageType

```go
type MessageType string
```

MessageType identifies a hook message type.

