package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentflare-ai/amux/internal/config"
)

func TestAddAgent(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create fake git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	
	newAgent := config.AgentConfig{
		Name:    "Test Agent",
		Adapter: "test-adapter",
		Location: config.LocationConfig{
			Type:     "local",
			RepoPath: tmpDir,
		},
	}

	if err := AddAgent(&cfg, newAgent); err != nil {
		t.Fatalf("AddAgent failed: %v", err)
	}

	// Verify in memory
	if len(cfg.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(cfg.Agents))
	}
	if cfg.Agents[0].Name != "Test Agent" {
		t.Errorf("Expected agent name 'Test Agent', got %q", cfg.Agents[0].Name)
	}

	// Verify persisted file
	projectCfgPath := filepath.Join(tmpDir, ".amux", "config.toml")
	if _, err := os.Stat(projectCfgPath); os.IsNotExist(err) {
		t.Error("Project config file not created")
	}

	// Test Duplicate
	if err := AddAgent(&cfg, newAgent); err == nil {
		t.Error("Expected error adding duplicate agent, got nil")
	}

	// Test Not Git Repo
	notGitDir := t.TempDir()
	badAgent := config.AgentConfig{
		Name:    "Bad Agent",
		Adapter: "test-adapter",
		Location: config.LocationConfig{
			Type:     "local",
			RepoPath: notGitDir,
		},
	}
	if err := AddAgent(&cfg, badAgent); err == nil {
		t.Error("Expected error adding agent to non-git repo, got nil")
	}
}
