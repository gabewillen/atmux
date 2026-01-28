package adapter

// Action represents an action returned by an adapter.
type Action struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// Matcher is the interface for pattern matching.
type Matcher interface {
	// Match returns actions for the given input.
	Match(input []byte) ([]Action, error)
}

// Runtime manages adapter instances.
type Runtime interface {
	Start() error
	Stop() error
}
