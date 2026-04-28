# Agent

An agent is a tmux session running an AI CLI under `atmux`'s control.

The agent CLI lives at `bin/(atmux)/agent` and owns the full agent
lifecycle:
- `agent create` — provision worktree, set tmux session env, launch the
  adapter via the in-pane `_run-adapter` bootstrap, and (when interactive
  with no manager) attach to the new session.
- `agent attach` — re-attach to an existing agent session from outside
  tmux.
- `agent list|kill|capture|watch|resolve` — manage running agents.

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `<ATMUX_HOME>/agents/{{repo}}-{{name}}`
- Default worktree creation initializes submodules with:
  `git submodule update --init --recursive`
