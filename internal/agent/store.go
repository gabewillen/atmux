package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/pkg/api"
)

// AddAgent validates and persists a new agent configuration.
// It requires the location.repo_path to be a valid git repository.
func AddAgent(cfg *config.Config, newAgent config.AgentConfig) error {
	if err := ValidateAgentConfig(newAgent); err != nil {
		return fmt.Errorf("invalid agent config: %w", err)
	}

	repoPath := newAgent.Location.RepoPath
	if repoPath == "" {
		return fmt.Errorf("agent location.repo_path is required")
	}

	// Canonicalize
	canonicalPath, err := paths.CanonicalizeRepoRoot(repoPath)
	if err != nil {
		return fmt.Errorf("failed to canonicalize repo path: %w", err)
	}
	newAgent.Location.RepoPath = canonicalPath

	// Validate it's a git repo
	if !isGitRepo(canonicalPath) {
		return fmt.Errorf("path %s is not a git repository", canonicalPath)
	}

	// Check for duplicates
	slug := api.NormalizeAgentSlug(newAgent.Name)
	for _, a := range cfg.Agents {
		if api.NormalizeAgentSlug(a.Name) == slug {
			return fmt.Errorf("agent with name %q (slug: %s) already exists", newAgent.Name, slug)
		}
	}

	// Append and save
	cfg.Agents = append(cfg.Agents, newAgent)
	
	// We save to Project Config
	if err := config.SaveProjectConfig(cfg, canonicalPath); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	return nil
}

// ValidateAgentConfig checks required fields.
func ValidateAgentConfig(c config.AgentConfig) error {
	if c.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if c.Adapter == "" {
		return fmt.Errorf("agent adapter is required")
	}
	return nil
}

func isGitRepo(path string) bool {
	// Simple check: .git directory exists
	// ideally use git rev-parse --is-inside-work-tree
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}