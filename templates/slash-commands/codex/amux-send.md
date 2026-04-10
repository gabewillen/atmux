Use the shell tool to run an AMUX send command.

Arguments are provided in this form:

`target message...`

Behavior:

1. Parse the first token as the AMUX target.
2. Treat the remaining text as the message body.
3. Run `amux send --to "<target>" "<message>"`.
4. Return the delivery result briefly.
