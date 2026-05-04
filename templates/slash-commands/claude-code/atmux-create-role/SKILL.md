---
name: atmux-create-role
description: Scaffold a new ATMUX role directory with `atmux role create`. Use when the user wants to author a reusable role for `atmux agent create --role <name>`.
allowed-tools: Bash(atmux role create*, mktemp*, rm*)
---
# ATMUX Create Role

Use `atmux role create` to scaffold a new role directory. The created role is discoverable by `atmux role resolve --kind agent <name>` and usable as `atmux agent create <agent> --role <name>`.

## Arguments

`$ARGUMENTS` should contain fields separated by `|`:

`name | description | intelligence? | adapters? | hooks? | scope? | parent-team?`

| # | Field | Required | Notes |
|---|-------|----------|-------|
| 1 | `name` | yes | Role identifier (`^[A-Za-z0-9._-]+$`). |
| 2 | `description` | yes | One- to few-sentence statement of what the role does. |
| 3 | `intelligence` | no | Integer 0–100. |
| 4 | `adapters` | no | Comma-separated subset (e.g. `claude-code,codex`). |
| 5 | `hooks` | no | Comma-separated subset of `start,stop` (or `none`). |
| 6 | `scope` | no | `repo` \| `global` \| `auto` (default `auto`). |
| 7 | `parent-team` | no | Team name for a private member role under `roles/teams/<team>/agents/<name>`. |

Examples:

- `/atmux-create-role pr-reviewer | reviews incoming GitHub PRs and posts structured feedback | 75 | claude-code,codex | start,stop | repo`
- `/atmux-create-role tester | runs the test suite for assigned changes and reports failures | 55`

## Behavior

1. Split `$ARGUMENTS` on `|` and trim whitespace from each field.
2. If `description` is non-trivial enough to warrant a hand-authored `role.md`, draft the body yourself (heading, "You are…" framing, `## Workflow` steps grounded in the description, `## Tools available` pointing at `atmux`/`git`/`gh` as relevant). Write it to a temp file via `mktemp`. Otherwise, skip this step and let `atmux role create --description ...` generate a stub.
3. Run `atmux role create <name>` with:
   - `--from-file <tmpfile>` if you authored the body, OR `--description "<text>"` to use the built-in stub.
   - `--intelligence <n>` if provided.
   - `--adapters <list>` if provided.
   - `--hooks <list>` if provided (or `--hooks none` to be explicit).
   - `--scope <repo|global|auto>` if provided (defaults to `auto`).
   - `--parent-team <team>` if provided.
4. Clean up the temp file with `rm -f`.
5. Report the path printed by `atmux role create` and the suggested invocation:
   `atmux agent create <suggested-agent-name> --role <name> --intelligence <n> [--adapter <first-adapter>]`

`atmux role create` validates the name, refuses to overwrite without `--force`, strips reserved `<atmux>` fence lines from supplied bodies, writes the manifest with `INTELLIGENCE`/`ADAPTERS=()`, and `chmod +x`'s any hooks. Trust it for those concerns.
