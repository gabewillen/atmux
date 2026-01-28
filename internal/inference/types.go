package inference

import "context"

// Engine defines the local inference interface.
type Engine interface {
	// Version returns the engine version or commit identifier.
	Version() string
	// Models returns the available model mappings.
	Models(ctx context.Context) ([]ModelInfo, error)
	// Infer executes a local inference request.
	Infer(ctx context.Context, req Request) (Response, error)
}

// ModelInfo describes a logical model ID and its artifact path.
type ModelInfo struct {
	ID          string
	ArtifactPath string
}

// Request describes a local inference request.
type Request struct {
	ModelID string
	Prompt  string
}

// Response contains the inference output.
type Response struct {
	Output string
}
