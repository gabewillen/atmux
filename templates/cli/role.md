```sh
atmux role list
atmux role show <name>
atmux role resolve <name>
atmux role create <name> (--from-file <path> | --from-stdin | --description <text>) \
                         [--intelligence <0-100>] [--adapters <a,b,...>] \
                         [--hooks <start,stop>] [--scope repo|global|auto] [--force]
```

Roles are adapter-agnostic. A role is a directory containing any of:

- `role.md` — prompt body, appended under `# Role` in the agent's control file
- `manifest` — optional, sourced bash: `INTELLIGENCE=<0-100>`, `ADAPTERS=(name ...)`
- `start` — runs before the adapter starts (at agent-create time)
- `stop` — runs after the adapter exits (at agent-kill time)

Resolution precedence (first match wins): `<repo>/.atmux/roles/<name>` → `~/.atmux/roles/<name>` → `<source-root>/roles/<name>`.

`create` writes the role to `~/.atmux/roles/<name>` by default. `--scope repo` writes under `<repo>/.atmux/roles/<name>`; `--scope auto` picks repo if inside a git repo with an existing `.atmux/`, otherwise global.
