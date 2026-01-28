// Package cli implements the amux CLI client command handling.
//
// The CLI communicates with the amux daemon over JSON-RPC 2.0 via a Unix
// socket. This package provides the command parsing, dispatching, and
// output formatting for all CLI commands.
//
// See spec §12 for the full CLI specification.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/agentflare-ai/amux/internal/cli/test"
	"github.com/agentflare-ai/amux/pkg/api"
)

// Version is the CLI version string.
const Version = "0.1.0-dev"

// Run executes the CLI with the given arguments.
func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return showHelp()
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "help", "--help", "-h":
		return showHelp()

	case "version", "--version", "-v":
		return showVersion()

	case "test":
		return test.Run(ctx, cmdArgs)

	case "agent":
		return runAgent(ctx, cmdArgs)

	case "plugin":
		return runPlugin(ctx, cmdArgs)

	default:
		return fmt.Errorf("unknown command: %s\n\nRun 'amux help' for usage", cmd)
	}
}

func showHelp() error {
	help := `amux - Agent Multiplexer CLI

Usage:
  amux <command> [options]

Commands:
  agent     Manage agents
  plugin    Manage plugins
  test      Run Go verification suite
  version   Show version information
  help      Show this help

Run 'amux <command> --help' for more information on a command.

Spec version: ` + api.SpecVersion

	fmt.Println(help)
	return nil
}

func showVersion() error {
	fmt.Printf("amux version %s\n", Version)
	fmt.Printf("spec version %s\n", api.SpecVersion)
	return nil
}

func runAgent(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return showAgentHelp()
	}

	cmd := args[0]

	switch cmd {
	case "help", "--help", "-h":
		return showAgentHelp()

	case "add":
		return agentAdd(ctx, args[1:])

	case "list", "ls":
		return agentList(ctx, args[1:])

	case "remove", "rm":
		return agentRemove(ctx, args[1:])

	case "start":
		return agentStart(ctx, args[1:])

	case "stop":
		return agentStop(ctx, args[1:])

	default:
		return fmt.Errorf("unknown agent command: %s", cmd)
	}
}

func showAgentHelp() error {
	help := `amux agent - Manage agents

Usage:
  amux agent <command> [options]

Commands:
  add       Add a new agent
  list      List agents
  remove    Remove an agent
  start     Start an agent
  stop      Stop an agent
`
	fmt.Println(help)
	return nil
}

func agentAdd(ctx context.Context, args []string) error {
	fmt.Println("agent add: not yet implemented")
	return nil
}

func agentList(ctx context.Context, args []string) error {
	fmt.Println("agent list: not yet implemented")
	return nil
}

func agentRemove(ctx context.Context, args []string) error {
	fmt.Println("agent remove: not yet implemented")
	return nil
}

func agentStart(ctx context.Context, args []string) error {
	fmt.Println("agent start: not yet implemented")
	return nil
}

func agentStop(ctx context.Context, args []string) error {
	fmt.Println("agent stop: not yet implemented")
	return nil
}

func runPlugin(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return showPluginHelp()
	}

	cmd := args[0]

	switch cmd {
	case "help", "--help", "-h":
		return showPluginHelp()

	case "install":
		return pluginInstall(ctx, args[1:])

	case "list", "ls":
		return pluginList(ctx, args[1:])

	case "remove", "rm":
		return pluginRemove(ctx, args[1:])

	default:
		return fmt.Errorf("unknown plugin command: %s", cmd)
	}
}

func showPluginHelp() error {
	help := `amux plugin - Manage CLI plugins

Usage:
  amux plugin <command> [options]

Commands:
  install   Install a plugin
  list      List installed plugins
  remove    Remove a plugin
`
	fmt.Println(help)
	return nil
}

func pluginInstall(ctx context.Context, args []string) error {
	fmt.Println("plugin install: not yet implemented")
	return nil
}

func pluginList(ctx context.Context, args []string) error {
	fmt.Println("plugin list: not yet implemented")
	return nil
}

func pluginRemove(ctx context.Context, args []string) error {
	fmt.Println("plugin remove: not yet implemented")
	return nil
}

// ParseFlags is a simple flag parser for CLI commands.
func ParseFlags(args []string) (flags map[string]string, positional []string) {
	flags = make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			// Long flag
			key := arg[2:]
			if idx := strings.Index(key, "="); idx >= 0 {
				flags[key[:idx]] = key[idx+1:]
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags[key] = args[i+1]
				i++
			} else {
				flags[key] = "true"
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			// Short flag
			key := arg[1:]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags[key] = args[i+1]
				i++
			} else {
				flags[key] = "true"
			}
		} else {
			positional = append(positional, arg)
		}
	}

	return flags, positional
}

// GetFlag gets a flag value with a default.
func GetFlag(flags map[string]string, names []string, defaultValue string) string {
	for _, name := range names {
		if v, ok := flags[name]; ok {
			return v
		}
	}
	return defaultValue
}

// HasFlag checks if any of the flag names are set.
func HasFlag(flags map[string]string, names ...string) bool {
	for _, name := range names {
		if _, ok := flags[name]; ok {
			return true
		}
	}
	return false
}

// Stdout returns the standard output writer.
func Stdout() *os.File {
	return os.Stdout
}

// Stderr returns the standard error writer.
func Stderr() *os.File {
	return os.Stderr
}
