# Pair-Programming Team

A driver-and-navigator experiment: a fast low-intelligence model writes
code while a high-intelligence model watches the shared worktree in real
time and steers the driver when it goes off-track. Goal: see how far a
fast model can go with smart oversight.

## Roles

- **driver** — `cursor-agent` at intelligence 20 (composer-2-fast). Picks
  up the task from the user, writes code, runs tests, iterates quickly.
- **navigator** — `codex` at intelligence 80 (gpt-5.5 high). Watches the
  shared worktree via the `git watch` shim. On every diff, reviews against
  the task, the agent file (`AGENTS.md` / `CLAUDE.md`), and any planning
  docs in the repo. Sends `atmux send --to <team>-driver --interrupt`
  with a corrective note when the driver veers, errors, takes a
  shortcut, or drifts off-scope.

## Setup

`atmux team create` creates a fresh git worktree by default at
`$ATMUX_HOME/teams/<repo>/<team>/worktree` (override with
`--worktree <path>`). Both members run in that worktree via
`--shared-worktree`, so the navigator sees the driver's edits without
copying — and your actual checkout stays untouched.

The navigator's `start` hook arms the `git watch` shim against the worktree at
a 10s poll interval, with `--coalesce 0` so each detected change
becomes its own review trigger before more changes pile on top. Diffs
are rolling — each review message contains only what changed since the
previous emit, never the cumulative diff.

Agent names are prefixed with the team name (e.g. team `feat-a` →
`feat-a-driver` / `feat-a-navigator`), so multiple pair-program teams
can coexist in one repo. The role.md prompts use `${ATMUX_TEAM}-driver`
/ `${ATMUX_TEAM}-navigator` placeholders that atmux substitutes at
agent-create time, so each LLM sees its actual peer's name.

## Usage

```sh
atmux team create <name> --role pair-program
atmux send --to <name>-driver "<your task description>"
```

The navigator wakes up automatically when the driver edits files. Send
the navigator the same task (or any planning context) so it has a
baseline to review against:

```sh
atmux send --to <name>-navigator "<your task description, plus any planning notes>"
```

To shut down (kills both members and tears down the watcher):

```sh
atmux team kill <name>
```
