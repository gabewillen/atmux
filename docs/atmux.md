# atmux

`atmux` is a bash CLI for starting and interacting with repo-scoped tmux sessions.

## Install Layout
When `atmux` runs, it ensures this home layout exists:
- `<ATMUX_HOME>/`
- `<ATMUX_HOME>/bin/`
- `<ATMUX_HOME>/agents/`
- `<ATMUX_HOME>/adapters/`

For project installs, `ATMUX_HOME` defaults to `<project>/.atmux`. For system installs, it defaults to `~/.atmux`.

## Session Model
- Session name: `atmux-{{repo}}-{{name}}`
- Initial window/tab name: `atmux-{{repo}}-{{name}}`

## Commands

### `atmux session start`
- Must run inside a git repo.
- Reuses an existing repo session if one already exists (`atmux-{{repo}}-*`).
- Otherwise creates `atmux-{{repo}}-{{name}}` (default `name=manager`).
- Creates a git worktree at `<ATMUX_HOME>/agents/{{repo}}-{{name}}` on branch `atmux-{{repo}}-{{name}}`, then initializes submodules recursively.
- Starts the selected adapter in that session (`--adapter`, default `codex`).

### `atmux session list`
- Lists atmux sessions (`atmux-*`), one per line.

### `atmux agent list`
- Lists agents owned by this agent (`ATMUX_AGENT_NAME`) via `ATMUX_MANAGER`.
- Outputs XML:
  `<atmux><agents><agent ... /></agents></atmux>`

### `atmux agent list -a`
- Lists all agents in the org (`atmux-{{repo}}-{{name}}` sessions).

### `atmux agent list --status`
- Includes nested status per agent:
  `<agent ...><status ... /></agent>`

### `atmux create --agent <name> --role <role> --intelligence <0-100> [--team <team>]`
- Creates a new agent session/worktree.
- `--intelligence` is adapter-portable and maps to a model plus reasoning level via the adapter manifest.

Current built-in mapping:

| Adapter | Intelligence | Model | Reasoning |
|---------|--------------|-------|-----------|
| `claude-code` | 0-39 | `sonnet` | `low` |
| `claude-code` | 40-74 | `sonnet` | `medium` |
| `claude-code` | 75-89 | `sonnet` | `high` |
| `claude-code` | 90-100 | `opus` | `high` |
| `codex` | 0-29 | `gpt-5.5` | `low` |
| `codex` | 30-59 | `gpt-5.5` | `medium` |
| `codex` | 60-84 | `gpt-5.5` | `high` |
| `codex` | 85-100 | `gpt-5.5` | `extra-high` |
| `cursor-agent` | 0-39 | `composer-2-fast` | `low` |
| `cursor-agent` | 40-74 | `composer-2` | `medium` |
| `cursor-agent` | 75-89 | `gpt-5.3-codex-high` | `high` |
| `cursor-agent` | 90-100 | `gpt-5.3-codex-xhigh` | `extra-high` |
| `gemini` | 0-39 | `gemini-3.1-flash-lite-preview` | `low` |
| `gemini` | 40-74 | `gemini-3-flash-preview` | `medium` |
| `gemini` | 75-89 | `gemini-3.1-pro-preview` | `medium` |
| `gemini` | 90-100 | `gemini-3.1-pro-preview` | `high` |

### `atmux create --team <name>`
- Creates a team session `atmux-{{repo}}-team-{{name}}`.

### `atmux create --issue --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue in `<ATMUX_HOME>/issues/{{repo}}/`.

### `atmux send --to <name|session> [--reply-required] "message"`
- Sends a message to a specific agent session or every agent in a team.
- Resolution order for `--to`:
  1) Team session/name
  2) Agent session/name

### `atmux schedule (--interval <duration> | --once <duration>) --notification "text"`
### `atmux schedule (--interval <duration> | --once <duration>) -- <command> [args...]`
- Schedules either:
  - a direct ATMUX notification via `--notification`
  - or an arbitrary command via `-- <command...>`
