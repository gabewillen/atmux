package adapter

import (
	"context"
	"errors"
)

// ErrAdapterNotFound is returned when a named adapter cannot be loaded.
var ErrAdapterNotFound = errors.New("adapter not found")

// PatternMatch describes a detected pattern match.
type PatternMatch struct {
	Pattern string
	Text    string
}

// PatternMatcher scans output and returns matches.
type PatternMatcher interface {
	Match(ctx context.Context, output []byte) ([]PatternMatch, error)
}

// AdapterPatterns defines adapter output detection patterns.
type AdapterPatterns struct {
	// Prompt detects readiness for input.
	Prompt string `json:"prompt"`
	// RateLimit detects rate limiting.
	RateLimit string `json:"rate_limit"`
	// Error detects error output.
	Error string `json:"error"`
	// Completion detects task completion.
	Completion string `json:"completion"`
	// Message detects outbound messages.
	Message string `json:"message,omitempty"`
}

// CLIRequirement describes the adapter CLI requirements.
type CLIRequirement struct {
	// Binary is the CLI binary name.
	Binary string `json:"binary"`
	// VersionCmd is the command used to fetch the CLI version.
	VersionCmd string `json:"version_cmd"`
	// VersionRe is the regex to parse the version.
	VersionRe string `json:"version_re"`
	// Constraint is the semantic version constraint.
	Constraint string `json:"constraint"`
}

// AdapterCommands describes commands used to control the agent.
type AdapterCommands struct {
	// Start is the argv used to start the agent.
	Start []string `json:"start"`
	// SendMessage formats an outbound message.
	SendMessage string `json:"send_message"`
}

// Manifest describes adapter capabilities and requirements.
type Manifest struct {
	// Name is the adapter identifier.
	Name string `json:"name"`
	// Version is the adapter version string.
	Version string `json:"version"`
	// Description is a human-readable description.
	Description string `json:"description,omitempty"`
	// CLI defines CLI requirements.
	CLI CLIRequirement `json:"cli"`
	// Patterns defines output patterns.
	Patterns AdapterPatterns `json:"patterns"`
	// Commands defines adapter commands.
	Commands AdapterCommands `json:"commands"`
}

// ActionFormatter converts a high-level action into agent input.
type ActionFormatter interface {
	Format(ctx context.Context, input string) (string, error)
}

// Adapter is the runtime-facing interface to a loaded adapter.
type Adapter interface {
	Name() string
	Manifest() Manifest
	Matcher() PatternMatcher
	Formatter() ActionFormatter
}

// Registry loads adapters by name.
type Registry interface {
	Load(ctx context.Context, name string) (Adapter, error)
}
