// Package agent provides the agent actor model and local management helpers.
//
// This file implements Phase 2 local agent management for:
// - Agent add flow (validation, repo required, config persistence) per spec §5.2
// - Worktree isolation and slug-based path layout per spec §5.3.1
// - Git merge strategy selection scaffolding per spec §5.7
package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"

	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/errors"
	"github.com/stateforward/amux/internal/paths"
	"github.com/stateforward/amux/pkg/api"
)

// AddLocalAgentOptions holds input parameters for adding a local agent.
//
// MergeStrategy represents supported git merge strategies for agent worktrees.
type MergeStrategy string

const (
	MergeStrategyMergeCommit MergeStrategy = "merge-commit"
	MergeStrategySquash      MergeStrategy = "squash"
	MergeStrategyRebase      MergeStrategy = "rebase"
	MergeStrategyFFOnly      MergeStrategy = "ff-only"
)

// SelectMergeStrategy returns the effective merge strategy based on config,
// defaulting to squash when an unknown or empty value is provided.
func SelectMergeStrategy(cfg *config.Config) MergeStrategy {
	if cfg == nil {
		return MergeStrategySquash
	}

	s := cfg.Git.Merge.Strategy
	switch s {
	case string(MergeStrategyMergeCommit):
		return MergeStrategyMergeCommit
	case string(MergeStrategyRebase):
		return MergeStrategyRebase
	case string(MergeStrategyFFOnly):
		return MergeStrategyFFOnly
	case "", string(MergeStrategySquash):
		fallthrough
	default:
		return MergeStrategySquash
	}
}

// AddLocalAgentOptions holds input parameters for adding a local agent.
type AddLocalAgentOptions struct {
	Name    string
	About   string
	Adapter string
	RepoRoot string
}

// AddLocalAgent adds a new local agent for the given repository root.
//
// Responsibilities (Phase 2, spec §5.2, §5.3.1, §5.7.1):
// - Validate input and ensure the repoRoot is a git repository
// - Derive a unique agent_slug from Name using NormalizeAgentSlug
// - Ensure a git worktree exists at .amux/worktrees/{agent_slug}/ under repoRoot
// - Append the agent to the provided configuration (Agents slice)
// - Return the instantiated api.Agent with canonical RepoRoot and Worktree
func AddLocalAgent(ctx context.Context, cfg *config.Config, opts AddLocalAgentOptions) (*api.Agent, string, error) {
	if cfg == nil {
		return nil, "", errors.Wrap(errors.ErrInvalidInput, "config must not be nil")
	}

	if opts.Name == "" {
		return nil, "", errors.Wrap(errors.ErrInvalidInput, "agent name is required")
	}

	if opts.Adapter == "" {
		return nil, "", errors.Wrap(errors.ErrInvalidInput, "adapter name is required")
	}

	if opts.RepoRoot == "" {
		return nil, "", errors.Wrap(errors.ErrInvalidInput, "repo root is required")
	}

	canonicalRepoRoot, err := api.CanonicalizeRepoRoot(opts.RepoRoot)
	if err != nil {
		return nil, "", errors.Wrap(err, "canonicalize repo root")
	}

	// Validate that canonicalRepoRoot is a git repository per spec §5.3.4.
	if err := verifyGitRepository(ctx, canonicalRepoRoot); err != nil {
		return nil, "", err
	}

	resolver, err := paths.NewResolver(canonicalRepoRoot)
	if err != nil {
		return nil, "", errors.Wrap(err, "create path resolver")
	}

	slug := deriveUniqueAgentSlug(cfg, opts.Name)
	worktreePath := resolver.WorktreePath(slug)

	if err := ensureLocalWorktree(ctx, canonicalRepoRoot, slug, worktreePath); err != nil {
		return nil, "", err
	}

	// Append agent configuration.
	cfg.Agents = append(cfg.Agents, config.AgentConfig{
		Name:   opts.Name,
		About:  opts.About,
		Adapter: opts.Adapter,
		Location: config.LocationConfig{
			Type:     "local",
			RepoPath: canonicalRepoRoot,
		},
	})

	agent := &api.Agent{
		ID:       api.GenerateID(),
		Name:     opts.Name,
		About:    opts.About,
		Adapter:  opts.Adapter,
		RepoRoot: canonicalRepoRoot,
		Worktree: worktreePath,
		Location: api.Location{
			Type:     api.LocationLocal,
			RepoPath: canonicalRepoRoot,
		},
	}

	return agent, slug, nil
}