- Notification mode always targets the current session; use it for self reminders, ticks, and status checks.
- Only schedule `atmux send` when the target is another agent or team:
  `atmux schedule --once 10m -- atmux send --to worker "status check"`
- Duration suffixes:
  `ms`, `s`, `m`, `h`, `d`

### `atmux exec [--] <command> [args...]`
- Executes a command with passthrough stdio and the wrapped command's original exit code.
- After the command exits or is interrupted, sends an ATMUX notification back to the current agent pane:
  `<notification type="exec" from="..." cmd="..." exit_code="..." />`
- Tracks each launched child process under `<ATMUX_HOME>/exec/<repo>/<pid>/`.

### `atmux watch`
- Wait until text appears, a tracked process exits, an issue updates, new GitHub issues appear, a PR discussion updates, or an agent goes idle.
- `atmux watch --issues <repo|url>` polls a GitHub repository for newly created issues and queues notifications to the current pane until the watcher is stopped.
- `watch --issues` registration output includes a `watcher_id`, which can be removed via `atmux kill --watcher <id>`.
- `atmux watch --pr <url>` polls GitHub PR comments/reviews and queues notifications to the current pane until the PR closes/merges or the watcher is stopped.
- `watch --pr` registration output includes a `watcher_id`, which can be removed via `atmux kill --watcher <id>`.

### `atmux kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]`
- Stops the tracked child for this repo (same `exec` metadata as `watch --pid`).
- After the executor finishes notifications (including watcher fan-out), removes `<ATMUX_HOME>/exec/<repo>/<pid>/`.
- Default `TERM` and `--timeout` 60s; escalates to `KILL` if the process is still alive after the timeout.

### `atmux kill --watcher <id> [--timeout <seconds>]`
- Removes a watcher registration by id.
- Supports watcher ids emitted by `atmux watch --pr` and `atmux watch --issues`.

### `atmux assign --to <agent> --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue and assigns it to the target agent/session.

### `atmux assign --issue <id> --to <agent>`
- Assigns an existing filesystem issue id to the target agent/session.

### `atmux install [--project|--system] [--project-root <dir>] [--no-slash-commands]`
- Installs atmux into `<project>/.atmux` by default. The installer prompts for project vs system scope when interactive, defaulting to project.
- Project installs write Claude Code, Gemini CLI, and Codex commands under project-local `.claude/`, `.gemini/`, and `.codex/` directories and do not modify shell profiles.
- Project installs include `.atmux/.gitignore` so the launcher and source can be committed while runtime state stays ignored.
- System installs use `~/.atmux` and user-level CLI command directories.
- Use `--no-slash-commands` to skip that step.

### `atmux adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `<ATMUX_HOME>/adapters/<adapter>`.

### Script Subcommands
- Any executable in `<ATMUX_HOME>/bin/scripts/<name>` is available as:
  `atmux <name>`
- Examples:
  - `atmux env`
  - `atmux agent`

## Environment
When `atmux session start` creates a session, it sets:
- `ATMUX_REPO`
- `ATMUX_SESSION_ID`
- `ATMUX_WORKTREE`
- `ATMUX_AGENT_NAME`
- If `ATMUX_MANAGER` is set, the agent is treated as manager-owned.

## Adapter Contract
- Adapter entrypoint:
  `adapter/{{name}}/cmd`
- Required adapter commands:
  - `status`
  - `model list|get|set <value>`
  - `reasoning-level list|get|set <value>`
  - `control-file`
  - `start`
- `status` output (XML):
  - Adapter script payload: `<status ... />` (attributes)
  - `atmux` command wraps payload in `<atmux>...</atmux>`
  - Required attributes:
    - `owner` (`agent|manager`)
    - `model`
    - `reasoning_level`
    - `state` (`idle|busy|errored|usage-limit`)
