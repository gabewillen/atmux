# Agent

An agent can be a directory.

For now, an agent directory can contain:
- `bin/(atmux)/session` (`session start` handles adapter launch)

Current example:
- `bin/(atmux)/session`

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `<ATMUX_HOME>/agents/{{repo}}-{{name}}`
- Default worktree creation initializes submodules with:
  `git submodule update --init --recursive`
