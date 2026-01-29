package config

import (
	"fmt"
	"os"
	"strings"
)

// EnvOverridePrefix is the required environment variable prefix.
const EnvOverridePrefix = "AMUX__"

// EnvOverrides converts environment variables into a TOML-like map overlay.
func EnvOverrides(env map[string]string) (map[string]any, error) {
	result := make(map[string]any)
	for key, value := range env {
		if !strings.HasPrefix(key, EnvOverridePrefix) {
			continue
		}
		path := strings.Split(strings.TrimPrefix(key, EnvOverridePrefix), "__")
		if len(path) == 0 {
			continue
		}
		for i := range path {
			path[i] = strings.ToLower(path[i])
		}
		if len(path) > 1 && path[0] == "adapters" {
			path[1] = strings.ReplaceAll(path[1], "_", "-")
		}
		parsed, err := parseEnvValue(value)
		if err != nil {
			return nil, fmt.Errorf("parse env %s: %w", key, err)
		}
		if err := setPath(result, path, parsed); err != nil {
			return nil, fmt.Errorf("parse env %s: %w", key, err)
		}
	}
	return result, nil
}

// EnvMap returns a map of the current process environment.
func EnvMap() map[string]string {
	items := make(map[string]string)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		items[parts[0]] = parts[1]
	}
	return items
}

func parseEnvValue(raw string) (any, error) {
	doc := "v = " + raw
	parsed, err := ParseTOML([]byte(doc))
	if err == nil {
		if value, ok := parsed["v"]; ok {
			return value, nil
		}
	}
	return raw, nil
}
