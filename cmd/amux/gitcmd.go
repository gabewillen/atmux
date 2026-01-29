package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func runGit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: amux git <merge>")
	}
	sub := args[0]
	switch sub {
	case "merge":
		return runGitMerge(args[1:])
	default:
		return fmt.Errorf("unknown git command: %s", sub)
	}
}

func runGitMerge(args []string) error {
	flags := flag.NewFlagSet("git merge", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var id, name, strategy, targetBranch string
	flags.StringVar(&id, "id", "", "agent id")
	flags.StringVar(&name, "name", "", "agent name")
	flags.StringVar(&strategy, "strategy", "", "merge strategy")
	flags.StringVar(&targetBranch, "target-branch", "", "target branch")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("git merge: %w", err)
	}
	if strings.TrimSpace(id) == "" && strings.TrimSpace(name) == "" {
		return fmt.Errorf("git merge: --id or --name required")
	}
	params := map[string]any{}
	if strings.TrimSpace(id) != "" {
		params["agent_id"] = id
	}
	if strings.TrimSpace(name) != "" {
		params["name"] = name
	}
	if strings.TrimSpace(strategy) != "" {
		params["strategy"] = strategy
	}
	if strings.TrimSpace(targetBranch) != "" {
		params["target_branch"] = targetBranch
	}
	ctx := context.Background()
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	var result map[string]any
	if err := client.Call(ctx, "git.merge", params, &result); err != nil {
		return fmt.Errorf("git merge: %w", err)
	}
	fmt.Fprintln(os.Stdout, "merge completed")
	return nil
}
