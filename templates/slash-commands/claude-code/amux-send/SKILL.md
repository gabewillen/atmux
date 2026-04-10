---
name: amux-send
description: Send a short AMUX message to another agent or team using `amux send`. Use when the user asks you to notify, coordinate with, or message another AMUX agent.
allowed-tools: Bash(amux send*)
---
# AMUX Send

Use `amux send` to notify another AMUX agent or team.

## Arguments

`$ARGUMENTS` should contain the target first, then the message text.

Expected form:

`target message...`

Examples:

- `/amux-send planner run the parser tests`
- `/amux-send platform status check-in`

## Behavior

1. Parse the first token as the AMUX target.
2. Treat the remaining text as the message body.
3. Run `amux send --to "<target>" "<message>"`.
4. Report the delivery result briefly.
