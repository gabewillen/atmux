# package inference

`import "github.com/agentflare-ai/amux/internal/inference"`

- `func init()`
- `type LiquidgenEngine`
- `type LiquidgenImpl`
- `type LiquidgenRequest`
- `type LiquidgenStream`

### Functions

#### init

```go
func init()
```


## type LiquidgenEngine

```go
type LiquidgenEngine interface {
	// Generate produces a completion; implementations SHOULD stream tokens.
	Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}
```

## type LiquidgenImpl

```go
type LiquidgenImpl struct{}
```

### Methods

#### LiquidgenImpl.Generate

```go
func () Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
```


## type LiquidgenRequest

```go
type LiquidgenRequest struct {
	Model       string // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
	Prompt      string
	MaxTokens   int
	Temperature float64
}
```

## type LiquidgenStream

```go
type LiquidgenStream interface {
	// Next returns the next token chunk; io.EOF indicates end of stream.
	Next() (token string, err error)
	Close() error
}
```

