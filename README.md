# Agent TMUX (**atmux**)

**atmux** is a tmux-first toolkit for running and coordinating multiple AI coding agents in parallel. It handles session lifecycle, inter-agent messaging, work assignment, output capture, and scheduling — across different AI CLIs (Claude Code, Gemini, Codex, Cursor) and git repos.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/gabrielwillen/atmux/main/install.sh | sh
```

Or from a local checkout:

```sh
./install.sh
```

This installs `atmux` to `~/.atmux/bin` and adds it to your shell `PATH`. Slash commands for Claude Code, Gemini CLI, and Codex are installed by default (pass `--no-slash-commands` to skip).

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

Each agent runs in a named tmux session: `atmux-<repo>-<agent>`. Agents get their own git worktree at `~/.atmux/agents/<repo>-<name>`, keeping their changes isolated.

### Teams

Agents can be grouped into teams (capped at 4 per team). Teams share a layout and can be messaged collectively.

```sh
atmux create --team platform
atmux create --agent reviewer --role reviewer --team platform --intelligence 80
atmux create --agent tester   --role tester   --team platform --intelligence 55
```

### Intelligence scale

The `--intelligence 0–100` flag selects a model and reasoning level automatically via the adapter's manifest. Higher values use more capable (and slower/costlier) models.

| Range  | Example mapping (Claude Code)  |
|--------|-------------------------------|
| 0–39   | sonnet + low reasoning        |
| 40–74  | sonnet + medium reasoning     |
| 75–89  | sonnet + high reasoning       |
| 90–100 | opus + high reasoning         |

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
  [--team <team>] [--adapter <adapter>] [-- <adapter-args...>]

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

### `schedule`

Send a message after a delay or on a repeating interval.

```sh
atmux schedule --to <name> --once <duration>     "message"
atmux schedule --to <name> --interval <duration> "message"
```

Duration suffixes: `ms`, `s`, `m`, `h`, `d`.

### `exec`

Run a shell command with tracked exit status. Sends an ATMUX notification when the process finishes.

```sh
atmux exec -- <command> [args...]
```

### `session`

```sh
atmux session list
atmux session start [--name <name>] [--adapter <adapter>] [-- <adapter-args...>]
atmux session attach <name|session>
```

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

### `agent destroy`

```sh
atmux agent destroy <session-pattern> [pattern...]
# e.g. atmux agent destroy 'atmux-myrepo-worker-*'
```

## Environment variables

| Variable           | Description                                   |
|--------------------|-----------------------------------------------|
| `ATMUX_HOME`       | Installation root (default: `~/.atmux`)       |
| `ATMUX_REPO`       | Repository name for the current session       |
| `ATMUX_AGENT_NAME` | Current agent's name                          |
| `ATMUX_MANAGER`    | Parent manager agent name                     |
| `ATMUX_WORKTREE`   | Git worktree path for this agent              |
| `ATMUX_TEAM`       | Team this agent belongs to                    |
| `ATMUX_SESSION_ID` | Unique session identifier                     |
| `ATMUX_SESSION_KIND` | `agent` or `team`                           |

## Agent coordination rules

Agents running inside `atmux` sessions are expected to:

- Acknowledge manager messages quickly with a short plan.
- Report completion with validation evidence (not just "done").
- Message their manager when stuck: `atmux send --to "$ATMUX_MANAGER" "..."`
- Escalate blockers immediately — never leave them unreported.
- Reuse idle agents before spawning new ones.
- Check `atmux list teams` before creating new team members.
- Never silently change scope — ask the manager first.

## Docs

- `docs/atmux.md` — full command and session reference
- `docs/agent.md` — agent and worktree concepts
- `docs/architecture.md` — internal architecture overview
- `docs/cli.md` — CLI implementation notes
