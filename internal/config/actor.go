package config

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/stateforward/hsm-go"
)

// ConfigChange describes a single configuration value change.
// Per spec §4.2.8.8, Path is a dot-separated key (e.g. "timeouts.stuck").
type ConfigChange struct {
	Path     string
	OldValue any
	NewValue any
}

// Config actor event names used with hsm-go.
const (
	ConfigFileChanged = "config.file_changed"
	ConfigReloaded    = "config.reloaded"
	ConfigUpdated     = "config.updated"
)

// Actor manages configuration state as an HSM actor per spec §4.2.8.7–§4.2.8.9.
//
// It keeps the current Config value and notifies subscribers via ConfigChange
// events when values change as a result of a reload.
type Actor struct {
	hsm.HSM

	mu   sync.RWMutex
	cfg  *Config
	load func() (*Config, error)

	subscribersMu sync.RWMutex
	subscribers   []chan ConfigChange

	hsmStarted bool // tracks whether HSM has been initialized
}

// Model defines the configuration actor state machine. It is intentionally
// minimal for Phase 0: it supports loading and reloading, and emits the
// ConfigReloaded event; file watching is left to the caller, which should
// dispatch ConfigFileChanged as appropriate.
var Model = hsm.Define("config",
	hsm.State("loading",
		hsm.Entry(func(ctx context.Context, a *Actor, e hsm.Event) {
			// Load config and emit per-key changes, but don't dispatch ConfigReloaded
			// (that's handled by the transition).
			_, _ = a.reloadInternal(ctx)
		}),
	),

	hsm.State("ready"),
	hsm.State("reloading",
		hsm.Entry(func(ctx context.Context, a *Actor, e hsm.Event) {
			// Load config and emit per-key changes, but don't dispatch ConfigReloaded
			// (that's handled by the transition).
			_, _ = a.reloadInternal(ctx)
		}),
	),

	hsm.Transition(hsm.On(hsm.Event{Name: ConfigFileChanged}), hsm.Source("ready"), hsm.Target("reloading")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("loading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("reloading"), hsm.Target("ready")),

	hsm.Initial(hsm.Target("loading")),
)

// NewActor constructs and starts a configuration actor using the provided
// loader function, which should load and merge configuration from all sources
// (files + env).
func NewActor(ctx context.Context, loader func() (*Config, error)) (*Actor, error) {
	actor := &Actor{
		load:       loader,
		hsmStarted: true, // Set before hsm.Started so reloadInternal() can dispatch per-key events
	}
	// Start the HSM using the Model; Actor is both the model owner and data.
	// The HSM will enter "loading" state and call reloadInternal() via entry action,
	// then auto-transition to "ready" via ConfigReloaded.
	actor = hsm.Started(ctx, actor, &Model)
	// Manually dispatch ConfigReloaded to transition from "loading" to "ready".
	done := hsm.Dispatch(ctx, actor, hsm.Event{Name: ConfigReloaded})
	<-done
	// Check if initial load succeeded
	if actor.cfg == nil {
		return nil, fmt.Errorf("initial config load failed")
	}
	return actor, nil
}

// Current returns a snapshot of the current configuration.
func (a *Actor) Current() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

// Subscribe returns a channel that will receive ConfigChange events produced
// when the actor reloads configuration and detects changed values. The caller
// is responsible for draining and closing the channel when done.
func (a *Actor) Subscribe() <-chan ConfigChange {
	ch := make(chan ConfigChange, 16)
	a.subscribersMu.Lock()
	a.subscribers = append(a.subscribers, ch)
	a.subscribersMu.Unlock()
	return ch
}

// NotifyFileChanged is a helper for external watchers to inform the actor
// that configuration files have changed. It dispatches ConfigFileChanged
// into the HSM; the actor will reload and emit ConfigReloaded/ConfigUpdated
// as appropriate.
func (a *Actor) NotifyFileChanged(ctx context.Context) {
	done := hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigFileChanged})
	<-done
	// After entering "reloading" and calling reloadInternal(), manually dispatch
	// ConfigReloaded to transition back to "ready".
	done = hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigReloaded})
	<-done
}

