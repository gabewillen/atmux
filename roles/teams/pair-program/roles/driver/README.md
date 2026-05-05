# Driver

The `driver` role is the fast implementation half of the pair-programming
workflow. It writes code, runs tests, and responds to navigator feedback.

## Defaults

- Kind: `agent`
- Default adapter: `cursor-agent`
- Default intelligence: `20`

## Usage

The driver is normally created by the pair-program team role:

```sh
atmux team create <name> --role pair-program
atmux send --to <name>-driver "<your task description>"
```

When used inside a pair-program team, the driver shares the team's
worktree with the navigator. The navigator watches changes and sends
corrective messages when the implementation goes off-track.

See [role.md](./role.md) for the role prompt.
