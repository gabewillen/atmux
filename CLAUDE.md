






























<atmux>
# Role
- ROLE: implementer

# atmux Rules
- Use plain `atmux ...` commands; do not prefix them with inherited session environment.

# Managed Agent Rules
- ALWAYS acknowledge manager messages quickly with a short plan.
- ALWAYS send a message to your manager when stuck or after completing any task.
- ALWAYS message your manager with `atmux send --to manager "..."`.
- ALWAYS coordinate with peer agents using `atmux send --to <agent> "..."`.
- ALWAYS check `atmux agent list --all --status` before creating new agents.
- ALWAYS reuse idle capable agents before creating new ones.
- ALWAYS spawn agents to decompose your todos if necessary.
- ALWAYS use `--reply-required` when a manager decision is needed.
- NEVER poll agent panes unless absolutely necessary.
- NEVER silently change scope; ask your manager first.
- NEVER report task completion without validation evidence.
- NEVER leave blockers unreported; escalate immediately.

<!-- ATMUX-HELP-BEGIN -->
# atmux help

Generated from each command's `usage()` heredoc by
`bin/(atmux)/(internal)/render-docs`. Edit the source `usage()` and re-run.

```text
Usage:
  atmux <noun> <verb> [args...]    # resource operations
  atmux <verb> [args...]           # cross-cutting verbs

atmux is an agent multiplexer across adapters.

Resources (use `atmux <noun> --help` for verbs):
  agent      manage agent sessions (create, attach, list, kill, capture, watch, resolve)
  team       manage team sessions (create, list, kill, capture, resolve)
  role       manage role definitions (create, list, show, resolve)
  issue      filesystem issue tickets (create, list, show, assign, claim, comment, watch)
  pr         filesystem pull request tickets (create, list, show, assign, claim, comment, watch)
  message    inter-agent messages (read, list)
  config     atmux configuration (get, set, unset, list)
  env        environment vars (get, list)
  adapter    AI CLI adapters (install, ...)
  shim       executable shims for agent sessions (install, list, show, resolve)
  watcher    background watcher registrations (list, kill)
  process    atmux exec-tracked processes (watch, kill)
  pane       tmux pane operations (watch)
  path       filesystem path operations (watch)
  git        git working tree operations (watch)
  gh         GitHub CLI operations with PR notification hooks

Verbs (no resource home):
  send       message another agent or team
  exec       run a command with notification on exit
  schedule   schedule a one-shot or recurring action
  notify     low-level notification CLI (mostly internal)
  update     update atmux to the latest version
  install    install atmux into a project or system

Run:
  atmux <subcommand> --help
for more information.
```

## Resources

#### `agent`

```sh
atmux agent create [name] --role <role> --intelligence <0-100>
                   [--team <team>] [--adapter <adapter>] [--adapters <list>]
                   [--model <id>] [--reasoning <low|medium|high|extra-high>]
                   [--set <key>=<value>]...
                   [--shared-worktree] [--start <cmd>] [--stop <cmd>]
                   [--task --description <desc> --todo <todo>...]
                   [-- <adapter-args...>]
  (or `--name <name>` instead of positional; if omitted, atmux
   auto-generates `agent-N`.)
  `--shared-worktree` runs the new agent in the caller's current
  worktree instead of creating a private one for it (deprecated
  alias: `--no-worktree`).
  `--model` / `--reasoning` skip the role's intelligence_map lookup
  and force a specific model/reasoning level on the adapter. `--set
  <key>=<value>` is a generic override; known keys (model, reasoning,
  adapter, intelligence) map onto the corresponding flag, unknown
  keys emit a warning and are passed through unchanged.
atmux agent list [--all] [--status]
atmux agent status [<name>]
atmux agent attach <name|session>
atmux agent kill <name|pattern> [<name|pattern>...]
atmux agent kill --all [--yes]
atmux agent capture <name> [--lines <n>]
atmux agent capture --all [--lines <n>]
atmux agent watch <name> [--idle <seconds>] [--timeout <seconds>]
                         [--interval <seconds>] [--lines <n>]
atmux agent resolve <name> [<repo_name>]
```

Manage atmux agents — sessions running an AI CLI under tmux. Agents are
scoped to the current repo. `create` works both as a top-level command
(no manager required) and from inside a manager agent. When run
interactively without a manager, the new agent is attached to.

agent attach must be run outside tmux.

#### `team`

