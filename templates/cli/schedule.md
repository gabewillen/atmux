```sh
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] --notification "<text>"
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] -- <command> [args...]
```

Schedule a future or recurring action.

- **Notification mode** (`--notification`) — queues an ATMUX notification back to the current agent's session. Use this for self-reminders, ticks, and status checks.
- **Command mode** (`-- <command...>`) — runs the command in the current environment. Only schedule `atmux send` when the target is **another** agent or team; never schedule `atmux send --to <self>` (use notification mode instead).
- **`--no-detach`** — run in the foreground (blocking). By default the scheduled task runs in a detached tmux window so the command returns immediately.

Durations accept a unit suffix: `ms`, `s`, `m`, `h`, `d`. No suffix means seconds. Examples: `45s`, `30m`, `2h`.
