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

### `./bin/amux session start`
- Must run inside a git repo.
- Reuses an existing repo session if one already exists (`amux-{{repo}}-*`).
- Otherwise creates `amux-{{repo}}-{{name}}` (default `name=manager`).
- Creates a git worktree at `~/.amux/agents/{{repo}}-{{name}}` on branch `amux-{{repo}}-{{name}}`.
- Starts the selected adapter in that session (`--adapter`, default `codex`).

### `./bin/amux session list`
- Lists amux sessions (`amux-*`), one per line.

### `./bin/amux agent list`
- Lists agents owned by this agent (`AMUX_AGENT_NAME`) via `AMUX_MANAGER`.

### `./bin/amux agent list -a`
- Lists all agents in the org (`amux-{{repo}}-{{name}}` sessions).

### `./bin/amux manager send "message"`
- Sends a message to your manager session.

### `./bin/amux agent send --to <agent> "message"`
- Sends a message to a specific agent session.

### `./bin/amux adapter install <owner/repo|github-url>`
- Installs (or updates) adapter repos under `~/.amux/adapters/<adapter>`.

### Script Subcommands
- Any executable in `~/.amux/bin/scripts/<name>` is available as:
  `amux <name>`
- Examples:
  - `amux env`
  - `amux agent`
  - `amux manager`
- `install` is reserved as a core `amux` command, not a script subcommand.

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
- `status` output fields:
  - `owner` (`agent|manager`)
  - `model`
  - `reasoning_level`
  - `state` (`idle|busy|errored|usage-limit`)
