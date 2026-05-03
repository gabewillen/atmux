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

> **Experimental** — this project is under active development (current version: `0.13.0`). APIs, commands, and behavior may change without notice. Use at your own risk.

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

Each agent runs in a named tmux session: `atmux-<repo>-<agent>`. By default, agents get their own git worktree at `ATMUX_HOME/agents/<repo>-<name>`, keeping their changes isolated. Worktree creation initializes submodules with `git submodule update --init --recursive`. Pass `--shared-worktree` to skip worktree creation and run the new agent in the caller's current worktree instead (`--no-worktree` is a deprecated alias).

### Teams

Agents can be grouped into teams (capped at 4 per team). Teams share a layout and can be messaged collectively.

```sh
atmux team create platform
atmux agent create reviewer --role reviewer --team platform --intelligence 80
atmux agent create tester   --role tester   --team platform --intelligence 55
```

### Intelligence scale

The `--intelligence 0–100` flag selects a model and reasoning level automatically via the adapter's manifest. Higher values use more capable (and slower/costlier) models.

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

## Command reference

Most commands take the form `atmux <noun> <verb> [args]`. A handful of cross-cutting verbs (`send`, `exec`, `schedule`, `notify`, `update`, `install`) have no resource home and stay verb-shaped. Run `atmux <noun> --help` (or `atmux <verb> --help`) for the full flag list of any single command.

The top-level entrypoint summarises the surface:

```text
Usage:
  atmux <noun> <verb> [args...]    # resource operations
  atmux <verb> [args...]           # cross-cutting verbs

atmux is an agent multiplexer across adapters.

Resources (use `atmux <noun> --help` for verbs):
  agent      manage agent sessions (create, attach, list, kill, capture, watch, resolve)
  team       manage team sessions (create, list, kill, capture, resolve)
  role       manage role definitions (create, list, show, resolve)
  issue      filesystem issue tickets (create, list, show, assign, claim, comment, watch)
  pr         filesystem pull request tickets (create, list, show, assign, claim, comment, watch)
  message    inter-agent messages (read, list)
  config     atmux configuration (get, set, unset, list)
  env        environment vars (get, list)
  adapter    AI CLI adapters (install, ...)
  watcher    background watcher registrations (list, kill)
  process    atmux exec-tracked processes (watch, kill)
  pane       tmux pane operations (watch)
  path       filesystem path operations (watch)

Verbs (no resource home):
  send       message another agent or team
  exec       run a command with notification on exit
  schedule   schedule a one-shot or recurring action
  notify     low-level notification CLI (mostly internal)
  update     update atmux to the latest version
  install    install atmux into a project or system

Run:
  atmux <subcommand> --help
for more information.
```

### Resources

#### `agent`

```sh
atmux agent create [name] --role <role> --intelligence <0-100>
                   [--team <team>] [--adapter <adapter>] [--adapters <list>]
                   [--shared-worktree] [--start <cmd>] [--stop <cmd>]
                   [--task --description <desc> --todo <todo>...]
                   [-- <adapter-args...>]
  (or `--name <name>` instead of positional; if omitted, atmux
   auto-generates `agent-N`.)
  `--shared-worktree` runs the new agent in the caller's current
  worktree instead of creating a private one for it (deprecated
  alias: `--no-worktree`).
atmux agent list [--all] [--status]
atmux agent status [<name>]
atmux agent attach <name|session>
atmux agent kill <name|pattern> [<name|pattern>...]
atmux agent kill --all [--yes]
atmux agent capture <name> [--lines <n>]
atmux agent capture --all [--lines <n>]
atmux agent watch <name> [--idle <seconds>] [--timeout <seconds>]
                         [--interval <seconds>] [--lines <n>]
atmux agent resolve <name> [<repo_name>]
```

Manage atmux agents — sessions running an AI CLI under tmux. Agents are
scoped to the current repo. `create` works both as a top-level command
(no manager required) and from inside a manager agent. When run
interactively without a manager, the new agent is attached to.

agent attach must be run outside tmux.

#### `team`

