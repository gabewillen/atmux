# Agent

An agent can be a directory.

For now, an agent directory can contain:
- `bin/amux/session` (`session start` handles adapter launch)

Current example:
- `bin/amux/session`

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `~/.amux/agents/{{repo}}-{{name}}`
