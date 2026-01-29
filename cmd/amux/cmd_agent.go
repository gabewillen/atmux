package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/agentflare-ai/amux/internal/agent"
	"github.com/agentflare-ai/amux/internal/config"
	"github.com/agentflare-ai/amux/internal/errors"
	"github.com/agentflare-ai/amux/internal/paths"
	"github.com/agentflare-ai/amux/internal/worktree"
	"github.com/agentflare-ai/amux/pkg/api"
)

var adapterName string

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}

var agentAddCmd = &cobra.Command{
	Use:   "add <name> <repo_root>",
	Short: "Add a new agent",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		repoPath := args[1]

		resolver, err := paths.NewResolver()
		if err != nil {
			return err
		}

		repoRoot, err := resolver.CanonicalizeRepoRoot(repoPath)
		if err != nil {
			return errors.Wrapf(err, "invalid repository path %q", repoPath)
		}

		if !worktree.IsRepo(repoRoot) {
			return fmt.Errorf("path %s is not a git repository", repoRoot)
		}

		def := config.AgentDef{
			Name:    name,
			Adapter: adapterName,
			Location: config.AgentLocation{
				Type:     "local",
				RepoPath: repoRoot,
			},
		}

		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		targetConfigRoot := ""
		if projectRoot != "" {
			targetConfigRoot = projectRoot
			cmd.Printf("Adding agent to project config: %s\n", projectRoot)
		} else {
			cmd.Printf("Adding agent to user config (~/.config/amux)\n")
		}

		if err := config.AddAgent(targetConfigRoot, def); err != nil {
			return err
		}

		cmd.Printf("Agent %q added successfully.\n", name)
		return nil
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "NAME\tADAPTER\tLOCATION")
		for _, a := range cfg.Agents {
			loc := a.Location.RepoPath
			if loc == "" {
				loc = a.Location.Host
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", a.Name, a.Adapter, loc)
		}
		return nil
	},
}

var agentRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)

		removed := false
		if projectRoot != "" {
			if err := config.RemoveAgent(projectRoot, name); err == nil {
				cmd.Printf("Removed from project config.\n")
				removed = true
			}
		}

		if !removed {
			if err := config.RemoveAgent("", name); err == nil {
				cmd.Printf("Removed from user config.\n")
				removed = true
			}
		}

		if !removed {
			return fmt.Errorf("agent %q not found", name)
		}
		return nil
	},
}

var agentStartCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start an agent (creates worktree)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		var agentDef *config.AgentDef
		for _, a := range cfg.Agents {
			if a.Name == name {
				agentDef = &a
				break
			}
		}
		if agentDef == nil {
			return fmt.Errorf("agent %q not found", name)
		}

		mgr, err := worktree.NewManager()
		if err != nil {
			return err
		}

		slug := api.NewAgentSlug(name)
		agentData := api.Agent{
			Name:     name,
			Slug:     slug,
			RepoRoot: agentDef.Location.RepoPath,
		}

		wtPath, err := mgr.Ensure(agentData)
		if err != nil {
			return err
		}

		// We need to run the agent.
		// For Phase 2 verification, we'll run it in the foreground and attach IO.

		// Create AgentActor
		a := agent.NewAgent(name, agentDef.Adapter, agentDef.Location.RepoPath, mgr)

		// Start (Async)
		a.Start()

		// Wait for PTY to be ready (hacky poll or need event?)
		// Since Start() is async, a.PtyFile() might be nil immediately.
		// Use a polling loop for Phase 2 strict local.
		cmd.Printf("Waiting for agent to start...\n")
		// In real impl we'd subscribe to events.
		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			if a.PtyFile() != nil {
				break
			}
		}

		ptmx := a.PtyFile()
		if ptmx == nil {
			return fmt.Errorf("agent failed to start (no PTY)")
		}

		cmd.Printf("Agent %q started (worktree: %s)\n", name, wtPath)
		cmd.Printf("Attached. Press Ctrl+C to exit.\n")

		// Handle signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// I/O Copy
		// Make raw terminal? (skip for now to avoid complexity in test wrapper)
		// io.Copy(os.Stdout, ptmx)

		go func() {
			io.Copy(os.Stdout, ptmx)
		}()
		go func() {
			io.Copy(ptmx, os.Stdin)
		}()

		<-sigCh
		cmd.Printf("\nStopping agent...\n")
		a.Stop()
		time.Sleep(500 * time.Millisecond) // Give it time to cleanup
		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentAddCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentRemoveCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentMergeCmd)

	agentAddCmd.Flags().StringVar(&adapterName, "adapter", "default", "Adapter name")
}

var agentMergeCmd = &cobra.Command{
	Use:   "merge <name>",
	Short: "Merge agent worktree back to base branch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()
		projectRoot := findProjectRoot(cwd)
		cfg, err := config.Load(projectRoot)
		if err != nil {
			return err
		}

		var agentDef *config.AgentDef
		for _, a := range cfg.Agents {
			if a.Name == name {
				agentDef = &a
				break
			}
		}
		if agentDef == nil {
			return fmt.Errorf("agent %q not found", name)
		}

		mgr, err := worktree.NewManager()
		if err != nil {
			return err
		}

		slug := api.NewAgentSlug(name)
		agentData := api.Agent{
			Name:     name,
			Slug:     slug,
			RepoRoot: agentDef.Location.RepoPath,
		}

		// Defaults
		strategy := cfg.Git.Merge.Strategy
		if strategy == "" {
			strategy = "squash" // Default per spec/plan
		}
		allowDirty := cfg.Git.Merge.AllowDirty

		cmd.Printf("Merging agent %q using strategy %q (allow_dirty=%v)...\n", name, strategy, allowDirty)
		if err := mgr.MergeAgent(agentData, strategy, allowDirty); err != nil {
			return err
		}

		cmd.Printf("Successfully merged agent %q.\n", name)
		return nil
	},
}

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, ".amux")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
