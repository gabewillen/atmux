# atmux

`atmux` is a bash CLI for starting and interacting with repo-scoped tmux sessions.

## Install Layout
When `atmux` runs, it ensures this home layout exists:
- `~/.atmux/`
- `~/.atmux/bin/`
- `~/.atmux/agents/`
- `~/.atmux/adapters/`

## Session Model
- Session name: `atmux-{{repo}}-{{name}}`
- Initial window/tab name: `atmux-{{repo}}-{{name}}`

## Commands

### `./bin/atmux.sh session start`
- Must run inside a git repo.
- Reuses an existing repo session if one already exists (`atmux-{{repo}}-*`).
- Otherwise creates `atmux-{{repo}}-{{name}}` (default `name=manager`).
- Creates a git worktree at `~/.atmux/agents/{{repo}}-{{name}}` on branch `atmux-{{repo}}-{{name}}`.
- Starts the selected adapter in that session (`--adapter`, default `codex`).

### `./bin/atmux.sh session list`
- Lists atmux sessions (`atmux-*`), one per line.

### `./bin/atmux.sh agent list`
- Lists agents owned by this agent (`ATMUX_AGENT_NAME`) via `ATMUX_MANAGER`.
- Outputs XML:
  `<atmux><agents><agent ... /></agents></atmux>`

### `./bin/atmux.sh agent list -a`
- Lists all agents in the org (`atmux-{{repo}}-{{name}}` sessions).

### `./bin/atmux.sh agent list --status`
- Includes nested status per agent:
  `<agent ...><status ... /></agent>`

### `./bin/atmux.sh create --agent <name> --role <role> --intelligence <0-100> [--team <team>]`
- Creates a new agent session/worktree.

### `./bin/atmux.sh create --team <name>`
- Creates a team session `atmux-{{repo}}-team-{{name}}`.

### `./bin/atmux.sh create --issue --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue in `~/.atmux/issues/{{repo}}/`.

### `./bin/atmux.sh manager send [--reply-required] "message"`
- Sends a message to your manager session.
- Message XML includes:
  `<message ... reply_required="true|false">`.

### `./bin/atmux.sh send --to <name|session> [--reply-required] "message"`
- Sends a message to a specific agent session or every agent in a team.
- Resolution order for `--to`:
  1) Team session/name
  2) Agent session/name

### `./bin/atmux.sh schedule (--interval <duration> | --once <duration>) --notification "text"`
### `./bin/atmux.sh schedule (--interval <duration> | --once <duration>) -- <command> [args...]`
- Schedules either:
  - a direct ATMUX notification via `--notification`
  - or an arbitrary command via `-- <command...>`
- Notification mode always targets the current session.
- To schedule a message, schedule the command explicitly:
  `atmux schedule --once 10m -- atmux send --to worker "status check"`
- Duration suffixes:
  `ms`, `s`, `m`, `h`, `d`

### `./bin/atmux.sh exec [--] <command> [args...]`
- Executes a command with passthrough stdio and the wrapped command's original exit code.
- After the command exits or is interrupted, sends an ATMUX notification back to the current agent pane:
  `<notification from="exec ..." timestamp="..." exitcode="..." />`
- Tracks each launched child process under `~/.atmux/exec/<repo>/<pid>/`.

### `./bin/atmux.sh kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]`
- Stops the tracked child for this repo (same `exec` metadata as `watch --pid`).
- After the executor finishes notifications (including watcher fan-out), removes `~/.atmux/exec/<repo>/<pid>/`.
- Default `TERM` and `--timeout` 60s; escalates to `KILL` if the process is still alive after the timeout.

### `./bin/atmux.sh assign --to <agent> --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue and assigns it to the target agent/session.

### `./bin/atmux.sh assign --issue <id> --to <agent>`
- Assigns an existing filesystem issue id to the target agent/session.

### `./bin/atmux.sh install [--no-slash-commands]`
- Installs atmux into `~/.atmux`.
- By default also installs harness-specific slash/custom commands for Claude Code, Gemini CLI, and Codex.
- Use `--no-slash-commands` to skip that step.

### `./bin/atmux.sh adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `~/.atmux/adapters/<adapter>`.

### Script Subcommands
- Any executable in `~/.atmux/bin/scripts/<name>` is available as:
  `atmux <name>`
- Examples:
  - `atmux env`
  - `atmux agent`
  - `atmux manager`

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