```sh
atmux team create <name>
atmux team list
atmux team ls
atmux team status [<name>]
atmux team capture <name> [--lines <n>]
atmux team kill <name|pattern> [...]
atmux team resolve <name> [<repo_name>]
```

Manage repo-scoped team tmux sessions.
Team session format: atmux-<repo>-team-<name>

#### `role`

```sh
atmux role list
atmux role show <name>
atmux role resolve <name>
atmux role create <name> (--from-file <path> | --from-stdin | --description <text>) \
                         [--intelligence <0-100>] [--adapters <a,b,...>] \
                         [--hooks <start,stop>] [--scope repo|global|auto] [--force]
```

Roles are adapter-agnostic. A role is a directory containing any of:

- `role.md` — prompt body, appended under `# Role` in the agent's control file
- `manifest` — optional, sourced bash: `INTELLIGENCE=<0-100>`, `ADAPTERS=(name ...)`
- `start` — runs before the adapter starts (at agent-create time)
- `stop` — runs after the adapter exits (at agent-kill time)

Resolution precedence (first match wins): `<repo>/.atmux/roles/<name>` → `~/.atmux/roles/<name>` → `<source-root>/roles/<name>`.

`create` writes the role to `~/.atmux/roles/<name>` by default. `--scope repo` writes under `<repo>/.atmux/roles/<name>`; `--scope auto` picks repo if inside a git repo with an existing `.atmux/`, otherwise global.

#### `issue`

```sh
atmux issue create --title <title> [--description <description>]
                   [--given <context>] [--when <action>] [--then <outcome>]
                   [--todo <todo>]... [--repo <repo>]
                   [--assign-to <agent|session>]
atmux issue list [--repo <repo>]
atmux issue ls [--repo <repo>]
atmux issue get <id> [--repo <repo>]
atmux issue show <id> [--repo <repo>]
atmux issue assign <id> --to <agent|session> [--repo <repo>]
atmux issue claim <id> [--by <agent|session>] [--repo <repo>]
atmux issue comment <id> "message" [--repo <repo>]
atmux issue watch <id> [--repo <repo>] [--timeout <s>] [--interval <s>]
atmux issue watch --feed <repo|url> [--timeout <s>] [--interval <s>]
```

Repo-scoped issue tickets on filesystem.
Issues are stored at: ~/.atmux/issues/<repo>/<id>/

`create --assign-to` creates the issue and immediately assigns it.
`watch <id>` waits for the next update on a single issue.
`watch --feed <repo>` watches a GitHub repo for newly-filed issues.

#### `pr`

```sh
atmux pr create --title <title> [--description <description>]
                [--source <branch>] [--target <branch>]
                [--todo <todo>]... [--repo <repo>]
atmux pr list [--repo <repo>]
atmux pr ls [--repo <repo>]
atmux pr get <id> [--repo <repo>]
atmux pr show <id> [--repo <repo>]
atmux pr assign <id> --to <agent|session> [--repo <repo>]
atmux pr claim <id> [--by <agent|session>] [--repo <repo>]
atmux pr comment <id> "message" [--repo <repo>]
atmux pr watch <id|atmux-uri|github-url> [--repo <repo>] [--timeout <s>] [--interval <s>]
atmux pr watch --feed <repo|url> [--timeout <s>] [--interval <s>]
```

Repo-scoped pull request tickets on filesystem.
Pull requests are stored at: ~/.atmux/pull-requests/<repo>/<id>/

`watch <id>` watches a single pull request (filesystem ticket or GitHub URL).
`watch --feed <repo>` watches a GitHub repo for newly-filed pull requests.

#### `message`

```sh
atmux message read <id> [--repo <repo>]
atmux message list [--unread]
```

Read or list filesystem-backed messages.
Messages are stored at: ~/.atmux/messages/<repo>/<id>/

#### `config`

```sh
atmux config get   <key>          [--global]
atmux config set   <key> <value>  [--global]
atmux config unset <key>          [--global]
atmux config list                 [--global]
```

