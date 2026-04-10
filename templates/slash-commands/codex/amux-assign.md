Use the shell tool to run an AMUX assign command.

Arguments are provided in this form:

`target | title | optional todo 1 | optional todo 2 ...`

Behavior:

1. Split the text on `|`.
2. Trim whitespace around each field.
3. Use the first field as `--to` and the second as `--title`.
4. Add each remaining field as `--todo`.
5. Run the matching `amux assign` command.
6. Return the assigned issue id or the error briefly.
