# atmux

**atmux** is a tmux-first toolkit for running multiple AI coding agents in one repo. It manages agents, isolated worktrees, messaging, work assignment, output capture, and notifications across Claude Code, Gemini, Codex, and Cursor.

The default install is project-local: atmux lives in `<project>/.atmux`, its source is plain shell, and agents can patch the tool in the same repo they are working on. There is no daemon, server, database, build step, or language runtime beyond ordinary Unix tools (`bash`, `git`, `tmux`).

## Why Use It

- **Visible agents.** Every agent runs in tmux, so you can attach, inspect, and kill it with familiar tools.
- **Isolated work.** Agents get their own git worktree by default, so parallel edits do not collide.
- **Adapter portable.** The same workflow can run Claude Code, Gemini, Codex, Cursor, or a third-party adapter.
- **Role and team aware.** Roles define reusable prompts/hooks; teams spawn coordinated member agents.
- **Agent-editable.** A project install keeps atmux source next to your code, so missing workflow glue can be patched locally.

> **Experimental** — this project is under active development (current version: `0.23.1`). APIs, commands, and behavior may change without notice. Use at your own risk.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/gabewillen/atmux/main/install.sh | sh
```

Or from a local checkout:

```sh
./install.sh
```

By default this installs `atmux` into `<project>/.atmux`, installs project-local Claude/Gemini/Codex commands, and leaves shell profiles untouched. Add the project launcher to your shell when you want to use it:

```sh
export PATH="$PWD/.atmux/bin:$PATH"
```

Use `./install.sh --system` for a user-level install in `~/.atmux`. Slash commands are installed by default in either scope; pass `--no-slash-commands` to skip them.

Then start atmux:

```sh
atmux
```

## Quick start

```sh
# Create an agent with a role and intelligence level (0–100)
atmux agent create planner --role planner --intelligence 80

# Send it a message
atmux send --to planner "analyze the codebase and create a task list"

# Capture its output
atmux agent capture planner

# Assign structured work
atmux issue create --title "stabilize parser" --assign-to planner \
  --todo "write failing test" \
  --todo "fix root cause" \
  --todo "verify green"
