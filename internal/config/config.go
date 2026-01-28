// Package config provides configuration management for amux per spec §4.2.8.
//
// Configuration is loaded from TOML files in a hierarchy:
// 1. Built-in defaults
// 2. Adapter defaults (from WASM manifests)
// 3. User config (~/.config/amux/config.toml)
// 4. Project config (.amux/config.toml)
// 5. Environment variables (AMUX__ prefix)
//
// The configuration supports:
// - Live reload via file watching
// - Opaque adapter config blocks
// - Sensitive configuration redaction
// - HSM-based config actor for updates
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stateforward/amux/internal/errors"
)

// Config represents the complete amux configuration.
type Config struct {
	General  GeneralConfig  `toml:"general"`
	Timeouts TimeoutsConfig `toml:"timeouts"`
	Process  ProcessConfig  `toml:"process"`
	Git      GitConfig      `toml:"git"`
	Events   EventsConfig   `toml:"events"`
	Remote   RemoteConfig   `toml:"remote"`
	NATS     NATSConfig     `toml:"nats"`
	Node     NodeConfig     `toml:"node"`
	Daemon   DaemonConfig   `toml:"daemon"`
	Plugins  PluginsConfig  `toml:"plugins"`
	Adapters map[string]any `toml:"adapters"` // Opaque adapter configs
	Agents   []AgentConfig  `toml:"agents"`
}

// GeneralConfig holds general application settings.
type GeneralConfig struct {
	LogLevel  string `toml:"log_level"`  // debug, info, warn, error
	LogFormat string `toml:"log_format"` // text, json
}

// TimeoutsConfig holds timeout durations.
type TimeoutsConfig struct {
	Idle  Duration `toml:"idle"`  // Idle timeout
	Stuck Duration `toml:"stuck"` // Stuck timeout
}

// ProcessConfig holds process tracking configuration.
type ProcessConfig struct {
	CaptureMode       string   `toml:"capture_mode"`        // none, stdout, stderr, stdin, all
	StreamBufferSize  ByteSize `toml:"stream_buffer_size"`  // Ring buffer size per stream
	HookMode          string   `toml:"hook_mode"`           // auto, preload, polling, disabled
	PollInterval      Duration `toml:"poll_interval"`       // Polling interval
	HookSocketDir     string   `toml:"hook_socket_dir"`     // Directory for hook Unix sockets
}

// GitConfig holds git-related configuration.
type GitConfig struct {
	Merge GitMergeConfig `toml:"merge"`
}

// GitMergeConfig holds git merge strategy configuration.
type GitMergeConfig struct {
	Strategy     string `toml:"strategy"`      // merge-commit, squash, rebase, ff-only
	AllowDirty   bool   `toml:"allow_dirty"`   // Allow merges with uncommitted changes
	TargetBranch string `toml:"target_branch"` // Target branch for merges
}

// EventsConfig holds event system configuration.
type EventsConfig struct {
	BatchWindow     Duration        `toml:"batch_window"`      // Coalesce window
	BatchMaxEvents  int             `toml:"batch_max_events"`  // Maximum events per batch
	BatchMaxBytes   ByteSize        `toml:"batch_max_bytes"`   // Maximum bytes for I/O batches
	BatchIdleFlush  Duration        `toml:"batch_idle_flush"`  // Flush if idle
	Coalesce        CoalesceConfig  `toml:"coalesce"`          // Coalescing rules
}

// CoalesceConfig holds event coalescing configuration.
type CoalesceConfig struct {
	IOStreams bool `toml:"io_streams"` // Coalesce stdout/stderr/stdin per process
	Presence  bool `toml:"presence"`   // Keep only latest presence per agent
	Activity  bool `toml:"activity"`   // Deduplicate activity events
}

