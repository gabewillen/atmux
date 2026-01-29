package agent

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stateforward/hsm-go/muid"

	"github.com/agentflare-ai/amux/internal/event"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

func TestPersisterSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	agents := []api.Agent{
		{
			ID:       muid.MUID(100),
			Name:     "agent-one",
			Slug:     "agent-one",
			Adapter:  "claude-code",
			RepoRoot: "/tmp/repo1",
		},
		{
			ID:       muid.MUID(200),
			Name:     "agent-two",
			Slug:     "agent-two",
			Adapter:  "cursor",
			RepoRoot: "/tmp/repo2",
		},
	}

	// Save
	if err := p.save(agents); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	// Verify file exists
	path := p.filePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("persistence file should exist at %s", path)
	}

	// Load
	loaded, err := p.load()
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("loaded %d agents, want 2", len(loaded))
	}

	if loaded[0].ID != muid.MUID(100) {
		t.Errorf("loaded[0].ID = %d, want 100", loaded[0].ID)
	}
	if loaded[0].Name != "agent-one" {
		t.Errorf("loaded[0].Name = %q, want %q", loaded[0].Name, "agent-one")
	}
	if loaded[1].ID != muid.MUID(200) {
		t.Errorf("loaded[1].ID = %d, want 200", loaded[1].ID)
	}
	if loaded[1].Adapter != "cursor" {
		t.Errorf("loaded[1].Adapter = %q, want %q", loaded[1].Adapter, "cursor")
	}
}

func TestPersisterLoadNonExistent(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	loaded, err := p.load()
	if err != nil {
		t.Fatalf("load() should not error for non-existent file: %v", err)
	}
	if loaded != nil {
		t.Errorf("load() should return nil for non-existent file, got %v", loaded)
	}
}

func TestPersisterLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	// Create empty file
	if err := os.WriteFile(p.filePath(), []byte{}, 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	loaded, err := p.load()
	if err != nil {
		t.Fatalf("load() should not error for empty file: %v", err)
	}
	if loaded != nil {
		t.Errorf("load() should return nil for empty file, got %v", loaded)
	}
}

func TestPersisterLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	// Create invalid JSON
	if err := os.WriteFile(p.filePath(), []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("write invalid file: %v", err)
	}

	_, err := p.load()
	if err == nil {
		t.Error("load() should error for invalid JSON")
	}
}

func TestPersisterAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	agents := []api.Agent{
		{
			ID:       muid.MUID(100),
			Name:     "test",
			Slug:     "test",
			Adapter:  "claude-code",
			RepoRoot: "/tmp/repo",
		},
	}

	if err := p.save(agents); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	// Verify no tmp file remains
	tmpPath := p.filePath() + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("tmp file should not exist after successful save")
	}

	// Verify content is valid JSON
	data, err := os.ReadFile(p.filePath())
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var loaded []api.Agent
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
}

func TestPersisterFilePath(t *testing.T) {
	dir := t.TempDir()
	r := paths.NewResolverWithDataDir(dir)
	p := newPersister(r)

	expected := filepath.Join(dir, "agents.json")
	if p.filePath() != expected {
		t.Errorf("filePath() = %q, want %q", p.filePath(), expected)
	}
}

func TestManagerPersistsOnAdd(t *testing.T) {
	repoRoot := initTestRepo(t)
	dataDir := t.TempDir()
	resolver := paths.NewResolverWithDataDir(dataDir)
	_ = resolver.SetRepoRoot(repoRoot)

	mgr := NewManagerWithResolver(event.NewNoopDispatcher(), resolver)
	ctx := context.Background()

	_, err := mgr.Add(ctx, api.Agent{
		Name:     "persist-test",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Verify file was created
	path := filepath.Join(dataDir, "agents.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("persistence file not created: %v", err)
	}

	var agents []api.Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		t.Fatalf("invalid JSON in persistence file: %v", err)
	}

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent in file, got %d", len(agents))
	}
	if agents[0].Name != "persist-test" {
		t.Errorf("persisted agent name = %q, want %q", agents[0].Name, "persist-test")
	}
}

func TestManagerPersistsOnRemove(t *testing.T) {
	repoRoot := initTestRepo(t)
	dataDir := t.TempDir()
	resolver := paths.NewResolverWithDataDir(dataDir)
	_ = resolver.SetRepoRoot(repoRoot)

	mgr := NewManagerWithResolver(event.NewNoopDispatcher(), resolver)
	ctx := context.Background()

	agent, err := mgr.Add(ctx, api.Agent{
		Name:     "persist-remove",
		Adapter:  "claude-code",
		RepoRoot: repoRoot,
	})
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Remove the agent
	if err := mgr.Remove(ctx, agent.ID, false); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify file now has 0 agents
	path := filepath.Join(dataDir, "agents.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("persistence file missing: %v", err)
	}

	var agents []api.Agent
	if err := json.Unmarshal(data, &agents); err != nil {
		t.Fatalf("invalid JSON in persistence file: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("expected 0 agents after remove, got %d", len(agents))
	}
}

func TestManagerLoadPersisted(t *testing.T) {
	repoRoot := initTestRepo(t)
	dataDir := t.TempDir()

	// Pre-populate the persistence file
	agents := []api.Agent{
		{
			ID:       muid.MUID(42),
			Name:     "preloaded",
			Slug:     "preloaded",
			Adapter:  "claude-code",
			RepoRoot: repoRoot,
		},
	}
	data, err := json.Marshal(agents)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(dataDir, "agents.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write persistence file: %v", err)
	}

	// Create manager and load persisted agents
	resolver := paths.NewResolverWithDataDir(dataDir)
	_ = resolver.SetRepoRoot(repoRoot)
	mgr := NewManagerWithResolver(event.NewNoopDispatcher(), resolver)
	ctx := context.Background()

	if err := mgr.LoadPersisted(ctx); err != nil {
		t.Fatalf("LoadPersisted() failed: %v", err)
	}

	// The agent should be loaded
	list := mgr.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 agent after LoadPersisted, got %d", len(list))
	}

	if list[0].Name != "preloaded" {
		t.Errorf("loaded agent name = %q, want %q", list[0].Name, "preloaded")
	}
}
