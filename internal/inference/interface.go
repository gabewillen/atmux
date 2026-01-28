package inference

import "context"

type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}

type LiquidgenRequest struct {
	Model       string   // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}

type LiquidgenStream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}