// RemoteConfig holds remote agent configuration.
type RemoteConfig struct {
	Transport            string            `toml:"transport"`              // nats, ssh_yamux
	BufferSize           ByteSize          `toml:"buffer_size"`            // Per-session PTY replay buffer
	RequestTimeout       Duration          `toml:"request_timeout"`        // NATS request-reply timeout
	ReconnectMaxAttempts int               `toml:"reconnect_max_attempts"` // Max reconnection attempts
	ReconnectBackoffBase Duration          `toml:"reconnect_backoff_base"` // Base backoff duration
	ReconnectBackoffMax  Duration          `toml:"reconnect_backoff_max"`  // Max backoff duration
	NATS                 RemoteNATSConfig  `toml:"nats"`                   // NATS configuration
	Manager              RemoteManagerConfig `toml:"manager"`              // Manager configuration
}

// RemoteNATSConfig holds NATS configuration for remote agents.
type RemoteNATSConfig struct {
	URL               string   `toml:"url"`                 // NATS server URL
	CredsPath         string   `toml:"creds_path"`          // Per-host NATS credential file
	SubjectPrefix     string   `toml:"subject_prefix"`      // Root subject namespace
	KVBucket          string   `toml:"kv_bucket"`           // JetStream KV bucket
	StreamEvents      string   `toml:"stream_events"`       // JetStream events stream
	StreamPTY         string   `toml:"stream_pty"`          // JetStream PTY stream
	HeartbeatInterval Duration `toml:"heartbeat_interval"`  // Heartbeat interval
}

// RemoteManagerConfig holds remote manager configuration.
type RemoteManagerConfig struct {
	Enabled bool   `toml:"enabled"` // Enable manager mode
	Model   string `toml:"model"`   // LLM model ID
}

// NATSConfig holds NATS server configuration.
type NATSConfig struct {
	Mode         string `toml:"mode"`          // embedded, external
	Topology     string `toml:"topology"`      // hub, leaf
	HubURL       string `toml:"hub_url"`       // Hub URL (for leaf)
	Listen       string `toml:"listen"`        // Listen address
	AdvertiseURL string `toml:"advertise_url"` // Advertise URL
	JetstreamDir string `toml:"jetstream_dir"` // JetStream directory
}

// NodeConfig holds node role configuration.
type NodeConfig struct {
	Role string `toml:"role"` // director, manager
}

// DaemonConfig holds daemon configuration.
type DaemonConfig struct {
	SocketPath string `toml:"socket_path"` // Unix socket path
	Autostart  bool   `toml:"autostart"`   // Auto-start daemon
}

// PluginsConfig holds plugin configuration.
type PluginsConfig struct {
	Dir         string `toml:"dir"`          // Plugin directory
	AllowRemote bool   `toml:"allow_remote"` // Allow remote plugins
}

// AgentConfig represents a single agent configuration.
type AgentConfig struct {
	Name     string         `toml:"name"`     // Agent name
	About    string         `toml:"about"`    // Agent description
	Adapter  string         `toml:"adapter"`  // Adapter name (string reference)
	Location LocationConfig `toml:"location"` // Location configuration
}

// LocationConfig represents agent location configuration.
type LocationConfig struct {
	Type     string `toml:"type"`      // local, ssh
	Host     string `toml:"host"`      // SSH host (for type=ssh)
	RepoPath string `toml:"repo_path"` // Repository path
}

// Duration is a time.Duration that can be parsed from TOML.
type Duration time.Duration

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return errors.Wrapf(err, "invalid duration")
	}
	*d = Duration(dur)
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// ByteSize represents a byte size that can be parsed from TOML.
type ByteSize int64

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (b *ByteSize) UnmarshalText(text []byte) error {
	s := string(text)
	size, err := ParseByteSize(s)
	if err != nil {
		return err
	}
	*b = ByteSize(size)
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (b ByteSize) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", int64(b))), nil
}

