package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agentflare-ai/amux/internal/paths"
)

// AdapterDefault provides adapter-scoped default configuration in TOML.
type AdapterDefault struct {
	Name   string
	Source string
	Data   []byte
}

// AdapterDefaultsProvider supplies adapter defaults for configuration loading.
type AdapterDefaultsProvider interface {
	AdapterDefaults() ([]AdapterDefault, error)
}

// LoadOptions controls configuration loading.
type LoadOptions struct {
	Resolver        *paths.Resolver
	AdapterDefaults AdapterDefaultsProvider
	Env             map[string]string
	Logger          *log.Logger
	// WatchPollInterval overrides the config file polling interval.
	WatchPollInterval time.Duration
}

// Load reads and merges configuration sources in spec order.
func Load(opts LoadOptions) (Config, error) {
	if opts.Resolver == nil {
		return Config{}, fmt.Errorf("load config: resolver is required")
	}
	logger := opts.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "amux-config ", log.LstdFlags)
	}
	merged := make(map[string]any)
	if err := mergeAdapterDefaults(merged, opts.AdapterDefaults); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	userConfig := opts.Resolver.UserConfigPath()
	if err := mergeFile(merged, userConfig, logger); err != nil {
		return Config{}, err
	}
	adapterNames := adapterNamesFromMap(merged)
	if err := mergeAdapterFiles(merged, adapterNames, opts.Resolver.UserAdapterConfigPath, logger); err != nil {
		return Config{}, err
	}
	projectConfig := opts.Resolver.ProjectConfigPath()
	if err := mergeFile(merged, projectConfig, logger); err != nil {
		return Config{}, err
	}
	adapterNames = adapterNamesFromMap(merged)
	if err := mergeAdapterFiles(merged, adapterNames, opts.Resolver.ProjectAdapterConfigPath, logger); err != nil {
		return Config{}, err
	}
	env := opts.Env
	if env == nil {
		env = EnvMap()
	}
	envOverrides, err := EnvOverrides(env)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	if err := MergeMaps(merged, envOverrides); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	defaults := DefaultConfig(opts.Resolver)
	cfg, err := DecodeConfig(defaults, merged, opts.Resolver)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

// DefaultConfig returns built-in default configuration.
func DefaultConfig(resolver *paths.Resolver) Config {
	path := func(value string) string {
		if resolver == nil {
			return value
		}
		return resolver.ExpandHome(value)
	}
	return Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  mustDuration("30s"),
			Stuck: mustDuration("5m"),
		},
		Process: ProcessConfig{
			CaptureMode:      "all",
			StreamBufferSize: mustBytes("1MB"),
			HookMode:         "auto",
			PollInterval:     mustDuration("100ms"),
			HookSocketDir:    "/tmp",
		},
		Git: GitConfig{
			Merge: GitMergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Shutdown: ShutdownConfig{
			DrainTimeout:     mustDuration("30s"),
			CleanupWorktrees: false,
		},
		Events: EventsConfig{
			BatchWindow:    mustDuration("50ms"),
			BatchMaxEvents: 100,
			BatchMaxBytes:  mustBytes("64KB"),
			BatchIdleFlush: mustDuration("10ms"),
			Coalesce: EventsCoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:           mustBytes("10MB"),
			RequestTimeout:       mustDuration("5s"),
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: mustDuration("1s"),
			ReconnectBackoffMax:  mustDuration("30s"),
			NATS: RemoteNATSConfig{
				URL:               "nats://amux-host:7422",
				CredsPath:         path("~/.config/amux/nats.creds"),
				SubjectPrefix:     "amux",
				KVBucket:          "AMUX_KV",
				StreamEvents:      "AMUX_EVENTS",
				StreamPTY:         "AMUX_PTY",
				HeartbeatInterval: mustDuration("5s"),
			},
			Manager: RemoteManagerConfig{
				Enabled: true,
				Model:   "lfm2.5-thinking",
				HostID:  "",
			},
		},
		NATS: NATSConfig{
			Mode:             "embedded",
			Topology:         "hub",
			HubURL:           "nats://amux-host:4222",
			Listen:           "0.0.0.0:4222",
			LeafListen:       "0.0.0.0:7422",
			AdvertiseURL:     "nats://amux-host:4222",
			LeafAdvertiseURL: "nats://amux-host:7422",
			JetStreamDir:     path("~/.amux/nats"),
		},
		Node: NodeConfig{
			Role: "director",
		},
		Daemon: DaemonConfig{
			SocketPath: path("~/.amux/amuxd.sock"),
			Autostart:  true,
		},
		Plugins: PluginsConfig{
			Dir:         path("~/.config/amux/plugins"),
			AllowRemote: true,
		},
		Telemetry: TelemetryConfig{
			Enabled:     false,
			ServiceName: "amux",
			Exporter: TelemetryExporterConfig{
				Protocol: "grpc",
			},
			Traces: TelemetryTracesConfig{
				Enabled: true,
				Sampler: "parentbased_traceidratio",
			},
			Metrics: TelemetryMetricsConfig{
				Enabled:  true,
				Interval: mustDuration("60s"),
			},
			Logs: TelemetryLogsConfig{
				Enabled: true,
				Level:   "info",
			},
		},
		Adapters: make(map[string]AdapterConfig),
	}
}

func mustDuration(raw string) time.Duration {
	value, err := parseDurationValue(raw)
	if err != nil {
		panic(err)
	}
	return value
}

