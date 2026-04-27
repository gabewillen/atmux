---
name: atmux-capture
description: Capture recent output from ATMUX agents or teams using `atmux agent capture` / `atmux team capture`. Use when the user asks for another agent's current output or status evidence.
allowed-tools: Bash(atmux agent capture*, atmux team capture*)
---
# ATMUX Capture

Use `atmux agent capture <name>`, `atmux team capture <name>`, or `atmux agent capture --all` to inspect another ATMUX target.

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
2. Run the corresponding command:
   - `agent <name>` → `atmux agent capture <name> [--lines N]`
   - `team <name>`  → `atmux team capture <name> [--lines N]`
   - `all`          → `atmux agent capture --all [--lines N]`
3. Return the important captured output, not a raw dump unless asked.
