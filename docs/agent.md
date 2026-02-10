# Agent

An agent can be a directory.

For now, an agent directory can contain:
- `scripts/start`

Current example:
- `bin/scripts/start`

## Worktree Convention
- When an agent starts in a repo, it uses a git worktree at:
  `~/.amux/agents/{{repo}}-{{name}}`
- For the manager session, `{{name}}` is always `manager`.
- Manager worktree path is therefore:
  `~/.amux/agents/{{repo}}-manager`
