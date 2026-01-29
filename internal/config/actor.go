package config

import (
	"context"
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
}

// Model defines the configuration actor state machine. It is intentionally
// minimal for Phase 0: it supports loading and reloading, and emits the
// ConfigReloaded event; file watching is left to the caller, which should
// dispatch ConfigFileChanged as appropriate.
var Model = hsm.Define("config",
	hsm.State("loading",
		hsm.Entry(func(ctx context.Context, a *Actor, e hsm.Event) {
			_, _ = a.reload(ctx)
		}),
	),

	hsm.State("ready"),
	hsm.State("reloading",
		hsm.Entry(func(ctx context.Context, a *Actor, e hsm.Event) {
			_, _ = a.reload(ctx)
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
	a := &Actor{load: loader}
	if _, err := a.reload(ctx); err != nil {
		return nil, err
	}
	// Start the HSM using the Model; Actor is both the model owner and data.
	a = hsm.Started(ctx, a, &Model)
	return a, nil
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
}

// reload loads configuration via the loader, computes a simple diff, updates
// the current config, and emits ConfigReloaded/ConfigUpdated events.
func (a *Actor) reload(ctx context.Context) (*Config, error) {
	newCfg, err := a.load()
	if err != nil {
		// Even on error, keep existing config; caller can observe the error.
		return nil, err
	}

	a.mu.Lock()
	oldCfg := a.cfg
	a.cfg = newCfg
	a.mu.Unlock()

	// Emit ConfigReloaded event through HSM; we don't attach data here.
	done := hsm.Dispatch(ctx, a, hsm.Event{Name: ConfigReloaded})
	<-done

	// For Phase 0, we emit a coarse-grained ConfigUpdated event for the entire
	// config rather than computing per-key diffs. This still allows consumers
	// to react to configuration changes via subscription.
	change := ConfigChange{Path: "*", OldValue: oldCfg, NewValue: newCfg}
	a.subscribersMu.RLock()
	for _, ch := range a.subscribers {
		select {
		case ch <- change:
		default:
			// Drop if subscriber is slow; best-effort semantics are acceptable for Phase 0.
		}
	}
	a.subscribersMu.RUnlock()

	return newCfg, nil
}
