Use the shell tool to run an ATMUX assign command.

Arguments are provided in this form:

`target | title | given | when | then | optional todo 1 | optional todo 2 ...`

Behavior:

1. Split the text on `|`.
2. Trim whitespace around each field.
3. Map fields: `--to` (1st), `--title` (2nd), `--given` (3rd), `--when` (4th), `--then` (5th).
4. Add each remaining field (6th+) as `--todo`.
5. Omit empty structured flags.
6. Run the matching `atmux assign` command.
7. Return the assigned issue id or the error briefly.