```sh
atmux team create <name> [--role <role>] [--start <cmd>] [--stop <cmd>]
                         [--worktree <path>] [--shared-worktree]
                         [--adapter <adapter>] [--intelligence <0-100>]
                         [--model <id>] [--reasoning <level>]
                         [--set <key>=<value>]...
                         [--set <member>.<key>=<value>]...
atmux team list
atmux team ls
atmux team status [<name>]
atmux team capture <name> [--lines <n>]
atmux team kill <name|pattern> [...]
atmux team resolve <name> [<repo_name>]
```

Manage repo-scoped team tmux sessions.
Team session format: atmux-<repo>-team-<name>

Every team gets its own fresh git worktree by default — members spawned
with `--shared-worktree` inherit it via ATMUX_WORKTREE, so a pair of
agents can edit the same worktree without stomping the user's checkout.
`team kill` removes the worktree and branch.

--role <name>       Apply a KIND=team role: opens the session, spawns
                    each MEMBERS entry from the role manifest, then
                    runs the role's optional `start` hook for any
                    cross-cutting wiring. `team kill` auto-kills the
                    spawned members and runs the matching `stop` hook.

--worktree <path>   Override the worktree location. Default is
                    $ATMUX_HOME/teams/<repo>/<team>/worktree on a new
                    branch atmux-<repo>-team-<team>.

--shared-worktree   Opt out of the fresh-worktree default. Use the
                    caller's pwd (or repo root) instead. Useful for
                    coordination teams that just talk and never edit.

--adapter / --intelligence / --model / --reasoning
                    Apply to every member spawned from MEMBERS=(...).
                    Per-member overrides via `--set <member>.<field>=
                    <value>` take precedence (member name matches the
                    last `-`-segment of the member's first token, e.g.
                    `${ATMUX_TEAM}-driver` matches on `driver`).

#### `role`

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

Resolution precedence is kind-aware. Agent roles resolve from `roles/agents/<name>`, team roles resolve from `roles/teams/<name>`, and team member spawning temporarily prepends the parent team's `roles/teams/<team>/roles/` directory so private members like `driver` are only reachable from that team.

`create` writes agent roles to `roles/agents/<name>` and team roles to `roles/teams/<name>`. `--parent-team <team>` with `--kind agent` writes a team-private member role to `roles/teams/<team>/roles/<name>`. `--scope repo` writes under `<repo>/.atmux/roles/...`; `--scope auto` picks repo if inside a git repo with an existing `.atmux/`, otherwise global.

#### `issue`

```sh
atmux issue create --title <title> [--description <description>]
                   [--given <context>] [--when <action>] [--then <outcome>]
                   [--todo <todo>]... [--repo <repo>]
                   [--assign-to <agent|session>]
atmux issue list [--repo <repo>]
atmux issue ls [--repo <repo>]
atmux issue get <id> [--repo <repo>]
atmux issue show <id> [--repo <repo>]
atmux issue assign <id> --to <agent|session> [--repo <repo>]
atmux issue claim <id> [--by <agent|session>] [--repo <repo>]
atmux issue comment <id> "message" [--repo <repo>]
atmux issue watch <id> [--repo <repo>] [--timeout <s>] [--interval <s>]
atmux issue watch --feed <repo|url> [--timeout <s>] [--interval <s>]
```

Repo-scoped issue tickets on filesystem.
Issues are stored at: ~/.atmux/issues/<repo>/<id>/

`create --assign-to` creates the issue and immediately assigns it.
`watch <id>` waits for the next update on a single issue.
`watch --feed <repo>` watches a GitHub repo for newly-filed issues.

#### `pr`

```sh
atmux pr create --title <title> [--description <description>]
                [--source <branch>] [--target <branch>]
                [--todo <todo>]... [--repo <repo>]
atmux pr list [--repo <repo>]
atmux pr ls [--repo <repo>]
atmux pr get <id> [--repo <repo>]
atmux pr show <id> [--repo <repo>]
atmux pr assign <id> --to <agent|session> [--repo <repo>]
atmux pr claim <id> [--by <agent|session>] [--repo <repo>]
atmux pr comment <id> "message" [--repo <repo>]
atmux pr watch <id|atmux-uri|github-url> [--repo <repo>] [--timeout <s>] [--interval <s>]
atmux pr watch --feed <repo|url> [--timeout <s>] [--interval <s>]
```

Repo-scoped pull request tickets on filesystem.
Pull requests are stored at: ~/.atmux/pull-requests/<repo>/<id>/

