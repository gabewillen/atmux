```sh
atmux process watch <pid> [--timeout <seconds>] [--interval <seconds>]
atmux process watch <pid> --stdio [--duration <seconds>] [--timeout <seconds>] \
                                  [--interval <seconds>] [--lines <n>] \
                                  [--coalesce <seconds>]
atmux process kill  <pid> [--timeout <seconds>] [--signal <NAME>]
```

Operates on `atmux exec`-tracked child processes by pid (state at `~/.atmux/exec/<repo>/<pid>/`).

- **`watch`** — wait for the process to finish; receive the same exit-notification XML the executor would have sent.
- **`watch --stdio`** — monitor a detached exec process pane for output changes. Output events are batched into one digest notification per `--coalesce` window (default `60s`; pass `0` for per-event delivery). Pending output flushes when the process exits or `--duration` / `--timeout` expires.
- **`kill`** — stop the tracked process, wait for executor and watcher fan-out notifications to drain, then remove its metadata.
