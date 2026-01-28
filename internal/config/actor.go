// Package config provides the configuration actor and live config updates.
//
// The configuration actor manages configuration loading, file watching,
// and dispatches config change events per spec §4.2.8.7-§4.2.8.9.

package config

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/paths"
)

// ConfigChange represents a change to a configuration value.
// Used in config.updated events per spec §4.2.8.8.
type ConfigChange struct {
	// Path is the config key path (e.g., "timeouts.idle").
	Path string `json:"path"`

	// OldValue is the previous value.
	OldValue any `json:"old_value,omitempty"`

	// NewValue is the new value.
	NewValue any `json:"new_value"`
}

// State represents the config actor state.
type State string

const (
	// StateLoading is the initial state while loading configuration.
	StateLoading State = "loading"

	// StateReady is the state when configuration is loaded and watching for changes.
	StateReady State = "ready"

	// StateReloading is the state while reloading configuration after a file change.
	StateReloading State = "reloading"
)

// Actor is the configuration actor that manages config loading and live updates.
// It implements an HSM-like state machine per spec §4.2.8.7.
type Actor struct {
	mu sync.RWMutex

	// id is the actor's unique identifier.
	id muid.MUID

	// state is the current actor state.
	state State

	// loader is the configuration loader.
	loader *Loader

	// config is the current configuration.
	config *Config

	// dispatcher is used to emit config events.
	dispatcher event.Dispatcher

	// watchCtx controls the file watcher goroutine.
	watchCtx    context.Context
	watchCancel context.CancelFunc

	// watchedFiles is the list of files being watched.
	watchedFiles []string

	// modTimes tracks file modification times.
	modTimes map[string]time.Time

	// pollInterval is how often to check for file changes.
	pollInterval time.Duration

	// closed indicates the actor has been closed.
	closed bool
}

// NewActor creates a new configuration actor.
func NewActor(resolver *paths.Resolver, dispatcher event.Dispatcher) *Actor {
	if dispatcher == nil {
		dispatcher = event.GetDefaultDispatcher()
	}

	return &Actor{
		id:           muid.Make(),
		state:        StateLoading,
		loader:       NewLoader(resolver),
		dispatcher:   dispatcher,
		modTimes:     make(map[string]time.Time),
		pollInterval: 2 * time.Second,
	}
}

// ID returns the actor's unique identifier.
func (a *Actor) ID() muid.MUID {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

// State returns the current state of the actor.
func (a *Actor) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Config returns the current configuration.
func (a *Actor) Config() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// Start loads configuration and transitions to the ready state.
// This begins file watching for live updates.
func (a *Actor) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return ErrActorClosed
	}

	// Transition to loading state
	a.state = StateLoading

	// Load configuration
	config, err := a.loader.Load()
	if err != nil {
		return err
	}
	a.config = config

	// Determine files to watch
	a.watchedFiles = a.collectWatchedFiles()
	a.updateModTimes()

	// Start file watcher
	a.watchCtx, a.watchCancel = context.WithCancel(ctx)
	go a.watchLoop()

	// Transition to ready state
	a.state = StateReady

	// Dispatch config.loaded event (implicit via transition to ready)
	return nil
}

// Close stops the configuration actor and file watcher.
func (a *Actor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}

	a.closed = true
	if a.watchCancel != nil {
		a.watchCancel()
	}

	return nil
}

// Reload forces a configuration reload.
func (a *Actor) Reload(ctx context.Context) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return ErrActorClosed
	}

	oldConfig := a.config
	a.state = StateReloading
	a.mu.Unlock()

	// Load new configuration
	newConfig, err := a.loader.Load()
	if err != nil {
		// Dispatch reload failed event
		a.dispatchReloadFailed(ctx, err)

		a.mu.Lock()
		a.state = StateReady
		a.mu.Unlock()
		return err
	}

	a.mu.Lock()
	a.config = newConfig
	a.updateModTimes()
	a.state = StateReady
	a.mu.Unlock()

	// Dispatch change events
	a.dispatchChanges(ctx, oldConfig, newConfig)

	// Dispatch config.reloaded event
	a.dispatchReloaded(ctx)

	return nil
}

// collectWatchedFiles returns the list of config files to watch.
func (a *Actor) collectWatchedFiles() []string {
	var files []string

	// User config
	userConfig := a.loader.resolver.UserConfigFile()
	if userConfig != "" {
		files = append(files, userConfig)
	}

	// Project config
	projectConfig := a.loader.resolver.ProjectConfigFile()
	if projectConfig != "" {
		files = append(files, projectConfig)
	}

	return files
}

// updateModTimes updates the modification times for watched files.
func (a *Actor) updateModTimes() {
	for _, f := range a.watchedFiles {
		info, err := os.Stat(f)
		if err == nil {
			a.modTimes[f] = info.ModTime()
		} else {
			delete(a.modTimes, f)
		}
	}
}