`watch <id>` watches a single pull request (filesystem ticket or GitHub URL).
`watch --feed <repo>` watches a GitHub repo for newly-filed pull requests.

#### `message`

```sh
atmux message read <id> [--repo <repo>]
atmux message list [--unread]
```

Read or list filesystem-backed messages.
Messages are stored at: ~/.atmux/messages/<repo>/<id>/

#### `config`

```sh
atmux config get   <key>          [--global]
atmux config set   <key> <value>  [--global]
atmux config unset <key>          [--global]
atmux config list                 [--global]
```

One file per key under `<repo>/.atmux/config/<key>` (local) or `~/.atmux/config/<key>` (with `--global`). Key paths are hierarchical with `/`, e.g. `update/auto`. `get` without `--global` resolves local first, then global.

#### `env`

```sh
atmux env
atmux env get <key>
```

Inspect ATMUX_* environment variables in the current process.

```sh
atmux env
atmux env get repo
atmux env get ATMUX_WORKTREE
```

#### `adapter`

```sh
atmux adapter install <owner/repo|github-url>
atmux adapter <name> <command> [args...]
```

Install adapters and run adapter contract commands.

```sh
atmux adapter install org/my-adapter
atmux adapter codex status
atmux adapter codex model list
```

#### `shim`

```sh
atmux shim install <name|owner/repo|github-url>
atmux shim list
atmux shim show <name>
atmux shim resolve <name>
```

Manage executable shims activated for agent sessions.

A shim is a directory containing:
  manifest      sourced bash: NAME, KIND=path-prefix, DESCRIPTION, BINARIES=(...)
  <binary>      one executable wrapper per BINARIES entry

Resolution precedence:
<repo>/.atmux/shims/<name>
~/.atmux/shims/<name>
<source_root>/shims/<name>

Installed shims are active for every adapter by default. An adapter can opt
out by setting SHIMS=off in its manifest.

#### `watcher`

```sh
atmux watcher list
atmux watcher kill <id> [--timeout <seconds>]
```

List or remove background watcher registrations created by `atmux pr watch
--feed`, `atmux issue watch --feed`, and `atmux pr watch <id|url>` (the
long-running fan-out modes).

Watcher ids have the form <kind>:<key>:<watcher_name>, e.g.:
  pr:owner_repo_pr_123:atmux-myrepo-worker-_12
  issues:owner_repo_issues:atmux-myrepo-worker-_12
  prs:owner_repo_prs:atmux-myrepo-worker-_12

#### `process`

```sh
atmux process watch <pid> [--timeout <seconds>] [--interval <seconds>]
atmux process watch <pid> --stdio [--duration <seconds>] [--timeout <seconds>] \
                                  [--interval <seconds>] [--lines <n>] \
                                  [--coalesce <seconds>]
atmux process kill  <pid> [--timeout <seconds>] [--signal <NAME>]
```

Operates on `atmux exec`-tracked child processes by pid (state at `~/.atmux/exec/<repo>/<pid>/`).

- **`watch`** — wait for the process to finish; receive the same exit-notification XML the executor would have sent.
- **`watch --stdio`** — monitor a detached exec process pane for output changes. Output events are batched into one digest notification per `--coalesce` window (default `60s`; pass `0` for per-event delivery). Pending output flushes when the process exits or `--duration` / `--timeout` expires.
- **`kill`** — stop the tracked process, wait for executor and watcher fan-out notifications to drain, then remove its metadata.

#### `pane`

```sh
atmux pane watch <target> --text <needle> [--scope pane|window|session]
                          [--timeout <seconds>] [--interval <seconds>]
                          [--lines <n>]
```

Operate on a tmux pane by id (e.g. %12, @3:1.0). The only verb today is
`watch`, which polls the pane and exits 0 when <needle> appears, non-zero
on timeout.

```sh
atmux pane watch %12 --scope pane --text "Select Model and Effort"
atmux pane watch @3 --scope window --text "running tests" --timeout 20
```

#### `path`

```sh
atmux path watch <glob> [--timeout <seconds>] [--duration <seconds>]
                        [--interval <seconds>] [--coalesce <seconds>]
                        [--exec <cmd>] [--once]
```

Operate on filesystem paths matched by a glob. The only verb today is
`watch`, which streams change events continuously, emitting one digest XML
line per --coalesce window (default 60s). Uses fswatch or inotifywait when
available; falls back to polling.

--timeout    Idle exit: return 124 if no change is observed for N seconds
             (0 = disabled, default 0).
