# package api

`import "github.com/agentflare-ai/amux/pkg/api"`

Package api provides public API types for amux.
All types in this package are agent-agnostic and contain no adapter-specific knowledge.

- `ErrNotFound, ErrInvalidConfig, ErrNotReady` — Sentinel errors for common failure modes.

### Variables

#### ErrNotFound, ErrInvalidConfig, ErrNotReady

```go
var (
	// ErrNotFound indicates a requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidConfig indicates configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrNotReady indicates the system is not ready for the requested operation.
	ErrNotReady = errors.New("system not ready")
)
```

Sentinel errors for common failure modes.


