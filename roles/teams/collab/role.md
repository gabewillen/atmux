# Collab Team

This team runs a structured deliberation session.

Members:
- `${ATMUX_TEAM}-leader` facilitates topics, phases, time budgets, and session completion.
- `${ATMUX_TEAM}-arbitrator` resolves each topic into a clear state and conclusion.
- `${ATMUX_TEAM}-recorder` records the discussion and writes the durable artifact.
- `${ATMUX_TEAM}-codex`, `${ATMUX_TEAM}-claude`, and `${ATMUX_TEAM}-gemini` deliberate substantively.

The live discussion uses `atmux send --to ${ATMUX_TEAM} ...`. Durable output belongs in `${ATMUX_TEAM_DOC_DIR}`, especially `${ATMUX_TEAM_DOC_DIR}/final.md`.