--duration   Hard cap: return 124 after N seconds total (0 = no cap).
--coalesce   Batch change events into one digest line per N seconds
             (default 60). Set 0 to emit per-event lines.
--once       Single-shot: emit one XML line on the first change and exit 0.
             Disables coalescing.
--exec       Pipe each emitted XML line into `bash -c <cmd>`.

```sh
atmux path watch 'src/**/*.sh'
atmux path watch '/tmp/build/*.log' --once --timeout 60
atmux path watch 'docs/**/*.md' --exec ./on-docs-changed --coalesce 30
```

#### `git`

```sh
atmux git <git-args...>
atmux git snapshot [--id <id>] [--exec <cmd>]
atmux git watch [--id <id>] [--coalesce <seconds>] [--interval <seconds>]
                [--timeout <seconds>] [--duration <seconds>]
                [--exec <cmd>] [--once]
```

Transparent wrapper around `git`. Every verb except `watch` is
forwarded to the real git executable — `atmux git status`, `atmux git
log --oneline`, `atmux git worktree add <path>` all behave exactly
like the underlying git command and exit with git's exit code.

`atmux git watch` is the only verb atmux owns: a rolling-diff watcher
that emits XML notifications on each change since the last emit. See
`atmux git watch --help` for details.

Wrapping git lets atmux react to specific operations. The current set
of hooks is hardcoded; a configurable hooks system is planned.

Hooks (current):
worktree add  Records the new worktree's absolute path at
              $ATMUX_HOME/agents/<repo>/<agent>/git-hooks/last-worktree-add
              on success. Pair-program's navigator uses this to
              follow the driver's worktree.

#### `gh`

```sh
atmux gh <gh-args...>
```

Transparent wrapper around GitHub CLI (`gh`). Every invocation is forwarded
to the real `gh` executable with the same stdout, stderr, and exit code.

When run through the agent shim, successful `gh pr create` commands
automatically register the creating agent's pane for `atmux pr watch` on the
created GitHub pull request. The existing PR watcher then delivers comment
notifications until the PR closes or merges.
## Cross-cutting verbs

#### `send`

```sh
atmux send --to <name|session> [--reply-required] [--interrupt] "message"
```

Send XML messages to a single agent or every agent in a team. Without
--interrupt, the message is queued and delivered when the receiving agent
is at its idle prompt.
Resolution order for --to:
  1) Team session/name
  2) Agent session/name
--interrupt  Hard interrupt: send the adapter's abort key sequence
             (`SUBMIT_KEYS_INTERRUPT` in the manifest) to stop the current
             operation, then submit the message. Use sparingly — this aborts
             whatever the agent is doing.

```sh
atmux send --to planner "run tests"
atmux send --to platform --reply-required "status check-in"
atmux send --to worker --interrupt "stop, that's wrong"
```

#### `exec`

```sh
atmux exec [--detach | --shared] [--] <command> [args...]
```

Execute a command with passthrough stdio and unchanged exit behavior.
After the command exits or is interrupted, send an ATMUX notification back
to the current agent pane with the exit code.

--detach   Run the command in a new tmux window inside the current
           agent's session. Returns immediately. Tmux ties the window
           to the session — when the session is killed, the window
           (and the command) die with it. No nohup, no orphans.
--shared   Run in a new tmux window inside the per-repo
           `atmux-<repo>-workers` session (lazy-created if missing).
           Use for long-running workers that aren't owned by any
           single agent (e.g. PR/issue feed watchers fanning events
           out to multiple subscribers). Implies --detach.

```sh
atmux exec sleep 30
atmux exec -- make test
atmux exec --detach -- make test
atmux exec --shared -- atmux pr watch 123
```

#### `schedule`

```sh
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] --notification "<text>"
atmux schedule (--once <duration> | --interval <duration>) [--no-detach] -- <command> [args...]
```

Schedule a future or recurring action.

- **Notification mode** (`--notification`) — queues an ATMUX notification back to the current agent's session. Use this for self-reminders, ticks, and status checks.
- **Command mode** (`-- <command...>`) — runs the command in the current environment. Only schedule `atmux send` when the target is **another** agent or team; never schedule `atmux send --to <self>` (use notification mode instead).
- **`--no-detach`** — run in the foreground (blocking). By default the scheduled task runs in a detached tmux window so the command returns immediately.

Durations accept a unit suffix: `ms`, `s`, `m`, `h`, `d`. No suffix means seconds. Examples: `45s`, `30m`, `2h`.

