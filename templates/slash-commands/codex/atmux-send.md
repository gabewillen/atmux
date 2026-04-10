Use the shell tool to run an ATMUX send command.

Arguments are provided in this form:

`target message...`

Behavior:

1. Parse the first token as the ATMUX target.
2. Treat the remaining text as the message body.
3. Run `atmux send --to "<target>" "<message>"`.
4. Return the delivery result briefly.
