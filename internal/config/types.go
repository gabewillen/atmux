package config

import "time"

// Config is the top-level amux configuration.
type Config struct {
	// General controls logging output.
	General GeneralConfig
	// Timeouts controls idle/stuck timeouts.
	Timeouts TimeoutsConfig
	// Process controls process capture behavior.
	Process ProcessConfig
	// Git controls merge behavior.
	Git GitConfig
	// Events configures event batching and coalescing.
	Events EventsConfig
	// Remote controls remote transport configuration.
	Remote RemoteConfig
	// NATS controls embedded NATS settings.
	NATS NATSConfig
	// Node controls daemon role selection.
	Node NodeConfig
	// Daemon configures the local daemon.
	Daemon DaemonConfig
	// Plugins configures plugin discovery.
	Plugins PluginsConfig
	// Telemetry configures OpenTelemetry.
	Telemetry TelemetryConfig
	// Adapters holds adapter-specific configuration.
	Adapters map[string]AdapterConfig
	// Agents declares agent definitions.
	Agents []AgentConfig
}

// GeneralConfig holds logging defaults.
type GeneralConfig struct {
	// LogLevel sets the default logging level.
	LogLevel string
	// LogFormat sets the logging format.
	LogFormat string
}

// TimeoutsConfig holds idle/stuck timeouts.
type TimeoutsConfig struct {
	// Idle is the idle timeout duration.
	Idle time.Duration
	// Stuck is the stuck timeout duration.
	Stuck time.Duration
}

// ProcessConfig controls process tracking and capture.
type ProcessConfig struct {
	// CaptureMode configures which streams to capture.
	CaptureMode string
	// StreamBufferSize is the per-stream ring buffer size.
	StreamBufferSize ByteSize
	// HookMode configures hook behavior.
	HookMode string
	// PollInterval is the polling interval for polling mode.
	PollInterval time.Duration
	// HookSocketDir is the directory for hook sockets.
	HookSocketDir string
}

// GitConfig controls merge behavior.
type GitConfig struct {
	// Merge configures merge strategy behavior.
	Merge GitMergeConfig
}

// GitMergeConfig configures git merge strategy.
type GitMergeConfig struct {
	// Strategy controls the merge strategy.
	Strategy string
	// AllowDirty controls merges with uncommitted changes.
	AllowDirty bool
	// TargetBranch overrides the base branch.
	TargetBranch string
}

// EventsConfig configures event batching/coalescing.
type EventsConfig struct {
	// BatchWindow is the batching window duration.
	BatchWindow time.Duration
	// BatchMaxEvents is the max events per batch.
	BatchMaxEvents int
	// BatchMaxBytes is the max bytes per batch.
	BatchMaxBytes ByteSize
	// BatchIdleFlush forces a flush after idle.
	BatchIdleFlush time.Duration
	// Coalesce toggles event coalescing.
	Coalesce EventsCoalesceConfig
}

// EventsCoalesceConfig configures coalescing options.
type EventsCoalesceConfig struct {
	// IOStreams controls stream coalescing.
	IOStreams bool
	// Presence controls presence coalescing.
	Presence bool
	// Activity controls activity deduplication.
	Activity bool
}

// RemoteConfig configures remote transport settings.
type RemoteConfig struct {
	// Transport selects the transport.
	Transport string
	// BufferSize caps PTY replay buffering.
	BufferSize ByteSize
	// RequestTimeout is the control request timeout.
	RequestTimeout time.Duration
	// ReconnectMaxAttempts limits reconnect attempts.
	ReconnectMaxAttempts int
	// ReconnectBackoffBase is the base backoff.
	ReconnectBackoffBase time.Duration
	// ReconnectBackoffMax is the max backoff.
	ReconnectBackoffMax time.Duration
	// NATS contains NATS-specific settings.
	NATS RemoteNATSConfig
	// Manager contains manager-specific settings.
	Manager RemoteManagerConfig
}