// watchLoop polls for file changes.
func (a *Actor) watchLoop() {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.watchCtx.Done():
			return
		case <-ticker.C:
			if a.checkFileChanges() {
				// Dispatch file changed event
				a.dispatchFileChanged(a.watchCtx)

				// Trigger reload
				_ = a.Reload(a.watchCtx)
			}
		}
	}
}

// checkFileChanges checks if any watched files have changed.
func (a *Actor) checkFileChanges() bool {
	a.mu.RLock()
	files := a.watchedFiles
	oldTimes := make(map[string]time.Time)
	for k, v := range a.modTimes {
		oldTimes[k] = v
	}
	a.mu.RUnlock()

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			// File removed or inaccessible
			if _, existed := oldTimes[f]; existed {
				return true
			}
			continue
		}

		oldTime, existed := oldTimes[f]
		if !existed {
			// New file
			return true
		}

		if !info.ModTime().Equal(oldTime) {
			return true
		}
	}

	return false
}

// dispatchFileChanged dispatches a config.file_changed event.
func (a *Actor) dispatchFileChanged(ctx context.Context) {
	evt := event.NewEvent(event.TypeConfigFileChanged, a.id, nil)
	_ = a.dispatcher.Dispatch(ctx, evt)
}

// dispatchReloaded dispatches a config.reloaded event.
func (a *Actor) dispatchReloaded(ctx context.Context) {
	evt := event.NewEvent(event.TypeConfigReloaded, a.id, nil)
	_ = a.dispatcher.Dispatch(ctx, evt)
}

// dispatchReloadFailed dispatches a config.reload_failed event.
func (a *Actor) dispatchReloadFailed(ctx context.Context, err error) {
	evt := event.NewEvent(event.TypeConfigReloadFailed, a.id, map[string]any{
		"error": err.Error(),
	})
	_ = a.dispatcher.Dispatch(ctx, evt)
}

// dispatchChanges compares old and new configs and dispatches config.updated events.
func (a *Actor) dispatchChanges(ctx context.Context, oldConfig, newConfig *Config) {
	changes := compareConfigs(oldConfig, newConfig)
	for _, change := range changes {
		evt := event.NewEvent(event.TypeConfigUpdated, a.id, change)
		_ = a.dispatcher.Dispatch(ctx, evt)
	}
}

// compareConfigs compares two configs and returns a list of changes.
func compareConfigs(old, new *Config) []ConfigChange {
	var changes []ConfigChange

	// Compare General
	if old.General.LogLevel != new.General.LogLevel {
		changes = append(changes, ConfigChange{
			Path:     "general.log_level",
			OldValue: old.General.LogLevel,
			NewValue: new.General.LogLevel,
		})
	}
	if old.General.LogFormat != new.General.LogFormat {
		changes = append(changes, ConfigChange{
			Path:     "general.log_format",
			OldValue: old.General.LogFormat,
			NewValue: new.General.LogFormat,
		})
	}

	// Compare Timeouts
	if old.Timeouts.Idle.Duration != new.Timeouts.Idle.Duration {
		changes = append(changes, ConfigChange{
			Path:     "timeouts.idle",
			OldValue: old.Timeouts.Idle.String(),
			NewValue: new.Timeouts.Idle.String(),
		})
	}
	if old.Timeouts.Stuck.Duration != new.Timeouts.Stuck.Duration {
		changes = append(changes, ConfigChange{
			Path:     "timeouts.stuck",
			OldValue: old.Timeouts.Stuck.String(),
			NewValue: new.Timeouts.Stuck.String(),
		})
	}

	// Compare Telemetry
	if old.Telemetry.Enabled != new.Telemetry.Enabled {
		changes = append(changes, ConfigChange{
			Path:     "telemetry.enabled",
			OldValue: old.Telemetry.Enabled,
			NewValue: new.Telemetry.Enabled,
		})
	}

	// Compare Events batching
	if old.Events.BatchWindow.Duration != new.Events.BatchWindow.Duration {
		changes = append(changes, ConfigChange{
			Path:     "events.batch_window",
			OldValue: old.Events.BatchWindow.String(),
			NewValue: new.Events.BatchWindow.String(),
		})
	}
	if old.Events.BatchMaxEvents != new.Events.BatchMaxEvents {
		changes = append(changes, ConfigChange{
			Path:     "events.batch_max_events",
			OldValue: old.Events.BatchMaxEvents,
			NewValue: new.Events.BatchMaxEvents,
		})
	}

	// Compare adapter configs (deep comparison)
	for name, newAdapterConfig := range new.Adapters {
		oldAdapterConfig, exists := old.Adapters[name]
		if !exists {
			changes = append(changes, ConfigChange{
				Path:     "adapters." + name,
				OldValue: nil,
				NewValue: newAdapterConfig,
			})
		} else if !reflect.DeepEqual(oldAdapterConfig, newAdapterConfig) {
			changes = append(changes, ConfigChange{
				Path:     "adapters." + name,
				OldValue: oldAdapterConfig,
				NewValue: newAdapterConfig,
			})
		}
	}

	return changes
}

