---
name: atmux-assign
description: Create and assign an ATMUX filesystem issue with `atmux assign`. Use when the user wants work handed to another ATMUX agent.
allowed-tools: Bash(atmux assign*)
---
# ATMUX Assign

Use `atmux assign` to create and assign work to another ATMUX agent.

## Arguments

`$ARGUMENTS` should contain fields separated by `|`:

`target | title | given | when | then | optional todo 1 | optional todo 2 ...`

Example:

`/atmux-assign planner | stabilize parser | a token stream containing nulls | the parser encounters a null token | it returns an error instead of panicking | write failing test | fix null handling`

## Behavior

1. Split the input on `|`.
2. Trim whitespace around each field.
3. Map fields: `--to` (1st), `--title` (2nd), `--given` (3rd), `--when` (4th), `--then` (5th).
4. Add each remaining field (6th+) as `--todo`.
5. Omit empty structured flags.
6. Run the matching `atmux assign` command.
7. Report the assigned issue id or the error.
