# Pair-Programming Team

A driver-and-navigator team role where a fast model writes code while a
stronger model watches the shared worktree and interrupts with review
notes when the implementation drifts.

![Pair-program team demo](./demo.gif)

## Usage

```sh
atmux team create <name> --role pair-program
atmux send --to <name>-driver "<your task description>"
atmux send --to <name>-navigator "<your task description, plus any planning notes>"
```

The driver and navigator run in the same team worktree by default, so
the navigator can review each rolling diff as the driver edits files.

See [role.md](./role.md) for the full role behavior and setup notes.
