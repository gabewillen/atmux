# atmux CLI

`atmux` is a bash CLI for managing repo-scoped tmux agent sessions, routing messages, and coordinating work between adapters such as Codex, Claude Code, Gemini CLI, and Cursor Agent.

## Common Commands

### `atmux agent create <name> --role <role> --intelligence <0-100> [--team <team>] [--adapter <adapter>]`

Creates an agent session and worktree. The `--intelligence` value is adapter-portable: the selected adapter maps it to the correct model and reasoning level.

Example:

```sh
atmux agent create planner --role planner --intelligence 80
```

### `atmux team create <name>`

Creates filesystem-backed team state. Agents created with `--team <name>` join that team as tmux-backed members. Use `atmux team view <name>` when you want an optional multiagent tmux view.

### `atmux pr create --title <title> [--description <description>] [--source <branch>] [--target <branch>]`

Creates a filesystem-backed pull request under `<ATMUX_HOME>/pull-requests/<repo>/`.

### `atmux send --to <agent|session|team> [--reply-required] [--interrupt] "message"`

Queues a notification into another agent pane. Target resolution prefers teams, then agents/sessions. Team targets are stored under `<ATMUX_HOME>/team-messages/<repo>/<team>/`, notify every current team member, and also notify outside panes registered with `atmux message subscribe --team <team>`. `--interrupt` uses the adapter interrupt submit key when available.

### `atmux message read <id> [--repo <repo>] [--team <team>]`
### `atmux message list [--unread] [--repo <repo>] [--team <team>]`
### `atmux message subscribe --team <team> [--repo <repo>]`

Reads and lists filesystem-backed direct messages by default. Pass `--team <team>` to read or list team messages. Team members can read team messages automatically; outside agents can subscribe for future team-message notifications and read access with `atmux message subscribe --team <team>`. Subscriptions can be removed with `atmux message unsubscribe --team <team>` or `atmux watcher kill <watcher_id>`.

### `atmux issue create --title <title> --assign-to <agent|session> [--description <description>] [--todo <todo>]...`

Creates and assigns a filesystem-backed issue in one shot. For an existing
issue, use `atmux issue assign <id> --to <agent|session>`.

### `atmux issue comment <id> "message"` / `atmux pr comment <id> "message"`

Posts a comment on a filesystem-backed issue or pull request; notifies
watchers, the assignee, and the assigner.

### `atmux schedule (--interval <duration> | --once <duration>) --notification <text>`
### `atmux schedule (--interval <duration> | --once <duration>) -- <command> [args...]`

Schedules a notification or command. Use `--notification` for self reminders,
ticks, and status checks. Only schedule `atmux send` when the target is another
agent or team:

```sh
atmux schedule --once 10m -- atmux send --to worker "status check"
```

### `atmux exec [--detach] -- <command> [args...]`

Runs a command and sends an exec notification when it exits. Detached execs run in a tmux window and are tracked under `<ATMUX_HOME>/exec/<repo>/<pid>/`.

### Watch (per-resource)

Watch verbs live on each resource. Available forms: `atmux agent watch`, `atmux pane watch`, `atmux process watch`, `atmux path watch`, `atmux issue watch`, `atmux pr watch`. They wait for text, process completion, issue updates, local PR updates, new GitHub issues, new GitHub pull requests, GitHub PR discussion updates, stdio output, or agent idleness.

Examples:

```sh
atmux agent watch worker --idle 20 --timeout 120
atmux process watch 12345 --timeout 60
atmux path watch 'src/**/*.sh' --timeout 60
atmux issue watch --feed owner/repo --timeout 600
atmux pr watch --feed owner/repo --timeout 600
atmux pr watch atmux://pull-request/myrepo/AbCdEfGhIjKlMnOp --timeout 120
atmux pr watch https://github.com/owner/repo/pull/123 --timeout 600
atmux pane watch %1 --text "ready" --timeout 30
```

`atmux issue watch --feed` is long-lived: it keeps notifying on newly created GitHub issues until stopped. Its registration output includes `watcher_id="..."` for use with `atmux watcher kill <id>`.

`atmux pr watch --feed` is long-lived: it keeps notifying on newly created GitHub pull requests until stopped. Its registration output includes `watcher_id="..."` for use with `atmux watcher kill <id>`.

`atmux path watch` exits when a filesystem glob's matched set or file metadata changes. It uses `fswatch` or `inotifywait` when available, otherwise it falls back to polling.

`atmux pr watch <url>` accepts both local PR ids/URIs and GitHub PR URLs. For GitHub URLs it is long-lived: it keeps notifying on new discussion until stopped or the PR closes/merges. Its registration output includes `watcher_id="..."` for use with `atmux watcher kill <id>`.

### `atmux process kill <pid> [--timeout <seconds>] [--signal <NAME>]`

Stops an `atmux exec` tracked process, waits for completion notifications and watcher fan-out, then clears metadata.

### `atmux watcher kill <id> [--timeout <seconds>]`

Removes a watcher registration by id. Supports watcher ids emitted by `atmux pr watch <url>`, `atmux issue watch --feed`, `atmux pr watch --feed`, and `atmux message subscribe --team`.

### `atmux install [--project|--system] [--project-root <dir>] [--no-slash-commands]`

Installs project-local by default into `<project>/.atmux`. Project installs write Claude Code, Gemini CLI, and Codex command files under project-local `.claude/`, `.gemini/`, and `.codex/` directories without modifying shell profiles.

## Intelligence Mapping

The `--intelligence 0-100` flag selects a model and reasoning level through the adapter manifest.

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

## Adapter Commands

Every adapter exposes:

- `status`
- `model list|get|set <value>`
- `reasoning-level list|get|set <value>`
- `control-file`
- `start`

Adapter status returns XML with `owner`, `state`, `model`, `reasoning_level`, and `control_file`.
