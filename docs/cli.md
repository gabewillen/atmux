# atmux CLI

`atmux` is a bash CLI for managing repo-scoped tmux agent sessions, routing messages, and coordinating work between adapters such as Codex, Claude Code, Gemini CLI, and Cursor Agent.

## Common Commands

### `atmux create --agent <name> --role <role> --intelligence <0-100> [--team <team>] [--adapter <adapter>]`

Creates an agent session and worktree. The `--intelligence` value is adapter-portable: the selected adapter maps it to the correct model and reasoning level.

Example:

```sh
atmux create --agent planner --role planner --intelligence 80
```

### `atmux create --team <name>`

Creates a team session. Agents created while `ATMUX_TEAM` is set join that team.

### `atmux send --to <agent|session|team> [--reply-required] [--interrupt] "message"`

Queues a notification into another agent pane. Target resolution prefers teams, then agents/sessions. `--interrupt` uses the adapter interrupt submit key when available.

### `atmux assign --to <agent|session> --title <title> [--description <description>] [--todo <todo>]...`

Creates and assigns a filesystem-backed issue.

### `atmux schedule (--interval <duration> | --once <duration>) --notification <text>`
### `atmux schedule (--interval <duration> | --once <duration>) -- <command> [args...]`

Schedules a notification or command. To schedule a future message, schedule `atmux send` directly:

```sh
atmux schedule --once 10m -- atmux send --to worker "status check"
```

### `atmux exec [--detach] -- <command> [args...]`

Runs a command and sends an exec notification when it exits. Detached execs run in a tmux window and are tracked under `<ATMUX_HOME>/exec/<repo>/<pid>/`.

### `atmux watch`

Waits for text, process completion, issue updates, stdio output, or agent idleness.

Examples:

```sh
atmux watch --agent worker --idle 20 --timeout 120
atmux watch --pid 12345 --timeout 60
atmux watch --target %1 --text "ready" --timeout 30
```

### `atmux kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]`

Stops an `atmux exec` tracked process, waits for completion notifications and watcher fan-out, then clears metadata.

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
| `codex` | 0-29 | `gpt-5.4` | `low` |
| `codex` | 30-59 | `gpt-5.4` | `medium` |
| `codex` | 60-84 | `gpt-5.4` | `high` |
| `codex` | 85-100 | `gpt-5.4` | `extra-high` |
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
