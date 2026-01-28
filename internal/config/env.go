package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/agentflare-ai/amux/internal/errors"
)

// loadEnv loads environment variables starting with AMUX__ and overrides configuration.
// Spec §4.2.8.3
func loadEnv(cfg *Config) error {
	environ := os.Environ()
	envMap := make(map[string]any)

	for _, result := range environ {
		parts := strings.SplitN(result, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		if !strings.HasPrefix(key, "AMUX__") {
			continue
		}

		// Parse key segments
		// AMUX__GENERAL__LOG_LEVEL -> [general, log_level]
		// AMUX__ADAPTERS__CLAUDE_CODE__CLI__CONSTRAINT -> [adapters, claude-code, cli, constraint]

		trimmed := strings.TrimPrefix(key, "AMUX__")
		segments := strings.Split(trimmed, "__")

		// Normalize segments
		path := make([]string, len(segments))
		for i, seg := range segments {
			lower := strings.ToLower(seg)
			if i == 1 && strings.ToLower(segments[0]) == "adapters" {
				// Special handling for adapter name: CLAUDE_CODE -> claude-code
				lower = strings.ReplaceAll(lower, "_", "-")
			}
			path[i] = lower
		}

		// Parse value (try TOML, else string)
		val := parseEnvValue(value)

		// Set in nested map
		if err := setPath(envMap, path, val); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to set env var %s", key))
		}
	}

	if len(envMap) == 0 {
		return nil
	}

	// Apply overrides by round-tripping through TOML
	// This merges the env map into the existing config struct
	// Wait, go-toml Unmarshal replaces usually.
	// But if we marshal the *existing* config to a map, merge, and unmarshal back?
	// Or just unmarshal the env TOML *over* the struct?
	// go-toml v2 Unmarshal over an existing struct *should* merge if we are careful,
	// but often it resets fields not present if we unmarshal into a zero struct.
	// However, if we unmarshal into the *existing* pointer `cfg`,
	// go-toml v2 behaves like json.Unmarshal: it updates fields that are present in the input.

	// Encode envMap to TOML
	envBytes, err := toml.Marshal(envMap)
	if err != nil {
		return errors.Wrap(err, "failed to marshal env overrides")
	}

	// Decode over existing cfg
	if err := toml.Unmarshal(envBytes, cfg); err != nil {
		return errors.Wrap(err, "failed to apply env overrides")
	}

	return nil
}

func parseEnvValue(s string) any {
	// embed in TOML and parse
	doc := fmt.Sprintf("v = %s", s)
	var dest struct {
		V any `toml:"v"`
	}
	if err := toml.Unmarshal([]byte(doc), &dest); err == nil {
		return dest.V
	}
	// Fallback to string
	return s
}

func setPath(m map[string]any, path []string, value any) error {
	if len(path) == 0 {
		return nil
	}
	last := len(path) - 1
	current := m

	for i := 0; i < last; i++ {
		key := path[i]
		existing, ok := current[key]
		if !ok {
			next := make(map[string]any)
			current[key] = next
			current = next
		} else {
			next, ok := existing.(map[string]any)
			if !ok {
				return errors.New("conflict in env path structure")
			}
			current = next
		}
	}

	current[path[last]] = value
	return nil
}
