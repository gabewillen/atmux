# package config

`import "github.com/agentflare-ai/amux/internal/config"`

Package config implements configuration loading, parsing, and live updates.

It enforces the amux configuration hierarchy, TOML parsing, environment
overrides, and value parsing conventions defined by the spec.

- `ConfigFileChanged, ConfigLoaded, ConfigReloaded, ConfigReloadFailed, ConfigUpdated` — Config event names.
- `ConfigModel` — ConfigModel defines the configuration actor state machine.
- `EnvOverridePrefix` — EnvOverridePrefix is the required environment variable prefix.
- `byteSizePattern`
- `func AdapterNames(cfg Config) []string` — AdapterNames returns a sorted list of adapter names in config.
- `func EncodeTOML(data map[string]any) ([]byte, error)` — EncodeTOML encodes a nested map into TOML.
- `func EnvMap() map[string]string` — EnvMap returns a map of the current process environment.
- `func EnvOverrides(env map[string]string) (map[string]any, error)` — EnvOverrides converts environment variables into a TOML-like map overlay.
- `func FindSensitiveKeys(node map[string]any, prefix string) []string` — FindSensitiveKeys returns dot-paths for keys that look sensitive.
- `func LoadConfigFile(path string) (map[string]any, error)` — LoadConfigFile reads and parses a TOML config file.
- `func MergeMaps(base map[string]any, override map[string]any) error` — MergeMaps overlays override onto base recursively.
- `func ParseTOML(data []byte) (map[string]any, error)` — ParseTOML parses a TOML v1.0.0 document into a nested map.
- `func RedactSensitive(node map[string]any) map[string]any` — RedactSensitive returns a copy of the map with sensitive values replaced.
- `func ResolveConfigPath(root string, path string) (string, error)` — ResolveConfigPath ensures config paths exist and are within repo when needed.
- `func ValidateSemverConstraint(expr string) error` — ValidateSemverConstraint validates a conjunction of semver comparisons.
- `func WriteConfigFile(path string, data map[string]any) error` — WriteConfigFile writes a TOML config file.
- `func adapterNamesFromMap(root map[string]any) []string`
- `func adapterPatterns(cfg AdapterConfig) map[string]any`
- `func applyAdapters(cfg *Config, raw map[string]any) error`
- `func applyAgents(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyDaemon(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyEvents(cfg *Config, raw map[string]any) error`
- `func applyGeneral(cfg *Config, raw map[string]any) error`
- `func applyGit(cfg *Config, raw map[string]any) error`
- `func applyNATS(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyNode(cfg *Config, raw map[string]any) error`
- `func applyPlugins(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyProcess(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyRemote(cfg *Config, raw map[string]any, resolver *paths.Resolver) error`
- `func applyShutdown(cfg *Config, raw map[string]any) error`
- `func applyTelemetry(cfg *Config, raw map[string]any) error`
- `func applyTimeouts(cfg *Config, raw map[string]any) error`
- `func cloneMap(source map[string]any) map[string]any`
- `func compare(changes *[]ConfigChange, path string, oldVal any, newVal any)`
- `func expandPath(resolver *paths.Resolver, value string) string`
- `func flattenEntry(prefix string, value any, out map[string]any) error`
- `func formatArray(values []any) (string, error)`
- `func formatValue(value any) (string, error)`
- `func getOrCreateArrayTable(root map[string]any, path []string) ([]any, error)`
- `func getOrCreateTable(root map[string]any, path []string) (map[string]any, error)`
- `func hasTriple(runes []rune, idx int, target rune) bool`
- `func isArrayTable(value any) bool`
- `func isSensitiveKey(key string) bool`
- `func isTable(value any) bool`
- `func mergeAdapterDefaults(target map[string]any, provider AdapterDefaultsProvider) error`
- `func mergeAdapterFiles(target map[string]any, adapters []string, pathFn func(string) string, logger *log.Logger) error`
- `func mergeFile(target map[string]any, path string, logger *log.Logger) error`
- `func modTime(path string) (time.Time, error)`
- `func mustDuration(raw string) time.Duration`
- `func parseBool(value any) (bool, bool)`
- `func parseDateTime(raw string) (time.Time, bool)`
- `func parseDurationValue(value any) (time.Duration, error)`
- `func parseEnvValue(raw string) (any, error)`
- `func parseFloat(raw string) (float64, error)`
- `func parseInt(value any) (int, bool)`
- `func parseInteger(raw string) (int64, error)`
- `func parseKeyPath(raw string) ([]string, error)`
- `func parseString(value any) (string, bool)`
- `func parseStringValue(raw string) (string, error)`
- `func parseTablePath(line string, brackets int) ([]string, error)`
- `func parseValue(raw string) (any, error)`
- `func setInlineKey(root map[string]any, key string, value any) error`
- `func setKey(current map[string]any, key string, value any) error`
- `func setPath(root map[string]any, path []string, value any) error`
- `func splitKeyValue(line string) (string, string, error)`
- `func splitOnDots(raw string) []string`
- `func splitOnEquals(line string) int`
- `func stripComments(line string) string`
- `func uniqueStrings(values []string) []string`
- `func validateAdapterConstraint(name string, section map[string]any) error`
- `func validateAdapterDefaults(name string, parsed map[string]any) error`
- `func warnSensitive(logger *log.Logger, path string, parsed map[string]any)`
- `func writeArrayTable(b *strings.Builder, path []string, entries []any) error`
- `func writeTable(b *strings.Builder, path []string, table map[string]any) error`
- `semverConstraintPattern`
- `type AdapterConfig` — AdapterConfig holds adapter-specific configuration.
- `type AdapterDefault` — AdapterDefault provides adapter-scoped default configuration in TOML.
- `type AdapterDefaultsProvider` — AdapterDefaultsProvider supplies adapter defaults for configuration loading.
- `type AgentConfig` — AgentConfig describes an agent definition.
- `type AgentLocationConfig` — AgentLocationConfig describes the agent location.
- `type ByteSize` — ByteSize represents a byte count.
- `type ConfigActor` — ConfigActor manages live configuration reloading.
- `type ConfigChange` — ConfigChange describes an individual configuration change.
- `type Config` — Config is the top-level amux configuration.
- `type DaemonConfig` — DaemonConfig configures the local daemon.
- `type EventsCoalesceConfig` — EventsCoalesceConfig configures coalescing options.
- `type EventsConfig` — EventsConfig configures event batching/coalescing.
- `type GeneralConfig` — GeneralConfig holds logging defaults.
- `type GitConfig` — GitConfig controls merge behavior.
- `type GitMergeConfig` — GitMergeConfig configures git merge strategy.
- `type LoadOptions` — LoadOptions controls configuration loading.
- `type NATSConfig` — NATSConfig configures embedded NATS for local deployments.
- `type NodeConfig` — NodeConfig configures the daemon role.
- `type PluginsConfig` — PluginsConfig configures plugin directory and policy.
- `type ProcessConfig` — ProcessConfig controls process tracking and capture.
- `type RemoteConfig` — RemoteConfig configures remote transport settings.
- `type RemoteManagerConfig` — RemoteManagerConfig configures manager behavior.
- `type RemoteNATSConfig` — RemoteNATSConfig configures NATS transport.
- `type ShutdownConfig` — ShutdownConfig controls graceful shutdown behavior.
- `type TelemetryConfig` — TelemetryConfig configures OpenTelemetry.
- `type TelemetryExporterConfig` — TelemetryExporterConfig configures the OTLP exporter.
- `type TelemetryLogsConfig` — TelemetryLogsConfig configures logs.
- `type TelemetryMetricsConfig` — TelemetryMetricsConfig configures metrics.
- `type TelemetryTracesConfig` — TelemetryTracesConfig configures traces.
- `type TimeoutsConfig` — TimeoutsConfig holds idle/stuck timeouts.
- `type statementState`
- `type tomlStatement`
- `type valueParser`
- `type watcher`