// RemoteNATSConfig configures NATS transport.
type RemoteNATSConfig struct {
	// URL is the NATS server URL.
	URL string
	// CredsPath is the NATS credentials path.
	CredsPath string
	// SubjectPrefix is the root subject namespace.
	SubjectPrefix string
	// KVBucket is the JetStream KV bucket name.
	KVBucket string
	// StreamEvents is the events stream name.
	StreamEvents string
	// StreamPTY is the PTY stream name.
	StreamPTY string
	// HeartbeatInterval is the heartbeat interval.
	HeartbeatInterval time.Duration
}

// RemoteManagerConfig configures manager behavior.
type RemoteManagerConfig struct {
	// Enabled toggles manager behavior.
	Enabled bool
	// Model selects the default model.
	Model string
}

// NATSConfig configures embedded NATS for local deployments.
type NATSConfig struct {
	// Mode selects embedded or external.
	Mode string
	// Topology sets hub or leaf.
	Topology string
	// HubURL is the hub URL when in leaf mode.
	HubURL string
	// Listen is the listen address.
	Listen string
	// AdvertiseURL is the advertised URL.
	AdvertiseURL string
	// JetStreamDir is the JetStream storage directory.
	JetStreamDir string
}

// NodeConfig configures the daemon role.
type NodeConfig struct {
	// Role is the node role.
	Role string
}

// DaemonConfig configures the local daemon.
type DaemonConfig struct {
	// SocketPath is the daemon socket path.
	SocketPath string
	// Autostart toggles autostart.
	Autostart bool
}

// PluginsConfig configures plugin directory and policy.
type PluginsConfig struct {
	// Dir is the plugin directory.
	Dir string
	// AllowRemote allows remote plugins.
	AllowRemote bool
}

// TelemetryConfig configures OpenTelemetry.
type TelemetryConfig struct {
	// Enabled toggles telemetry.
	Enabled bool
	// ServiceName sets the service name.
	ServiceName string
	// Exporter configures OTLP exporter settings.
	Exporter TelemetryExporterConfig
	// Traces configures tracing.
	Traces TelemetryTracesConfig
	// Metrics configures metrics.
	Metrics TelemetryMetricsConfig
	// Logs configures log export.
	Logs TelemetryLogsConfig
}

// TelemetryExporterConfig configures the OTLP exporter.
type TelemetryExporterConfig struct {
	// Endpoint is the OTLP endpoint.
	Endpoint string
	// Protocol is the OTLP protocol.
	Protocol string
}

// TelemetryTracesConfig configures traces.
type TelemetryTracesConfig struct {
	// Enabled toggles tracing.
	Enabled bool
	// Sampler configures the sampler.
	Sampler string
	// SamplerArg configures the sampler argument.
	SamplerArg float64
}

// TelemetryMetricsConfig configures metrics.
type TelemetryMetricsConfig struct {
	// Enabled toggles metrics.
	Enabled bool
	// Interval sets the export interval.
	Interval time.Duration
}

// TelemetryLogsConfig configures logs.
type TelemetryLogsConfig struct {
	// Enabled toggles log export.
	Enabled bool
	// Level sets the log level.
	Level string
}

// AdapterConfig holds adapter-specific configuration.
type AdapterConfig map[string]any

// AgentConfig describes an agent definition.
type AgentConfig struct {
	// Name is the agent name.
	Name string
	// About is the agent description.
	About string
	// Adapter names the adapter.
	Adapter string
	// Location describes the agent location.
	Location AgentLocationConfig
}

// AgentLocationConfig describes the agent location.
type AgentLocationConfig struct {
	// Type is the location type.
	Type string
	// Host is the SSH host when applicable.
	Host string
	// RepoPath is the repository path.
	RepoPath string
}

// ByteSize represents a byte count.
type ByteSize int64

// Bytes returns the byte size as int64.
func (b ByteSize) Bytes() int64 {
	return int64(b)
}
