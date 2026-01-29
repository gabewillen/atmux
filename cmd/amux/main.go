// Package main implements the amux CLI client per spec §12.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/stateforward/amux/internal/agent"
	"github.com/stateforward/amux/internal/config"
	"github.com/stateforward/amux/internal/errors"
	"github.com/stateforward/amux/internal/paths"
	"github.com/stateforward/amux/internal/snapshot"
	"github.com/stateforward/amux/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "amux: command required")
		fmt.Fprintln(os.Stderr, "Usage: amux <command> [args...]")
		os.Exit(1)
	}

	// Phase 0/2: CLI stub with test and basic agent management
	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("amux v0.1.0-phase0")
	case "test":
		handleTestCommand()
	case "agent":
		handleAgentCommand()
	default:
		fmt.Fprintf(os.Stderr, "amux: unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleTestCommand() {
	// Parse flags
	testFlags := flag.NewFlagSet("test", flag.ExitOnError)
	regression := testFlags.Bool("regression", false, "Compare against previous snapshot")
	noSnapshot := testFlags.Bool("no-snapshot", false, "Write snapshot to stdout")
	testFlags.Parse(os.Args[2:])

	moduleRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: get working directory: %v\n", err)
		os.Exit(1)
	}

	// Create snapshot
	fmt.Fprintln(os.Stderr, "Running amux test suite...")
	snap, err := snapshot.Create(moduleRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: create snapshot: %v\n", err)
		os.Exit(1)
	}

	// Helper to determine if any step failed.
	anyFailed := func() bool {
		statuses := []string{snap.TidyStatus, snap.VetStatus, snap.LintStatus, snap.TestStatus, snap.RaceStatus}
		for _, s := range statuses {
			if s == "fail" {
				return true
			}
		}
		return false
	}

	if *regression {
		// Regression mode: compare with latest snapshot
		latestPath, err := snapshot.FindLatestSnapshot(moduleRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: find baseline snapshot: %v\n", err)
			os.Exit(1)
		}

		baseline, err := snapshot.Read(latestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: read baseline snapshot: %v\n", err)
			os.Exit(1)
		}

		passed, report := snapshot.Compare(baseline, snap)
		fmt.Fprintln(os.Stderr, "\nRegression report:")
		fmt.Fprintln(os.Stderr, report)

		if !passed {
			fmt.Fprintln(os.Stderr, "\nREGRESSION DETECTED")
			os.Exit(1)
		}

		fmt.Fprintln(os.Stderr, "\nNo regressions detected")

		// Write new snapshot
		outPath := snapshot.GenerateSnapshotPath(moduleRoot)
		if err := snapshot.Write(snap, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write snapshot: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Snapshot written: %s\n", outPath)
	} else if *noSnapshot {
		// No-snapshot mode: write TOML snapshot to stdout; logs remain on stderr.
		data, err := toml.Marshal(snap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: marshal snapshot: %v\n", err)
			os.Exit(1)
		}
		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write snapshot to stdout: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Normal mode: write snapshot
		outPath := snapshot.GenerateSnapshotPath(moduleRoot)
		if err := snapshot.Write(snap, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: write snapshot: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Snapshot written: %s\n", outPath)
	}

	if anyFailed() {
		fmt.Fprintln(os.Stderr, "One or more verification steps failed; see logs above.")
		os.Exit(1)
	}
}

func handleAgentCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "amux agent: subcommand required")
		fmt.Fprintln(os.Stderr, "Usage: amux agent <add> [args...]")
		os.Exit(1)
	}

	sub := os.Args[2]
	switch sub {
	case "add":
		handleAgentAddCommand()
	default:
		fmt.Fprintf(os.Stderr, "amux agent: unknown subcommand: %s\n", sub)
		os.Exit(1)
	}
}

func handleAgentAddCommand() {
	addFlags := flag.NewFlagSet("agent add", flag.ExitOnError)
	name := addFlags.String("name", "", "Agent name (required)")
	about := addFlags.String("about", "", "Agent description")
	adapter := addFlags.String("adapter", "", "Adapter name (e.g., claude-code)")
	repo := addFlags.String("repo", "", "Repository root (optional, defaults to current git repo)")

	addFlags.Parse(os.Args[3:])

	if strings.TrimSpace(*name) == "" {
		fmt.Fprintln(os.Stderr, "Error: -name is required")
		os.Exit(1)
	}

	if strings.TrimSpace(*adapter) == "" {
		fmt.Fprintln(os.Stderr, "Error: -adapter is required")
		os.Exit(1)
	}

	repoRoot := strings.TrimSpace(*repo)
	var err error
	if repoRoot == "" {
		// Discover repo root from current working directory.
		repoRoot, err = discoverGitRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: determine git repo root: %v\n", err)
			os.Exit(1)
		}
	}

	ctx := context.Background()

	// Load user + project configuration per spec §4.2.8.2 hierarchy.
	resolver, err := paths.NewResolver(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: create path resolver: %v\n", err)
		os.Exit(1)
	}

	userConfigPath, err := paths.UserConfigFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: determine user config path: %v\n", err)
		os.Exit(1)
	}

	projectConfigPath := resolver.ProjectConfig()
	cfg, err := config.Load(userConfigPath, projectConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: load project config: %v\n", err)
		os.Exit(1)
	}

	agentOpts := agent.AddLocalAgentOptions{
		Name:    *name,
		About:   *about,
		Adapter: *adapter,
		RepoRoot: repoRoot,
	}

	ag, slug, err := agent.AddLocalAgent(ctx, cfg, agentOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: add agent: %v\n", err)
		os.Exit(1)
	}

	if err := config.Save(projectConfigPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: save project config: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Added agent %q (slug %q) for repo %s\n", ag.Name, slug, ag.RepoRoot)
}

// discoverGitRepoRoot determines the git repository root for the current working directory.
// It runs `git rev-parse --show-toplevel` and canonicalizes the result per spec §3.23.
func discoverGitRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(errors.ErrInvalidInput, "current directory is not inside a git repository (%s)", out.String())
	}

	root := strings.TrimSpace(out.String())

	canonical, err := api.CanonicalizeRepoRoot(root)
	if err != nil {
		return "", errors.Wrap(err, "canonicalize repo root")
	}

	return canonical, nil
}
