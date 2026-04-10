---
name: amux-capture
description: Capture recent output from AMUX agents or teams using `amux capture`. Use when the user asks for another agent's current output or status evidence.
allowed-tools: Bash(amux capture*)
---
# AMUX Capture

Use `amux capture` to inspect another AMUX target.

## Arguments

Expected forms:

- `agent <name-or-session>`
- `team <name-or-session>`
- `all`

Optional suffix:

- `| <lines>`

Examples:

- `/amux-capture agent planner`
- `/amux-capture team platform | 300`
- `/amux-capture all | 200`

## Behavior

1. Parse the selector and optional line count.
2. Run the corresponding `amux capture` command.
3. Return the important captured output, not a raw dump unless asked.
