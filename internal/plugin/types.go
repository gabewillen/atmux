package plugin

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// Manifest describes a CLI plugin.
type Manifest struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Description string   `toml:"description"`
	Permissions []string `toml:"permissions"`
	Entrypoint  string   `toml:"entrypoint"` // Path to WASM or executable
}

// ParseManifest parses a plugin manifest.
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
		return fmt.Errorf("plugin missing name")
	}
	if m.Version == "" {
		return fmt.Errorf("plugin missing version")
	}
	// Permissions are optional (default none)
	// Entrypoint is optional if it's a metadata-only plugin (e.g. alias pack)?
	// Spec says "CLI plugin system (WASM and remote)".
	// Let's require entrypoint for now.
	if m.Entrypoint == "" {
		return fmt.Errorf("plugin missing entrypoint")
	}
	return nil
}
