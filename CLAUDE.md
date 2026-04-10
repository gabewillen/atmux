<amux>
# Amux Env
AMUX_REPO=test-repo
AMUX_MANAGER=agent-0

# Managed Agent Rules
- ALWAYS acknowledge manager messages quickly with a short plan.
- ALWAYS send a message to your manager when stuck or after completing any task.
- ALWAYS message your manager with `amux send --to "$AMUX_MANAGER" "..."`.
- ALWAYS coordinate with peer agents using `amux send --to <agent> "..."`.
- ALWAYS use team broadcast when needed: `amux send --to <team> "..."`.
- ALWAYS check `amux list teams` before creating new team members.
- ALWAYS reuse idle capable agents before creating new ones.
- ALWAYS spawn agents to decompose your todos if necessary.
- ALWAYS use `--reply-required` when a manager decision is needed.
- NEVER silently change scope; ask your manager first.
- NEVER report task completion without validation evidence.
- NEVER leave blockers unreported; escalate immediately.

# Amux Help
## agent
Usage:
  amux create --agent <name>
    [--name <name>]
    --role <role>
    [--team|-t <team>]
    [--adapter <adapter>]
    # 0-100, choose based on task complexity
    --intelligence <0-100>
    [-- <adapter-args...>]
  amux agent destroy <session-pattern> [session-pattern...]

Description:
  Create agents or remove agent sessions/worktrees/branches.

  Notes:
  create requires AMUX_AGENT_NAME (used as AMUX_MANAGER for child agents).
  create requires --role.
  create requires --intelligence.
  create optionally accepts --team/-t to group agents.
  team members are capped at 4.
  Why: keeps team layouts readable (2x2), reduces context/capture bloat, and
  avoids oversized detached tmux grids that degrade reliability.
  create defaults --adapter to auto.
  Target session format: amux-<repo>-<agent>
Examples:
  amux create --agent reviewer --role reviewer --intelligence 80
  amux create --agent reviewer --role reviewer --team platform --intelligence 80
  amux create --agent worker-1 --role implementer --intelligence 55 --adapter claude-code -- --dangerously-skip-permissions
  amux agent destroy 'amux-myrepo-agent-*'

## send
Usage:
  amux send --to <name|session> [--reply-required] "message"

Description:
  Send XML messages to a single agent or every agent in a team.
  Resolution order for --to:
    1) Team session/name
    2) Agent session/name

Examples:
  amux send --to planner "run tests"
  amux send --to platform --reply-required "status check-in"

## assign
Usage:
  amux assign --to <agent|session> --title <title> [--description <description>] [--todo <todo>]...
  amux assign --issue <id> --to <agent|session>

Description:
  Create-and-assign (or assign existing) filesystem issues.

Examples:
  amux assign --to planner --title "stabilize parser" --todo "write failing test first"
  amux assign --issue 123e4567-e89b-12d3-a456-426614174000 --to amux-myrepo-planner

## capture
Usage:
  amux capture --agent <name|session> [--lines <n>]
  amux capture --team <name|session> [--lines <n>]
  amux capture --all [--lines <n>]

Description:
  Capture tmux pane output for agents or team sessions.

Examples:
  amux capture --agent planner
  amux capture --team platform --lines 300
  amux capture --all --lines 200

## env
Usage:
  amux env
  amux env get <key>

Description:
  Inspect AMUX_* environment variables in the current process.

Examples:
  amux env
  amux env get repo
  amux env get AMUX_WORKTREE
</amux>