#### `notify`

```sh
atmux notify --pane <tmux-pane-id> --xml <payload> [--interrupt]
```

Send an ATMUX XML notification to a tmux pane. Without --interrupt, the
notification is queued and delivered when the pane's adapter is at its
idle prompt.
--interrupt  Hard interrupt: send the adapter's abort key sequence
             (`SUBMIT_KEYS_INTERRUPT` in the manifest) before the message,
             then submit. Resolves adapter from the pane's session, or from
             --session if provided. Use sparingly — this aborts whatever
             the agent is doing.

```sh
atmux notify --pane %12 --xml '<notification type="test" from="manual" cmd="atmux message read abc" />'
atmux notify --pane %12 --xml '<notification type="abort" from="mgr" />' --interrupt
```

#### `update`

```sh
atmux update [--check] [--version <version>]
atmux update --auto
atmux update --no-auto
```

**Options**

```
--check              Only report whether an update is available; do not install.
--version <version>  Install a specific version (e.g. 0.2.0). Defaults to latest.
--auto               Enable background auto-update on every atmux command (hourly throttle).
--no-auto            Disable background auto-update.
```

```sh
atmux update
atmux update --check
atmux update --version 0.2.0
atmux update --auto
atmux update --no-auto
```

#### `install`

```sh
atmux install [--project|--system] [--project-root <dir>] [--no-slash-commands]
```

Install atmux for a project by default, or system-wide when requested. Project
installs write to <project>/.atmux and project-local CLI command directories.

**Options**

```
--project            Install into <project>/.atmux and project-local CLI command dirs.
--system             Install into ~/.atmux and user-level CLI command dirs.
--scope <scope>      Same as --project or --system. Values: project, system.
--project-root <dir> Project directory for --project (default: git root or cwd).
--no-slash-commands  Skip installing Claude/Gemini/Codex slash commands.
```

```sh
atmux install
atmux install --project
atmux install --system
atmux install --no-slash-commands
```
## Adapters

| Adapter | CLI |
|---------|-----|
| `claude-code` | Claude Code |
| `codex` | OpenAI Codex |
| `cursor-agent` | Cursor AI |
| `gemini` | Gemini CLI |

## Intelligence scale

| Adapter | Intelligence | Model | Reasoning level |
|---------|--------------|-------|-----------------|
| `claude-code` | 0–39 | `sonnet` | `low` |
| `claude-code` | 40–74 | `sonnet` | `medium` |
| `claude-code` | 75–89 | `sonnet` | `high` |
| `claude-code` | 90–100 | `opus` | `high` |
| `codex` | 0–29 | `gpt-5.5` | `low` |
| `codex` | 30–59 | `gpt-5.5` | `medium` |
| `codex` | 60–84 | `gpt-5.5` | `high` |
| `codex` | 85–100 | `gpt-5.5` | `extra-high` |
| `cursor-agent` | 0–39 | `composer-2-fast` | `low` |
| `cursor-agent` | 40–74 | `composer-2` | `medium` |
| `cursor-agent` | 75–89 | `gpt-5.3-codex-high` | `high` |
| `cursor-agent` | 90–100 | `gpt-5.3-codex-xhigh` | `extra-high` |
| `gemini` | 0–39 | `gemini-3.1-flash-lite-preview` | `low` |
| `gemini` | 40–74 | `gemini-3-flash-preview` | `medium` |
| `gemini` | 75–89 | `gemini-3.1-pro-preview` | `medium` |
| `gemini` | 90–100 | `gemini-3.1-pro-preview` | `high` |

## Environment variables

| Variable | Description |
|----------|-------------|
| `ATMUX_HOME` | Installation/state root (default: `<project>/.atmux` for project installs, `~/.atmux` for system installs) |
| `ATMUX_REPO` | Repository name for the current session |
| `ATMUX_AGENT_NAME` | Current agent's name |
| `ATMUX_MANAGER` | Parent manager agent name |
| `ATMUX_WORKTREE` | Working directory (worktree or repo root) |
| `ATMUX_TEAM` | Team this agent belongs to |
| `ATMUX_SESSION_ID` | Unique session identifier |
| `ATMUX_SESSION_KIND` | `agent` or `team` |
| `ATMUX_TMUX_MOUSE` | Tmux mouse mode for managed sessions, `off` by default to avoid accidental copy-mode blocking notifications; set `on` to restore tmux mouse scrolling |
<!-- ATMUX-HELP-END -->
</atmux>