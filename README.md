# Agent TMUX (**atmux**)

**atmux** is a tmux-first toolkit for running and coordinating multiple AI coding agents in parallel. It handles session lifecycle, inter-agent messaging, work assignment, output capture, and notifications — across different AI CLIs (Claude Code, Gemini, Codex, Cursor) and git repos.

## No frills. No dependencies. No build.

- **No build step.** Pure shell scripts — nothing to compile, bundle, or transpile. `git clone` and run.
- **No runtime.** No Node, Python, Go, or Rust toolchain required. No package manager, no lockfiles, no `node_modules`.
- **No dependencies** beyond what's already on every Unix box you'd run an agent on: `bash`, `git`, and `tmux`. That's it.
- **No daemon, no server, no database.** State lives in plain files under `ATMUX_HOME` (`<project>/.atmux` by default). Inspect it with `ls` and `cat`.
- **No frills.** One shell script per command. Read the source, patch it in place, move on.

Install by piping `curl` into `sh`, or clone the repo and run `./install.sh`. The installer defaults to a project-local install so it does not modify system-level agent config unless you choose `--system`. Uninstall by deleting the install directory.

## What makes it different

- **CLI-agnostic via adapters.** Run Claude Code, Gemini, Codex, and Cursor side-by-side in the same session. Swap vendors without rewriting your workflow. Third-party adapters install with `atmux adapter install owner/repo`.
- **Intelligence scale, not model names.** Say `--intelligence 80` and the adapter picks the right model and reasoning level. Portable across vendors, survives model renames — no more hardcoding `claude-opus-4-7` or `gpt-5-codex` across your scripts.
- **You can actually see the agents work.** It's just tmux. Attach to any session, watch the agent think in real time, detach and come back later. No custom TUI, no web dashboard, no log tailing.
- **Git worktree per agent.** Each agent gets its own branch and working directory under `ATMUX_HOME/agents/`. Parallel agents can't stomp each other's changes, and cleanup is a single `atmux kill --agent`.

> **Experimental** — this project is under active development. APIs, commands, and behavior may change without notice. Use at your own risk.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/gabrielwillen/atmux/main/install.sh | sh
```

Or from a local checkout:

```sh
./install.sh
```

By default this installs `atmux` into `<project>/.atmux`, installs Claude Code/Gemini/Codex commands under the project-local `.claude`, `.gemini`, and `.codex` directories, and leaves shell profiles untouched. The project `.atmux/.gitignore` keeps runtime state out of git while allowing the launcher and installed source to be committed. Add the project launcher to your shell when you want to use it:

```sh
export PATH="$PWD/.atmux/bin:$PATH"
```

Use `./install.sh --system` for the legacy user-level install into `~/.atmux` and user-level CLI command directories. Slash commands are installed by default in either scope; pass `--no-slash-commands` to skip them.

Then launch a tmux session with `atmux` on the path:

```sh
atmux
```

## Quick start

```sh
# Create an agent with a role and intelligence level (0–100)
atmux create --agent planner --role planner --intelligence 80

# Send it a message
atmux send --to planner "analyze the codebase and create a task list"

# Capture its output
atmux capture --agent planner

# Assign structured work
atmux assign --to planner --title "stabilize parser" \
  --todo "write failing test" \
  --todo "fix root cause" \
  --todo "verify green"
```

## Concepts

### Sessions and agents

Each agent runs in a named tmux session: `atmux-<repo>-<agent>`. By default, agents get their own git worktree at `ATMUX_HOME/agents/<repo>-<name>`, keeping their changes isolated. Worktree creation initializes submodules with `git submodule update --init --recursive`. Pass `--no-worktree` to skip worktree creation and run in the repo root instead.

### Teams

Agents can be grouped into teams (capped at 4 per team). Teams share a layout and can be messaged collectively.

```sh
atmux create --team platform
atmux create --agent reviewer --role reviewer --team platform --intelligence 80
atmux create --agent tester   --role tester   --team platform --intelligence 55
```

### Intelligence scale

The `--intelligence 0–100` flag selects a model and reasoning level automatically via the adapter's manifest. Higher values use more capable (and slower/costlier) models.

| Adapter        | Intelligence | Model                  | Reasoning level |
|----------------|--------------|------------------------|-----------------|
| `claude-code`  | 0–39         | `sonnet`               | `low`           |
| `claude-code`  | 40–74        | `sonnet`               | `medium`        |
| `claude-code`  | 75–89        | `sonnet`               | `high`          |
| `claude-code`  | 90–100       | `opus`                 | `high`          |
| `codex`        | 0–29         | `gpt-5.4`              | `low`           |
| `codex`        | 30–59        | `gpt-5.4`              | `medium`        |
| `codex`        | 60–84        | `gpt-5.4`              | `high`          |
| `codex`        | 85–100       | `gpt-5.4`              | `extra-high`    |
| `cursor-agent` | 0–39         | `composer-2-fast`      | `low`           |
| `cursor-agent` | 40–74        | `composer-2`           | `medium`        |
| `cursor-agent` | 75–89        | `gpt-5.3-codex-high`  | `high`          |
| `cursor-agent` | 90–100       | `gpt-5.3-codex-xhigh` | `extra-high`    |
| `gemini`       | 0–39         | `gemini-3.1-flash-lite-preview` | `low`  |
| `gemini`       | 40–74        | `gemini-3-flash-preview` | `medium`      |
| `gemini`       | 75–89        | `gemini-3.1-pro-preview` | `medium`     |
| `gemini`       | 90–100       | `gemini-3.1-pro-preview` | `high`       |

### Adapters

Adapters translate `atmux` commands to a specific AI CLI. The default adapter is `auto` (auto-detected). Available adapters:

| Adapter        | CLI           |
|----------------|---------------|
| `claude-code`  | Claude Code   |
| `gemini`       | Gemini CLI    |
| `codex`        | OpenAI Codex  |
| `cursor-agent` | Cursor AI     |

```sh
# Use a specific adapter
atmux create --agent worker --role implementer --adapter claude-code --intelligence 60 \
  -- --dangerously-skip-permissions
