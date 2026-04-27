```sh
atmux process watch <pid> [--timeout <seconds>] [--interval <seconds>]
atmux process watch <pid> --stdio [--duration <seconds>] [--timeout <seconds>] \
                                  [--interval <seconds>] [--lines <n>]
atmux process kill  <pid> [--timeout <seconds>] [--signal <NAME>]
```

Operates on `atmux exec`-tracked child processes by pid (state at `~/.atmux/exec/<repo>/<pid>/`).

- **`watch`** — wait for the process to finish; receive the same exit-notification XML the executor would have sent.
- **`watch --stdio`** — monitor a detached exec process pane for output changes; sends a notification each time new output appears. Exits when `--duration` expires, `--timeout` (no new output) expires, or the process exits.
- **`kill`** — stop the tracked process, wait for executor and watcher fan-out notifications to drain, then remove its metadata.
