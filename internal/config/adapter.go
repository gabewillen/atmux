// Package config provides configuration management for amux.
package config

import (
	"fmt"
	"strings"
)

// AdapterConfig represents adapter-specific configuration.
// Adapter configs are opaque to the core system.
type AdapterConfig map[string]interface{}

// GetAdapterConfig returns the configuration for a specific adapter.
func (c *Config) GetAdapterConfig(adapterName string) AdapterConfig {
	if c.Adapters == nil {
		return make(AdapterConfig)
	}

	// Get adapter config from the map
	if cfg, ok := c.Adapters[adapterName]; ok {
		if adapterCfg, ok := cfg.(map[string]interface{}); ok {
			return AdapterConfig(adapterCfg)
		}
		// If it's not a map, wrap it
		return AdapterConfig{"_raw": cfg}
	}

	return make(AdapterConfig)
}

// RedactSensitiveFields redacts sensitive configuration fields for logging.
// Per spec §4.2.8.6, sensitive values should not appear in logs.
func (c *Config) RedactSensitiveFields() *Config {
	redacted := *c
	redacted.Adapters = make(map[string]interface{})

	// Copy adapter configs with sensitive fields redacted
	for name, cfg := range c.Adapters {
		if adapterCfg, ok := cfg.(map[string]interface{}); ok {
			redactedCfg := make(map[string]interface{})
			for k, v := range adapterCfg {
				if isSensitiveKey(k) {
					redactedCfg[k] = "[REDACTED]"
				} else {
					redactedCfg[k] = v
				}
			}
			redacted.Adapters[name] = redactedCfg
		} else {
			redacted.Adapters[name] = cfg
		}
	}

	return &redacted
}

// isSensitiveKey determines if a configuration key contains sensitive information.
func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	sensitivePatterns := []string{
		"key",
		"secret",
		"token",
		"password",
		"credential",
		"api_key",
		"auth",
		"private",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// ValidateAdapterConfig validates that adapter configuration is properly scoped.
func (l *Loader) ValidateAdapterConfig(adapterName string, cfg map[string]interface{}) error {
	// Per spec §4.2.8.2, adapter configs must be scoped under [adapters.<adapter_name>]
	// This is enforced during loading, but we can validate the structure here
	for key := range cfg {
		// Check if key looks like it belongs to this adapter
		// In practice, the TOML parser should enforce this, but we validate as a safety check
		if strings.HasPrefix(key, "adapters.") && !strings.HasPrefix(key, fmt.Sprintf("adapters.%s.", adapterName)) {
			return fmt.Errorf("adapter config key %q is not scoped to adapter %q", key, adapterName)
		}
	}

	return nil
}