// HotReloadableKeys returns the list of config keys that can be hot-reloaded.
// Per spec §4.2.8.9, these keys can be updated without restart.
func HotReloadableKeys() []string {
	return []string{
		"timeouts.idle",
		"timeouts.stuck",
		"events.batch_window",
		"events.batch_max_events",
		"events.batch_max_bytes",
		"events.batch_idle_flush",
		"events.coalesce.io_streams",
		"events.coalesce.presence",
		"events.coalesce.activity",
		"telemetry.enabled",
		"telemetry.traces.enabled",
		"telemetry.traces.sampler",
		"telemetry.traces.sampler_arg",
		"telemetry.metrics.enabled",
		"telemetry.metrics.interval",
		"telemetry.logs.enabled",
		"telemetry.logs.level",
		"adapters.*", // Adapter patterns are hot-reloadable
	}
}

// IsHotReloadable checks if a config key path is hot-reloadable.
func IsHotReloadable(path string) bool {
	for _, key := range HotReloadableKeys() {
		if key == path {
			return true
		}
		// Handle wildcard patterns
		if len(key) > 0 && key[len(key)-1] == '*' {
			prefix := key[:len(key)-1]
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// ErrActorClosed is returned when operating on a closed actor.
var ErrActorClosed = &ActorClosedError{}

// ActorClosedError indicates the actor is closed.
type ActorClosedError struct{}

func (e *ActorClosedError) Error() string {
	return "config actor is closed"
}

// DefaultActor is the global config actor.
var (
	defaultActor *Actor
	actorMu      sync.RWMutex
)

// SetDefaultActor sets the global config actor.
func SetDefaultActor(a *Actor) {
	actorMu.Lock()
	defer actorMu.Unlock()
	defaultActor = a
}

// GetDefaultActor returns the global config actor.
func GetDefaultActor() *Actor {
	actorMu.RLock()
	defer actorMu.RUnlock()
	return defaultActor
}

// Subscribe registers a handler for config change events using the default dispatcher.
// Returns an unsubscribe function.
func Subscribe(handler func(ctx context.Context, change ConfigChange)) func() {
	return event.Subscribe(event.Subscription{
		Types: []event.Type{event.TypeConfigUpdated},
		Handler: func(ctx context.Context, evt event.Event) error {
			if change, ok := evt.Data.(ConfigChange); ok {
				handler(ctx, change)
			} else if changeMap, ok := evt.Data.(map[string]any); ok {
				// Handle deserialized event data
				change := ConfigChange{}
				if p, ok := changeMap["path"].(string); ok {
					change.Path = p
				}
				change.OldValue = changeMap["old_value"]
				change.NewValue = changeMap["new_value"]
				handler(ctx, change)
			}
			return nil
		},
	})
}

// SubscribeAll registers handlers for all config events.
// Returns an unsubscribe function.
func SubscribeAll(
	onFileChanged func(ctx context.Context),
	onReloaded func(ctx context.Context),
	onUpdated func(ctx context.Context, change ConfigChange),
	onReloadFailed func(ctx context.Context, err string),
) func() {
	return event.Subscribe(event.Subscription{
		Types: []event.Type{
			event.TypeConfigFileChanged,
			event.TypeConfigReloaded,
			event.TypeConfigUpdated,
			event.TypeConfigReloadFailed,
		},
		Handler: func(ctx context.Context, evt event.Event) error {
			switch evt.Type {
			case event.TypeConfigFileChanged:
				if onFileChanged != nil {
					onFileChanged(ctx)
				}
			case event.TypeConfigReloaded:
				if onReloaded != nil {
					onReloaded(ctx)
				}
			case event.TypeConfigUpdated:
				if onUpdated != nil {
					if change, ok := evt.Data.(ConfigChange); ok {
						onUpdated(ctx, change)
					}
				}
			case event.TypeConfigReloadFailed:
				if onReloadFailed != nil {
					if data, ok := evt.Data.(map[string]any); ok {
						if errStr, ok := data["error"].(string); ok {
							onReloadFailed(ctx, errStr)
						}
					}
				}
			}
			return nil
		},
	})
}

// WatchFile adds a file to the watch list for the default actor.
func WatchFile(path string) {
	a := GetDefaultActor()
	if a == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if already watching
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	for _, f := range a.watchedFiles {
		if f == absPath {
			return
		}
	}

	a.watchedFiles = append(a.watchedFiles, absPath)

	// Update mod time for new file
	if info, err := os.Stat(absPath); err == nil {
		a.modTimes[absPath] = info.ModTime()
	}
}
