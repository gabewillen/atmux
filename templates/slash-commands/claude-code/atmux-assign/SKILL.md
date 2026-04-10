---
name: atmux-assign
description: Create and assign an ATMUX filesystem issue with `atmux assign`. Use when the user wants work handed to another ATMUX agent.
allowed-tools: Bash(atmux assign*)
---
# ATMUX Assign

Use `atmux assign` to create and assign work to another ATMUX agent.

## Arguments

`$ARGUMENTS` should contain:

`target | title | optional todo 1 | optional todo 2 ...`

Example:

`/atmux-assign planner | stabilize parser | write failing test first | fix null handling`

## Behavior

1. Split the input on `|`.
2. Trim whitespace around each field.
3. Use the first field as `--to` and the second as `--title`.
4. Add each remaining field as `--todo`.
5. Run the matching `atmux assign` command.
6. Report the assigned issue id or the error.
