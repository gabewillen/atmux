// Package main implements the amux CLI client.
// agent.go implements amux agent add (spec §5.2).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/agentflare-ai/amux/pkg/api"
)

func runAgent(args []string) error {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: amux agent <add|list|remove> [options]\n")
		return nil
	}
	sub := args[0]
	switch sub {
	case "add":
		return runAgentAdd(args[1:])
	default:
		return fmt.Errorf("unknown agent subcommand: %s", sub)
	}
}

func runAgentAdd(args []string) error {
	fs := flag.NewFlagSet("amux agent add", flag.ExitOnError)
	name := fs.String("name", "", "Agent name (required)")
	about := fs.String("about", "", "Agent description")
	adapter := fs.String("adapter", "", "Adapter name (required)")
	repoPath := fs.String("repo-path", "", "Repository root path (default: current directory's git root)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	repoRoot, err := agent.ResolveRepoRoot(homeDir, cwd, *repoPath)
	if err != nil {
		return fmt.Errorf("repo root: %w", err)
	}

	loc := config.AgentLocationConfig{Type: "local"}
	if *repoPath != "" {
		loc.RepoPath = *repoPath
	}
	in, err := agent.ValidateAddInput(repoRoot, *name, *about, *adapter, loc)
	if err != nil {
		return err
	}

	// Load existing agents to uniquify slug
	proj, err := config.LoadProjectFile(repoRoot)
	if err != nil {
		return fmt.Errorf("load project config: %w", err)
	}
	existing := make(map[string]struct{})
	for _, a := range proj.Agents {
		slug := a.Slug
		if slug == "" {
			slug = api.NormalizeAgentSlug(a.Name)
		}
		existing[slug] = struct{}{}
	}
	baseSlug := api.NormalizeAgentSlug(in.Name)
	agentSlug := api.UniquifyAgentSlug(baseSlug, existing)

	// Create worktree (spec §5.2 step 5)
	wtPath, err := worktree.Create(repoRoot, agentSlug)
	if err != nil {
		return fmt.Errorf("create worktree: %w", err)
	}
	_ = wtPath

	// Persist agent config (spec §5.2)
	ac := agent.BuildAgentConfig(in, agentSlug)
	if err := config.AddAgentToProject(repoRoot, ac); err != nil {
		return fmt.Errorf("persist agent config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Added agent %q (slug %s)\n", in.Name, agentSlug)
	return nil
}
