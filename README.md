# Agent TMUX (**atmux**)

**atmux** is a tmux-first toolkit for running and coordinating multiple AI coding agents in parallel. It handles session lifecycle, inter-agent messaging, work assignment, output capture, and notifications — across different AI CLIs (Claude Code, Gemini, Codex, Cursor) and git repos.

The whole point: because the install lives inside your project and every command is plain shell, the agents you run can read, patch, and extend atmux itself. When the tool doesn't do what your workflow needs, the agent fixes the tool — in the same commit as the work.

## No frills. No dependencies. No build.

- **No build step.** Pure shell scripts — nothing to compile, bundle, or transpile. `git clone` and run.
- **No runtime.** No Node, Python, Go, or Rust toolchain required. No package manager, no lockfiles, no `node_modules`.
- **No dependencies** beyond what's already on every Unix box you'd run an agent on: `bash`, `git`, and `tmux`. That's it.
- **No daemon, no server, no database.** State lives in plain files under `ATMUX_HOME` (`<project>/.atmux` by default). Inspect it with `ls` and `cat`.
- **No frills.** One shell script per command. Read the source, patch it in place, move on.

Install by piping `curl` into `sh`, or clone the repo and run `./install.sh`. The installer defaults to a project-local install so it does not modify system-level agent config unless you choose `--system`. Uninstall by deleting the install directory.

## What makes it different

- **The agents can change atmux itself.** This is the headline feature, not a side effect. Project-local install + plain shell + no build step means the source sits next to your code, in the same git history, editable by the same agent that's doing the work. Hit a missing flag? The agent adds it. Found a bug in `pr watch`? The agent patches it and the next command picks up the fix. No fork, no rebuild, no upstream wait, no "file an issue and hope." The tool adapts to your project at the speed of the project.
- **CLI-agnostic via adapters.** Run Claude Code, Gemini, Codex, and Cursor side-by-side in the same session. Swap vendors without rewriting your workflow. Third-party adapters install with `atmux adapter install owner/repo`.
- **Intelligence scale, not model names.** Say `--intelligence 80` and the adapter picks the right model and reasoning level. Portable across vendors, survives model renames — no more hardcoding `claude-opus-4-7` or `gpt-5-codex` across your scripts.
- **You can actually see the agents work.** It's just tmux. Attach to any session, watch the agent think in real time, detach and come back later. No custom TUI, no web dashboard, no log tailing.
- **Git worktree per agent.** Each agent gets its own branch and working directory under `ATMUX_HOME/agents/`. Parallel agents can't stomp each other's changes, and cleanup is a single `atmux agent kill`.

> **Experimental** — this project is under active development. APIs, commands, and behavior may change without notice. Use at your own risk.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/gabewillen/atmux/main/install.sh | sh
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

## Concepts

### Sessions and agents

Each agent runs in a named tmux session: `atmux-<repo>-<agent>`. By default, agents get their own git worktree at `ATMUX_HOME/agents/<repo>-<name>`, keeping their changes isolated. Worktree creation initializes submodules with `git submodule update --init --recursive`. Pass `--no-worktree` to skip worktree creation and run in the repo root instead.

### Teams

Agents can be grouped into teams (capped at 4 per team). Teams share a layout and can be messaged collectively.

```sh
atmux team create platform
atmux agent create reviewer --role reviewer --team platform --intelligence 80
atmux agent create tester   --role tester   --team platform --intelligence 55
```

### Intelligence scale

The `--intelligence 0–100` flag selects a model and reasoning level automatically via the adapter's manifest. Higher values use more capable (and slower/costlier) models.

| Adapter        | Intelligence | Model                  | Reasoning level |
|----------------|--------------|------------------------|-----------------|
| `claude-code`  | 0–39         | `sonnet`               | `low`           |
| `claude-code`  | 40–74        | `sonnet`               | `medium`        |
| `claude-code`  | 75–89        | `sonnet`               | `high`          |
| `claude-code`  | 90–100       | `opus`                 | `high`          |
| `codex`        | 0–29         | `gpt-5.5`              | `low`           |
| `codex`        | 30–59        | `gpt-5.5`              | `medium`        |
| `codex`        | 60–84        | `gpt-5.5`              | `high`          |
| `codex`        | 85–100       | `gpt-5.5`              | `extra-high`    |
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
atmux agent create worker --role implementer --adapter claude-code --intelligence 60 \
  -- --dangerously-skip-permissions
```

Install a third-party adapter:

```sh
atmux adapter install owner/repo
```

## Command reference

Most commands take the form `atmux <noun> <verb> [args]`. A handful of
cross-cutting verbs (`send`, `exec`, `schedule`, `update`, `install`,
`env`) have no resource home and stay verb-shaped.

| Resource  | Verbs                                            |
|-----------|--------------------------------------------------|
| `agent`   | `create`, `list`, `capture`, `watch`, `kill`     |
| `team`    | `create`, `list`, `capture`, `kill`              |
| `session` | `list`, `start`, `attach`                        |
| `issue`   | `create`, `list`, `assign`, `comment`, `watch`   |
| `pr`      | `create`, `list`, `assign`, `comment`, `watch`   |
| `message` | `list`, `read`                                   |
| `role`    | `create`, `list`, `show`, `resolve`              |
| `process` | `watch`, `kill`                                  |
| `pane`    | `watch`                                          |
| `path`    | `watch`                                          |
| `watcher` | `list`, `kill`                                   |
| `adapter` | `install`                                        |
| `config`  | `get`, `set`, `list`                             |

Cross-cutting verbs: `send`, `exec`, `schedule`, `env`, `update`, `install`.

### Resources

#### `agent`

```sh
atmux agent create <name> --role <role> --intelligence <0-100> \
  [--team <team>] [--adapter <adapter>] [--no-worktree] [-- <adapter-args...>]