// reloadInternal loads configuration via the loader, computes per-key diffs, updates
// the current config, and sends ConfigChange events to subscribers. It does NOT
// dispatch any HSM events (to avoid deadlock when called from HSM entry actions).
func (a *Actor) reloadInternal(ctx context.Context) (*Config, error) {
	newCfg, err := a.load()
	if err != nil {
		// Even on error, keep existing config; caller can observe the error.
		return nil, err
	}

	a.mu.Lock()
	oldCfg := a.cfg
	a.cfg = newCfg
	a.mu.Unlock()

	// Compute per-key diffs when an old config exists.
	changes := diffConfigs(oldCfg, newCfg)

	// Send ConfigChange events to subscribers per spec §4.2.8.8.
	if len(changes) > 0 {
		// Snapshot subscribers under lock for best-effort fanout.
		a.subscribersMu.RLock()
		subs := append([]chan ConfigChange(nil), a.subscribers...)
		a.subscribersMu.RUnlock()

		// Fan out to subscribers; drop if a subscriber is slow.
		for _, change := range changes {
			for _, ch := range subs {
				select {
				case ch <- change:
				default:
				}
			}
		}
	}

	return newCfg, nil
}

// diffConfigs computes a list of ConfigChange entries describing differences
// between the old and new configuration. When old is nil, no changes are
// reported (initial load).
func diffConfigs(oldCfg, newCfg *Config) []ConfigChange {
	if oldCfg == nil || newCfg == nil {
		return nil
	}

	var changes []ConfigChange

	add := func(path string, oldVal, newVal any) {
		if path == "" {
			return
		}
		if reflect.DeepEqual(oldVal, newVal) {
			return
		}
		changes = append(changes, ConfigChange{
			Path:     path,
			OldValue: oldVal,
			NewValue: newVal,
		})
	}

	// General
	add("general.log_level", oldCfg.General.LogLevel, newCfg.General.LogLevel)
	add("general.log_format", oldCfg.General.LogFormat, newCfg.General.LogFormat)

	// Timeouts
	add("timeouts.idle", oldCfg.Timeouts.Idle, newCfg.Timeouts.Idle)
	add("timeouts.stuck", oldCfg.Timeouts.Stuck, newCfg.Timeouts.Stuck)

	// Process
	add("process.capture_mode", oldCfg.Process.CaptureMode, newCfg.Process.CaptureMode)
	add("process.stream_buffer_size", oldCfg.Process.StreamBufferSize, newCfg.Process.StreamBufferSize)
	add("process.hook_mode", oldCfg.Process.HookMode, newCfg.Process.HookMode)
	add("process.poll_interval", oldCfg.Process.PollInterval, newCfg.Process.PollInterval)
	add("process.hook_socket_dir", oldCfg.Process.HookSocketDir, newCfg.Process.HookSocketDir)

	// Git merge
	add("git.merge.strategy", oldCfg.Git.Merge.Strategy, newCfg.Git.Merge.Strategy)
	add("git.merge.allow_dirty", oldCfg.Git.Merge.AllowDirty, newCfg.Git.Merge.AllowDirty)
	add("git.merge.target_branch", oldCfg.Git.Merge.TargetBranch, newCfg.Git.Merge.TargetBranch)

	// Events
	add("events.batch_window", oldCfg.Events.BatchWindow, newCfg.Events.BatchWindow)
	add("events.batch_max_events", oldCfg.Events.BatchMaxEvents, newCfg.Events.BatchMaxEvents)
	add("events.batch_max_bytes", oldCfg.Events.BatchMaxBytes, newCfg.Events.BatchMaxBytes)
	add("events.batch_idle_flush", oldCfg.Events.BatchIdleFlush, newCfg.Events.BatchIdleFlush)
	add("events.coalesce.io_streams", oldCfg.Events.Coalesce.IOStreams, newCfg.Events.Coalesce.IOStreams)
	add("events.coalesce.presence", oldCfg.Events.Coalesce.Presence, newCfg.Events.Coalesce.Presence)
	add("events.coalesce.activity", oldCfg.Events.Coalesce.Activity, newCfg.Events.Coalesce.Activity)

	// Remote
	add("remote.transport", oldCfg.Remote.Transport, newCfg.Remote.Transport)
	add("remote.buffer_size", oldCfg.Remote.BufferSize, newCfg.Remote.BufferSize)
	add("remote.request_timeout", oldCfg.Remote.RequestTimeout, newCfg.Remote.RequestTimeout)
	add("remote.reconnect_max_attempts", oldCfg.Remote.ReconnectMaxAttempts, newCfg.Remote.ReconnectMaxAttempts)
	add("remote.reconnect_backoff_base", oldCfg.Remote.ReconnectBackoffBase, newCfg.Remote.ReconnectBackoffBase)
	add("remote.reconnect_backoff_max", oldCfg.Remote.ReconnectBackoffMax, newCfg.Remote.ReconnectBackoffMax)

	// Remote NATS
	add("remote.nats.url", oldCfg.Remote.NATS.URL, newCfg.Remote.NATS.URL)
	add("remote.nats.creds_path", oldCfg.Remote.NATS.CredsPath, newCfg.Remote.NATS.CredsPath)
	add("remote.nats.subject_prefix", oldCfg.Remote.NATS.SubjectPrefix, newCfg.Remote.NATS.SubjectPrefix)
	add("remote.nats.kv_bucket", oldCfg.Remote.NATS.KVBucket, newCfg.Remote.NATS.KVBucket)
	add("remote.nats.stream_events", oldCfg.Remote.NATS.StreamEvents, newCfg.Remote.NATS.StreamEvents)
	add("remote.nats.stream_pty", oldCfg.Remote.NATS.StreamPTY, newCfg.Remote.NATS.StreamPTY)
	add("remote.nats.heartbeat_interval", oldCfg.Remote.NATS.HeartbeatInterval, newCfg.Remote.NATS.HeartbeatInterval)

	// Remote manager
	add("remote.manager.enabled", oldCfg.Remote.Manager.Enabled, newCfg.Remote.Manager.Enabled)
	add("remote.manager.model", oldCfg.Remote.Manager.Model, newCfg.Remote.Manager.Model)

	// NATS
	add("nats.mode", oldCfg.NATS.Mode, newCfg.NATS.Mode)
	add("nats.topology", oldCfg.NATS.Topology, newCfg.NATS.Topology)
	add("nats.hub_url", oldCfg.NATS.HubURL, newCfg.NATS.HubURL)
	add("nats.listen", oldCfg.NATS.Listen, newCfg.NATS.Listen)
	add("nats.advertise_url", oldCfg.NATS.AdvertiseURL, newCfg.NATS.AdvertiseURL)
	add("nats.jetstream_dir", oldCfg.NATS.JetstreamDir, newCfg.NATS.JetstreamDir)

	// Node
	add("node.role", oldCfg.Node.Role, newCfg.Node.Role)

	// Daemon
	add("daemon.socket_path", oldCfg.Daemon.SocketPath, newCfg.Daemon.SocketPath)
	add("daemon.autostart", oldCfg.Daemon.Autostart, newCfg.Daemon.Autostart)

	// Plugins
	add("plugins.dir", oldCfg.Plugins.Dir, newCfg.Plugins.Dir)
	add("plugins.allow_remote", oldCfg.Plugins.AllowRemote, newCfg.Plugins.AllowRemote)

	// Adapters and agents (treated as opaque blocks).
	add("adapters", oldCfg.Adapters, newCfg.Adapters)
	add("agents", oldCfg.Agents, newCfg.Agents)

	return changes
}
