package config

import (
	"strings"
)

// Redacted returns a copy of the configuration with sensitive fields redacted.
func (c Config) Redacted() Config {
	// Start with a shallow copy
	cfgCopy := c

	// Deep copy and redact Adapters
	if c.Adapters != nil {
		cfgCopy.Adapters = make(map[string]map[string]any, len(c.Adapters))
		for adapterName, adapterConfig := range c.Adapters {
			newConfig := make(map[string]any, len(adapterConfig))
			for k, v := range adapterConfig {
				if isSensitiveKey(k) {
					newConfig[k] = "[REDACTED]"
				} else {
					newConfig[k] = v
				}
			}
			cfgCopy.Adapters[adapterName] = newConfig
		}
	}

	// Agents slice deep copy
	if c.Agents != nil {
		cfgCopy.Agents = make([]AgentConfig, len(c.Agents))
		copy(cfgCopy.Agents, c.Agents)
	}

	return cfgCopy
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "key") ||
		strings.Contains(k, "secret") ||
		strings.Contains(k, "token") ||
		strings.Contains(k, "password") ||
		strings.Contains(k, "credential") ||
		strings.Contains(k, "auth")
}
