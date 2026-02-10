# Agent

An agent can be a directory.

For now, an agent directory can contain:
- `bin/scripts/session` (`session start` handles adapter launch)

Current example:
- `bin/scripts/session`

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `~/.amux/agents/{{repo}}-{{name}}`
