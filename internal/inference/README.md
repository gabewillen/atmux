# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

- `type LiquidgenEngine` — LiquidgenEngine defines the interface for local model inference.
- `type LiquidgenRequest` — LiquidgenRequest represents an inference request.
- `type LiquidgenStream` — LiquidgenStream provides access to streamed response tokens.

## type LiquidgenEngine

```go
type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}
```

LiquidgenEngine defines the interface for local model inference.
Spec §4.2.10

## type LiquidgenRequest

```go
type LiquidgenRequest struct {
	Model       string // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}
```

LiquidgenRequest represents an inference request.

## type LiquidgenStream

```go
type LiquidgenStream interface {
	// Next returns the next token or an error.
	// Returns io.EOF when stream is complete.
	Next() (string, error)
	// Close releases resources associated with the stream.
	Close() error
}
```

LiquidgenStream provides access to streamed response tokens.

