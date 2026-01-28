# package adapter

`import "github.com/agentflare-ai/amux/internal/adapter"`

Package adapter manages the WASM runtime and adapter loading.

- `type Action` — Action represents an action returned by an adapter.
- `type Matcher` — Matcher is the interface for pattern matching.
- `type Runtime` — Runtime manages adapter instances.

## type Action

```go
type Action struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}
```

Action represents an action returned by an adapter.

## type Matcher

```go
type Matcher interface {
	// Match returns actions for the given input.
	Match(input []byte) ([]Action, error)
}
```

Matcher is the interface for pattern matching.

## type Runtime

```go
type Runtime interface {
	Start() error
	Stop() error
}
```

Runtime manages adapter instances.

