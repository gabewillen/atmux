Use the shell tool to run an ATMUX capture command.

Expected inputs:

- `agent <name-or-session>`
- `team <name-or-session>`
- `all`

Optional suffix:

- `| <lines>`

Behavior:

1. Parse the selector and optional line count.
2. Run the corresponding `atmux capture` command.
3. Return the important captured output, not a raw dump unless the user asks for it.
