# Changelog

## 0.19.0 — Per-field overrides on `agent create` / `team create`

Promotes `--model` and `--reasoning` to first-class flags and adds a generic repeatable `--set <key>=<value>` so you can override anything declared in a role manifest at launch time without editing the role on disk. On teams, `--set <member>.<field>=<value>` targets a single member spawned from the role's `MEMBERS=(...)` array.

```sh
atmux agent create planner --role architect --model gpt-5.5 --reasoning extra-high
atmux team create alpha --role pair-program \
  --set driver.intelligence=95 \
  --set navigator.adapter=codex
```

Member name match is the last `-`-segment of the member's first token, so role-templated names like `${ATMUX_TEAM}-driver` match on `driver`. Precedence (low → high): role manifest defaults → role `MEMBERS` flags → first-class flags → undotted `--set` → dotted per-member `--set`. Unknown `--set` keys warn and pass through (forward-compat: future adapter manifest fields land without atmux changes).

### Adapter coverage

`agent create` stamps `ATMUX_MODEL` / `ATMUX_REASONING_LEVEL` on the new tmux session, and each adapter's start script honors them after the intelligence_map lookup:

- **claude-code**: `--model <id> --effort <level>` argv flags. Generic override beats both `ATMUX_INTELLIGENCE` and the legacy `ATMUX_CLAUDE_MODEL` / `ATMUX_CLAUDE_EFFORT` envvars.
- **codex**: drives the same `/model` TUI menu used by intelligence_map, so explicit overrides take effect post-boot the same way.
- **cursor-agent**: `--model <id>` argv. Reasoning level is encoded in the cursor-agent model id itself (`composer-2-fast` vs `gpt-5.3-codex-xhigh`), so `--reasoning` updates state.env for status output but `--model` is the knob that moves runtime behavior.
- **gemini**: `--model <id>` argv. The current Gemini CLI exposes no reasoning/effort flag; `--reasoning` is preserved on the session for parity.

## 0.18.2 — Pair-program role docs and demo

- Added README docs for the built-in roles and a GIF demo for the
  `pair-program` team role.
- Added an auto-generated built-in role table to the project README,
  split by agent roles and team roles, including links to role docs and
  demos when present.

## 0.18.0 — Pair-programming team role

Adds three new roles that compose into a driver/navigator pair-programming team. Goal: see how far a fast low-intelligence model can go with smart oversight from a high-intelligence model.

**`pair-program`** (KIND=team). Spawns two members in the same fresh worktree (the team-worktree default from 0.17.0):

- **`driver`** — `cursor-agent` at intelligence 20 (composer-2-fast). Picks up the task, writes code, runs tests, iterates quickly.
- **`navigator`** — `codex` at intelligence 80 (gpt-5.5 high). Watches the worktree via `atmux git watch` and routes each rolling diff back to its own pane via `atmux notify --pane`. Reviews each diff against the task, the agent file (`AGENTS.md`/`CLAUDE.md`), and any planning docs; sends `atmux send --to <team>-driver --interrupt` with a corrective note when the driver veers, errors, takes a shortcut, or drifts off-scope.

Usage:

```sh
atmux team create <name> --role pair-program
atmux send --to <name>-driver "<your task>"
```

Multiple pair-program teams can coexist in one repo. Member names are templated as `${ATMUX_TEAM}-driver` / `${ATMUX_TEAM}-navigator` (e.g. team `feat-a` → agents `feat-a-driver` / `feat-a-navigator`). The role manifests' `MEMBERS` arrays are sourced bash; atmux pre-exports `ATMUX_TEAM` and `ATMUX_REPO` before sourcing so the prefix expands at parse time. A team-of-teams primitive ([#17](https://github.com/gabewillen/atmux/issues/17)) is the next-step generalization.

### Framework changes used by this role
- **`role.md` templating**: the agent script now substitutes a curated set of `${ATMUX_*}` variables (`ATMUX_TEAM`, `ATMUX_REPO`, `ATMUX_AGENT_NAME`, `ATMUX_WORKTREE`, `ATMUX_ROLE`) when it dumps `role.md` into the agent's control file. Lets prompts reference peer agent names like `${ATMUX_TEAM}-driver` so each LLM sees its actual peer's name instead of a placeholder. `${VAR}` form only — bare `$VAR` isn't supported (avoids accidental matches against unrelated env vars).
- **Team-manifest sourcing env**: `atmux team create` now pre-exports `ATMUX_TEAM` and `ATMUX_REPO` before sourcing the role's manifest, so team-kind manifests can parameterize `MEMBERS` with `${ATMUX_TEAM}` to get team-unique session names without templating support inside atmux's spawn path.

