---
name: atmux-send
description: Send a short ATMUX message to another agent or team using `atmux send`. Use when the user asks you to notify, coordinate with, or message another ATMUX agent.
allowed-tools: Bash(atmux send*)
---
# ATMUX Send

Use `atmux send` to notify another ATMUX agent or team.

## Arguments

`$ARGUMENTS` should contain the target first, then the message text.

Expected form:

`target message...`

Examples:

- `/atmux-send planner run the parser tests`
- `/atmux-send platform status check-in`

## Behavior

1. Parse the first token as the ATMUX target.
2. Treat the remaining text as the message body.
3. Run `atmux send --to "<target>" "<message>"`.
4. Report the delivery result briefly.
