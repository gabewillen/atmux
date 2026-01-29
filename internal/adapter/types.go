package adapter

import "encoding/json"

// Manifest represents the adapter configuration.
type Manifest struct {
	Name        string    `toml:"name"`
	Version     string    `toml:"version"`
	Description string    `toml:"description"`
	CLI         CLIConfig `toml:"cli"`
}

// CLIConfig defines CLI version requirements.
type CLIConfig struct {
	MinVersion string `toml:"min_version"`
}

// Action represents an action returned by the adapter.
type Action struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