The team's fresh worktree, `--shared-worktree` env-aware behavior, cursor-agent workspace pre-trust, and rolling `atmux git watch` are all 0.17.0/0.16.0 primitives this role composes.

## 0.17.0 — Teams own a worktree by default

**Behavior change**: `atmux team create <name>` now creates a fresh git worktree by default — at `$ATMUX_HOME/teams/<repo>/<team>/worktree` on a new branch `atmux-<repo>-team-<name>`. Members spawned with `--shared-worktree` inherit it via `ATMUX_WORKTREE`, so a pair of agents can edit the same worktree without stomping the user's checkout. `team kill` removes the worktree and branch.

To opt out (e.g. a coordination team that just talks and never edits files), pass `--shared-worktree` to `team create`. To override only the path, pass `--worktree <path>`.

Setup work for the upcoming pair-programming team — driver and navigator share the same fresh worktree, navigator's `git watch` reviews the driver's edits in real time.

**`atmux agent create --shared-worktree`** is now env-aware: it picks up `ATMUX_WORKTREE` from env when set, falling back to the repo root only when unset. Members spawned by `team create` inherit the team's worktree automatically — the team's spawn path injects the worktree env into each child `agent create` call. Standalone `--shared-worktree` from a bare shell behaves the same as before.

The team session's env carries `ATMUX_WORKTREE=<path>` and an `ATMUX_TEAM_WORKTREE_OWNED=1` marker so `team kill` knows which worktrees it owns. Removal also runs `git worktree prune` after the directory is gone, so the subsequent `git branch -D` doesn't trip over stale `.git/worktrees/<name>` metadata.

### Other changes
- `team create`'s teardown path (when member spawn or start hook fails) also removes the worktree if the failed create_team had already created it.
- `cursor-agent` adapter pre-trusts the workspace via `<git_root>/.workspace-trusted` at start time. cursor-agent's first run in an unfamiliar directory otherwise blocks on a TUI "Workspace Trust Required" dialog that no autonomous agent can dismiss — every fresh team / agent worktree was an unfamiliar directory until now. The fix uses cursor-agent's own trust-state mechanism (no `--yolo` overreach).
- `bin/atmux` no longer gates commands behind a "must be run inside tmux" check. Calls into `tmux` happen lazily inside the underlying scripts, which surface their own error if no tmux server is reachable. Removes the friction of prefixing every `send` / `notify` / `schedule` from a non-tmux shell with `ATMUX_ALLOW_OUTSIDE_TMUX=1` (the env var stays recognized for back-compat).
- `atmux send --to <team>` no longer crashes under bash 3.2 strict-mode when the team has zero members; falls through to the normal "no targets" error.
- Tests: `117_team_worktree_lifecycle` (covers default + `--shared-worktree` opt-out paths).

## 0.16.0 — `atmux git watch` rolling-diff worktree watcher

Adds a new `git` resource so an agent can subscribe to worktree changes and receive **rolling diffs** — each notification contains only the diff *since the previous notification*, never the full dirty state. Setup work for the upcoming pair-programming team's navigator agent, which will watch the shared worktree and steer the driver when it goes off-track.

**`atmux git watch [--id <id>] [--coalesce <s>] [--interval <s>] [--timeout <s>] [--duration <s>] [--exec <cmd>] [--once]`**. Polls the current repo on `--interval` (default 10s), snapshots the index+worktree via `git stash create` (non-mutating — produces a commit object referencing the dirty state without touching the working tree, the index, or the stash list), resolves it to its tree SHA, and emits a digest per detected change (or per `--coalesce` window if set):

```xml
<watch type="git" id="navigator" prev="<tree>" new="<tree>" events="1" window="0s" reason="change">
  <diff>diff --git a/f b/f
  ...</diff>
</watch>
```

The baseline persists at `$ATMUX_HOME/git/watch/<repo>/<id>/last`, so a watcher restart resumes from its last emit instead of replaying the cumulative diff. Empty diffs (e.g. a change reverted within the coalesce window) are suppressed.

Snapshots are content-addressed at the tree level — comparing tree SHAs (not stash *commit* SHAs, whose timestamp metadata changes every call) means equality genuinely means "no change." Safe to run against a worktree shared with other agents.

### Other changes
- `bin/atmux`'s `resolve_script_subcommand` now falls back to `$ATMUX_SOURCE_ROOT/bin/(atmux)/<cmd>` after the cwd-repo and `~/.atmux` install lookups. New subcommands work the moment they land in the source tree, regardless of cwd or whether the user has run `atmux install`. Fixes a latent gap that would have hit any future top-level noun.
- Tests: `116_watch_git_rolling_diff`.

