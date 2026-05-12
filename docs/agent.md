# Agent

An agent is a tmux session running an AI CLI under `atmux`'s control.

The agent CLI lives at `bin/(atmux)/agent` and owns the full agent
lifecycle:
- `agent create` — provision worktree, set tmux session env, launch the
  adapter via the in-pane `_run-adapter` bootstrap, and attach when
  interactive.
- `agent attach` — re-attach to an existing agent session from outside
  tmux.
- `agent list|kill|capture|watch|resolve` — manage running agents.

## Kill Semantics
- `agent kill <name|pattern>` is lifecycle-only: it stops the tmux session,
  removes agent idle watchers, and runs the role stop hook while preserving the
  agent's worktree and branch.
- Pass `--cleanup` to remove the agent-owned worktree and branch:
  `agent kill --cleanup <name|pattern>`.
- `agent kill --all --yes` follows the same rule; add `--cleanup` only when
  deleting all matching worktrees and branches is intended.

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `<ATMUX_HOME>/agents/{{repo}}-{{name}}`
- Default worktree creation initializes submodules with:
  `git submodule update --init --recursive`
