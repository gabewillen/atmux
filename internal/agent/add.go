// Package agent provides agent orchestration: lifecycle, presence, and messaging.
// add.go implements agent add validation and config building (spec §5.2, §5.3.1).
package agent

import (
	"errors"
	"fmt"
	"strings"

	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/git"
	"github.com/agentflare-ai/amux/internal/paths"
)

// ErrNotInRepo is returned when adding an agent outside a git repository.
var ErrNotInRepo = errors.New("not in a git repository")

// ErrInvalidLocation is returned when location type is invalid.
var ErrInvalidLocation = errors.New("invalid location type")

// AddInput holds validated inputs for adding an agent (spec §5.2).
type AddInput struct {
	Name     string
	About    string
	Adapter  string
	RepoRoot string // Canonical repo root; must be a git repo
	Location config.AgentLocationConfig
}

// ValidateAddInput validates inputs for adding an agent.
// Adding an agent outside a git repo fails (spec §1.3, §5.2).
// For local agents, repoRoot is used; if empty, the caller must resolve from cwd.
func ValidateAddInput(repoRoot, name, about, adapter string, location config.AgentLocationConfig) (*AddInput, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("repo root required: %w", ErrNotInRepo)
	}
	ok, err := git.IsRepo(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("check git repo: %w", err)
	}
	if !ok {
		return nil, ErrNotInRepo
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	adapter = strings.TrimSpace(adapter)
	if adapter == "" {
		return nil, fmt.Errorf("adapter is required")
	}
	locType := strings.TrimSpace(strings.ToLower(location.Type))
	if locType != "local" && locType != "ssh" {
		return nil, fmt.Errorf("location type must be local or ssh: %w", ErrInvalidLocation)
	}
	return &AddInput{
		Name:     name,
		About:    strings.TrimSpace(about),
		Adapter:  adapter,
		RepoRoot: repoRoot,
		Location: config.AgentLocationConfig{
			Type:     locType,
			Host:     location.Host,
			User:     location.User,
			Port:     location.Port,
			RepoPath: location.RepoPath,
		},
	}, nil
}

// BuildAgentConfig builds an AgentConfig for persistence from AddInput.
// agentSlug is the uniquified slug assigned to this agent (for worktree path and persistence).
func BuildAgentConfig(in *AddInput, agentSlug string) config.AgentConfig {
	return config.AgentConfig{
		Name:    in.Name,
		About:   in.About,
		Adapter: in.Adapter,
		Slug:    agentSlug,
		Location: config.AgentLocationConfig{
			Type:     in.Location.Type,
			Host:     in.Location.Host,
			User:     in.Location.User,
			Port:     in.Location.Port,
			RepoPath: in.Location.RepoPath,
		},
	}
}

// ResolveRepoRoot returns the canonical repo root for a local add.
// If repoPath is non-empty, canonicalizes it; otherwise uses git.Root(cwd).
func ResolveRepoRoot(homeDir, cwd, repoPath string) (string, error) {
	if repoPath != "" {
		return paths.CanonicalizeRepoRoot(homeDir, repoPath)
	}
	return git.Root(cwd)
}
