// Package config implements a configuration actor with live updates and subscriptions
package config

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// Actor manages configuration with live updates and subscriptions
type Actor struct {
	mu         sync.RWMutex
	config     *Config
	subscribers map[string]chan Config
	nextID     int
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewActor creates a new configuration actor
func NewActor(initialConfig *Config) *Actor {
	ctx, cancel := context.WithCancel(context.Background())
	actor := &Actor{
		config:      initialConfig,
		subscribers: make(map[string]chan Config),
		ctx:         ctx,
		cancel:      cancel,
	}
	return actor
}

// Get returns the current configuration
func (a *Actor) Get() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// Update updates the configuration and notifies subscribers
func (a *Actor) Update(newConfig *Config) error {
	// Validate the new configuration
	if err := newConfig.Validate(); err != nil {
		return err
	}

	// Create a copy of the config to avoid race conditions
	configCopy := *newConfig

	a.mu.Lock()
	defer a.mu.Unlock()

	// Update the config
	a.config = newConfig

	// Notify all subscribers
	for _, ch := range a.subscribers {
		// Use a timeout to prevent indefinite blocking
		select {
		case ch <- configCopy:
		case <-time.After(100 * time.Millisecond): // 100ms timeout
			// Skip if we can't send within the timeout
		}
	}

	return nil
}

// Subscribe registers a subscriber to receive configuration updates
func (a *Actor) Subscribe() (<-chan Config, func()) {
	a.mu.Lock()
	defer a.mu.Unlock()

	id := strconv.Itoa(a.nextID)
	a.nextID++

	// Use a buffered channel with sufficient capacity to hold initial config and updates
	ch := make(chan Config, 10)
	a.subscribers[id] = ch

	unsubscribe := func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		delete(a.subscribers, id)
		close(ch)
	}

	// Send the current config immediately
	// Make a copy to avoid race conditions
	configCopy := *a.config
	// Use a non-blocking send for the initial value
	select {
	case ch <- configCopy:
	default:
		// If the channel is full, skip sending the initial value
	}

	return ch, unsubscribe
}

// LoadAndWatch loads configuration from a file and watches for changes
func (a *Actor) LoadAndWatch(ctx context.Context, configPath string) error {
	// Initial load
	newConfig, err := LoadConfig(configPath)
	if err != nil {
		return err
	}
	
	if err := a.Update(newConfig); err != nil {
		return err
	}
	
	// TODO: Implement file watching logic using fsnotify or similar
	// For now, we'll simulate periodic reloading in a goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				updatedConfig, err := LoadConfig(configPath)
				if err != nil {
					// Log error but continue watching
					continue
				}
				
				// Only update if the config actually changed
				currentConfig := a.Get()
				if configsEqual(currentConfig, updatedConfig) {
					continue
				}
				
				if err := a.Update(updatedConfig); err != nil {
					// Log error but continue watching
					continue
				}
			}
		}
	}()
	
	return nil
}

// configsEqual compares two config objects for equality
func configsEqual(a, b *Config) bool {
	// Simple comparison - in a real implementation, we'd need a deep comparison
	// For now, we'll just compare a few key fields
	return a.Core.RepoRoot == b.Core.RepoRoot &&
		   a.Core.Debug == b.Core.Debug &&
		   a.Server.SocketPath == b.Server.SocketPath &&
		   a.Logging.Level == b.Logging.Level
}

// Close shuts down the actor and all subscriptions
func (a *Actor) Close() {
	a.cancel()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Close all subscriber channels
	for id, ch := range a.subscribers {
		delete(a.subscribers, id)
		close(ch)
	}
}