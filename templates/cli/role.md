```sh
atmux role list
atmux role show <name>
atmux role resolve [--kind agent|team] <name>
atmux role kind <name>
atmux role create <name> (--from-file <path> | --from-stdin | --description <text>) \
                         [--kind agent|team] \
                         [--parent-team <name>] \
                         [--intelligence <0-100>] [--adapters <a,b,...>] \
                         [--hooks <start,stop>] [--scope repo|global|auto] [--force]
```

Roles are adapter-agnostic. A role is a directory containing any of:

- `role.md` — prompt body, appended under `# Role` in the agent's control file
- `manifest` — optional, sourced bash: `KIND=<agent|team>`, `INTELLIGENCE=<0-100>`, `ADAPTERS=(name ...)`, `MEMBERS=("<agent-create args>" ...)` (team kind only)
- `start` — runs before the adapter starts (at agent-create time) or after the team session opens (at team-create time)
- `stop` — runs after the adapter exits (at agent-kill time) or after a team is killed (at team-kill time)

`KIND` defaults to `agent`. `KIND=team` roles are consumed by `atmux team create --role <name>`. For team-kind roles, `MEMBERS` is an array of strings; each entry is shell-tokenized (so `--description "multi word"` survives) and appended to `atmux agent create` to spawn one member with `--team <team>` already injected. The optional `start` hook runs *after* members spawn, for cross-cutting wiring (watchers, message buses) — most team roles won't need it. Team kill auto-kills any agent whose `ATMUX_TEAM` matches and runs the optional `stop` hook.

The team start hook receives: `ATMUX_TEAM`, `ATMUX_REPO`, `ATMUX_WORKTREE`, `ATMUX_ROLE`, `ATMUX_ROLE_DIR`, `ATMUX_ROLE_STATE_DIR` (`~/.atmux/teams/<repo>/<team>/role/`).

Resolution precedence is kind-aware. Agent roles resolve from `roles/agents/<name>`, team roles resolve from `roles/teams/<name>`, and team member spawning temporarily prepends the parent team's `roles/teams/<team>/agents/` directory so private members like `driver` are only reachable from that team.

`create` writes agent roles to `roles/agents/<name>` and team roles to `roles/teams/<name>`. `--parent-team <team>` with `--kind agent` writes a team-private member role to `roles/teams/<team>/agents/<name>`. `--scope repo` writes under `<repo>/.atmux/roles/...`; `--scope auto` picks repo if inside a git repo with an existing `.atmux/`, otherwise global.
