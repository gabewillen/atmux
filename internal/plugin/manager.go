package plugin

import (
	"fmt"
	"sync"
)

// Manager manages installed plugins.
type Manager struct {
	mu       sync.RWMutex
	registry map[string]*Plugin
}

// Plugin represents an installed plugin.
type Plugin struct {
	Manifest Manifest
	Enabled  bool
	Path     string // Installation path
}

// NewManager creates a new plugin manager.
func NewManager() *Manager {
	return &Manager{
		registry: make(map[string]*Plugin),
	}
}

// Install registers a plugin (simulated installation).
func (m *Manager) Install(manifest Manifest, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.registry[manifest.Name]; exists {
		return fmt.Errorf("plugin %s already installed", manifest.Name)
	}

	m.registry[manifest.Name] = &Plugin{
		Manifest: manifest,
		Enabled:  true, // Auto-enable on install
		Path:     path,
	}
	return nil
}

// List returns all installed plugins.
func (m *Manager) List() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	list := make([]*Plugin, 0, len(m.registry))
	for _, p := range m.registry {
		list = append(list, p)
	}
	return list
}

// Enable enables a plugin.
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	p, ok := m.registry[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}
	p.Enabled = true
	return nil
}

// Disable disables a plugin.
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	p, ok := m.registry[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}
	p.Enabled = false
	return nil
}

// Remove removes a plugin.
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, ok := m.registry[name]; !ok {
		return fmt.Errorf("plugin %s not found", name)
	}
	delete(m.registry, name)
	return nil
}
