// Package config provides configuration management for amux.
package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/stateforward/hsm-go"
)

// Actor manages configuration state and live updates.
type Actor struct {
	mu       sync.RWMutex
	config   *Config
	loader   *Loader
	subs     []chan ConfigChange
	stopWatching func() // Function to stop file watching
}

// ConfigChange represents a configuration value change.
type ConfigChange struct {
	Path     string // Config key path: "coordination.interval"
	OldValue any
	NewValue any
}

// Config events
const (
	ConfigFileChanged = "config.file_changed" // File modified on disk
	ConfigReloaded    = "config.reloaded"     // Reload complete
	ConfigUpdated     = "config.updated"      // Specific value changed
)

// NewActor creates a new configuration actor.
func NewActor(loader *Loader) *Actor {
	return &Actor{
		loader: loader,
		subs:   make([]chan ConfigChange, 0),
	}
}

// Load loads the initial configuration.
func (a *Actor) Load() error {
	cfg, err := a.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	a.mu.Lock()
	a.config = cfg
	a.mu.Unlock()

	return nil
}

// Get returns the current configuration.
func (a *Actor) Get() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// Subscribe returns a channel that receives configuration change notifications.
func (a *Actor) Subscribe() <-chan ConfigChange {
	ch := make(chan ConfigChange, 10)
	a.mu.Lock()
	a.subs = append(a.subs, ch)
	a.mu.Unlock()
	return ch
}

// notifySubscribers notifies all subscribers of a configuration change.
func (a *Actor) notifySubscribers(change ConfigChange) {
	a.mu.RLock()
	subs := make([]chan ConfigChange, len(a.subs))
	copy(subs, a.subs)
	a.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- change:
		default:
			// Channel full, skip
		}
	}
}

// Reload reloads configuration from disk.
func (a *Actor) Reload() error {
	oldCfg := a.Get()
	cfg, err := a.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	a.mu.Lock()
	a.config = cfg
	a.mu.Unlock()

	// Compare old and new config and emit ConfigChange events
	// Phase 0: Basic comparison - full diff would be implemented in later phases
	changes := a.compareConfigs(oldCfg, cfg)
	for _, change := range changes {
		a.notifySubscribers(change)
	}

	return nil
}

// compareConfigs compares two configs and returns a list of changes.
// Phase 0: Simplified comparison - full deep diff would be implemented later.
func (a *Actor) compareConfigs(old, new *Config) []ConfigChange {
	var changes []ConfigChange

	// Compare general settings
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

	// TODO: Compare other fields as needed
	// For Phase 0, this is a minimal implementation

	return changes
}

// StartWatching starts watching configuration files for changes.
// Phase 0: Placeholder - actual file watching would use fsnotify or similar.
func (a *Actor) StartWatching() {
	// Phase 0: File watching is a placeholder
	// In a full implementation, this would:
	// 1. Watch ~/.config/amux/config.toml
	// 2. Watch .amux/config.toml (if repo root is set)
	// 3. Watch adapter config files
	// 4. On change, dispatch ConfigFileChanged event
	a.stopWatching = func() {
		// Stop watching
	}
}

// StopWatching stops watching configuration files.
func (a *Actor) StopWatching() {
	if a.stopWatching != nil {
		a.stopWatching()
		a.stopWatching = nil
	}
}

// ConfigModel defines the HSM model for configuration management.
var ConfigModel = hsm.Define("config",
	hsm.State("loading"),
	hsm.State("ready"),
	hsm.State("reloading"),

	hsm.Transition(hsm.On(hsm.Event{Name: "config.loaded"}), hsm.Source("loading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigFileChanged}), hsm.Source("ready"), hsm.Target("reloading")),
	hsm.Transition(hsm.On(hsm.Event{Name: ConfigReloaded}), hsm.Source("reloading"), hsm.Target("ready")),
	hsm.Transition(hsm.On(hsm.Event{Name: "config.reload_failed"}), hsm.Source("reloading"), hsm.Target("ready")),

	hsm.Initial(hsm.Target("loading")),
)

// Start initializes the config actor and starts the HSM.
func (a *Actor) Start(ctx context.Context) error {
	// Load initial config
	if err := a.Load(); err != nil {
		return fmt.Errorf("failed to load initial config: %w", err)
	}

	// Start file watching when entering ready state
	// This would be done via HSM entry actions in a full implementation
	a.StartWatching()

	return nil
}
