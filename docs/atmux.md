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

### `atmux agent list`
- Lists repo-scoped agent sessions.
- Outputs XML:
  `<atmux><agents><agent ... /></agents></atmux>`

### `atmux agent list -a`
- Lists all agents in the org (`atmux-{{repo}}-{{name}}` sessions).

### `atmux agent list --status`
- Includes nested status per agent:
  `<agent ...><status ... /></agent>`

### `atmux agent create [name] --role <role> --intelligence <0-100> [--team <team>]`
- Creates a new agent session/worktree. If `name` is omitted, `atmux` auto-generates a name such as `agent-N`.
- When run interactively, attaches to the new session.
- Creates a git worktree at `<ATMUX_HOME>/agents/{{repo}}-{{name}}` on branch `atmux-{{repo}}-{{name}}` (skip with `--shared-worktree` to run in the caller's current worktree; `--no-worktree` is a deprecated alias), then initializes submodules recursively.
- Starts the selected adapter in that session (`--adapter`, default `auto`; restrict candidates with `--adapters a,b,...`).
- `--intelligence` is adapter-portable and maps to a model plus reasoning level via the adapter manifest.

### `atmux agent attach <name|session>`
- Attaches to an existing agent session. Must run outside tmux.

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

### `atmux team create <name>`
- Creates filesystem-backed team state under `<ATMUX_HOME>/teams/{{repo}}/{{name}}/`.
- Team members are still tmux-backed agent sessions, but membership is not inferred from an optional tmux view.
- Use `atmux team view <name>` to create an optional multiagent tmux view.

### `atmux issue create --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue in `<ATMUX_HOME>/issues/{{repo}}/`.

### `atmux pr create --title <title> [--description "..."] [--source <branch>] [--target <branch>] [--todo "..."]`
- Creates a filesystem pull request in `<ATMUX_HOME>/pull-requests/{{repo}}/`.

### `atmux send --to <name|session> [--reply-required] "message"`
- Sends a direct message to a specific agent session, or a team message to every agent in a team.
- Team messages are stored in `<ATMUX_HOME>/team-messages/{{repo}}/{{team}}/{{id}}/`.
- Team-message notifications include `atmux message read <id> --team <team> --repo <repo>`.
- Resolution order for `--to`:
  1) Team name
  2) Agent session/name

### `atmux message read <id> [--repo <repo>] [--team <team>]`
### `atmux message list [--unread] [--repo <repo>] [--team <team>]`
### `atmux message subscribe --team <team> [--repo <repo>]`
- Without `--team`, reads or lists direct messages under `<ATMUX_HOME>/messages/{{repo}}/{{id}}/`.
- With `--team`, reads or lists team messages visible to current team members or subscribed outside agents.
- `message subscribe --team` registers the current pane for future team-message notifications. Remove it with `message unsubscribe --team` or `watcher kill <watcher_id>`.

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

### Watch (per-resource)
- Watch verbs live on each resource: `atmux agent watch`, `atmux pane watch`, `atmux process watch`, `atmux path watch`, `atmux issue watch`, `atmux pr watch`.
- Wait until text appears, a tracked process exits, an issue updates, a local pull request updates, new GitHub issues appear, new GitHub pull requests appear, a PR discussion updates, or an agent goes idle.
- `atmux path watch <glob>` watches filesystem paths matching a glob and exits when the matched set or file metadata changes. It uses `fswatch` or `inotifywait` when available, otherwise it falls back to polling.
- `atmux pr watch <id|atmux-uri|github-url>` watches either a filesystem pull request or a GitHub pull request based on the URI. Local PRs accept an id with `--repo`, or `atmux://pull-request/<repo>/<id>`.
- `atmux issue watch --feed <repo|url>` polls a GitHub repository for newly created issues and queues notifications to the current pane until the watcher is stopped. Registration output includes a `watcher_id`, removable via `atmux watcher kill <id>`.
- `atmux pr watch --feed <repo|url>` polls a GitHub repository for newly created pull requests and queues notifications to the current pane until the watcher is stopped. Registration output includes a `watcher_id`, removable via `atmux watcher kill <id>`.
- For GitHub PR URLs, `atmux pr watch <url>` polls comments/reviews, failing checks, and merge conflicts, then queues notifications until the PR closes/merges or the watcher is stopped. Remote watcher registration output includes a `watcher_id`, removable via `atmux watcher kill <id>`.

### `atmux process kill <pid> [--timeout <seconds>] [--signal <NAME>]`
- Stops the tracked child for this repo (same `exec` metadata as `atmux process watch <pid>`).
- After the executor finishes notifications (including watcher fan-out), removes `<ATMUX_HOME>/exec/<repo>/<pid>/`.
- Default `TERM` and `--timeout` 60s; escalates to `KILL` if the process is still alive after the timeout.

### `atmux watcher kill <id> [--timeout <seconds>]`
- Removes a watcher registration by id.
- Supports watcher ids emitted by `atmux pr watch`, `atmux issue watch --feed`, and `atmux pr watch --feed`.

### `atmux issue create --title <title> --assign-to <agent> [--description "..."] [--todo "..."]`
- Creates a filesystem issue and assigns it to the target agent/session in one shot.

### `atmux issue assign <id> --to <agent>`
- Assigns an existing filesystem issue id to the target agent/session.

### `atmux issue comment <id> "message"` / `atmux pr comment <id> "message"`
- Adds a comment to a filesystem-backed issue or pull request and notifies watchers/assignee/assigner.

### `atmux install [--project|--system] [--project-root <dir>] [--no-slash-commands]`
- Installs atmux into `<project>/.atmux` by default. The installer prompts for project vs system scope when interactive, defaulting to project.
- Project installs write Claude Code, Gemini CLI, and Codex commands under project-local `.claude/`, `.gemini/`, and `.codex/` directories and do not modify shell profiles.
- Project installs include `.atmux/.gitignore` so the launcher and source can be committed while runtime state stays ignored.
- System installs use `~/.atmux` and user-level CLI command directories.
- Use `--no-slash-commands` to skip that step.

### `atmux adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `<ATMUX_HOME>/adapters/<adapter>`.

### `atmux shim install <name|owner/repo|github-url>`
- Installs a shipped shim by name or a single-shim GitHub repo whose root contains `manifest` and executable wrappers.
- Installed shims live under `<ATMUX_HOME>/shims/<shim>`.
- Use `atmux shim list`, `atmux shim show <name>`, and `atmux shim resolve <name>` to inspect installed and shipped shims.

### Script Subcommands
- Any executable in `<ATMUX_HOME>/bin/scripts/<name>` is available as:
  `atmux <name>`
- Examples:
  - `atmux env`
  - `atmux agent`

## Environment
When `atmux agent create` creates a session, it sets:
- `ATMUX_REPO`
- `ATMUX_SESSION_ID`
- `ATMUX_WORKTREE`
- `ATMUX_AGENT_NAME`

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
    - `owner` (`agent`)
    - `model`
    - `reasoning_level`
    - `state` (`idle|busy|errored|usage-limit`)
