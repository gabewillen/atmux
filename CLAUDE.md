<atmux>
# Agent TMUX env
ATMUX_REPO=test-repo
ATMUX_MANAGER=agent-0

# Managed Agent Rules
- ALWAYS acknowledge manager messages quickly with a short plan.
- ALWAYS send a message to your manager when stuck or after completing any task.
- ALWAYS message your manager with `atmux send --to "$ATMUX_MANAGER" "..."`.
- ALWAYS coordinate with peer agents using `atmux send --to <agent> "..."`.
- ALWAYS use team broadcast when needed: `atmux send --to <team> "..."`.
- ALWAYS check `atmux list teams` before creating new team members.
- ALWAYS reuse idle capable agents before creating new ones.
- ALWAYS spawn agents to decompose your todos if necessary.
- ALWAYS use `--reply-required` when a manager decision is needed.
- NEVER silently change scope; ask your manager first.
- NEVER report task completion without validation evidence.
- NEVER leave blockers unreported; escalate immediately.

# atmux help
## agent
Usage:
  atmux create --agent <name>
    [--name <name>]
    --role <role>
    [--team|-t <team>]
    [--adapter <adapter>]
    # 0-100, choose based on task complexity
    --intelligence <0-100>
    [-- <adapter-args...>]

Description:
  Create agents.

  Notes:
  create requires ATMUX_AGENT_NAME (used as ATMUX_MANAGER for child agents).
  create requires --role.
  create requires --intelligence.
  create optionally accepts --team/-t to group agents.
  team members are capped at 4.
  Why: keeps team layouts readable (2x2), reduces context/capture bloat, and
  avoids oversized detached tmux grids that degrade reliability.
  create defaults --adapter to auto.
  Target session format: atmux-<repo>-<agent>
Examples:
  atmux create --agent reviewer --role reviewer --intelligence 80
  atmux create --agent reviewer --role reviewer --team platform --intelligence 80
  atmux create --agent worker-1 --role implementer --intelligence 55 --adapter claude-code -- --dangerously-skip-permissions

## send
Usage:
  atmux send --to <name|session> [--reply-required] "message"

Description:
  Send a message to a single agent or every agent in a team.
  Message body is stored on disk; agent receives a one-line notification:
    <notification type="message" from="..." cmd="atmux message read <id>" />
  Agent runs the cmd to read the full message.
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

Examples:
  atmux message read 123e4567-e89b-12d3-a456-426614174000

## assign
Usage:
  atmux assign --to <agent|session> --title <title> [--given <context>] [--when <action>] [--then <outcome>] [--todo <todo>]... [--description <description>]
  atmux assign --issue <id> --to <agent|session>

Description:
  Create-and-assign (or assign existing) filesystem issues.
  Structured flags (--given, --when, --then, --todo) render via templates/issue.md as BDD scenarios.
  Use --description for freeform text (bypasses template).

Examples:
  atmux assign --to planner --title "stabilize parser" --given "a token stream containing nulls" --when "the parser encounters a null token" --then "it returns an error instead of panicking" --todo "write failing test" --todo "fix null check in parse()"
  atmux assign --issue 123e4567-e89b-12d3-a456-426614174000 --to atmux-myrepo-planner

## comment
Usage:
  atmux comment "message" --issue <id>

Description:
  Add a comment to a filesystem issue.
  Notifies watchers, assignee, and assigner.

Examples:
  atmux comment "blocking on upstream API" --issue 123e4567-e89b-12d3-a456-426614174000

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
