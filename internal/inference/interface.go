package inference

import (
	"context"
)

// LiquidgenEngine defines the interface for local model inference.
// Spec §4.2.10
type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}

// LiquidgenRequest represents an inference request.
type LiquidgenRequest struct {
	Model       string // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}

// LiquidgenStream provides access to streamed response tokens.
type LiquidgenStream interface {
	// Next returns the next token or an error.
	// Returns io.EOF when stream is complete.
	Next() (string, error)
	// Close releases resources associated with the stream.
	Close() error
}
