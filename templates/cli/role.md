```sh
atmux role list
atmux role show <name>
atmux role resolve <name>
atmux role kind <name>
atmux role create <name> (--from-file <path> | --from-stdin | --description <text>) \
                         [--kind agent|team] \
                         [--intelligence <0-100>] [--adapters <a,b,...>] \
                         [--hooks <start,stop>] [--scope repo|global|auto] [--force]
```

Roles are adapter-agnostic. A role is a directory containing any of:

- `role.md` — prompt body, appended under `# Role` in the agent's control file
- `manifest` — optional, sourced bash: `KIND=<agent|team>`, `INTELLIGENCE=<0-100>`, `ADAPTERS=(name ...)`, `MEMBERS=("<agent-create args>" ...)` (team kind only)
- `start` — runs before the adapter starts (at agent-create time) or after the team session opens (at team-create time)
- `stop` — runs after the adapter exits (at agent-kill time) or after a team is killed (at team-kill time)

`KIND` defaults to `agent`. `KIND=team` roles are consumed by `atmux team create --role <name>`. For team-kind roles, `MEMBERS` is an array of strings; each entry is appended verbatim to `atmux agent create` to spawn one member with `--team <team>` already injected. Team kill auto-kills any agent whose `ATMUX_TEAM` matches, so `start`/`stop` are reserved for cross-cutting wiring (watchers, message buses) — most team roles won't need them.

The team start hook receives: `ATMUX_TEAM`, `ATMUX_REPO`, `ATMUX_WORKTREE`, `ATMUX_ROLE`, `ATMUX_ROLE_DIR`, `ATMUX_ROLE_STATE_DIR` (`~/.atmux/teams/<repo>/<team>/role/`).

Resolution precedence (first match wins): `<repo>/.atmux/roles/<name>` → `~/.atmux/roles/<name>` → `<source-root>/roles/<name>`.

`create` writes the role to `~/.atmux/roles/<name>` by default. `--scope repo` writes under `<repo>/.atmux/roles/<name>`; `--scope auto` picks repo if inside a git repo with an existing `.atmux/`, otherwise global.
