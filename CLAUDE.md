<atmux>
# Agent TMUX env
ATMUX_REPO=test-repo
ATMUX_MANAGER=agent-0

# Managed Agent Rules
- ALWAYS acknowledge manager messages quickly with a short plan.
- ALWAYS send a message to your manager when stuck or after completing any task.
- ALWAYS message your manager with `atmux send --to "$ATMUX_MANAGER" "..."`.
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

## list
Usage:
  atmux list teams
  atmux list sessions
  atmux list agents [--all] [--status]
  atmux list issues [--repo <repo>]

## send
Usage:
  atmux send --to <name|session> [--reply-required] [--interrupt] "message"

Description:
  Send XML messages to a single agent or every agent in a team.
  Resolution order for --to:
    1) Team session/name
    2) Agent session/name

Examples:
  atmux send --to planner "run tests"
  atmux send --to platform --reply-required "status check-in"

## message
Usage:
  atmux message read <id> [--repo <repo>]
  atmux message list [--repo <repo>]

Description:
  Read or list filesystem-backed messages.
  Messages are stored at: ~/.atmux/messages/<repo>/<id>/

## schedule
Usage:
  atmux schedule [--to <name|session>] [--reply-required] (--interval <duration> | --once <duration>) [--no-detach] "message"

Description:
  Schedule a future or repeating send. Runs detached by default.
  If --to is omitted, the current agent session is targeted.

Examples:
  atmux schedule --to planner --once 10m "status check"
  atmux schedule --to platform --interval 30m --reply-required "heartbeat"

## assign
Usage:
  atmux assign --to <agent|session> --title <title> [--description <description>] [--given <context>] [--when <action>] [--then <outcome>] [--todo <todo>]... [--repo <repo>]
  atmux assign --issue <id> --to <agent|session> [--repo <repo>]

Description:
  Assign work using filesystem issues.

Examples:
  atmux assign --to planner --title "stabilize parser" --todo "write failing test" --todo "fix null check"
  atmux assign --issue 123e4567 --to atmux-myrepo-planner

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
  atmux capture --team platform --lines 300
  atmux capture --all --lines 200

## kill
Usage:
  atmux kill --pid <pid> [--timeout <seconds>] [--signal <NAME>]
  atmux kill --agent <name|pattern> [name|pattern...]

Description:
  --pid    Stop an atmux exec-tracked child process for this repo.
  --agent  Kill agent sessions and clean up their worktrees and branches.

Examples:
  atmux kill --pid 12345
  atmux kill --agent worker
  atmux kill --agent 'agent-*'

## exec
Usage:
  atmux exec [--detach] [--] <command> [args...]

Description:
  Execute a command with notification on exit.
  --detach  Run in a new tmux window. Returns immediately.

Examples:
  atmux exec -- make test
  atmux exec --detach -- make test

## watch
Usage:
  atmux watch --target <tmux-target> --text <needle> [--scope pane|window|session] [--timeout <seconds>]
  atmux watch --pid <pid> [--timeout <seconds>]
  atmux watch --pid <pid> --stdio [--duration <seconds>]
  atmux watch --issue <id> [--repo <repo>] [--timeout <seconds>]
  atmux watch --agent <name|session> [--idle <seconds>] [--timeout <seconds>]

Description:
  Poll/wait for text, process exit, issue updates, or agent idle state.

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