## 0.15.0 — Roles can now apply to teams

Roles previously only described agents. This release extends the role mechanism so a single role definition can target either an agent (default) or a team — same directory contract (`role.md` + `manifest` + optional `start`/`stop`), one new `KIND` field in the manifest. Setup work for the upcoming driver+navigator paired-programming team, which lands in a follow-up release.

**Manifest gains `KIND=agent|team`** (default `agent`). For `KIND=team` roles, the manifest can also declare `MEMBERS=("<agent create args>" ...)` — each entry is the tail of an `atmux agent create` call. atmux shell-tokenizes each entry (so `--description "multi word"` survives) and runs `atmux agent create <tokens> --team <team>`. Conditionals fall out for free because the manifest is already sourced bash:

```bash
KIND=team
MEMBERS=(
  "driver    --role driver    --intelligence ${DRIVER_IQ:-20} --shared-worktree"
  "navigator --role navigator --intelligence ${NAV_IQ:-95}    --shared-worktree"
)
```

**`atmux team create <name> --role <role>`** resolves a `KIND=team` role, opens the team session, spawns each `MEMBERS` entry, then runs the optional `start` hook for cross-cutting wiring. Hook env contract: `ATMUX_TEAM`, `ATMUX_REPO`, `ATMUX_WORKTREE`, `ATMUX_ROLE`, `ATMUX_ROLE_DIR`, `ATMUX_ROLE_STATE_DIR` (`~/.atmux/teams/<repo>/<team>/role/`).

**Auto-kill of team members.** `atmux team kill` enumerates every agent session whose `ATMUX_TEAM` matches and forwards them through the regular agent-kill flow, so each member's own role-stop hook still fires. The team role's `start`/`stop` are reserved for scripted services (watchers, message buses) — most team roles won't need either, since member-level reactive logic belongs in member roles.

**Kind-mismatch rejection.** `atmux team create --role <agent-kind-role>` fails with `role 'X' is KIND=agent; use 'atmux agent create --role X' instead`. Symmetric error from `atmux agent create --role <team-kind-role>`.

### Other changes
- New `atmux role kind <name>` subcommand reports the resolved kind. `atmux role list` gains a kind column.
- `atmux role create --kind team` scaffolds a team-kind role; the generated `start`/`stop` stubs differ from the agent-kind stubs (members + watchers vs. adapter setup).
- `atmux role create` rejects `--intelligence` and `--adapters` for `--kind team` (those flags only mean something for agents).
- `atmux team create` now accepts `--start <cmd>` / `--stop <cmd>` directly, mirroring `atmux agent create`. `--role` is the usual path; the explicit flags are for one-off teams that don't deserve a named role.
- Internal: `bin/(atmux)/(internal)/kill` propagates `ATMUX_TEAM` and `ATMUX_WORKTREE` to the role-stop hook for team sessions, and skips the adapter wind-down for `ATMUX_SESSION_KIND=team` (no adapter to signal).
- Tests: `113_team_role_kind_validation`, `114_team_role_lifecycle`, `115_team_role_start_failure_tears_down`.

## 0.14.0 — Breaking: `--interrupt` is now a hard abort

`atmux send --interrupt` and `atmux notify --interrupt` previously meant "submit during work using the adapter's interrupt key" — a soft steer. The keys themselves were inconsistent across adapters (`Tab` for codex, `Escape+Enter` for claude-code, plain `Enter` for cursor-agent and gemini), so the flag's behavior also varied. Setup work for an upcoming driver+navigator paired-programming feature needs a real abort, so the flag's semantics are being repurposed.

**New `--interrupt` semantics**: send the adapter's abort key sequence to stop the current operation, then submit the notification.

| adapter      | `submit_keys.interrupt` (abort prefix) | `submit_keys.steer` (submit) |
|--------------|----------------------------------------|------------------------------|
| claude-code  | `["Escape"]`                           | `["Enter"]`                  |
| codex        | `["Escape"]`                           | `["Enter"]`                  |
| cursor-agent | `["C-c"]`                              | `["Enter"]`                  |
| gemini       | `["C-c"]`                              | `["Enter"]`                  |

**Manifest schema rename**: `submit_keys.queue` → `submit_keys.steer`. The previous name implied "deliver after the agent finishes all its current todos" (true queue), but the field is actually used to submit at the next idle prompt — i.e., the steer path. True queue mode doesn't exist yet; calling it `steer` matches its real semantics. The default fallback mode in `resolve_submit_keys` is now `steer` instead of `queue`.

