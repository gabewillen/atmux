package config

import (
	"context"
	"sync"
)

type ConfigActor struct {
	mu          sync.RWMutex
	current     *Config
	subscribers []chan<- Config
}

func NewActor(initial *Config) *ConfigActor {
	return &ConfigActor{
		current: initial,
	}
}

func (a *ConfigActor) Start(ctx context.Context) error {
	// TODO: Implement file watching and reloading
	return nil
}

func (a *ConfigActor) Subscribe() <-chan Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	ch := make(chan Config, 1)
	a.subscribers = append(a.subscribers, ch)
	if a.current != nil {
		ch <- *a.current
	}
	return ch
}

func (a *ConfigActor) Update(cfg *Config) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.current = cfg
	for _, ch := range a.subscribers {
		select {
		case ch <- *cfg:
		default:
		}
	}
}