### Constants

#### ConfigFileChanged, ConfigLoaded, ConfigReloaded, ConfigReloadFailed, ConfigUpdated

```go
const (
	ConfigFileChanged  = "config.file_changed"
	ConfigLoaded       = "config.loaded"
	ConfigReloaded     = "config.reloaded"
	ConfigReloadFailed = "config.reload_failed"
	ConfigUpdated      = "config.updated"
)
```

Config event names.

#### EnvOverridePrefix

```go
const EnvOverridePrefix = "AMUX__"
```

EnvOverridePrefix is the required environment variable prefix.


### Variables

#### ConfigModel

```go
var ConfigModel = hsm.Define(
	"config",
	hsm.State(
		"loading",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.loadAll(ctx)
		}),
	),
	hsm.State(
		"ready",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.startWatching(ctx)
		}),
		hsm.Exit(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.stopWatching()
		}),
	),
	hsm.State(
		"reloading",
		hsm.Entry(func(ctx context.Context, actor *ConfigActor, event hsm.Event) {
			actor.reload(ctx)
		}),
	),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigLoaded}), hsm.Source("loading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigFileChanged}), hsm.Source("ready"), hsm.Target("reloading")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloadFailed}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Initial(hsm.Target("loading")),
)
```

ConfigModel defines the configuration actor state machine.

