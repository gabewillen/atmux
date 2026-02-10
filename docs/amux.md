# amux

`amux` is a bash CLI for starting and interacting with repo-scoped tmux sessions.

## Install Layout
When `amux` runs, it ensures this home layout exists:
- `~/.amux/`
- `~/.amux/bin/`
- `~/.amux/roles/`
- `~/.amux/agents/`
- `~/.amux/adapters/`

## Session Model
- Session name: `amux-{{repo}}-{{name}}`
- Initial window/tab name: `amux-{{repo}}-{{name}}`

## Commands

### `./bin/amux start`
- Must run inside a git repo.
- Reuses existing session `amux-{{repo}}-{{name}}` if present.
- Otherwise creates a new agent session and runs `scripts/start` in the repo root.
- `scripts/start` creates `~/.amux/agents/{{repo}}-{{name}}` on branch `amux-{{repo}}-{{name}}`, then runs `bin/scripts/start` inside that worktree.

### `./bin/amux ls`
- Lists amux sessions (`amux-*`), one per line.

### `./bin/amux agents list`
- Lists agents owned by this agent (`AMUX_AGENT_NAME`) via `AMUX_MANAGER`.

### `./bin/amux agents list -a`
- Lists all agents in the org (`amux-{{repo}}-{{name}}` sessions).

### `./bin/amux route --agent <name> --cmd "..."`
### `./bin/amux route --target <session:window.pane> --cmd "..."`
- Sends a command to a tmux pane.

### `./bin/amux adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `~/.amux/adapters/<adapter>`.

### Script Subcommands
- Any executable in `~/.amux/bin/scripts/<name>` is available as:
  `amux <name>`
- Examples:
  - `amux env`
  - `amux send-message`
- `install` is reserved as a core `amux` command, not a script subcommand.

## Environment
When `amux start` creates a session, it sets:
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
- `status` output fields:
  - `owner` (`agent|manager`)
  - `model`
  - `reasoning_level`
  - `state` (`idle|busy|errored|usage-limit`)
