// Package config implements configuration management with hierarchy, environment mapping,
// and parsing conventions as specified in the amux specification.
package config

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the main configuration structure
type Config struct {
	// Core settings
	Core CoreConfig `toml:"core" json:"core"`

	// Server settings for daemon
	Server ServerConfig `toml:"server" json:"server"`

	// Logging settings
	Logging LoggingConfig `toml:"logging" json:"logging"`

	// Telemetry settings
	Telemetry TelemetryConfig `toml:"telemetry" json:"telemetry"`

	// Remote settings
	Remote RemoteConfig `toml:"remote" json:"remote"`

	// Adapter-specific configurations (opaque to core)
	Adapters map[string]map[string]interface{} `toml:"adapters" json:"adapters"`
}

// CoreConfig holds core application settings
type CoreConfig struct {
	RepoRoot string `toml:"repo_root" json:"repo_root"`
	Debug    bool   `toml:"debug" json:"debug"`
}

// ServerConfig holds server/daemon settings
type ServerConfig struct {
	SocketPath string        `toml:"socket_path" json:"socket_path"`
	RPCTimeout time.Duration `toml:"rpc_timeout" json:"rpc_timeout"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string `toml:"level" json:"level"`
	Format string `toml:"format" json:"format"`
	File   string `toml:"file" json:"file"`
}

// TelemetryConfig holds OpenTelemetry settings
type TelemetryConfig struct {
	Enabled     bool   `toml:"enabled" json:"enabled"`
	Endpoint    string `toml:"endpoint" json:"endpoint"`
	ServiceName string `toml:"service_name" json:"service_name"`
}

// RemoteConfig holds remote orchestration settings
type RemoteConfig struct {
	Enabled       bool          `toml:"enabled" json:"enabled"`
	Transport     string        `toml:"transport" json:"transport"`               // nats or ssh_yamux
	RequestTimeout time.Duration `toml:"request_timeout" json:"request_timeout"` // Timeout for NATS request-reply control operations
	BufferSize    int64         `toml:"buffer_size" json:"buffer_size"`         // Size of replay buffer in bytes

	// NATS-specific settings
	NATS NATSConfig `toml:"nats" json:"nats"`

	// Manager-specific settings
	Manager ManagerConfig `toml:"manager" json:"manager"`
}

// NATSConfig holds NATS connection and server settings
type NATSConfig struct {
	URL           string `toml:"url" json:"url"`                       // NATS server URL
	CredsPath     string `toml:"creds_path" json:"creds_path"`       // Path to NATS credential file
	SubjectPrefix string `toml:"subject_prefix" json:"subject_prefix"` // Root subject namespace for all amux traffic
	KVBucket      string `toml:"kv_bucket" json:"kv_bucket"`         // JetStream KV bucket for remote state
	StreamEvents  string `toml:"stream_events" json:"stream_events"`   // JetStream stream for EventMessage envelopes
	StreamPTY     string `toml:"stream_pty" json:"stream_pty"`       // JetStream stream for PTY byte chunks
}

// ManagerConfig holds manager-specific settings
type ManagerConfig struct {
	Enabled bool `toml:"enabled" json:"enabled"` // Whether to run local supervisor loop
}

// LoadConfig loads configuration from multiple sources with precedence:
// 1. Built-in defaults
// 2. Config file
// 3. Environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := getDefaultConfig()

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		fileConfig, err := loadConfigFromFile(configPath)
		if err != nil {
			return nil, err
		}
		mergeConfig(&config, fileConfig)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	return &config, nil
}

// getDefaultConfig returns the default configuration values
func getDefaultConfig() Config {
	return Config{
		Core: CoreConfig{
			RepoRoot: ".",
			Debug:    false,
		},
		Server: ServerConfig{
			SocketPath: "~/.amux/amuxd.sock",
			RPCTimeout: 30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			File:   "",
		},
		Telemetry: TelemetryConfig{
			Enabled:     false,
			Endpoint:    "http://localhost:4317",
			ServiceName: "amux",
		},
		Remote: RemoteConfig{
			Enabled:      false,
			Transport:    "nats",
			RequestTimeout: 5 * time.Second,
			BufferSize:   10 * 1024 * 1024, // 10 MB

			NATS: NATSConfig{
				URL:           "nats://localhost:4222",
				CredsPath:     "~/.config/amux/nats.creds",
				SubjectPrefix: "amux",
				KVBucket:      "AMUX_KV",
				StreamEvents:  "AMUX_EVENTS",
				StreamPTY:     "AMUX_PTY",
			},

			Manager: ManagerConfig{
				Enabled: true,
			},
		},
		Adapters: make(map[string]map[string]interface{}),
	}
}

// loadConfigFromFile loads configuration from a TOML file
func loadConfigFromFile(path string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// applyEnvOverrides applies environment variable overrides to the config
func applyEnvOverrides(config *Config) {
	// Core settings
	if repoRoot := getEnvWithPrefix("AMUX_CORE_REPO_ROOT"); repoRoot != "" {
		config.Core.RepoRoot = repoRoot
	}
	if debugStr := getEnvWithPrefix("AMUX_CORE_DEBUG"); debugStr != "" {
		if strings.ToLower(debugStr) == "true" {
			config.Core.Debug = true
		}
	}

	// Server settings
	if socketPath := getEnvWithPrefix("AMUX_SERVER_SOCKET_PATH"); socketPath != "" {
		config.Server.SocketPath = socketPath
	}
	if rpcTimeoutStr := getEnvWithPrefix("AMUX_SERVER_RPC_TIMEOUT"); rpcTimeoutStr != "" {
		if dur, err := time.ParseDuration(rpcTimeoutStr); err == nil {
			config.Server.RPCTimeout = dur
		}
	}

	// Logging settings
	if level := getEnvWithPrefix("AMUX_LOGGING_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	if format := getEnvWithPrefix("AMUX_LOGGING_FORMAT"); format != "" {
		config.Logging.Format = format
	}
	if file := getEnvWithPrefix("AMUX_LOGGING_FILE"); file != "" {
		config.Logging.File = file
	}

	// Telemetry settings
	if enabledStr := getEnvWithPrefix("AMUX_TELEMETRY_ENABLED"); enabledStr != "" {
		if strings.ToLower(enabledStr) == "true" {
			config.Telemetry.Enabled = true
		}
	}
	if endpoint := getEnvWithPrefix("AMUX_TELEMETRY_ENDPOINT"); endpoint != "" {
		config.Telemetry.Endpoint = endpoint
	}
	if serviceName := getEnvWithPrefix("AMUX_TELEMETRY_SERVICE_NAME"); serviceName != "" {
		config.Telemetry.ServiceName = serviceName
	}

	// Remote settings
	if enabledStr := getEnvWithPrefix("AMUX_REMOTE_ENABLED"); enabledStr != "" {
		if strings.ToLower(enabledStr) == "true" {
			config.Remote.Enabled = true
		}
	}
	if transport := getEnvWithPrefix("AMUX_REMOTE_TRANSPORT"); transport != "" {
		config.Remote.Transport = transport
	}
	if requestTimeoutStr := getEnvWithPrefix("AMUX_REMOTE_REQUEST_TIMEOUT"); requestTimeoutStr != "" {
		if dur, err := time.ParseDuration(requestTimeoutStr); err == nil {
			config.Remote.RequestTimeout = dur
		}
	}
	if bufferSizeStr := getEnvWithPrefix("AMUX_REMOTE_BUFFER_SIZE"); bufferSizeStr != "" {
		if size, err := parseBytes(bufferSizeStr); err == nil {
			config.Remote.BufferSize = size
		}
	}

	// Remote.NATS settings
	if natsURL := getEnvWithPrefix("AMUX_REMOTE_NATS_URL"); natsURL != "" {
		config.Remote.NATS.URL = natsURL
	}
	if credsPath := getEnvWithPrefix("AMUX_REMOTE_NATS_CREDS_PATH"); credsPath != "" {
		config.Remote.NATS.CredsPath = credsPath
	}
	if subjectPrefix := getEnvWithPrefix("AMUX_REMOTE_NATS_SUBJECT_PREFIX"); subjectPrefix != "" {
		config.Remote.NATS.SubjectPrefix = subjectPrefix
	}
	if kvBucket := getEnvWithPrefix("AMUX_REMOTE_NATS_KV_BUCKET"); kvBucket != "" {
		config.Remote.NATS.KVBucket = kvBucket
	}
	if streamEvents := getEnvWithPrefix("AMUX_REMOTE_NATS_STREAM_EVENTS"); streamEvents != "" {
		config.Remote.NATS.StreamEvents = streamEvents
	}
	if streamPTY := getEnvWithPrefix("AMUX_REMOTE_NATS_STREAM_PTY"); streamPTY != "" {
		config.Remote.NATS.StreamPTY = streamPTY
	}

	// Remote.Manager settings
	if managerEnabledStr := getEnvWithPrefix("AMUX_REMOTE_MANAGER_ENABLED"); managerEnabledStr != "" {
		if strings.ToLower(managerEnabledStr) == "true" {
			config.Remote.Manager.Enabled = true
		}
	}
}

// getEnvWithPrefix gets an environment variable with the AMUX__ prefix
// The key should already include the full variable name (e.g., "AMUX_CORE_REPO_ROOT")
func getEnvWithPrefix(key string) string {
	return os.Getenv(key)
}

// parseBytes parses a string representation of bytes (e.g., "10MB", "1GB") into an int64
func parseBytes(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty byte string")
	}

	// Handle common suffixes
	var multiplier int64 = 1
	switch {
	case strings.HasSuffix(s, "KB"):
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "TB"):
		multiplier = 1024 * 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}

	value, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, err
	}

	return value * multiplier, nil
}

// mergeConfig merges source config into destination config
func mergeConfig(dst *Config, src *Config) {
	// Core
	if src.Core.RepoRoot != "" {
		dst.Core.RepoRoot = src.Core.RepoRoot
	}
	dst.Core.Debug = dst.Core.Debug || src.Core.Debug

	// Server
	if src.Server.SocketPath != "" {
		dst.Server.SocketPath = src.Server.SocketPath
	}
	if src.Server.RPCTimeout != 0 {
		dst.Server.RPCTimeout = src.Server.RPCTimeout
	}

	// Logging
	if src.Logging.Level != "" {
		dst.Logging.Level = src.Logging.Level
	}
	if src.Logging.Format != "" {
		dst.Logging.Format = src.Logging.Format
	}
	if src.Logging.File != "" {
		dst.Logging.File = src.Logging.File
	}

	// Telemetry
	dst.Telemetry.Enabled = dst.Telemetry.Enabled || src.Telemetry.Enabled
	if src.Telemetry.Endpoint != "" {
		dst.Telemetry.Endpoint = src.Telemetry.Endpoint
	}
	if src.Telemetry.ServiceName != "" {
		dst.Telemetry.ServiceName = src.Telemetry.ServiceName
	}

	// Remote
	dst.Remote.Enabled = dst.Remote.Enabled || src.Remote.Enabled
	if src.Remote.Transport != "" {
		dst.Remote.Transport = src.Remote.Transport
	}
	if src.Remote.RequestTimeout != 0 {
		dst.Remote.RequestTimeout = src.Remote.RequestTimeout
	}
	if src.Remote.BufferSize != 0 {
		dst.Remote.BufferSize = src.Remote.BufferSize
	}

	// Remote.NATS
	if src.Remote.NATS.URL != "" {
		dst.Remote.NATS.URL = src.Remote.NATS.URL
	}
	if src.Remote.NATS.CredsPath != "" {
		dst.Remote.NATS.CredsPath = src.Remote.NATS.CredsPath
	}
	if src.Remote.NATS.SubjectPrefix != "" {
		dst.Remote.NATS.SubjectPrefix = src.Remote.NATS.SubjectPrefix
	}
	if src.Remote.NATS.KVBucket != "" {
		dst.Remote.NATS.KVBucket = src.Remote.NATS.KVBucket
	}
	if src.Remote.NATS.StreamEvents != "" {
		dst.Remote.NATS.StreamEvents = src.Remote.NATS.StreamEvents
	}
	if src.Remote.NATS.StreamPTY != "" {
		dst.Remote.NATS.StreamPTY = src.Remote.NATS.StreamPTY
	}

	// Remote.Manager
	dst.Remote.Manager.Enabled = dst.Remote.Manager.Enabled || src.Remote.Manager.Enabled

	// Merge adapter configs
	for adapter, adapterCfg := range src.Adapters {
		if dst.Adapters[adapter] == nil {
			dst.Adapters[adapter] = make(map[string]interface{})
		}
		for k, v := range adapterCfg {
			dst.Adapters[adapter][k] = v
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Core.RepoRoot == "" {
		return fmt.Errorf("repo_root cannot be empty: %w", errors.New("validation error"))
	}

	// Expand ~ in paths
	c.Server.SocketPath = expandHomeDir(c.Server.SocketPath)

	return nil
}

// expandHomeDir expands the ~ symbol to the user's home directory
func expandHomeDir(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// GetAdapterConfig retrieves the configuration for a specific adapter
func (c *Config) GetAdapterConfig(adapterName string) map[string]interface{} {
	if cfg, exists := c.Adapters[adapterName]; exists {
		return cfg
	}
	return make(map[string]interface{})
}

// RedactSensitiveFields removes sensitive information from the config for logging/debugging
func (c *Config) RedactSensitiveFields() *Config {
	redacted := *c

	// Redact sensitive fields in adapters
	redacted.Adapters = make(map[string]map[string]interface{})
	for adapterName, adapterConfig := range c.Adapters {
		redacted.Adapters[adapterName] = make(map[string]interface{})

		for key, value := range adapterConfig {
			// Check if the key indicates sensitive data
			lowerKey := strings.ToLower(key)
			if isSensitiveField(lowerKey) {
				redacted.Adapters[adapterName][key] = "[REDACTED]"
			} else {
				redacted.Adapters[adapterName][key] = value
			}
		}
	}

	return &redacted
}

// isSensitiveField checks if a field name indicates sensitive data
func isSensitiveField(fieldName string) bool {
	lowerFieldName := strings.ToLower(fieldName)
	sensitivePatterns := []string{
		"password", "secret", "token", "key", "credential", "auth", "api_key",
		"access_token", "refresh_token", "client_secret", "private", "cert", "jwt",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerFieldName, pattern) {
			return true
		}
	}
	return false
}

// ValidateAdapterConfig validates the configuration for a specific adapter
func (c *Config) ValidateAdapterConfig(adapterName string, requiredFields []string) error {
	adapterConfig := c.GetAdapterConfig(adapterName)

	for _, field := range requiredFields {
		if _, exists := adapterConfig[field]; !exists {
			return fmt.Errorf("required field '%s' missing for adapter '%s'", field, adapterName)
		}
	}

	return nil
}

// SecureCompare compares sensitive values in a timing-attack resistant way
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// EncodeSensitiveData base64 encodes sensitive data for secure transmission/storage
func EncodeSensitiveData(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeSensitiveData decodes base64 encoded sensitive data
func DecodeSensitiveData(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}