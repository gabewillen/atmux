









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
- ALWAYS check `atmux list agents --all --status` before creating new agents.
- ALWAYS reuse idle capable agents before creating new ones.
- ALWAYS spawn agents to decompose your todos if necessary.
- ALWAYS use `--reply-required` when a manager decision is needed.
- NEVER poll agent panes unless absolutely necessary.
- NEVER silently change scope; ask your manager first.
- NEVER report task completion without validation evidence.
- NEVER leave blockers unreported; escalate immediately.

# atmux help
## create
Usage:
  atmux create --agent <name> --role <role> --intelligence <0-100> [--team <team>] [--adapter <adapter>] [--no-worktree] [--task --description <desc> --todo <todo>...] [-- <adapter-args...>]
  atmux create --team <name>
  atmux create --issue --title <title> [--description <description>] [--todo <todo>...] [--repo <repo>]

Description:
  Unified create entrypoint for agents, teams, and issues.
  For agents, --team defaults to ATMUX_TEAM when set (for example after `atmux create --team <name>` in a tmux session).

## list
Usage:
  atmux list teams
  atmux list sessions
  atmux list agents [--all] [--status]
  atmux list issues [--repo <repo>]
  atmux list messages [--unread]

Description:
  Listings are implemented by scripts under bin/(atmux)/(list)/.

## send
Usage:
  atmux send --to <name|session> [--reply-required] [--interrupt] "message"

Description:
  Send XML messages to a single agent or every agent in a team.
  Resolution order for --to:
    1) Team session/name
    2) Agent session/name
  --interrupt  Submit using the adapter's interrupt key (processed after current
               tool) instead of the default queue key (processed when idle).

Examples:
  atmux send --to planner "run tests"
  atmux send --to platform --reply-required "status check-in"
  atmux send --to worker --interrupt "stop and check this"

## message
Usage:
  atmux message read <id> [--repo <repo>]
  atmux message list [--unread]

Description:
  Read or list filesystem-backed messages.
  Messages are stored at: ~/.atmux/messages/<repo>/<id>/

## schedule
Usage:
  atmux schedule (--interval <duration> | --once <duration>) [--no-detach] --notification <text>
  atmux schedule (--interval <duration> | --once <duration>) [--no-detach] -- <command> [args...]

Description:
  Schedule a future or repeating action. Use `--notification` to queue an
  ATMUX notification to the current session, or use `-- <command...>`
  to run any command after the delay.

  Notification mode:
    - Always targets the current agent/session.

  Command mode:
    - Runs the provided command in the current environment.
    - If you want to schedule a message, schedule the command directly:
      `atmux schedule --once 10m -- atmux send --to worker "status check"`

  --no-detach  Run in the foreground (blocking). By default, the scheduled
               task runs in a detached tmux window and the command returns
               immediately.

Durations:
  Supports integer values with optional unit suffix:
    ms  milliseconds
    s   seconds
    m   minutes
    h   hours
    d   days
  If no suffix is provided, seconds are assumed.

Examples:
  atmux schedule --interval 30m --notification "check on long-running jobs"
  atmux schedule --once 45s -- atmux send --to atmux-myrepo-worker "follow up"

## assign
Usage:
  atmux assign --to <agent|session> --title <title> [--description <description>] [--given <context>] [--when <action>] [--then <outcome>] [--todo <todo>]... [--repo <repo>]
  atmux assign --issue <id> --to <agent|session> [--repo <repo>]

Description:
  Assign work using filesystem issues.
  - Without --issue: creates a new issue, then assigns it.
  - With --issue: assigns an existing issue id.

## comment
Usage:
  atmux comment "message" --issue <id> [--repo <repo>]

Description:
  Add a comment to a filesystem issue.
  Notifies watchers, assignee, and assigner.

## capture
Usage:
  atmux capture --agent <name|session> [--lines <n>]
  atmux capture --team <name|session> [--lines <n>]
  atmux capture --all [--lines <n>]

Description:
  Capture tmux pane output for agents or team sessions.

Examples:
  atmux capture --agent planner
  atmux capture --agent atmux-myrepo-planner --lines 300
  atmux capture --team platform
  atmux capture --team atmux-myrepo-team-platform --lines 500
  atmux capture --all --lines 200

## kill
Usage:
  atmux kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]
  atmux kill --agent <name|pattern> [name|pattern...]
  atmux kill --all [--yes]

Description:
  --pid    Stop an atmux exec-tracked child process for this repo, wait for
           executor notifications (including watcher fan-out) to finish, then
           remove metadata under ~/.atmux/exec/<repo>/<pid>/.
  --agent  Kill agent sessions and clean up their worktrees and branches.
           Accepts agent names, session names, or glob patterns.
  --all    Kill every atmux session, worktree, and branch for this repo.
           Prompts for y/N confirmation; use --yes to skip. Refuses to run
           from inside an atmux session.

Examples:
  atmux kill --pid 12345
  atmux kill --pid 12345 --timeout 30 --signal TERM
  atmux kill --agent worker
  atmux kill --agent 'agent-*'
  atmux kill --agent worker planner
  atmux kill --all
  atmux kill --all --yes

## exec
Usage:
  atmux exec [--detach] [--] <command> [args...]

Description:
  Execute a command with passthrough stdio and unchanged exit behavior.
  After the command exits or is interrupted, send an ATMUX notification back
  to the current agent pane with the exit code.

  --detach  Run the command in a new tmux window. Returns immediately.
            The process pane is stored so watchers can capture its output.
            Notification is sent to the agent pane when the process exits.

Examples:
  atmux exec sleep 30
  atmux exec -- make test
  atmux exec --detach -- make test

## watch
Usage:
  atmux watch --target <tmux-target> --text <needle> [--scope pane|window|session] [--timeout <seconds>] [--interval <seconds>] [--lines <n>]
  atmux watch --pid <pid> [--timeout <seconds>] [--interval <seconds>]
  atmux watch --pid <pid> --stdio [--duration <seconds>] [--timeout <seconds>] [--interval <seconds>] [--lines <n>]
  atmux watch --issue <id> [--repo <repo>] [--timeout <seconds>] [--interval <seconds>]
  atmux watch --agent <name|session> [--idle <seconds>] [--timeout <seconds>] [--interval <seconds>] [--lines <n>]

Description:
  Pane mode: poll tmux output until text appears, non-zero on timeout.
  PID mode: wait for an atmux exec process (see ~/.atmux/exec/<repo>/<pid>/) and
  receive the same exit notification XML as the executor when it finishes.
  Stdio mode: monitor a detached exec process pane for output changes. Sends
  a notification each time new output is detected. Exits when --duration
  expires, --timeout (no new output) expires, or the process exits.
  Issue mode: wait for the next filesystem issue update and receive the same
  notification XML as issue assign/claim fan-out.
  Agent mode: wait until an agent's pane output has been stable for --idle
  seconds (default 30). Exits 0 when idle, 124 on timeout.

  Implementations: bin/(atmux)/[watch]/text, [watch]/pid, [watch]/stdio, [watch]/issue, [watch]/agent.

## env
Usage:
  atmux env
  atmux env get <key>

Description:
  Inspect ATMUX_* environment variables in the current process.

Examples:
  atmux env
  atmux env get repo
  atmux env get ATMUX_WORKTREE
</atmux>

