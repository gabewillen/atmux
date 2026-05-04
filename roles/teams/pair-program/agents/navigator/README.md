# Navigator

The `navigator` role is the review half of the pair-programming workflow.
It watches a shared worktree, reviews rolling diffs, and steers the
driver without editing files directly.

## Defaults

- Kind: `agent`
- Default adapter: `codex`
- Default intelligence: `80`

## Usage

The navigator is normally created by the pair-program team role:

```sh
atmux team create <name> --role pair-program
atmux send --to <name>-navigator "<your task description, plus any planning notes>"
```

When used inside a pair-program team, the navigator receives worktree
change notifications and driver-idle notifications from the team's
hooks. It reviews the driver's changes and sends feedback with
`atmux send`.

See [role.md](./role.md) for the role prompt.
