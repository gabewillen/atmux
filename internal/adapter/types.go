package adapter

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// Manifest describes an adapter.
type Manifest struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	CLI         CLIReq `toml:"cli"`
}

// CLIReq defines CLI version requirements.
type CLIReq struct {
	MinVersion string `toml:"min_version"`
	MaxVersion string `toml:"max_version"` // Optional
}

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

// ParseManifest parses a TOML manifest.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks required fields.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("manifest missing name")
	}
	if m.Version == "" {
		return fmt.Errorf("manifest missing version")
	}
	if m.CLI.MinVersion == "" {
		return fmt.Errorf("manifest missing cli.min_version")
	}
	return nil
}