func mustBytes(raw string) ByteSize {
	value, err := ParseByteSize(raw)
	if err != nil {
		panic(err)
	}
	return value
}

func mergeAdapterDefaults(target map[string]any, provider AdapterDefaultsProvider) error {
	if provider == nil {
		return nil
	}
	defaults, err := provider.AdapterDefaults()
	if err != nil {
		return fmt.Errorf("adapter defaults: %w", err)
	}
	for _, def := range defaults {
		if def.Name == "" {
			return fmt.Errorf("adapter defaults: missing name")
		}
		prefix := def.Name
		if strings.TrimSpace(def.Source) != "" {
			prefix = fmt.Sprintf("%s (%s)", def.Name, def.Source)
		}
		parsed, err := ParseTOML(def.Data)
		if err != nil {
			return fmt.Errorf("adapter defaults: %s: %w", prefix, err)
		}
		if err := validateAdapterDefaults(def.Name, parsed); err != nil {
			return fmt.Errorf("adapter defaults: %s: %w", prefix, err)
		}
		if err := MergeMaps(target, parsed); err != nil {
			return fmt.Errorf("adapter defaults: %s: %w", prefix, err)
		}
	}
	return nil
}

func mergeFile(target map[string]any, path string, logger *log.Logger) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("load config: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("load config: %s is a directory", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	parsed, err := ParseTOML(data)
	if err != nil {
		return fmt.Errorf("load config: %s: %w", path, err)
	}
	warnSensitive(logger, path, parsed)
	if err := MergeMaps(target, parsed); err != nil {
		return fmt.Errorf("load config: %s: %w", path, err)
	}
	return nil
}

func mergeAdapterFiles(target map[string]any, adapters []string, pathFn func(string) string, logger *log.Logger) error {
	for _, adapter := range adapters {
		if adapter == "" {
			continue
		}
		if err := mergeFile(target, pathFn(adapter), logger); err != nil {
			return err
		}
	}
	return nil
}

func adapterNamesFromMap(root map[string]any) []string {
	names := make(map[string]struct{})
	if adaptersRaw, ok := root["adapters"]; ok {
		if adapters, ok := adaptersRaw.(map[string]any); ok {
			for name := range adapters {
				names[name] = struct{}{}
			}
		}
	}
	if agentsRaw, ok := root["agents"]; ok {
		if agents, ok := agentsRaw.([]any); ok {
			for _, entry := range agents {
				agent, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if adapterRaw, ok := agent["adapter"]; ok {
					if adapter, ok := adapterRaw.(string); ok {
						names[adapter] = struct{}{}
					}
				}
			}
		}
	}
	list := make([]string, 0, len(names))
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

func validateAdapterDefaults(name string, parsed map[string]any) error {
	if len(parsed) == 0 {
		return nil
	}
	adaptersRaw, ok := parsed["adapters"]
	if !ok {
		return fmt.Errorf("defaults must be under [adapters.%s]", name)
	}
	adapters, ok := adaptersRaw.(map[string]any)
	if !ok {
		return fmt.Errorf("defaults.adapters is not a table")
	}
	for key := range parsed {
		if key != "adapters" {
			return fmt.Errorf("defaults must not set %s", key)
		}
	}
	for adapterName := range adapters {
		if adapterName != name {
			return fmt.Errorf("defaults for %s must not set %s", name, adapterName)
		}
	}
	return nil
}

// MergeMaps overlays override onto base recursively.
func MergeMaps(base map[string]any, override map[string]any) error {
	for key, value := range override {
		if existing, ok := base[key]; ok {
			left, okLeft := existing.(map[string]any)
			right, okRight := value.(map[string]any)
			if okLeft && okRight {
				if err := MergeMaps(left, right); err != nil {
					return err
				}
				continue
			}
		}
		base[key] = value
	}
	return nil
}

func warnSensitive(logger *log.Logger, path string, parsed map[string]any) {
	keys := FindSensitiveKeys(parsed, "")
	if len(keys) == 0 {
		return
	}
	logger.Printf("sensitive keys found in %s: %s", path, strings.Join(keys, ", "))
}

// FindSensitiveKeys returns dot-paths for keys that look sensitive.
func FindSensitiveKeys(node map[string]any, prefix string) []string {
	var keys []string
	for key, value := range node {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if isSensitiveKey(key) {
			keys = append(keys, path)
		}
		child, ok := value.(map[string]any)
		if ok {
			keys = append(keys, FindSensitiveKeys(child, path)...)
		}
	}
	sort.Strings(keys)
	return keys
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "secret") || strings.Contains(lower, "token") ||
		strings.Contains(lower, "password") || strings.Contains(lower, "api_key") ||
		strings.Contains(lower, "apikey") || strings.Contains(lower, "access_key") ||
		strings.Contains(lower, "creds") || strings.Contains(lower, "credential")
}

// RedactSensitive returns a copy of the map with sensitive values replaced.
func RedactSensitive(node map[string]any) map[string]any {
	clone := make(map[string]any)
	for key, value := range node {
		if isSensitiveKey(key) {
			clone[key] = "[redacted]"
			continue
		}
		child, ok := value.(map[string]any)
		if ok {
			clone[key] = RedactSensitive(child)
			continue
		}
		clone[key] = value
	}
	return clone
}

// ResolveConfigPath ensures config paths exist and are within repo when needed.
func ResolveConfigPath(root string, path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	resolved := filepath.Join(root, path)
	return resolved, nil
}
