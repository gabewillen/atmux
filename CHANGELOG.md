# Changelog

## 0.9.1

- Rename built-in role `example-pr-reviewer` → `gh-pr-reviewer`. It's a real role, not an example.

## 0.9.0 — Resource-first CLI (BREAKING)

The CLI is now noun-first. Most commands take the form `atmux <noun> <verb>`;
only a handful of cross-cutting verbs (`send`, `exec`, `schedule`, `notify`,
`update`, `install`) have no resource home and stay verb-shaped.

The motivation: the legacy `atmux create --agent X --role Y` /
`atmux create --team Z` / `atmux create --issue --title T` forms reused the
same flag names (`--description`, `--todo`, `--title`, `--repo`, `--team`)
to mean different things depending on which sibling flag was set. That
contextual disambiguation is hard for both humans and LLMs. Resource-first
makes the disambiguation structural: the noun in the command position
determines what the flags mean.

### Migration table

| Old | New |
|---|---|
| `atmux create --agent X --role R --intelligence N [...]` | `atmux agent create X --role R --intelligence N [...]` |
| `atmux create --team X` | `atmux team create X` |
| `atmux create --issue --title T [--description D] [--todo ...] [--repo R]` | `atmux issue create --title T [--description D] [--todo ...] [--repo R]` |
| `atmux create --pr --title T [--source B] [--target B] [...]` | `atmux pr create --title T [--source B] [--target B] [...]` |
| `atmux assign --to A --title T [--given/--when/--then ...]` | `atmux issue create --title T --assign-to A [--given/--when/--then ...]` |
| `atmux assign --issue ID --to A` | `atmux issue assign ID --to A` |
| `atmux comment "msg" --issue ID` | `atmux issue comment ID "msg"` |
| `atmux comment "msg" --pr ID` | `atmux pr comment ID "msg"` |
| `atmux capture --agent X [--lines N]` | `atmux agent capture X [--lines N]` |
| `atmux capture --team X [--lines N]` | `atmux team capture X [--lines N]` |
| `atmux capture --all [--lines N]` | `atmux agent capture --all [--lines N]` |
| `atmux kill --agent X [pat...]` | `atmux agent kill X [pat...]` |
| `atmux kill --pid N` | `atmux process kill N` |
| `atmux kill --watcher ID` | `atmux watcher kill ID` |
| `atmux kill --all [--yes]` | `atmux agent kill --all [--yes]` |
| `atmux watch --target T --text S [...]` | `atmux pane watch T --text S [...]` |
| `atmux watch --pid N [--stdio ...]` | `atmux process watch N [--stdio ...]` |
| `atmux watch --path GLOB` | `atmux path watch GLOB` |
| `atmux watch --agent X [--idle N]` | `atmux agent watch X [--idle N]` |
| `atmux watch --issue ID [--repo R]` | `atmux issue watch ID [--repo R]` |
| `atmux watch --pr ID_OR_URL` | `atmux pr watch ID_OR_URL` |
| `atmux watch --issues REPO` | `atmux issue watch --feed REPO` |
| `atmux watch --prs REPO` | `atmux pr watch --feed REPO` |
| `atmux list teams` | `atmux team list` |
| `atmux list sessions` | `atmux session list` |
| `atmux list agents [--all] [--status]` | `atmux agent list [--all] [--status]` |
| `atmux list issues [--repo R]` | `atmux issue list [--repo R]` |
| `atmux list prs [--repo R]` | `atmux pr list [--repo R]` |
| `atmux list messages [--unread]` | `atmux message list [--unread]` |
| `atmux start [adapter args]` | `atmux session start [adapter args]` |

### Added

- `atmux role create` — scaffold a new role directory in one shot. Body comes
  from `--from-file <path>`, `--from-stdin`, or a generated stub via
  `--description <text>`. Optional `--intelligence`, `--adapters`, `--hooks`,
  `--scope repo|global|auto`. Replaces the manual `mkdir`/`Write` pattern that
  the `atmux-create-role` SKILL used to do by hand — that skill is now a thin
  pipe-arg parser around the new command.
- `atmux issue create --assign-to <agent>` — one-shot create + assign,
  replacing the legacy `atmux assign --to A --title T` form.
- `atmux issue comment <id> "msg"` and `atmux pr comment <id> "msg"`.
- `atmux issue watch --feed <repo>` and `atmux pr watch --feed <repo>` —
  the long-running fan-out modes (replacing `atmux watch --issues|--prs`).
- `atmux watcher list` — enumerate background watcher registrations.
- `atmux agent create <name>` accepts a positional name (in addition to
  the legacy `--name <name>`).

### Removed

- `atmux create`, `atmux assign`, `atmux comment`, `atmux capture`,
  `atmux kill`, `atmux watch`, `atmux list`, `atmux start` — all deleted.
  Internal implementations preserved at `bin/(atmux)/(internal)/{kill,capture,comment}`
  for use by noun scripts; not user-facing.
- `tests/37_watch_router_mode_conflict` — validated mode-conflict rejection
  in the polymorphic `watch` dispatcher; obsolete under noun-first.

### Internal

- New noun scripts: `pane`, `path`, `process`, `watcher`. The `watch` and
  `kill` verbs in these (and in `agent`, `team`, `issue`, `pr`) currently
  delegate via `exec` to the `[watch]/<mode>` and `(internal)/{kill,capture}`
  helpers; the bodies stay where they are to keep this cutover focused on
  the user-facing CLI shape.
- Top-level dispatcher `bin/atmux` now renders help in two static groups
  (Resources, Verbs) and resolves the script BEFORE checking for tmux so
  unknown commands surface as "unknown command" rather than the confusing
  "must be run inside tmux".

### Fixed

- `bin/(atmux)/message`: `message_is_for_current_agent` returned early
  from inside `while ... done < <(producer)`, which intermittently
  triggered SIGABRT under macOS bash 3.2's process-substitution cleanup
  race. Switched to a here-string with the producer captured up front.
- `bin/(atmux)/session`: when a role start hook failed, `run_adapter`
  signaled `ATMUX_ROLE_START_STATE=failed` before cleaning up the
  worktree. The parent `agent _create` poll then killed the session in
  response, terminating cleanup mid-flight and leaking the worktree.
  Reordered to clean up worktree+branch first, then signal, then kill.
- `tests/runner`: glob `[0-9][0-9]_*` silently skipped tests with 3-digit
  prefixes (`100_`, `110_`–`112_`); broadened to `[0-9]*_*`. Live-test
  skip list replaced with a `*_live` suffix pattern so newly added live
  tests are gated by default.

## Earlier versions

See `git log` — no CHANGELOG was kept before 0.9.0.
