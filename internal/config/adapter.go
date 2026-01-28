package config

import (
	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// AdapterConfig represents adapter-specific configuration with sensitive field handling.
type AdapterConfig struct {
	// Adapter name
	Name string `toml:"name" json:"name"`

	// Adapter version requirements
	Version string `toml:"version" json:"version"`

	// Adapter-specific configuration (may contain sensitive data)
	Config map[string]interface{} `toml:"config" json:"config"`

	// Sensitive field names (to be redacted in logs/output)
	SensitiveFields []string `toml:"sensitive_fields" json:"sensitive_fields"`
}

// GetSensitiveValue returns a sensitive field value or error if field is not found.
func (ac *AdapterConfig) GetSensitiveValue(key string) (interface{}, error) {
	value, exists := ac.Config[key]
	if !exists {
		return nil, amuxerrors.Wrap("getting sensitive value", amuxerrors.ErrInvalidConfig)
	}

	return value, nil
}

// Redact returns a copy of the config with sensitive fields redacted.
func (ac *AdapterConfig) Redact() *AdapterConfig {
	if len(ac.SensitiveFields) == 0 {
		return ac
	}

	redacted := &AdapterConfig{
		Name:            ac.Name,
		Version:         ac.Version,
		Config:          make(map[string]interface{}),
		SensitiveFields: make([]string, len(ac.SensitiveFields)),
	}

	// Copy sensitive fields list
	copy(redacted.SensitiveFields, ac.SensitiveFields)

	// Copy config with redaction
	for k, v := range ac.Config {
		if ac.isSensitiveField(k) {
			redacted.Config[k] = "[REDACTED]"
		} else {
			redacted.Config[k] = v
		}
	}

	return redacted
}

// isSensitiveField checks if a field name is in the sensitive fields list.
func (ac *AdapterConfig) isSensitiveField(field string) bool {
	for _, sensitive := range ac.SensitiveFields {
		if field == sensitive {
			return true
		}
	}
	return false
}

// GetAdapterConfigs extracts and validates adapter configurations from the main config.
func (c *Config) GetAdapterConfigs() (map[string]*AdapterConfig, error) {
	if c.Adapters == nil {
		return make(map[string]*AdapterConfig), nil
	}

	result := make(map[string]*AdapterConfig)

	for adapterName, adapterData := range c.Adapters {
		// adapterData should already be map[string]interface{}
		configData := adapterData

		// Extract adapter-specific fields
		adapterConfig := &AdapterConfig{
			Name:   adapterName,
			Config: configData,
		}

		// Extract version if present
		if version, ok := configData["version"].(string); ok {
			adapterConfig.Version = version
			delete(configData, "version")
		}

		// Extract sensitive fields if present
		if sensitive, ok := configData["sensitive_fields"].([]interface{}); ok {
			sensitiveFields := make([]string, len(sensitive))
			for i, field := range sensitive {
				if str, ok := field.(string); ok {
					sensitiveFields[i] = str
				}
			}
			adapterConfig.SensitiveFields = sensitiveFields
			delete(configData, "sensitive_fields")
		}

		result[adapterName] = adapterConfig
	}

	return result, nil
}

// ValidateAdapterConfig validates a specific adapter configuration.
func ValidateAdapterConfig(config *AdapterConfig) error {
	if config.Name == "" {
		return amuxerrors.Wrap("validating adapter config", amuxerrors.ErrInvalidConfig)
	}

	// TODO: implement adapter-specific validation based on adapter name
	return nil
}