#### byteSizePattern

```go
var byteSizePattern = regexp.MustCompile(`^(\d+)(B|KB|MB|GB)?$`)
```

#### semverConstraintPattern

```go
var semverConstraintPattern = regexp.MustCompile(`^(=|==|!=|>=|<=|>|<)(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:[-+][0-9A-Za-z.-]+)?$`)
```


### Functions

#### AdapterNames

```go
func AdapterNames(cfg Config) []string
```

AdapterNames returns a sorted list of adapter names in config.

#### EncodeTOML

```go
func EncodeTOML(data map[string]any) ([]byte, error)
```

EncodeTOML encodes a nested map into TOML.

#### EnvMap

```go
func EnvMap() map[string]string
```

EnvMap returns a map of the current process environment.

#### EnvOverrides

```go
func EnvOverrides(env map[string]string) (map[string]any, error)
```

EnvOverrides converts environment variables into a TOML-like map overlay.

#### FindSensitiveKeys

```go
func FindSensitiveKeys(node map[string]any, prefix string) []string
```

FindSensitiveKeys returns dot-paths for keys that look sensitive.

#### LoadConfigFile

```go
func LoadConfigFile(path string) (map[string]any, error)
```

LoadConfigFile reads and parses a TOML config file.

#### MergeMaps

```go
func MergeMaps(base map[string]any, override map[string]any) error
```

MergeMaps overlays override onto base recursively.

#### ParseTOML

```go
func ParseTOML(data []byte) (map[string]any, error)
```

ParseTOML parses a TOML v1.0.0 document into a nested map.

#### RedactSensitive

```go
func RedactSensitive(node map[string]any) map[string]any
```

RedactSensitive returns a copy of the map with sensitive values replaced.

#### ResolveConfigPath

```go
func ResolveConfigPath(root string, path string) (string, error)
```

ResolveConfigPath ensures config paths exist and are within repo when needed.

#### ValidateSemverConstraint

```go
func ValidateSemverConstraint(expr string) error
```

ValidateSemverConstraint validates a conjunction of semver comparisons.

#### WriteConfigFile

```go
func WriteConfigFile(path string, data map[string]any) error
```

WriteConfigFile writes a TOML config file.

#### adapterNamesFromMap

```go
func adapterNamesFromMap(root map[string]any) []string
```

#### adapterPatterns

```go
func adapterPatterns(cfg AdapterConfig) map[string]any
```

#### applyAdapters

```go
func applyAdapters(cfg *Config, raw map[string]any) error
```

#### applyAgents

