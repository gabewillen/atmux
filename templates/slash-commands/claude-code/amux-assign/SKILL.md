---
name: amux-assign
description: Create and assign an AMUX filesystem issue with `amux assign`. Use when the user wants work handed to another AMUX agent.
allowed-tools: Bash(amux assign*)
---
# AMUX Assign

Use `amux assign` to create and assign work to another AMUX agent.

## Arguments

`$ARGUMENTS` should contain:

`target | title | optional todo 1 | optional todo 2 ...`

Example:

`/amux-assign planner | stabilize parser | write failing test first | fix null handling`

## Behavior

1. Split the input on `|`.
2. Trim whitespace around each field.
3. Use the first field as `--to` and the second as `--title`.
4. Add each remaining field as `--todo`.
5. Run the matching `amux assign` command.
6. Report the assigned issue id or the error.
