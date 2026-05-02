# Changelog

## 0.12.0 — Reporter system: humans get CLI text, agents get XML; remove read-only pane on agent create

Until now, every `atmux` command wrapped its output in an `<atmux command="…">` XML envelope and listings emitted hand-rolled XML records, regardless of who invoked them. Humans running `atmux pr list` from a normal terminal saw raw XML.

This release adds a reporter that picks the format based on the caller:

- **Agent context** (`ATMUX_AGENT_NAME` set) → unchanged: full `<atmux …>` envelope, structured XML records.
- **Human terminal** → plain-text CLI: `Usage:` heredoc for `--help`, column-aligned tables for listings, `(none)` for empty results.

Override env vars:

| Variable | Effect |
|---|---|
| `ATMUX_REPORTER=xml` | Force XML output regardless of caller |
| `ATMUX_REPORTER=cli` | Force CLI text output regardless of caller |
| `ATMUX_REPORTER=auto` (default) | Detect via `ATMUX_AGENT_NAME` |
| `ATMUX_NO_WRAP=1` | Legacy alias for `ATMUX_REPORTER=cli` |

Migrated commands: `agent list`, `team list`, `issue list`, `pr list`, `message list`. `role list`, `watcher list`, `config list`, and `env` were already plain text. Single-record commands (`pr create`, `agent create`, `pr show`, etc.) still emit XML in both modes — they're mostly agent-driven.

Tests that specifically assert XML output set `ATMUX_REPORTER=xml`.

### Read-only pane on agent create — removed

Creating an agent attached the new pane in tmux read-only mode (controlled by `ATMUX_ATTACH_READONLY` and `ATMUX_BOOTSTRAP_STATE=pending`), with a background `_bootstrap-unlock` worker watching for adapter readiness before unlocking input. This made the pane feel frozen for the first few seconds after creation.

`attach_or_switch` now always attaches read-write. The `ATMUX_ATTACH_READONLY` env var, the `bootstrap_unlock_*` helpers, the `_bootstrap-unlock` subcommand, and the read-only attach paths are gone. `ATMUX_BOOTSTRAP_STATE` is still set by adapter `start` scripts and still gates notification delivery in `notify` — that path is untouched.

## 0.11.1

- `atmux send` / notify delivery: large payloads sometimes had their submit Enter swallowed because tmux flushes big pastes in chunks (the captured input briefly looks stable while more bytes are still in flight) and adapters need additional time to ingest the full input before submit. Adds a length-scaled pause between paste and submit. Tunable via:
  - `ATMUX_SEND_PASTE_PER_KB_SECONDS` (default `0.05` — bonus seconds per KB)
  - `ATMUX_SEND_PASTE_BONUS_MIN_BYTES` (default `1024` — payloads under this skip the bonus)
  - `ATMUX_SEND_PASTE_BONUS_MAX_SECONDS` (default `5` — hard cap)

  Set `ATMUX_SEND_PASTE_PER_KB_SECONDS=0` to disable.

## 0.11.0 — Adapter `prepare-prompt` contract; stop mutating project memory files (BREAKING)

atmux no longer writes its `<atmux>` block into the project's
`AGENTS.md` / `CLAUDE.md` / `GEMINI.md`. Each adapter now carries a
`scripts/prepare-prompt` script that decides where atmux content goes —
purely in-memory for adapters that support it, sidecar files in
adapter-specific dotdirs for those that don't.

Per-adapter mechanism:

| Adapter | Mechanism | Mutates anything? |
|---|---|---|
| `claude-code` | `--append-system-prompt "<text>"` | No — pure CLI flag |
| `codex` | `--config developer_instructions=<TOML-quoted-text>` (appends after defaults) | No — pure CLI flag |
| `gemini` | sidecar at `$ATMUX_HOME/agents/<repo>/<agent>/gemini-system.md`, `GEMINI_SYSTEM_MD=<path>` env | Sidecar in `$ATMUX_HOME`, **not** `GEMINI.md` |
| `cursor-agent` | sidecar at `<worktree>/.cursor/rules/atmux.mdc` (cursor auto-discovers it) | Sidecar in `.cursor/rules/`, **not** `AGENTS.md` |

The new contract: each adapter's `cmd prepare-prompt` is invoked with the
rendered atmux block on stdin and two output paths via env
(`ATMUX_PREP_ARGS_FILE` for null-separated argv to prepend;
`ATMUX_PREP_ENV_FILE` for `KEY=VALUE` env directives). The `auto`
adapter delegates to the selected adapter, mirroring `control-file`.

Migration notes:

- Existing `<atmux>` blocks already injected into project `AGENTS.md` /
  `CLAUDE.md` / `GEMINI.md` are no longer touched by atmux. Delete them
  by hand if you want — they're harmless if left in place but they
  duplicate context atmux now passes through the adapter directly.
- For `gemini`, `GEMINI_SYSTEM_MD` is replace-only: the built-in gemini
  prompt is dropped in favor of the atmux block. Run
  `GEMINI_WRITE_SYSTEM_MD=1 gemini --version` once if you want to
  capture the built-in for reference.
- `atmux agent attach` no longer re-renders the prompt; the running
  adapter already has it loaded from `agent create`.
- The internal `inject_control_file` function is gone; an internal
  `atmux _render-prompt` diagnostic command now prints the rendered
  block for tests / debugging.

## 0.10.1

- `atmux process watch --stdio` now batches output events into a single digest notification per `--coalesce` window (default `60s`, pass `0` for the previous per-event behavior). Pending output flushes when the process exits or `--duration` / `--timeout` expires. Digest notifications carry `coalesced="true" events="N" window="<dur>s" reason="…"`.
- `atmux path watch` now streams change events continuously by default with the same `--coalesce` semantics. New `--once` flag preserves the prior single-shot exit-0-on-first-change contract. `--timeout` is now an idle exit (matching `--stdio`); `--duration` was added as a hard cap.
- Bug fix in the watcher helpers: never toggle `set -e` inside a function body — the new setting leaks to the caller, and a non-zero return then trips errexit before the rc can be read.

## 0.10.0 — Session command removed (BREAKING)

The `atmux session` command has been removed. Its functionality is folded
into `atmux agent`:

| Old | New |
|---|---|
| `atmux session start [--name N] [--adapter A] [--adapters ...] [--no-worktree] [-- ...]` | `atmux agent create [N] --role R --intelligence I [--adapter A] [--adapters ...] [--no-worktree] [-- ...]` |
| `atmux session attach <name\|session>` | `atmux agent attach <name\|session>` |
| `atmux session list` | `atmux agent list --all` (or `atmux team list` for teams) |

`atmux agent create` no longer requires a manager. When run as a top-level
command outside any agent session, it works just like the old
`atmux session start` did — it creates the session/worktree, launches the
adapter, and (when interactive) attaches you to it. From inside a manager
agent it keeps its existing sub-agent-spawn behavior and prints the
`<agent>...</agent>` XML envelope.

`--role` and `--intelligence` are now required for every `agent create`,
including top-level invocations that previously relied on `session
start`'s defaults.

## 0.9.2

- Fix `atmux install` failing with "must be run inside tmux" outside a tmux session. The 0.9.0 cutover added `install` to the require-tmux list by mistake — install is the bootstrap command and doesn't touch panes or notifications.
- Fix `atmux install` hanging silently when run interactively. The dispatcher was capturing install's stdout/stderr into temp files for `<atmux>` XML wrapping, so the project-vs-system scope prompt never reached the terminal and the script blocked on `read` with no visible output. Install now bypasses output wrapping (same as `exec`/`watch`/`notify`/`kill`).

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