One file per key under `<repo>/.atmux/config/<key>` (local) or `~/.atmux/config/<key>` (with `--global`). Key paths are hierarchical with `/`, e.g. `update/auto`. `get` without `--global` resolves local first, then global.

#### `env`

```sh
atmux env
atmux env get <key>
```

Inspect ATMUX_* environment variables in the current process.

```sh
atmux env
atmux env get repo
atmux env get ATMUX_WORKTREE
```

#### `adapter`

```sh
atmux adapter install <owner/repo|github-url>
atmux adapter <name> <command> [args...]
```

Install adapters and run adapter contract commands.

```sh
atmux adapter install org/my-adapter
atmux adapter codex status
atmux adapter codex model list
```

#### `watcher`

```sh
atmux watcher list
atmux watcher kill <id> [--timeout <seconds>]
```

List or remove background watcher registrations created by `atmux pr watch
--feed`, `atmux issue watch --feed`, and `atmux pr watch <id|url>` (the
long-running fan-out modes).

Watcher ids have the form <kind>:<key>:<watcher_name>, e.g.:
  pr:owner_repo_pr_123:atmux-myrepo-worker-_12
  issues:owner_repo_issues:atmux-myrepo-worker-_12
  prs:owner_repo_prs:atmux-myrepo-worker-_12

#### `process`

```sh
atmux process watch <pid> [--timeout <seconds>] [--interval <seconds>]
atmux process watch <pid> --stdio [--duration <seconds>] [--timeout <seconds>] \
                                  [--interval <seconds>] [--lines <n>] \
                                  [--coalesce <seconds>]
atmux process kill  <pid> [--timeout <seconds>] [--signal <NAME>]
```

Operates on `atmux exec`-tracked child processes by pid (state at `~/.atmux/exec/<repo>/<pid>/`).

- **`watch`** — wait for the process to finish; receive the same exit-notification XML the executor would have sent.
- **`watch --stdio`** — monitor a detached exec process pane for output changes. Output events are batched into one digest notification per `--coalesce` window (default `60s`; pass `0` for per-event delivery). Pending output flushes when the process exits or `--duration` / `--timeout` expires.
- **`kill`** — stop the tracked process, wait for executor and watcher fan-out notifications to drain, then remove its metadata.

#### `pane`

```sh
atmux pane watch <target> --text <needle> [--scope pane|window|session]
                          [--timeout <seconds>] [--interval <seconds>]
                          [--lines <n>]
```

Operate on a tmux pane by id (e.g. %12, @3:1.0). The only verb today is
`watch`, which polls the pane and exits 0 when <needle> appears, non-zero
on timeout.

```sh
atmux pane watch %12 --scope pane --text "Select Model and Effort"
atmux pane watch @3 --scope window --text "running tests" --timeout 20
```

#### `path`

```sh
atmux path watch <glob> [--timeout <seconds>] [--duration <seconds>]
                        [--interval <seconds>] [--coalesce <seconds>]
                        [--exec <cmd>] [--once]
```

Operate on filesystem paths matched by a glob. The only verb today is
`watch`, which streams change events continuously, emitting one digest XML
line per --coalesce window (default 60s). Uses fswatch or inotifywait when
available; falls back to polling.

--timeout    Idle exit: return 124 if no change is observed for N seconds
             (0 = disabled, default 0).
--duration   Hard cap: return 124 after N seconds total (0 = no cap).
--coalesce   Batch change events into one digest line per N seconds
             (default 60). Set 0 to emit per-event lines.
--once       Single-shot: emit one XML line on the first change and exit 0.
             Disables coalescing.
--exec       Pipe each emitted XML line into `bash -c <cmd>`.

```sh
atmux path watch 'src/**/*.sh'
atmux path watch '/tmp/build/*.log' --once --timeout 60
atmux path watch 'docs/**/*.md' --exec ./on-docs-changed --coalesce 30
```

### Cross-cutting verbs

#### `send`

```sh
atmux send --to <name|session> [--reply-required] [--interrupt] "message"
```

Send XML messages to a single agent or every agent in a team.
Resolution order for --to:
  1) Team session/name
  2) Agent session/name
