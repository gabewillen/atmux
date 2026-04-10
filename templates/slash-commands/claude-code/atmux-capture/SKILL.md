---
name: atmux-capture
description: Capture recent output from ATMUX agents or teams using `atmux capture`. Use when the user asks for another agent's current output or status evidence.
allowed-tools: Bash(atmux capture*)
---
# ATMUX Capture

Use `atmux capture` to inspect another ATMUX target.

## Arguments

Expected forms:

- `agent <name-or-session>`
- `team <name-or-session>`
- `all`

Optional suffix:

- `| <lines>`

Examples:

- `/atmux-capture agent planner`
- `/atmux-capture team platform | 300`
- `/atmux-capture all | 200`

## Behavior

1. Parse the selector and optional line count.
2. Run the corresponding `atmux capture` command.
3. Return the important captured output, not a raw dump unless asked.