// deriveUniqueAgentSlug derives a unique agent_slug given the existing config
// and desired agent name. Per spec §5.3.1, collisions are resolved by
// appending a numeric suffix -2, -3, ... until unique.
func deriveUniqueAgentSlug(cfg *config.Config, name string) string {
	base := api.NormalizeAgentSlug(name)
	candidate := base
	index := 2

	for {
		if !slugExists(cfg, candidate) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", base, index)
		index++
	}
}

// slugExists checks whether a slug is already in use by any configured agent.
func slugExists(cfg *config.Config, slug string) bool {
	for _, a := range cfg.Agents {
		if api.NormalizeAgentSlug(a.Name) == slug {
			return true
		}
	}
	return false
}

// verifyGitRepository ensures the given path is a git repository.
// It runs `git rev-parse --show-toplevel` and verifies success.
func verifyGitRepository(ctx context.Context, repoRoot string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "rev-parse", "--show-toplevel")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(errors.ErrInvalidInput, "not a git repository: %s (%s)", repoRoot, stdout.String())
	}

	return nil
}

// ensureLocalWorktree ensures a git worktree exists for the given agent slug.
// If the worktree already exists, the function is idempotent and returns nil.
func ensureLocalWorktree(ctx context.Context, repoRoot, slug, worktreePath string) error {
	// First, check if a worktree pointing to worktreePath already exists.
	listCmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "worktree", "list", "--porcelain")
	var out bytes.Buffer
	listCmd.Stdout = &out
	listCmd.Stderr = &out

	if err := listCmd.Run(); err != nil {
		return errors.Wrapf(err, "list git worktrees in %s", repoRoot)
	}

	if worktreeListed(out.String(), worktreePath) {
		// Worktree already registered; nothing to do.
		return nil
	}

	// Create a new worktree on branch amux/{agent_slug} from current HEAD.
	branchName := fmt.Sprintf("amux/%s", slug)
	addCmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "worktree", "add", "-b", branchName, worktreePath)
	var addOut bytes.Buffer
	addCmd.Stdout = &addOut
	addCmd.Stderr = &addOut

	if err := addCmd.Run(); err != nil {
		return errors.Wrapf(err, "create git worktree %s (output: %s)", worktreePath, addOut.String())
	}

	return nil
}

// worktreeListed checks whether a worktreePath appears in the
// `git worktree list --porcelain` output.
func worktreeListed(output, worktreePath string) bool {
	lines := bytes.Split([]byte(output), []byte("\n"))
	prefix := []byte("worktree ")

	for _, line := range lines {
		if bytes.HasPrefix(line, prefix) {
			path := bytes.TrimSpace(line[len(prefix):])
			if string(path) == worktreePath {
				return true
			}
		}
	}

	return false
}

// LocalSession represents a local PTY-backed agent session.
//
// Phase 2 provides basic PTY ownership for local agents; later phases add
// monitoring and process tracking.
type LocalSession struct {
	Cmd *exec.Cmd
	PTY *os.File
}

// StartLocalSession starts a new local PTY session for the given agent.
// The process working directory is set to agent.Worktree per spec §5.3.1.
func StartLocalSession(ctx context.Context, ag *api.Agent, command []string, env map[string]string) (*LocalSession, error) {
	if ag == nil {
		return nil, errors.Wrap(errors.ErrInvalidInput, "agent must not be nil")
	}
	if len(command) == 0 {
		return nil, errors.Wrap(errors.ErrInvalidInput, "command must not be empty")
	}

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = ag.Worktree

	if env != nil {
		cmd.Env = append(os.Environ(), flattenEnv(env)...)
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "start PTY session")
	}

	return &LocalSession{
		Cmd: cmd,
		PTY: ptmx,
	}, nil
}

// Stop terminates the local session and waits for the process to exit.
func (s *LocalSession) Stop() error {
	if s == nil {
		return nil
	}

	if s.PTY != nil {
		_ = s.PTY.Close()
	}

	if s.Cmd != nil && s.Cmd.Process != nil {
		// Best-effort graceful stop; if it fails, force kill.
		if err := s.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
			_ = s.Cmd.Process.Kill()
		}
		_ = s.Cmd.Wait()
	}

	return nil
}

// RestartLocalSession stops the previous session (if any) and starts a new one.
func RestartLocalSession(ctx context.Context, ag *api.Agent, prev *LocalSession, command []string, env map[string]string) (*LocalSession, error) {
	if prev != nil {
		_ = prev.Stop()
	}
	return StartLocalSession(ctx, ag, command, env)
}

// flattenEnv flattens a map[string]string into KEY=VALUE strings.
func flattenEnv(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
