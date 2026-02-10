# amux CLI Design

## Scope
`amux` is a Cobra-based CLI that routes agent commands into tmux for a POC.

Current commands only:
- `sessions`
- `route`

Implementation:
- Cobra (`github.com/spf13/cobra`)
- tmux integration via `github.com/gabefiori/gotmux`

## Command Model
Global form:
`amux <command> [flags]`

## Commands

### `amux sessions`
List active tmux sessions.

Example:
- `amux sessions`

### `amux route`
Send one command to an agent target.

Flags:
- `--agent <name>`
- `--target <session:window.pane>`
- `--cmd <string>`

Current behavior:
- `--cmd` is required.
- Use either `--agent` or `--target`.
- If `--agent` is used, the target is resolved from the in-memory map.

Examples:
- `amux route --agent agent-1 --cmd "echo hello"`
- `amux route --target agent-2:0.0 --cmd "pwd"`

## POC Agent Map
- `agent-0 -> agent-0:0.0`
- `agent-1 -> agent-1:0.0`
- `agent-2 -> agent-2:0.0`
- `agent-3 -> agent-3:0.0`
- `agent-4 -> agent-4:0.0`
