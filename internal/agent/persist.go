// Package agent - persist.go provides disk persistence for agent definitions.
//
// Per spec, agent definitions must survive daemon restarts. This file implements
// a persistence layer that saves agents to ~/.amux/agents.json using the
// paths.Resolver for directory resolution.
//
// The persistence format is JSON with api.Agent types. On startup, the manager
// loads persisted agents. On Add/Remove, the persistence file is updated.
package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// persistFilename is the name of the agents persistence file.
const persistFilename = "agents.json"

// persister handles reading and writing agent definitions to disk.
type persister struct {
	mu       sync.Mutex
	resolver *paths.Resolver
}

// newPersister creates a new persister with the given resolver.
func newPersister(resolver *paths.Resolver) *persister {
	if resolver == nil {
		resolver = paths.DefaultResolver
	}
	return &persister{resolver: resolver}
}

// filePath returns the full path to the persistence file.
func (p *persister) filePath() string {
	return filepath.Join(p.resolver.DataDir(), persistFilename)
}

// load reads persisted agent definitions from disk.
// Returns an empty slice (not an error) if the file does not exist.
func (p *persister) load() ([]api.Agent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("load agents: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var agents []api.Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		return nil, fmt.Errorf("load agents: unmarshal: %w", err)
	}
	return agents, nil
}

// save writes agent definitions to disk atomically.
// It creates the data directory if it does not exist.
func (p *persister) save(agents []api.Agent) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	path := p.filePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("save agents: create dir: %w", err)
	}

	data, err := json.MarshalIndent(agents, "", "  ")
	if err != nil {
		return fmt.Errorf("save agents: marshal: %w", err)
	}

	// Write atomically via temp file + rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("save agents: write tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up tmp on rename failure
		_ = os.Remove(tmpPath)
		return fmt.Errorf("save agents: rename: %w", err)
	}

	return nil
}