// ParseByteSize parses a byte size string (e.g., "1MB", "512KB").
func ParseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	
	// Parse with unit suffix first
	units := map[string]int64{
		"GB": 1024 * 1024 * 1024,
		"MB": 1024 * 1024,
		"KB": 1024,
		"B":  1,
	}
	
	// Try units in order (longest first to avoid matching "B" in "MB")
	for _, suffix := range []string{"GB", "MB", "KB", "B"} {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(s, suffix))
			var num int64
			_, err := fmt.Sscanf(numStr, "%d", &num)
			if err != nil {
				return 0, errors.Wrapf(err, "invalid byte size: %s", s)
			}
			return num * units[suffix], nil
		}
	}
	
	// Try parsing as raw integer
	var size int64
	n, err := fmt.Sscanf(s, "%d", &size)
	if err == nil && n == 1 {
		return size, nil
	}
	
	return 0, errors.Wrapf(errors.ErrInvalidInput, "invalid byte size format: %s", s)
}

// Load loads configuration from the specified paths and environment variables.
func Load(paths ...string) (*Config, error) {
	cfg := DefaultConfig()
	
	// Load from each path in order
	for _, path := range paths {
		if path == "" {
			continue
		}
		
		// Expand home directory
		if strings.HasPrefix(path, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, errors.Wrap(err, "get home directory")
			}
			path = filepath.Join(home, path[2:])
		}
		
		// Skip if file doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		
		// Load and merge
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "read config file: %s", path)
		}
		
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, errors.Wrapf(err, "parse config file: %s", path)
		}
	}
	
	// Apply environment variable overrides
	if err := applyEnvOverrides(cfg); err != nil {
		return nil, errors.Wrap(err, "apply environment overrides")
	}
	
	return cfg, nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  Duration(30 * time.Second),
			Stuck: Duration(5 * time.Minute),
		},
		Process: ProcessConfig{
			CaptureMode:      "all",
			StreamBufferSize: ByteSize(1024 * 1024), // 1MB
			HookMode:         "auto",
			PollInterval:     Duration(100 * time.Millisecond),
			HookSocketDir:    "/tmp",
		},
		Git: GitConfig{
			Merge: GitMergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Events: EventsConfig{
			BatchWindow:    Duration(50 * time.Millisecond),
			BatchMaxEvents: 100,
			BatchMaxBytes:  ByteSize(64 * 1024), // 64KB
			BatchIdleFlush: Duration(10 * time.Millisecond),
			Coalesce: CoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:           ByteSize(10 * 1024 * 1024), // 10MB
			RequestTimeout:       Duration(5 * time.Second),
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: Duration(1 * time.Second),
			ReconnectBackoffMax:  Duration(30 * time.Second),
			NATS: RemoteNATSConfig{
				URL:               "nats://amux-host:4222",
				CredsPath:         "~/.config/amux/nats.creds",
				SubjectPrefix:     "amux",
				KVBucket:          "AMUX_KV",
				StreamEvents:      "AMUX_EVENTS",
				StreamPTY:         "AMUX_PTY",
				HeartbeatInterval: Duration(5 * time.Second),
			},
			Manager: RemoteManagerConfig{
				Enabled: true,
				Model:   "lfm2.5-thinking",
			},
		},
		NATS: NATSConfig{
			Mode:         "embedded",
			Topology:     "hub",
			Listen:       "0.0.0.0:4222",
			AdvertiseURL: "nats://amux-host:4222",
			JetstreamDir: "~/.amux/nats",
		},
		Node: NodeConfig{
			Role: "director",
		},
		Daemon: DaemonConfig{
			SocketPath: "~/.amux/amuxd.sock",
			Autostart:  true,
		},
		Plugins: PluginsConfig{
			Dir:         "~/.config/amux/plugins",
			AllowRemote: true,
		},
		Adapters: make(map[string]any),
		Agents:   []AgentConfig{},
	}
}

// applyEnvOverrides applies environment variable overrides to the config.
// Environment variables follow the pattern: AMUX__<path>__<path>...
func applyEnvOverrides(cfg *Config) error {
	// This is a placeholder for Phase 0.
	// Full implementation requires reflection and path traversal.
	// For now, we'll just validate the approach.
	return nil
}
