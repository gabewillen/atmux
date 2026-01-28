package config

// DefaultConfig returns the built-in default configuration.
// Spec §4.2.8.2 (1. Built-in defaults)
func DefaultConfig() Config {
	return Config{
		General: GeneralConfig{
			LogLevel:  "info",
			LogFormat: "text",
		},
		Timeouts: TimeoutsConfig{
			Idle:  "30s",
			Stuck: "5m",
		},
		Process: ProcessConfig{
			CaptureMode:      "all",
			StreamBufferSize: "1MB",
			HookMode:         "auto",
			PollInterval:     "100ms",
			HookSocketDir:    "/tmp",
		},
		Git: GitConfig{
			Merge: MergeConfig{
				Strategy:   "squash",
				AllowDirty: false,
			},
		},
		Events: EventsConfig{
			BatchWindow:    "50ms",
			BatchMaxEvents: 100,
			BatchMaxBytes:  "64KB",
			BatchIdleFlush: "10ms",
			Coalesce: CoalesceConfig{
				IOStreams: true,
				Presence:  true,
				Activity:  true,
			},
		},
		Remote: RemoteConfig{
			Transport:            "nats",
			BufferSize:           "10MB",
			RequestTimeout:       "5s",
			ReconnectMaxAttempts: 10,
			ReconnectBackoffBase: "1s",
			ReconnectBackoffMax:  "30s",
			NATS: RemoteNATSConfig{
				SubjectPrefix:     "amux",
				KVBucket:          "AMUX_KV",
				StreamEvents:      "AMUX_EVENTS",
				StreamPTY:         "AMUX_PTY",
				HeartbeatInterval: "5s",
			},
			Manager: ManagerConfig{
				Enabled: true,
				Model:   "lfm2.5-thinking",
			},
		},
		NATS: NATSConfig{
			Mode:         "embedded",
			Topology:     "hub",
			Listen:       "0.0.0.0:4222",
			JetStreamDir: "~/.amux/nats",
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
		Telemetry: TelemetryConfig{
			Enabled:     true,
			ServiceName: "amux",
			Traces: TelemetrySignal{
				Enabled:    true,
				Sampler:    "parentbased_traceidratio",
				SamplerArg: 0.1,
			},
			Metrics: TelemetrySignal{
				Enabled:  true,
				Interval: "60s",
			},
			Logs: TelemetrySignal{
				Enabled: true,
				Level:   "info",
			},
			Exporter: TelemetryExporter{
				Protocol: "grpc",
			},
		},
		Adapters: make(map[string]AdapterConfig),
	}
}