```go
func applyAgents(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyDaemon

```go
func applyDaemon(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyEvents

```go
func applyEvents(cfg *Config, raw map[string]any) error
```

#### applyGeneral

```go
func applyGeneral(cfg *Config, raw map[string]any) error
```

#### applyGit

```go
func applyGit(cfg *Config, raw map[string]any) error
```

#### applyNATS

```go
func applyNATS(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyNode

```go
func applyNode(cfg *Config, raw map[string]any) error
```

#### applyPlugins

```go
func applyPlugins(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyProcess

```go
func applyProcess(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyRemote

```go
func applyRemote(cfg *Config, raw map[string]any, resolver *paths.Resolver) error
```

#### applyShutdown

```go
func applyShutdown(cfg *Config, raw map[string]any) error
```

#### applyTelemetry

```go
func applyTelemetry(cfg *Config, raw map[string]any) error
```

#### applyTimeouts

```go
func applyTimeouts(cfg *Config, raw map[string]any) error
```

#### cloneMap

```go
func cloneMap(source map[string]any) map[string]any
```

#### compare

```go
func compare(changes *[]ConfigChange, path string, oldVal any, newVal any)
```

#### expandPath

```go
func expandPath(resolver *paths.Resolver, value string) string
```

#### flattenEntry

```go
func flattenEntry(prefix string, value any, out map[string]any) error
```

#### formatArray

```go
func formatArray(values []any) (string, error)
```

#### formatValue

```go
func formatValue(value any) (string, error)
```

#### getOrCreateArrayTable

```go
func getOrCreateArrayTable(root map[string]any, path []string) ([]any, error)
```

#### getOrCreateTable

```go
func getOrCreateTable(root map[string]any, path []string) (map[string]any, error)
```

#### hasTriple

```go
func hasTriple(runes []rune, idx int, target rune) bool
```

#### isArrayTable

```go
func isArrayTable(value any) bool
```

#### isSensitiveKey

```go
func isSensitiveKey(key string) bool
```

#### isTable

```go
func isTable(value any) bool
```

#### mergeAdapterDefaults

```go
func mergeAdapterDefaults(target map[string]any, provider AdapterDefaultsProvider) error
```

#### mergeAdapterFiles

```go
func mergeAdapterFiles(target map[string]any, adapters []string, pathFn func(string) string, logger *log.Logger) error
```

#### mergeFile

```go
func mergeFile(target map[string]any, path string, logger *log.Logger) error
```

#### modTime

```go
func modTime(path string) (time.Time, error)
```

#### mustDuration

```go
func mustDuration(raw string) time.Duration
```

#### parseBool

```go
func parseBool(value any) (bool, bool)
```

#### parseDateTime

```go
func parseDateTime(raw string) (time.Time, bool)
```

#### parseDurationValue

```go
func parseDurationValue(value any) (time.Duration, error)
```

#### parseEnvValue

```go
func parseEnvValue(raw string) (any, error)
```

#### parseFloat

```go
func parseFloat(raw string) (float64, error)
```

#### parseInt

```go
func parseInt(value any) (int, bool)
```

#### parseInteger

```go
func parseInteger(raw string) (int64, error)
```

#### parseKeyPath

```go
func parseKeyPath(raw string) ([]string, error)
```

#### parseString

```go
func parseString(value any) (string, bool)
```

#### parseStringValue

```go
func parseStringValue(raw string) (string, error)
```

#### parseTablePath

```go
func parseTablePath(line string, brackets int) ([]string, error)
```

#### parseValue

```go
func parseValue(raw string) (any, error)
```

#### setInlineKey

```go
func setInlineKey(root map[string]any, key string, value any) error
```

#### setKey

```go
func setKey(current map[string]any, key string, value any) error
```

#### setPath

```go
func setPath(root map[string]any, path []string, value any) error
```

#### splitKeyValue

```go
func splitKeyValue(line string) (string, string, error)
```

#### splitOnDots

```go
func splitOnDots(raw string) []string
```

#### splitOnEquals

```go
func splitOnEquals(line string) int
```

#### stripComments

```go
func stripComments(line string) string
```

#### uniqueStrings

```go
func uniqueStrings(values []string) []string
```

#### validateAdapterConstraint

```go
func validateAdapterConstraint(name string, section map[string]any) error
```

#### validateAdapterDefaults

```go
func validateAdapterDefaults(name string, parsed map[string]any) error
```

#### warnSensitive

```go
func warnSensitive(logger *log.Logger, path string, parsed map[string]any)
```

#### writeArrayTable

```go
func writeArrayTable(b *strings.Builder, path []string, entries []any) error
```

#### writeTable

```go
func writeTable(b *strings.Builder, path []string, table map[string]any) error
```


## type AdapterConfig

```go
type AdapterConfig map[string]any
```

AdapterConfig holds adapter-specific configuration.

## type AdapterDefault

```go
type AdapterDefault struct {
	Name   string
	Source string
	Data   []byte
}
```

AdapterDefault provides adapter-scoped default configuration in TOML.

## type AdapterDefaultsProvider

```go
type AdapterDefaultsProvider interface {
	AdapterDefaults() ([]AdapterDefault, error)
}
```

AdapterDefaultsProvider supplies adapter defaults for configuration loading.

## type AgentConfig

```go
type AgentConfig struct {
	// Name is the agent name.
	Name string
	// About is the agent description.
	About string
	// Adapter names the adapter.
	Adapter string
	// ListenChannels are participant channels to mirror into the agent PTY.
	ListenChannels []string
	// Location describes the agent location.
	Location AgentLocationConfig
}
```

AgentConfig describes an agent definition.

## type AgentLocationConfig

```go
type AgentLocationConfig struct {
	// Type is the location type.
	Type string
	// Host is the SSH host when applicable.
	Host string
	// RepoPath is the repository path.
	RepoPath string
}
```

AgentLocationConfig describes the agent location.

## type ByteSize

```go
type ByteSize int64
```

ByteSize represents a byte count.

### Functions returning ByteSize

#### ParseByteSize

```go
func ParseByteSize(raw string) (ByteSize, error)
```

ParseByteSize parses a byte size string or integer.

#### ParseByteSizeValue

```go
func ParseByteSizeValue(value any) (ByteSize, error)
```

ParseByteSizeValue parses a byte size from an interface value.

#### mustBytes

```go
func mustBytes(raw string) ByteSize
```


### Methods

#### ByteSize.Bytes

```go
func () Bytes() int64
```

Bytes returns the byte size as int64.


## type Config

```go
type Config struct {
	// General controls logging output.
	General GeneralConfig
	// Timeouts controls idle/stuck timeouts.
	Timeouts TimeoutsConfig
	// Process controls process capture behavior.
	Process ProcessConfig
	// Git controls merge behavior.
	Git GitConfig
	// Shutdown controls graceful shutdown behavior.
	Shutdown ShutdownConfig
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
```

Config is the top-level amux configuration.

### Functions returning Config

#### DecodeConfig

```go
func DecodeConfig(defaults Config, raw map[string]any, resolver *paths.Resolver) (Config, error)
```

DecodeConfig applies a parsed TOML map onto the defaults.

#### DefaultConfig

```go
func DefaultConfig(resolver *paths.Resolver) Config
```

DefaultConfig returns built-in default configuration.

#### Load

```go
func Load(opts LoadOptions) (Config, error)
```

Load reads and merges configuration sources in spec order.


## type ConfigActor

```go
type ConfigActor struct {
	hsm.HSM
	opts        LoadOptions
	mu          sync.RWMutex
	current     Config
	watcher     *watcher
	subscribers map[uint64]func(ConfigChange)
	nextSubID   uint64
}
```

ConfigActor manages live configuration reloading.

### Functions returning ConfigActor

#### StartConfigActor

```go
func StartConfigActor(ctx context.Context, opts LoadOptions) (*ConfigActor, error)
```

StartConfigActor constructs and starts the configuration actor.


### Methods

#### ConfigActor.Current

```go
func () Current() Config
```

Current returns the current configuration snapshot.

#### ConfigActor.Subscribe

```go
func () Subscribe(callback func(ConfigChange)) uint64
```

Subscribe registers a callback for configuration updates.

#### ConfigActor.Unsubscribe

```go
func () Unsubscribe(id uint64)
```

Unsubscribe removes a configuration subscription.

#### ConfigActor.dispatchChange

```go
func () dispatchChange(ctx context.Context, change ConfigChange)
```

#### ConfigActor.loadAll

```go
func () loadAll(ctx context.Context)
```

#### ConfigActor.reload

```go
func () reload(ctx context.Context)
```

#### ConfigActor.startWatching

```go
func () startWatching(ctx context.Context)
```

#### ConfigActor.stopWatching

```go
func () stopWatching()
```

#### ConfigActor.watchPaths

```go
func () watchPaths() []string
```


## type ConfigChange

```go
type ConfigChange struct {
	Path     string
	OldValue any
	NewValue any
}
```

ConfigChange describes an individual configuration change.

### Functions returning ConfigChange

#### DiffConfig

```go
func DiffConfig(oldCfg Config, newCfg Config) []ConfigChange
```

DiffConfig returns a list of changes between two configs for reloadable keys.

#### diffAdapterPatterns

```go
func diffAdapterPatterns(oldAdapters map[string]AdapterConfig, newAdapters map[string]AdapterConfig) []ConfigChange
```


## type DaemonConfig

```go
type DaemonConfig struct {
	// SocketPath is the daemon socket path.
	SocketPath string
	// Autostart toggles autostart.
	Autostart bool
}
```

DaemonConfig configures the local daemon.

## type EventsCoalesceConfig

```go
type EventsCoalesceConfig struct {
	// IOStreams controls stream coalescing.
	IOStreams bool
	// Presence controls presence coalescing.
	Presence bool
	// Activity controls activity deduplication.
	Activity bool
}
```

EventsCoalesceConfig configures coalescing options.

## type EventsConfig

```go
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
```

EventsConfig configures event batching/coalescing.

## type GeneralConfig

```go
type GeneralConfig struct {
	// LogLevel sets the default logging level.
	LogLevel string
	// LogFormat sets the logging format.
	LogFormat string
}
```

GeneralConfig holds logging defaults.

## type GitConfig

```go
type GitConfig struct {
	// Merge configures merge strategy behavior.
	Merge GitMergeConfig
}
```

GitConfig controls merge behavior.

## type GitMergeConfig

```go
type GitMergeConfig struct {
	// Strategy controls the merge strategy.
	Strategy string
	// AllowDirty controls merges with uncommitted changes.
	AllowDirty bool
	// TargetBranch overrides the base branch.
	TargetBranch string
}
```

GitMergeConfig configures git merge strategy.

## type LoadOptions

```go
type LoadOptions struct {
	Resolver        *paths.Resolver
	AdapterDefaults AdapterDefaultsProvider
	Env             map[string]string
	Logger          *log.Logger
	// WatchPollInterval overrides the config file polling interval.
	WatchPollInterval time.Duration
}
```

LoadOptions controls configuration loading.

## type NATSConfig

```go
type NATSConfig struct {
	// Mode selects embedded or external.
	Mode string
	// Topology sets hub or leaf.
	Topology string
	// HubURL is the hub URL when in leaf mode.
	HubURL string
	// Listen is the listen address.
	Listen string
	// LeafListen is the leaf listen address.
	LeafListen string
	// AdvertiseURL is the advertised URL.
	AdvertiseURL string
	// LeafAdvertiseURL is the advertised leaf URL.
	LeafAdvertiseURL string
	// JetStreamDir is the JetStream storage directory.
	JetStreamDir string
}
```

NATSConfig configures embedded NATS for local deployments.

## type NodeConfig

```go
type NodeConfig struct {
	// Role is the node role.
	Role string
}
```

NodeConfig configures the daemon role.

## type PluginsConfig

```go
type PluginsConfig struct {
	// Dir is the plugin directory.
	Dir string
	// AllowRemote allows remote plugins.
	AllowRemote bool
}
```

PluginsConfig configures plugin directory and policy.

## type ProcessConfig

```go
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
```

ProcessConfig controls process tracking and capture.

## type RemoteConfig

```go
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
```

RemoteConfig configures remote transport settings.

## type RemoteManagerConfig

```go
type RemoteManagerConfig struct {
	// Enabled toggles manager behavior.
	Enabled bool
	// Model selects the default model.
	Model string
	// HostID sets the stable host_id for manager role.
	HostID string
}
```

RemoteManagerConfig configures manager behavior.

## type RemoteNATSConfig

```go
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
```

RemoteNATSConfig configures NATS transport.

## type ShutdownConfig

```go
type ShutdownConfig struct {
	// DrainTimeout is the duration before forcing termination.
	DrainTimeout time.Duration
	// CleanupWorktrees removes worktrees and branches on shutdown.
	CleanupWorktrees bool
}
```

ShutdownConfig controls graceful shutdown behavior.

## type TelemetryConfig

```go
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
```

TelemetryConfig configures OpenTelemetry.

## type TelemetryExporterConfig

```go
type TelemetryExporterConfig struct {
	// Endpoint is the OTLP endpoint.
	Endpoint string
	// Protocol is the OTLP protocol.
	Protocol string
}
```

TelemetryExporterConfig configures the OTLP exporter.

## type TelemetryLogsConfig

```go
type TelemetryLogsConfig struct {
	// Enabled toggles log export.
	Enabled bool
	// Level sets the log level.
	Level string
}
```

TelemetryLogsConfig configures logs.

## type TelemetryMetricsConfig

```go
type TelemetryMetricsConfig struct {
	// Enabled toggles metrics.
	Enabled bool
	// Interval sets the export interval.
	Interval time.Duration
}
```

TelemetryMetricsConfig configures metrics.

## type TelemetryTracesConfig

```go
type TelemetryTracesConfig struct {
	// Enabled toggles tracing.
	Enabled bool
	// Sampler configures the sampler.
	Sampler string
	// SamplerArg configures the sampler argument.
	SamplerArg float64
}
```

TelemetryTracesConfig configures traces.

## type TimeoutsConfig

```go
type TimeoutsConfig struct {
	// Idle is the idle timeout duration.
	Idle time.Duration
	// Stuck is the stuck timeout duration.
	Stuck time.Duration
}
```

TimeoutsConfig holds idle/stuck timeouts.

## type statementState

```go
type statementState struct {
	inBasic        bool
	inLiteral      bool
	inMultiBasic   bool
	inMultiLiteral bool
	bracketDepth   int
	braceDepth     int
	lastWasEscape  bool
}
```

### Methods

#### statementState.complete

```go
func () complete() bool
```

#### statementState.scan

```go
func () scan(line string)
```


## type tomlStatement

```go
type tomlStatement struct {
	text string
	line int
}
```

### Functions returning tomlStatement

#### splitStatements

```go
func splitStatements(data string) ([]tomlStatement, error)
```


## type valueParser

```go
type valueParser struct {
	data []rune
	pos  int
}
```

### Methods

#### valueParser.more

```go
func () more() bool
```

#### valueParser.next

```go
func () next() rune
```

#### valueParser.parseArray

```go
func () parseArray() ([]any, error)
```

#### valueParser.parseInlineTable

```go
func () parseInlineTable() (map[string]any, error)
```

#### valueParser.parseKey

```go
func () parseKey() (string, error)
```

#### valueParser.parseLiteral

```go
func () parseLiteral() (any, error)
```

#### valueParser.parseMultiLineString

```go
func () parseMultiLineString() (any, error)
```

#### valueParser.parseString

```go
func () parseString() (any, error)
```

#### valueParser.parseValue

```go
func () parseValue() (any, error)
```

#### valueParser.peek

```go
func () peek() rune
```

#### valueParser.readUnicode

```go
func () readUnicode(length int) (rune, error)
```

#### valueParser.skipComment

```go
func () skipComment()
```

#### valueParser.skipSpace

```go
func () skipSpace()
```


## type watcher

```go
type watcher struct {
	paths     []string
	onChange  func()
	pollEvery time.Duration
	cancel    context.CancelFunc
	lastMod   map[string]time.Time
}
```

### Functions returning watcher

#### newWatcher

```go
func newWatcher(paths []string, onChange func(), pollEvery time.Duration) *watcher
```


### Methods

#### watcher.check

```go
func () check() bool
```

#### watcher.start

```go
func () start(ctx context.Context)
```

#### watcher.stop

```go
func () stop()
```


