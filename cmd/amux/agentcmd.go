package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/agentflare-ai/amux/pkg/api"
)

func runAgent(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: amux agent <add|list|remove|start|stop|kill|restart|attach>")
	}
	sub := args[0]
	switch sub {
	case "add":
		return runAgentAdd(args[1:])
	case "list":
		return runAgentList(args[1:])
	case "remove":
		return runAgentRemove(args[1:])
	case "start":
		return runAgentStart(args[1:])
	case "stop":
		return runAgentStop(args[1:])
	case "kill":
		return runAgentKill(args[1:])
	case "restart":
		return runAgentRestart(args[1:])
	case "attach":
		return runAgentAttach(args[1:])
	default:
		return fmt.Errorf("unknown agent command: %s", sub)
	}
}

func runAgentAdd(args []string) error {
	flags := flag.NewFlagSet("agent add", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var name, about, adapter, locType, host, repoPath string
	flags.StringVar(&name, "name", "", "agent name")
	flags.StringVar(&about, "about", "", "agent description")
	flags.StringVar(&adapter, "adapter", "", "adapter name")
	flags.StringVar(&locType, "location", "local", "location type")
	flags.StringVar(&host, "host", "", "ssh host")
	flags.StringVar(&repoPath, "repo-path", "", "repo path")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("agent add: %w", err)
	}
	ctx := context.Background()
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("agent add: %w", err)
	}
	params := map[string]any{
		"name":    name,
		"about":   about,
		"adapter": adapter,
		"cwd":     cwd,
		"location": map[string]any{
			"type":      locType,
			"host":      host,
			"repo_path": repoPath,
		},
	}
	var result struct {
		AgentID string `json:"agent_id"`
	}
	if err := client.Call(ctx, "agent.add", params, &result); err != nil {
		return fmt.Errorf("agent add: %w", err)
	}
	fmt.Fprintln(os.Stdout, result.AgentID)
	return nil
}

func runAgentList(args []string) error {
	_ = args
	ctx := context.Background()
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	var result struct {
		Roster []api.RosterEntry `json:"roster"`
	}
	if err := client.Call(ctx, "agent.list", map[string]any{}, &result); err != nil {
		return fmt.Errorf("agent list: %w", err)
	}
	writer := bufio.NewWriter(os.Stdout)
	for _, entry := range result.Roster {
		id := entry.RuntimeID.String()
		if entry.AgentID != nil {
			id = entry.AgentID.String()
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", entry.Kind, id, entry.Name, entry.Adapter, entry.Presence, entry.RepoRoot)
	}
	return writer.Flush()
}

func runAgentRemove(args []string) error {
	return runAgentRefCommand(args, "agent.remove")
}

func runAgentStart(args []string) error {
	return runAgentRefCommand(args, "agent.start")
}

func runAgentStop(args []string) error {
	return runAgentRefCommand(args, "agent.stop")
}

func runAgentKill(args []string) error {
	return runAgentRefCommand(args, "agent.kill")
}

func runAgentRestart(args []string) error {
	return runAgentRefCommand(args, "agent.restart")
}

func runAgentAttach(args []string) error {
	params, err := parseAgentRefFlags("agent attach", args)
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	var result struct {
		SocketPath string `json:"socket_path"`
	}
	if err := client.Call(ctx, "agent.attach", params, &result); err != nil {
		return fmt.Errorf("agent attach: %w", err)
	}
	conn, err := net.Dial("unix", result.SocketPath)
	if err != nil {
		return fmt.Errorf("agent attach: %w", err)
	}
	defer func() { _ = conn.Close() }()
	go func() {
		_, _ = io.Copy(conn, os.Stdin)
		_ = conn.Close()
	}()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		return fmt.Errorf("agent attach: %w", err)
	}
	return nil
}

func runAgentRefCommand(args []string, method string) error {
	params, err := parseAgentRefFlags(method, args)
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, _, _, err := connectDaemon(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	var result map[string]any
	if err := client.Call(ctx, method, params, &result); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}

func parseAgentRefFlags(name string, args []string) (map[string]any, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var id, agentName string
	flags.StringVar(&id, "id", "", "agent id")
	flags.StringVar(&agentName, "name", "", "agent name")
	if err := flags.Parse(args); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if strings.TrimSpace(id) == "" && strings.TrimSpace(agentName) == "" {
		return nil, fmt.Errorf("%s: --id or --name required", name)
	}
	params := map[string]any{}
	if strings.TrimSpace(id) != "" {
		params["agent_id"] = id
	}
	if strings.TrimSpace(agentName) != "" {
		params["name"] = agentName
	}
	return params, nil
}