```

Install a third-party adapter:

```sh
atmux adapter install owner/repo
```

## Command reference

### `create`

```sh
atmux create --agent <name> --role <role> --intelligence <0-100> \
  [--team <team>] [--adapter <adapter>] [--no-worktree] [-- <adapter-args...>]

atmux create --team <name>

atmux create --issue --title <title> [--description <desc>] [--todo <todo>]...
```

### `send`

Send a message to an agent or every agent in a team.

```sh
atmux send --to <name|session> [--reply-required] "message"
```

`--reply-required` signals that the sender is blocked until the recipient responds.

### `assign`

Create and assign filesystem-tracked issues.

```sh
atmux assign --to <agent> --title <title> [--description <desc>] [--todo <todo>]...
atmux assign --issue <uuid> --to <agent>
```

### `capture`

Read tmux pane output from one or more agents.

```sh
atmux capture --agent <name>  [--lines <n>]
atmux capture --team <name>   [--lines <n>]
atmux capture --all           [--lines <n>]
```

### `exec`

Run a shell command with tracked exit status. Sends an ATMUX notification when the process finishes.

```sh
atmux exec [--detach] -- <command> [args...]
```

`--detach` runs the command in a new tmux window and returns immediately so the agent stays unblocked. Watchers can monitor the process via `watch --pid <pid> --stdio`.

### `watch`

Wait for a condition: process exit, pane text, output changes, issue updates, or agent idle state.

```sh
atmux watch --pid <pid> [--timeout <seconds>]
atmux watch --pid <pid> --stdio [--duration <seconds>] [--timeout <seconds>]
atmux watch --target <tmux-target> --text <needle> [--scope pane|window|session]
atmux watch --issue <id> [--timeout <seconds>]
atmux watch --agent <name> [--idle <seconds>] [--timeout <seconds>]
```

### `schedule`

Schedule a future or repeating action. Runs detached by default.

```sh
atmux schedule --once <duration> --notification "check on training"
atmux schedule --interval <duration> --notification "heartbeat"
atmux schedule --once <duration> -- atmux send --to <name> "message"
```

`--no-detach` runs in the foreground (blocking). Duration suffixes: `ms`, `s`, `m`, `h`, `d`.

### `kill`

Stop exec-tracked processes or remove agent sessions.

```sh
atmux kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]
atmux kill --agent <name|pattern> [name|pattern...]
```

`--pid` stops an exec process, notifies watchers, and cleans up metadata.
`--agent` kills agent sessions and removes their worktrees and branches.

### `session`

```sh
atmux session list
atmux session start [--name <name>] [--adapter <adapter>] [-- <adapter-args...>]
atmux session attach <name|session>
```

`atmux session attach` must be run outside tmux.

### `list`

```sh
atmux list agents
atmux list sessions
atmux list teams
atmux list issues
```

### `env`

```sh
atmux env            # show all ATMUX_* variables
atmux env get <key>  # get a single variable
```


## Environment variables

| Variable           | Description                                   |
|--------------------|-----------------------------------------------|
| `ATMUX_HOME`       | Installation/state root (default: `<project>/.atmux` for project installs, `~/.atmux` for system installs) |
| `ATMUX_REPO`       | Repository name for the current session       |
| `ATMUX_AGENT_NAME` | Current agent's name                          |
| `ATMUX_MANAGER`    | Parent manager agent name                     |
| `ATMUX_WORKTREE`   | Working directory (worktree or repo root)     |
| `ATMUX_TEAM`       | Team this agent belongs to                    |
| `ATMUX_SESSION_ID` | Unique session identifier                     |
| `ATMUX_SESSION_KIND` | `agent` or `team`                           |

## Agent coordination rules

Agents running inside `atmux` sessions are expected to:

- Acknowledge manager messages quickly with a short plan.
- Report completion with validation evidence (not just "done").
- Message their manager when stuck: `atmux send --to <manager> "..."`
- Escalate blockers immediately — never leave them unreported.
- Reuse idle agents before spawning new ones.
- Check `atmux list teams` before creating new team members.
- Never silently change scope — ask the manager first.

## Docs

- `docs/atmux.md` — full command and session reference
- `docs/agent.md` — agent and worktree concepts
- `docs/architecture.md` — internal architecture overview
- `docs/cli.md` — CLI implementation notes
