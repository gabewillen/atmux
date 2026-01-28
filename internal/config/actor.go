// Package config provides configuration management with live updates.
package config

import (
	"context"
	"sync"

	amuxerrors "github.com/agentflare-ai/amux/internal/errors"
)

// Actor manages configuration with live updates and subscriptions.
type Actor struct {
	mu       sync.RWMutex
	config   *Config
	watchers []chan *Config
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewActor creates a new configuration actor.
func NewActor(config *Config) *Actor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Actor{
		config:   config,
		watchers: make([]chan *Config, 0),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Get returns the current configuration.
func (a *Actor) Get() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.config
}

// Update updates the configuration and notifies all watchers.
func (a *Actor) Update(config *Config) error {
	if err := Validate(config); err != nil {
		return amuxerrors.Wrap("validating config", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	oldConfig := a.config
	a.config = config

	// Notify watchers in separate goroutines to avoid blocking
	for _, watcher := range a.watchers {
		go func(ch chan *Config) {
			select {
			case ch <- config:
			case <-a.ctx.Done():
			}
		}(watcher)
	}

	// TODO: implement selective notification based on what changed
	_ = oldConfig

	return nil
}

// Subscribe creates a new subscription for configuration changes.
func (a *Actor) Subscribe() <-chan *Config {
	a.mu.Lock()
	defer a.mu.Unlock()

	ch := make(chan *Config, 1) // Buffer 1 to prevent blocking
	a.watchers = append(a.watchers, ch)

	// Send current config immediately
	go func() {
		select {
		case ch <- a.config:
		case <-a.ctx.Done():
		}
	}()

	return ch
}

// Unsubscribe removes a subscription.
func (a *Actor) Unsubscribe(ch <-chan *Config) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, watcher := range a.watchers {
		if watcher == ch {
			// Remove watcher from slice
			a.watchers = append(a.watchers[:i], a.watchers[i+1:]...)
			close(watcher)
			break
		}
	}
}

// Shutdown gracefully shuts down the actor and closes all subscriptions.
func (a *Actor) Shutdown() {
	a.cancel()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Close all watchers
	for _, watcher := range a.watchers {
		close(watcher)
	}
	a.watchers = nil
}

// Watcher represents a configuration change subscriber.
type Watcher interface {
	OnConfigChange(oldConfig, newConfig *Config) error
}

// RegisterWatcher registers a watcher that receives config change callbacks.
func (a *Actor) RegisterWatcher(watcher Watcher) {
	ch := a.Subscribe()

	go func() {
		var oldConfig *Config
		for {
			select {
			case newConfig, ok := <-ch:
				if !ok {
					return
				}

				if oldConfig != nil {
					if err := watcher.OnConfigChange(oldConfig, newConfig); err != nil {
						// TODO: log error but don't crash
						_ = err
					}
				}
				oldConfig = newConfig

			case <-a.ctx.Done():
				return
			}
		}
	}()
}
