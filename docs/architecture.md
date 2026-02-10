# amux Architecture

## Overview
`amux` is a Go CLI for routing agent commands into tmux.

Constraints we agreed on:
- Implemented with Cobra.
- POC scope only.
- Drives tmux via `gotmux`.

## Current Scope
- List tmux sessions.
- Route one command to an agent target.

No additional orchestration behavior is part of this architecture yet.

## Components

1. CLI (`cmd/*`)
- Built with Cobra.
- Exposes the current commands.

2. Router (`internal/router/*`)
- Resolves agent name to tmux target.
- Sends command payload into the target pane.

3. tmux integration (`gotmux`)
- Uses `github.com/gabefiori/gotmux` for session checks and tmux command execution.

## Code Layout
- `main.go`
- `cmd/root.go`
- `cmd/sessions.go`
- `cmd/route.go`
- `internal/router/router.go`
- `docs/cli.md`
- `docs/architecture.md`

## Runtime Flow

### `amux sessions`
1. CLI command is invoked.
2. `gotmux` lists tmux sessions.
3. Session names are printed.

### `amux route --agent <name> --cmd <command>`
1. CLI validates required flags.
2. Router resolves agent to `session:window.pane`.
3. Router verifies the target session exists.
4. Router sends the command to the pane.

## Route Mapping (POC)
- In-memory default map:
  - `agent-0 -> agent-0:0.0`
  - `agent-1 -> agent-1:0.0`
  - `agent-2 -> agent-2:0.0`
  - `agent-3 -> agent-3:0.0`
  - `agent-4 -> agent-4:0.0`