```

## Core Concepts

### Agents

Each agent is a named worker running an AI CLI in tmux. By default, `atmux agent create` gives it a private worktree under `ATMUX_HOME/agents/`. Pass `--shared-worktree` when an agent should work in the caller's current checkout.

### Teams

Teams group any number of tmux-backed agents in filesystem state under `ATMUX_HOME`. A team role can spawn member agents, wire watchers, apply a time limit, and share a team worktree. `atmux team view <team>` creates an optional multiagent tmux view when you want to watch the members together.

```sh
atmux team create platform
atmux agent create reviewer --role reviewer --team platform --intelligence 80
atmux agent create tester   --role tester   --team platform --intelligence 55
```

### Built-in roles

#### Built-in agent roles

| Role | Description | Demo |
|------|-------------|------|
| [`gh-pr-reviewer`](roles/agents/gh-pr-reviewer/README.md) | The `gh-pr-reviewer` role reviews GitHub pull requests, looks for concrete risks in the diff, and posts structured review feedback with `gh`. |  |

#### Built-in team roles

| Role | Description | Members | Demo |
|------|-------------|---------|------|
| [`collab`](roles/teams/collab/README.md) | Create a multi-agent deliberation team: | [`arbiter`](roles/teams/collab/roles/arbiter/role.md): You are the Arbitrator for collab team `${ATMUX_TEAM}`.<br>[`collaborator`](roles/teams/collab/roles/collaborator/role.md): You are a substantive collaborator in collab team `${ATMUX_TEAM}`.<br>[`leader`](roles/teams/collab/roles/leader/role.md): You are the Leader for collab team `${ATMUX_TEAM}`.<br>[`recorder`](roles/teams/collab/roles/recorder/role.md): You are the Recorder for collab team `${ATMUX_TEAM}`. |  |
| [`pair-program`](roles/teams/pair-program/README.md) | A driver-and-navigator team role where a fast model writes code while a stronger model watches the shared worktree and interrupts with review notes when the implementation drifts. | [`driver`](roles/teams/pair-program/roles/driver/README.md): The `driver` role is the fast implementation half of the pair-programming workflow. It writes code, runs tests, and responds to navigator feedback.<br>[`navigator`](roles/teams/pair-program/roles/navigator/README.md): The `navigator` role is the review half of the pair-programming workflow. It watches a shared worktree, reviews rolling diffs, and steers the driver without editing files directly. | [demo](roles/teams/pair-program/demo.gif) |

### Intelligence scale

The `--intelligence 0–100` flag selects a model and reasoning level via the adapter manifest. Higher values use more capable and usually slower/costlier settings.

| Adapter | Intelligence | Model | Reasoning level |
|---------|--------------|-------|-----------------|
| `claude-code` | 0–39 | `sonnet` | `low` |
| `claude-code` | 40–74 | `sonnet` | `medium` |
| `claude-code` | 75–89 | `sonnet` | `high` |
| `claude-code` | 90–100 | `opus` | `high` |
| `codex` | 0–29 | `gpt-5.5` | `low` |
| `codex` | 30–59 | `gpt-5.5` | `medium` |
| `codex` | 60–84 | `gpt-5.5` | `high` |
| `codex` | 85–100 | `gpt-5.5` | `extra-high` |
| `cursor-agent` | 0–39 | `composer-2-fast` | `low` |
| `cursor-agent` | 40–74 | `composer-2` | `medium` |
| `cursor-agent` | 75–89 | `gpt-5.3-codex-high` | `high` |
| `cursor-agent` | 90–100 | `gpt-5.3-codex-xhigh` | `extra-high` |
| `gemini` | 0–39 | `gemini-3.1-flash-lite-preview` | `low` |
| `gemini` | 40–74 | `gemini-3-flash-preview` | `medium` |
| `gemini` | 75–89 | `gemini-3.1-pro-preview` | `medium` |
| `gemini` | 90–100 | `gemini-3.1-pro-preview` | `high` |

### Adapters

Adapters translate `atmux` commands to a specific AI CLI. The default adapter is `auto` (auto-detected). Available adapters:

| Adapter | CLI |
|---------|-----|
| `claude-code` | Claude Code |
| `codex` | OpenAI Codex |
| `cursor-agent` | Cursor AI |
| `gemini` | Gemini CLI |

```sh
# Use a specific adapter
atmux agent create worker --role implementer --adapter claude-code --intelligence 60 \
  -- --dangerously-skip-permissions
```

Install a third-party adapter:

```sh
atmux adapter install owner/repo
```

## Command Map

Most commands are resource-first: `atmux <noun> <verb> [args]`. Cross-cutting verbs such as `send`, `exec`, `schedule`, `update`, and `install` are direct. Run `atmux <command> --help` for complete options.

The top-level help is generated from `bin/atmux`:

```text
Usage:
  atmux <resource> <verb> [args...]
  atmux <verb> [args...]

tmux-first orchestration for AI coding agents.

Commands:
  adapter    Install adapters and run adapter contract commands.
  agent      Manage repo-scoped AI agents running under tmux.
  config     Read and write atmux configuration values.
  env        Inspect ATMUX_* environment variables in the current process.
  exec       Execute a command with passthrough stdio and unchanged exit behavior.
  install    Install atmux for a project by default, or system-wide when requested.
  issue      Repo-scoped issue tickets on filesystem.
  message    Read, list, or subscribe to filesystem-backed messages.
  notify     Send an ATMUX XML notification to a tmux pane.
  pane       Watch tmux pane text.
  path       Operate on filesystem paths matched by a glob.
  pr         Repo-scoped pull request tickets on filesystem.
  process    Watch or stop an atmux exec-tracked child process.
  role       Manage adapter-agnostic roles.
  schedule   Schedule a future or repeating action.
  send       Send XML messages to a single agent or every agent in a team.
  team       Manage repo-scoped teams of agents.
  update     Update atmux from the latest GitHub release.
  watcher    List or remove background watcher registrations.

