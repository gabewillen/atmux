# Agent TMUX (**atmux**)

**Agent TMUX** is a tmux-first toolkit for running and coordinating multiple AI agents (different CLIs, adapters, and repos) in parallel sessions with shared messaging, issues, capture, and scheduling.

The command-line entrypoint is **`atmux`**. Install it on your `PATH` (e.g. `~/.atmux/bin` via `./install.sh`), run **`atmux`** with no arguments to start tmux with `atmux` on `PATH`, then use subcommands like `atmux create`, `atmux send`, and `atmux capture`.

## Environment

- Default state lives under **`~/.atmux`** (`ATMUX_HOME`).
- Session and agent context use the **`ATMUX_*`** variables (see `atmux env` and `AGENTS.md`).

## Docs

- `docs/atmux.md` — overview
- `docs/cli.md` — CLI notes
- `docs/agent.md` — agent/session concepts

## License

See the repository license file if present.
