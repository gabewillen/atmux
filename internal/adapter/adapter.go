// Package adapter provides stable interfaces for adapters (to be fully implemented in Phase 8).
package adapter

import (
	"context"

	"github.com/agentflare-ai/amux/internal/event"
)

// Pattern represents a pattern that adapters can match against.
type Pattern struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "regex", "glob", "semantic"
	Pattern     string                 `json:"pattern"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// Match represents a successful pattern match.
type Match struct {
	PatternID  string                 `json:"pattern_id"`
	Offset     int                    `json:"offset"`
	Length     int                    `json:"length"`
	Groups     map[string]string      `json:"groups,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Action represents an action that adapters can perform.
type Action struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "input", "command", "message", "screenshot"
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority,omitempty"` // Higher = more important
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PTYSnapshot represents a snapshot of PTY output.
type PTYSnapshot struct {
	Content   string `json:"content"`
	Length    int    `json:"length"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp
}

// ProcessSnapshot represents a snapshot of process information.
type ProcessSnapshot struct {
	PID        int                    `json:"pid"`
	ParentPID  int                    `json:"parent_pid"`
	Command    string                 `json:"command"`
	Args       []string               `json:"args"`
	Env        map[string]string      `json:"env"`
	WorkingDir string                 `json:"working_dir"`
	Status     string                 `json:"status"` // "running", "stopped", "zombie"
	Children   []*ProcessSnapshot     `json:"children,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Runtime provides adapter pattern matching and action execution.
type Runtime interface {
	// Initialize the adapter runtime with configuration
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Match patterns against provided content
	Match(ctx context.Context, patterns []Pattern, content string) ([]Match, error)

	// Execute an action
	Execute(ctx context.Context, action *Action) error

	// Get adapter information
	Info() *RuntimeInfo

	// Shutdown the adapter runtime
	Shutdown(ctx context.Context) error
}

// RuntimeInfo provides information about the adapter runtime.
type RuntimeInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"` // e.g., "regex", "semantic", "input", "command"
	Config       map[string]interface{} `json:"config,omitempty"`
}

// Manager manages adapter lifecycles and interactions.
type Manager interface {
	// Load and initialize an adapter
	Load(ctx context.Context, adapterName string, config map[string]interface{}) error

	// Unload an adapter
	Unload(ctx context.Context, adapterName string) error

	// Get all loaded adapters
	GetAdapters() map[string]Runtime

	// Match patterns using all loaded adapters
	Match(ctx context.Context, content string) ([]Match, error)

	// Execute an action using appropriate adapter
	Execute(ctx context.Context, action *Action) error

	// Process event and generate potential actions
	ProcessEvent(ctx context.Context, ev *event.Event) ([]*Action, error)

	// Get manager information
	Info() *ManagerInfo

	// Shutdown all adapters
	Shutdown(ctx context.Context) error
}

// ManagerInfo provides information about the adapter manager.
type ManagerInfo struct {
	LoadedAdapters int      `json:"loaded_adapters"`
	ActiveMatchers int      `json:"active_matchers"`
	ActiveActions  int      `json:"active_actions"`
	Capabilities   []string `json:"capabilities"`
}

// NoopManager provides a no-op implementation for Phase 0.
type NoopManager struct {
	adapters map[string]Runtime
	config   map[string]map[string]interface{}
}

// NewNoopManager creates a new no-op adapter manager.
func NewNoopManager() *NoopManager {
	return &NoopManager{
		adapters: make(map[string]Runtime),
		config:   make(map[string]map[string]interface{}),
	}
}

// Load implements Manager interface.
func (m *NoopManager) Load(ctx context.Context, adapterName string, config map[string]interface{}) error {
	// TODO: implement actual adapter loading in Phase 8
	return nil
}

// Unload implements Manager interface.
func (m *NoopManager) Unload(ctx context.Context, adapterName string) error {
	delete(m.adapters, adapterName)
	return nil
}

// GetAdapters implements Manager interface.
func (m *NoopManager) GetAdapters() map[string]Runtime {
	return m.adapters
}

// Match implements Manager interface (no-op).
func (m *NoopManager) Match(ctx context.Context, content string) ([]Match, error) {
	// TODO: implement actual pattern matching in Phase 8
	return []Match{}, nil
}

// Execute implements Manager interface (no-op).
func (m *NoopManager) Execute(ctx context.Context, action *Action) error {
	// TODO: implement actual action execution in Phase 8
	return nil
}

// ProcessEvent implements Manager interface (no-op).
func (m *NoopManager) ProcessEvent(ctx context.Context, ev *event.Event) ([]*Action, error) {
	// TODO: implement actual event processing in Phase 8
	return []*Action{}, nil
}

// Info implements Manager interface.
func (m *NoopManager) Info() *ManagerInfo {
	return &ManagerInfo{
		LoadedAdapters: len(m.adapters),
		ActiveMatchers: 0,
		ActiveActions:  0,
		Capabilities:   []string{}, // Will be populated in Phase 8
	}
}

// Shutdown implements Manager interface.
func (m *NoopManager) Shutdown(ctx context.Context) error {
	m.adapters = make(map[string]Runtime)
	m.config = make(map[string]map[string]interface{})
	return nil
}
