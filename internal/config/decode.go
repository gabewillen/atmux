package config

import (
	"fmt"
	"sort"

	"github.com/agentflare-ai/amux/internal/paths"
)

// DecodeConfig applies a parsed TOML map onto the defaults.
func DecodeConfig(defaults Config, raw map[string]any, resolver *paths.Resolver) (Config, error) {
	cfg := defaults
	if err := applyGeneral(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyTimeouts(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyProcess(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	if err := applyGit(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyShutdown(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyEvents(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyRemote(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	if err := applyNATS(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	if err := applyNode(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyDaemon(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	if err := applyPlugins(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	if err := applyTelemetry(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyAdapters(&cfg, raw); err != nil {
		return Config{}, err
	}
	if err := applyAgents(&cfg, raw, resolver); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyGeneral(cfg *Config, raw map[string]any) error {
	section, ok := raw["general"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["log_level"]); ok {
		cfg.General.LogLevel = value
	}
	if value, ok := parseString(section["log_format"]); ok {
		cfg.General.LogFormat = value
	}
	return nil
}

func applyTimeouts(cfg *Config, raw map[string]any) error {
	section, ok := raw["timeouts"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := section["idle"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("timeouts.idle: %w", err)
		}
		cfg.Timeouts.Idle = parsed
	}
	if value, ok := section["stuck"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("timeouts.stuck: %w", err)
		}
		cfg.Timeouts.Stuck = parsed
	}
	return nil
}

func applyProcess(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	section, ok := raw["process"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["capture_mode"]); ok {
		cfg.Process.CaptureMode = value
	}
	if value, ok := section["stream_buffer_size"]; ok {
		parsed, err := ParseByteSizeValue(value)
		if err != nil {
			return fmt.Errorf("process.stream_buffer_size: %w", err)
		}
		cfg.Process.StreamBufferSize = parsed
	}
	if value, ok := parseString(section["hook_mode"]); ok {
		cfg.Process.HookMode = value
	}
	if value, ok := section["poll_interval"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("process.poll_interval: %w", err)
		}
		cfg.Process.PollInterval = parsed
	}
	if value, ok := parseString(section["hook_socket_dir"]); ok {
		cfg.Process.HookSocketDir = expandPath(resolver, value)
	}
	return nil
}

func applyGit(cfg *Config, raw map[string]any) error {
	section, ok := raw["git"].(map[string]any)
	if !ok {
		return nil
	}
	merge, ok := section["merge"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(merge["strategy"]); ok {
		cfg.Git.Merge.Strategy = value
	}
	if value, ok := parseBool(merge["allow_dirty"]); ok {
		cfg.Git.Merge.AllowDirty = value
	}
	if value, ok := parseString(merge["target_branch"]); ok {
		cfg.Git.Merge.TargetBranch = value
	}
	return nil
}

func applyShutdown(cfg *Config, raw map[string]any) error {
	section, ok := raw["shutdown"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := section["drain_timeout"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("shutdown.drain_timeout: %w", err)
		}
		cfg.Shutdown.DrainTimeout = parsed
	}
	if value, ok := parseBool(section["cleanup_worktrees"]); ok {
		cfg.Shutdown.CleanupWorktrees = value
	}
	return nil
}

func applyEvents(cfg *Config, raw map[string]any) error {
	section, ok := raw["events"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := section["batch_window"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("events.batch_window: %w", err)
		}
		cfg.Events.BatchWindow = parsed
	}
	if value, ok := parseInt(section["batch_max_events"]); ok {
		cfg.Events.BatchMaxEvents = value
	}
	if value, ok := section["batch_max_bytes"]; ok {
		parsed, err := ParseByteSizeValue(value)
		if err != nil {
			return fmt.Errorf("events.batch_max_bytes: %w", err)
		}
		cfg.Events.BatchMaxBytes = parsed
	}
	if value, ok := section["batch_idle_flush"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("events.batch_idle_flush: %w", err)
		}
		cfg.Events.BatchIdleFlush = parsed
	}
	coalesce, ok := section["coalesce"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseBool(coalesce["io_streams"]); ok {
		cfg.Events.Coalesce.IOStreams = value
	}
	if value, ok := parseBool(coalesce["presence"]); ok {
		cfg.Events.Coalesce.Presence = value
	}
	if value, ok := parseBool(coalesce["activity"]); ok {
		cfg.Events.Coalesce.Activity = value
	}
	return nil
}

func applyRemote(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	section, ok := raw["remote"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["transport"]); ok {
		cfg.Remote.Transport = value
	}
	if value, ok := section["buffer_size"]; ok {
		parsed, err := ParseByteSizeValue(value)
		if err != nil {
			return fmt.Errorf("remote.buffer_size: %w", err)
		}
		cfg.Remote.BufferSize = parsed
	}
	if value, ok := section["request_timeout"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("remote.request_timeout: %w", err)
		}
		cfg.Remote.RequestTimeout = parsed
	}
	if value, ok := parseInt(section["reconnect_max_attempts"]); ok {
		cfg.Remote.ReconnectMaxAttempts = value
	}
	if value, ok := section["reconnect_backoff_base"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("remote.reconnect_backoff_base: %w", err)
		}
		cfg.Remote.ReconnectBackoffBase = parsed
	}
	if value, ok := section["reconnect_backoff_max"]; ok {
		parsed, err := parseDurationValue(value)
		if err != nil {
			return fmt.Errorf("remote.reconnect_backoff_max: %w", err)
		}
		cfg.Remote.ReconnectBackoffMax = parsed
	}
	if natsRaw, ok := section["nats"].(map[string]any); ok {
		if value, ok := parseString(natsRaw["url"]); ok {
			cfg.Remote.NATS.URL = value
		}
		if value, ok := parseString(natsRaw["creds_path"]); ok {
			cfg.Remote.NATS.CredsPath = expandPath(resolver, value)
		}
		if value, ok := parseString(natsRaw["subject_prefix"]); ok {
			cfg.Remote.NATS.SubjectPrefix = value
		}
		if value, ok := parseString(natsRaw["kv_bucket"]); ok {
			cfg.Remote.NATS.KVBucket = value
		}
		if value, ok := parseString(natsRaw["stream_events"]); ok {
			cfg.Remote.NATS.StreamEvents = value
		}
		if value, ok := parseString(natsRaw["stream_pty"]); ok {
			cfg.Remote.NATS.StreamPTY = value
		}
		if value, ok := natsRaw["heartbeat_interval"]; ok {
			parsed, err := parseDurationValue(value)
			if err != nil {
				return fmt.Errorf("remote.nats.heartbeat_interval: %w", err)
			}
			cfg.Remote.NATS.HeartbeatInterval = parsed
		}
	}
	if managerRaw, ok := section["manager"].(map[string]any); ok {
		if value, ok := parseBool(managerRaw["enabled"]); ok {
			cfg.Remote.Manager.Enabled = value
		}
		if value, ok := parseString(managerRaw["model"]); ok {
			cfg.Remote.Manager.Model = value
		}
		if value, ok := parseString(managerRaw["host_id"]); ok {
			cfg.Remote.Manager.HostID = value
		}
	}
	return nil
}

func applyNATS(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	section, ok := raw["nats"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["mode"]); ok {
		cfg.NATS.Mode = value
	}
	if value, ok := parseString(section["topology"]); ok {
		cfg.NATS.Topology = value
	}
	if value, ok := parseString(section["hub_url"]); ok {
		cfg.NATS.HubURL = value
	}
	if value, ok := parseString(section["listen"]); ok {
		cfg.NATS.Listen = value
	}
	if value, ok := parseString(section["leaf_listen"]); ok {
		cfg.NATS.LeafListen = value
	}
	if value, ok := parseString(section["advertise_url"]); ok {
		cfg.NATS.AdvertiseURL = value
	}
	if value, ok := parseString(section["leaf_advertise_url"]); ok {
		cfg.NATS.LeafAdvertiseURL = value
	}
	if value, ok := parseString(section["jetstream_dir"]); ok {
		cfg.NATS.JetStreamDir = expandPath(resolver, value)
	}
	return nil
}

func applyNode(cfg *Config, raw map[string]any) error {
	section, ok := raw["node"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["role"]); ok {
		cfg.Node.Role = value
	}
	return nil
}

func applyDaemon(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	section, ok := raw["daemon"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["socket_path"]); ok {
		cfg.Daemon.SocketPath = expandPath(resolver, value)
	}
	if value, ok := parseBool(section["autostart"]); ok {
		cfg.Daemon.Autostart = value
	}
	return nil
}

func applyPlugins(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	section, ok := raw["plugins"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseString(section["dir"]); ok {
		cfg.Plugins.Dir = expandPath(resolver, value)
	}
	if value, ok := parseBool(section["allow_remote"]); ok {
		cfg.Plugins.AllowRemote = value
	}
	return nil
}

func applyTelemetry(cfg *Config, raw map[string]any) error {
	section, ok := raw["telemetry"].(map[string]any)
	if !ok {
		return nil
	}
	if value, ok := parseBool(section["enabled"]); ok {
		cfg.Telemetry.Enabled = value
	}
	if value, ok := parseString(section["service_name"]); ok {
		cfg.Telemetry.ServiceName = value
	}
	if exporterRaw, ok := section["exporter"].(map[string]any); ok {
		if value, ok := parseString(exporterRaw["endpoint"]); ok {
			cfg.Telemetry.Exporter.Endpoint = value
		}
		if value, ok := parseString(exporterRaw["protocol"]); ok {
			cfg.Telemetry.Exporter.Protocol = value
		}
	}
	if tracesRaw, ok := section["traces"].(map[string]any); ok {
		if value, ok := parseBool(tracesRaw["enabled"]); ok {
			cfg.Telemetry.Traces.Enabled = value
		}
		if value, ok := parseString(tracesRaw["sampler"]); ok {
			cfg.Telemetry.Traces.Sampler = value
		}
		if value, ok := tracesRaw["sampler_arg"]; ok {
			if parsed, ok := value.(float64); ok {
				cfg.Telemetry.Traces.SamplerArg = parsed
			}
		}
	}
	if metricsRaw, ok := section["metrics"].(map[string]any); ok {
		if value, ok := parseBool(metricsRaw["enabled"]); ok {
			cfg.Telemetry.Metrics.Enabled = value
		}
		if value, ok := metricsRaw["interval"]; ok {
			parsed, err := parseDurationValue(value)
			if err != nil {
				return fmt.Errorf("telemetry.metrics.interval: %w", err)
			}
			cfg.Telemetry.Metrics.Interval = parsed
		}
	}
	if logsRaw, ok := section["logs"].(map[string]any); ok {
		if value, ok := parseBool(logsRaw["enabled"]); ok {
			cfg.Telemetry.Logs.Enabled = value
		}
		if value, ok := parseString(logsRaw["level"]); ok {
			cfg.Telemetry.Logs.Level = value
		}
	}
	return nil
}

func applyAdapters(cfg *Config, raw map[string]any) error {
	adaptersRaw, ok := raw["adapters"].(map[string]any)
	if !ok {
		return nil
	}
	cfg.Adapters = make(map[string]AdapterConfig)
	for name, value := range adaptersRaw {
		section, ok := value.(map[string]any)
		if !ok {
			continue
		}
		cfg.Adapters[name] = cloneMap(section)
		if err := validateAdapterConstraint(name, section); err != nil {
			return err
		}
	}
	return nil
}

func applyAgents(cfg *Config, raw map[string]any, resolver *paths.Resolver) error {
	agentsRaw, ok := raw["agents"].([]any)
	if !ok {
		return nil
	}
	cfg.Agents = nil
	for _, entry := range agentsRaw {
		section, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		agent := AgentConfig{}
		if value, ok := parseString(section["name"]); ok {
			agent.Name = value
		}
		if value, ok := parseString(section["about"]); ok {
			agent.About = value
		}
		if value, ok := parseString(section["adapter"]); ok {
			agent.Adapter = value
		}
		if rawList, ok := section["listen_channels"].([]any); ok {
			for _, item := range rawList {
				if value, ok := parseString(item); ok {
					agent.ListenChannels = append(agent.ListenChannels, value)
				}
			}
		}
		if locRaw, ok := section["location"].(map[string]any); ok {
			if value, ok := parseString(locRaw["type"]); ok {
				agent.Location.Type = value
			}
			if value, ok := parseString(locRaw["host"]); ok {
				agent.Location.Host = value
			}
			if value, ok := parseString(locRaw["repo_path"]); ok {
				agent.Location.RepoPath = expandPath(resolver, value)
			}
		}
		cfg.Agents = append(cfg.Agents, agent)
	}
	return nil
}

func cloneMap(source map[string]any) map[string]any {
	clone := make(map[string]any)
	for key, value := range source {
		child, ok := value.(map[string]any)
		if ok {
			clone[key] = cloneMap(child)
			continue
		}
		clone[key] = value
	}
	return clone
}

func validateAdapterConstraint(name string, section map[string]any) error {
	cliRaw, ok := section["cli"].(map[string]any)
	if !ok {
		return nil
	}
	constraint, ok := parseString(cliRaw["constraint"])
	if !ok || constraint == "" {
		return nil
	}
	if err := ValidateSemverConstraint(constraint); err != nil {
		return fmt.Errorf("adapters.%s.cli.constraint: %w", name, err)
	}
	return nil
}

// AdapterNames returns a sorted list of adapter names in config.
func AdapterNames(cfg Config) []string {
	nameMap := make(map[string]struct{})
	for name := range cfg.Adapters {
		nameMap[name] = struct{}{}
	}
	for _, agent := range cfg.Agents {
		if agent.Adapter != "" {
			nameMap[agent.Adapter] = struct{}{}
		}
	}
	list := make([]string, 0, len(nameMap))
	for name := range nameMap {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}
