# amux

`amux` is a bash CLI for starting and interacting with repo-scoped tmux sessions.

## Install Layout
When `amux` runs, it ensures this home layout exists:
- `~/.amux/`
- `~/.amux/bin/`
- `~/.amux/agents/`
- `~/.amux/adapters/`

## Session Model
- Session name: `amux-{{repo}}-{{name}}`
- Initial window/tab name: `amux-{{repo}}-{{name}}`

## Commands

### `./bin/amux.sh session start`
- Must run inside a git repo.
- Reuses an existing repo session if one already exists (`amux-{{repo}}-*`).
- Otherwise creates `amux-{{repo}}-{{name}}` (default `name=manager`).
- Creates a git worktree at `~/.amux/agents/{{repo}}-{{name}}` on branch `amux-{{repo}}-{{name}}`.
- Starts the selected adapter in that session (`--adapter`, default `codex`).

### `./bin/amux.sh session list`
- Lists amux sessions (`amux-*`), one per line.

### `./bin/amux.sh agent list`
- Lists agents owned by this agent (`AMUX_AGENT_NAME`) via `AMUX_MANAGER`.
- Outputs XML:
  `<amux><agents><agent ... /></agents></amux>`

### `./bin/amux.sh agent list -a`
- Lists all agents in the org (`amux-{{repo}}-{{name}}` sessions).

### `./bin/amux.sh agent list --status`
- Includes nested status per agent:
  `<agent ...><status ... /></agent>`

### `./bin/amux.sh create --agent <name> --role <role> --intelligence <0-100> [--team <team>]`
- Creates a new agent session/worktree.

### `./bin/amux.sh create --team <name>`
- Creates a team session `amux-{{repo}}-team-{{name}}`.

### `./bin/amux.sh create --issue --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue in `~/.amux/issues/{{repo}}/`.

### `./bin/amux.sh manager send [--reply-required] "message"`
- Sends a message to your manager session.
- Message XML includes:
  `<message ... reply_required="true|false">`.

### `./bin/amux.sh send --to <name|session> [--reply-required] "message"`
- Sends a message to a specific agent session or every agent in a team.
- Resolution order for `--to`:
  1) Team session/name
  2) Agent session/name

### `./bin/amux.sh schedule --to <name|session> [--reply-required] (--interval <duration> | --once <duration>) "message"`
- Schedules a one-shot or repeating message using the same target matching as `send`.
- Resolution order for `--to`:
  1) Team session/name
  2) Agent session/name
- Duration suffixes:
  `ms`, `s`, `m`, `h`, `d`

### `./bin/amux.sh exec [--] <command> [args...]`
- Executes a command with passthrough stdio and the wrapped command's original exit code.
- After the command exits or is interrupted, sends an AMUX notification back to the current agent pane:
  `<notification from="exec ..." timestamp="..." exitcode="..." />`
- Tracks each launched child process under `~/.amux/exec/<repo>/<pid>/`.

### `./bin/amux.sh assign --to <agent> --title <title> [--description "..."] [--todo "..."]`
- Creates a filesystem issue and assigns it to the target agent/session.

### `./bin/amux.sh assign --issue <id> --to <agent>`
- Assigns an existing filesystem issue id to the target agent/session.

### `./bin/amux.sh install [--no-slash-commands]`
- Installs amux into `~/.amux`.
- By default also installs harness-specific slash/custom commands for Claude Code, Gemini CLI, and Codex.
- Use `--no-slash-commands` to skip that step.

### `./bin/amux.sh adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `~/.amux/adapters/<adapter>`.

### Script Subcommands
- Any executable in `~/.amux/bin/scripts/<name>` is available as:
  `amux <name>`
- Examples:
  - `amux env`
  - `amux agent`
  - `amux manager`

## Environment
When `amux session start` creates a session, it sets:
- `AMUX_REPO`
- `AMUX_SESSION_ID`
- `AMUX_WORKTREE`
- `AMUX_AGENT_NAME`
- If `AMUX_MANAGER` is set, the agent is treated as manager-owned.

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
  - `amux` command wraps payload in `<amux>...</amux>`
  - Required attributes:
    - `owner` (`agent|manager`)
    - `model`
    - `reasoning_level`
    - `state` (`idle|busy|errored|usage-limit`)