Run:
  atmux <command> --help
```

### Common Commands

| Command | Purpose |
|---------|---------|
| `atmux adapter install <owner/repo\|github-url>` | Install adapters and run adapter contract commands. |
| `atmux agent create [name] --role <role> --intelligence <0-100>` | Manage repo-scoped AI agents running under tmux. |
| `atmux config get <key> [--global]` | Read and write atmux configuration values. |
| `atmux env` | Inspect ATMUX_* environment variables in the current process. |
| `atmux exec [--detach \| --shared] [--] <command> [args...]` | Execute a command with passthrough stdio and unchanged exit behavior. |
| `atmux install [--project\|--system] [--project-root <dir>] [--no-slash-commands]` | Install atmux for a project by default, or system-wide when requested. |
| `atmux issue create --title <title> [--description <description>]` | Repo-scoped issue tickets on filesystem. |
| `atmux message read <id> [--repo <repo>] [--team <team>]` | Read, list, or subscribe to filesystem-backed messages. |
| `atmux notify --pane <tmux-pane-id> --xml <payload> [--interrupt]` | Send an ATMUX XML notification to a tmux pane. |
| `atmux pane watch <target> --text <needle> [--scope pane\|window\|session]` | Watch tmux pane text. |
| `atmux path watch <glob> [--timeout <seconds>] [--duration <seconds>]` | Operate on filesystem paths matched by a glob. |
| `atmux pr create --title <title> [--description <description>]` | Repo-scoped pull request tickets on filesystem. |
| `atmux process watch <pid> [--timeout <seconds>] [--interval <seconds>]` | Watch or stop an atmux exec-tracked child process. |
| `atmux role list` | Manage adapter-agnostic roles. |
| `atmux schedule (--interval <duration> \| --once <duration>) [--no-detach] --notification <text>` | Schedule a future or repeating action. |
| `atmux send --to <agent\|team> [--reply-required] [--interrupt] "message"` | Send XML messages to a single agent or every agent in a team. |
| `atmux team create <name> [--role <role>] [--start <cmd>] [--stop <cmd>]` | Manage repo-scoped teams of agents. |
| `atmux update [--project] [--project-root <dir>] [--check] [--version <version>]` | Update atmux from the latest GitHub release. |
| `atmux watcher list` | List or remove background watcher registrations. |

## Environment variables

| Variable | Description |
|----------|-------------|
| `ATMUX_HOME` | Installation/state root (default: `<project>/.atmux` for project installs, `~/.atmux` for system installs) |
| `ATMUX_REPO` | Repository name for the current agent or team |
| `ATMUX_AGENT_NAME` | Current agent's name |
| `ATMUX_WORKTREE` | Working directory (worktree or repo root) |
| `ATMUX_TEAM` | Team this agent belongs to |
| `ATMUX_SESSION_ID` | Unique agent run identifier |
| `ATMUX_SESSION_KIND` | `agent` or `team` |
| `ATMUX_TMUX_SOCKET` | Tmux server socket for the project; defaults to `$ATMUX_HOME/tmux/server.sock` when short enough, otherwise a stable per-project socket under `/tmp/atmux-tmux-<uid>/` |
| `ATMUX_TMUX_MOUSE` | Tmux mouse mode for agents and teams, `off` by default to avoid accidental copy-mode blocking notifications; set `on` to restore tmux mouse scrolling |

## More Docs

- `docs/atmux.md` — full command reference
- `docs/agent.md` — agent and worktree concepts
- `docs/architecture.md` — internal architecture overview
- `docs/cli.md` — CLI implementation notes

<!--
  This file is generated. Edit templates/README.md.tmpl and the underlying
  command `usage()` heredocs / adapter manifests, then run:
    bin/(atmux)/(internal)/render-docs readme
-->