atmux agent list [--all] [--status]
atmux agent capture <name|--all> [--lines <n>]
atmux agent watch <name> [--idle <seconds>] [--timeout <seconds>]
atmux agent kill <name|pattern>...
atmux agent kill --all [--yes]
```

Each agent runs in `atmux-<repo>-<name>` and gets a git worktree under
`ATMUX_HOME/agents/<repo>-<name>` (skip with `--no-worktree`).

#### `team`

```sh
atmux team create <name>
atmux team list
atmux team capture <name> [--lines <n>]
```

Up to 4 agents per team. New agents inherit `--team` from `ATMUX_TEAM` when set.

#### `session`

```sh
atmux session list
atmux session start [--name <name>] [--adapter <adapter>] [-- <adapter-args...>]
atmux session attach <name|session>
```

`session attach` must be run outside tmux.

#### `issue` / `pr`

Filesystem-tracked issues and pull requests share the same verbs.

```sh
atmux issue create --title <title> [--description <desc>] [--todo <todo>]... \
  [--given <ctx>] [--when <action>] [--then <outcome>] [--assign-to <agent>]
atmux issue list [--repo <repo>]
atmux issue assign <id> --to <agent>
atmux issue comment <id> "message"
atmux issue watch <id>                       # local update
atmux issue watch --feed <repo|url>          # GitHub fan-out

atmux pr create --title <title> [--description <desc>] \
  [--source <branch>] [--target <branch>] [--todo <todo>]...
atmux pr list [--repo <repo>]
atmux pr assign <id> --to <agent>
atmux pr comment <id> "message"
atmux pr watch <id|atmux-uri|github-url>
atmux pr watch --feed <repo|url>
```

`issue create --assign-to` is a one-shot create + assign. `pr watch <github-url>`
and `--feed` are long-running watchers — list/remove via `watcher`.

#### `message`

```sh
atmux message list [--unread]
atmux message read <id> [--repo <repo>]
```

Messages live at `~/.atmux/messages/<repo>/<id>/`.

#### `role`

```sh
atmux role create <name> --description "..." [--intelligence <0-100>] \
  [--adapters <csv>] [--hooks start,stop] [--scope repo|global|auto]
atmux role create <name> --from-file <path>      # alternate body source
atmux role create <name> --from-stdin
atmux role list
atmux role show <name>
atmux role resolve <name>
```

Roles are discovered (in order) under `<repo>/.atmux/roles/`,
`~/.atmux/roles/`, then `<atmux-source>/roles/`.

#### `process`

```sh
atmux process watch <pid> [--timeout <seconds>]
atmux process watch <pid> --stdio [--duration <seconds>] [--lines <n>]
atmux process kill <pid> [--timeout <seconds>] [--signal <NAME>]
```

Operates on `atmux exec`-tracked children (`~/.atmux/exec/<repo>/<pid>/`).

#### `pane` / `path`

```sh
atmux pane watch <tmux-target> --text <needle> [--scope pane|window|session]
atmux path watch <glob> [--timeout <seconds>] [--interval <seconds>]
```

`path watch` uses `fswatch` or `inotifywait` when available, otherwise polls.

#### `watcher`

```sh
atmux watcher list
atmux watcher kill <id> [--timeout <seconds>]
```

Lists and removes background watcher registrations created by `pr watch`,
`issue watch --feed`, and `pr watch --feed` (the long-running fan-out modes).

#### `adapter` / `config`

```sh
atmux adapter install <owner/repo>
atmux config get <key>
atmux config set <key> <value>
atmux config list
```

### Cross-cutting verbs

#### `send`

```sh
atmux send --to <name|session> [--reply-required] [--interrupt] "message"
```

`--reply-required` signals the sender is blocked until the recipient responds.
`--interrupt` submits via the adapter's interrupt key (processed after the
current tool) instead of the default queue key (processed when idle).

#### `exec`

```sh
atmux exec [--detach] -- <command> [args...]
```

Runs a command and notifies the agent when it exits. `--detach` runs in a
new tmux window — track it with `process watch <pid>` and `process kill <pid>`.

#### `schedule`

```sh
atmux schedule (--once <duration> | --interval <duration>) --notification "..."
atmux schedule (--once <duration> | --interval <duration>) -- <command> [args...]
```

Use `--notification` for self reminders, ticks, and status checks. Only
schedule `atmux send` when the target is another agent or team.
`--no-detach` runs in the foreground. Duration suffixes: `ms`, `s`, `m`, `h`, `d`.

#### `env`

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
- Check `atmux team list` before creating new team members.
- Never silently change scope — ask the manager first.

## Docs

- `docs/atmux.md` — full command and session reference
- `docs/agent.md` — agent and worktree concepts
- `docs/architecture.md` — internal architecture overview
- `docs/cli.md` — CLI implementation notes