**No replacement flag for the old soft-steer behavior** — what `--interrupt` used to do is now just `atmux send` with no flag. The default steer path was already the right shape for "deliver when the receiving agent is ready"; adapters that supported mid-tool injection (codex's Tab-queues, claude-code's Enter-while-busy) will inject through the same default path.

Internal plumbing: queued items now carry optional `pre_keys` (abort prefix sent before the payload) and `bypass_idle` (skip the worker's wait-for-idle gate). Both are set only by `--interrupt`. Adapter manifests gain new semantics for the existing `submit_keys.interrupt` field — it's the abort prefix now, not the post-message submit keys.

**Caveats per adapter**:
- **codex**: single `Esc` aborts the in-flight task and returns to the prompt — verified.
- **cursor-agent / gemini**: `C-c` may also exit the CLI session, depending on version. There are open upstream issues asking for double-press confirmation. Avoid `--interrupt` in long-lived production agents on these adapters until upstream behavior settles.

### Other changes
- Removed `resolve_interrupt_keys()` helper from `bin/(atmux)/send`. New helpers: `resolve_steer_keys()` and `resolve_abort_keys()` (the latter uses the strict resolver so missing manifest fields don't silently fall back to `Enter`).
- Added `resolve_submit_keys_strict()` to `bin/(atmux)/notify` — same as `resolve_submit_keys` but returns empty + non-zero when the requested mode is missing from the manifest.
- `tests/101_notify_adapter_submit_keys_matrix` asserts the new abort keys.
- `tests/102_notify_adapter_delivery_matrix_mock` exercises the `--interrupt` plumbing end-to-end against a SIGINT-resistant `cat` mock pane.

## 0.13.1 — `--shared-worktree` replaces `--no-worktree`

Renamed `agent create --no-worktree` to `--shared-worktree` to better describe what the flag does: the new agent runs in the caller's *current worktree* rather than getting its own. This is a setup step for an upcoming paired-programming feature where two agents intentionally share one worktree.

- **`atmux agent create --shared-worktree`** — preferred name. Skips worktree creation; the new tmux session opens with `pwd` set to the caller's `git rev-parse --show-toplevel`.
- **`atmux agent create --no-worktree`** — still works, prints `--no-worktree is deprecated; use --shared-worktree instead` to stderr, then proceeds. Existing scripts won't break.

No behavior change beyond the warning. Help text and prose docs updated to say "current worktree" instead of "repo root", which was inaccurate for callers running inside a worktree.

## 0.13.0 — `agent status` and `team status` for cross-agent observability

Two new verbs that read the same on-disk state the running system writes (`tmux` + the notify-queue under `ATMUX_HOME/notify-queue/`), so you can see across all agents in a repo without `capture-pane`-ing one at a time.

- **`atmux agent status [<name>]`** — per-agent runtime state: pane alive, notify-worker liveness (alive/dead/none), pending queue depth, age of the oldest queued notification, and the last worker error (truncated). With no name, lists every agent in the repo; with a name, drills in.
- **`atmux team status [<name>]`** — rollup verb. With no name, one row per team (members / alive / queue_total / errors). With a name, per-member breakdown — same columns as `agent status`. Membership comes from each pane's `@atmux_agent_session` tag.

Both honor the reporter system: human terminals get column-aligned tables, agent context (`ATMUX_AGENT_NAME` set) gets XML records.

This complements `agent list --status`, which shells out to each adapter's `status` hook to surface model/reasoning level. The new verbs answer "is the *plumbing* alive" — pane, worker, queue, errors — without round-tripping the adapter.

Existing tests:
- `tests/141_agent_status_mock` covers all worker-liveness branches and queue/error rendering.
- `tests/142_team_status_mock` covers the rollup math and drill-in for a 3-member team plus a team with an absent member.
- `tests/143_status_real_adapter_live` and `tests/144_team_status_real_adapter_live` exercise the same paths against real adapter sessions (`ATMUX_RUN_LIVE_ADAPTER_TESTS=1`).

## 0.12.1 — Auto-release on VERSION bump

CI workflow that auto-tags and cuts a GitHub release when a push to `main` changes the `VERSION` file. Replaces the manual `git tag vX.Y.Z && git push --tags` step that used to follow every release-PR merge.

The 0.12.0 release missed this path because the workflow itself landed after the 0.12.0 VERSION bump; 0.12.1 is the first release through the auto-tagger.

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
  `atmux agent _render-prompt` diagnostic command now prints the rendered
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
