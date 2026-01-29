package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/pkg/api"
)

// AddAgent validates and persists a new agent configuration.
// It requires the location.repo_path to be a valid git repository.
func AddAgent(cfg *config.Config, newAgent config.AgentConfig) error {
	if err := ValidateAgentConfig(newAgent); err != nil {
		return fmt.Errorf("invalid agent config: %w", err)
	}

	repoPath := newAgent.Location.RepoPath
	// If repo_path is missing, we can't persist project-scoped config correctly without knowing where.
	// However, the spec says "If location.repo_path is unset, the director MUST use the git repository root that contains the request working directory".
	// But AddAgent here is likely called by the CLI/Daemon which should have resolved the repo root.
	// For now, let's assume RepoPath is required for persistence or the caller must resolve it.
	
	if repoPath == "" {
		// Fallback: assume current directory is inside a repo?
		// But we need to write to .amux/config.toml inside the repo root.
		// Let's require the caller to provide resolved RepoPath in the config or we fail.
		return fmt.Errorf("agent location.repo_path is required")
	}

	// Validate it's a git repo
	if !isGitRepo(repoPath) {
		return fmt.Errorf("path %s is not a git repository", repoPath)
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
	if err := config.SaveProjectConfig(cfg, repoPath); err != nil {
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
	// Ideally we use git command or internal/paths logic if available.
	// Since we don't have a full git lib, checking .git is a reasonable heuristic for local repos.
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