--interrupt  Submit using the adapter's interrupt key (processed after current
             tool) instead of the default queue key (processed when idle).

```sh
atmux send --to planner "run tests"
atmux send --to platform --reply-required "status check-in"
atmux send --to worker --interrupt "stop and check this"
```

#### `exec`

```sh
atmux exec [--detach] [--] <command> [args...]
```

Execute a command with passthrough stdio and unchanged exit behavior.
After the command exits or is interrupted, send an ATMUX notification back
to the current agent pane with the exit code.

--detach  Run the command in a new tmux window. Returns immediately.
          The process pane is stored so watchers can capture its output.
          Notification is sent to the agent pane when the process exits.

```sh
atmux exec sleep 30
atmux exec -- make test
atmux exec --detach -- make test
```

#### `schedule`

```sh
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] --notification "<text>"
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] -- <command> [args...]
```

Schedule a future or recurring action.

- **Notification mode** (`--notification`) — queues an ATMUX notification back to the current agent's session. Use this for self-reminders, ticks, and status checks.
- **Command mode** (`-- <command...>`) — runs the command in the current environment. Only schedule `atmux send` when the target is **another** agent or team; never schedule `atmux send --to <self>` (use notification mode instead).
- **`--no-detach`** — run in the foreground (blocking). By default the scheduled task runs in a detached tmux window so the command returns immediately.

Durations accept a unit suffix: `ms`, `s`, `m`, `h`, `d`. No suffix means seconds. Examples: `45s`, `30m`, `2h`.

#### `notify`

```sh
atmux notify --pane <tmux-pane-id> --xml <payload> [--interrupt]
```

Send an ATMUX XML notification to a tmux pane.
--interrupt  Use the adapter's interrupt submit key instead of the default
             key (Enter). Resolves the adapter from the pane's session, or from
             --session if provided.

```sh
atmux notify --pane %12 --xml '<notification type="test" from="manual" cmd="atmux message read abc" />'
atmux notify --pane %12 --xml '<notification type="urgent" from="mgr" />' --interrupt
```

#### `update`

```sh
atmux update [--check] [--version <version>]
atmux update --auto
atmux update --no-auto
```

**Options**

```
--check              Only report whether an update is available; do not install.
--version <version>  Install a specific version (e.g. 0.2.0). Defaults to latest.
--auto               Enable background auto-update on every atmux command (hourly throttle).
--no-auto            Disable background auto-update.
```

```sh
atmux update
atmux update --check
atmux update --version 0.2.0
atmux update --auto
atmux update --no-auto
```

#### `install`

```sh
atmux install [--project|--system] [--project-root <dir>] [--no-slash-commands]
```

Install atmux for a project by default, or system-wide when requested. Project
installs write to <project>/.atmux and project-local CLI command directories.

**Options**

```
--project            Install into <project>/.atmux and project-local CLI command dirs.
--system             Install into ~/.atmux and user-level CLI command dirs.
--scope <scope>      Same as --project or --system. Values: project, system.
--project-root <dir> Project directory for --project (default: git root or cwd).
--no-slash-commands  Skip installing Claude/Gemini/Codex slash commands.
```

```sh
atmux install
atmux install --project
atmux install --system
atmux install --no-slash-commands
```

## Environment variables

| Variable | Description |
|----------|-------------|
| `ATMUX_HOME` | Installation/state root (default: `<project>/.atmux` for project installs, `~/.atmux` for system installs) |
| `ATMUX_REPO` | Repository name for the current session |
| `ATMUX_AGENT_NAME` | Current agent's name |
| `ATMUX_MANAGER` | Parent manager agent name |
| `ATMUX_WORKTREE` | Working directory (worktree or repo root) |
| `ATMUX_TEAM` | Team this agent belongs to |
| `ATMUX_SESSION_ID` | Unique session identifier |
| `ATMUX_SESSION_KIND` | `agent` or `team` |

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

<!--
  This file is generated. Edit templates/README.md.tmpl and the underlying
  command `usage()` heredocs / adapter manifests, then run:
    bin/(atmux)/(internal)/render-docs all
-->