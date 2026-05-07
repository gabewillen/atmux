# Collab Team

This team runs a structured deliberation session.

Members:
- `${ATMUX_TEAM}-facilitator` facilitates topics, resolves conclusions (including dissent), and writes `${ATMUX_TEAM_DOC_DIR}/final.md`.
- `${ATMUX_TEAM}-codex`, `${ATMUX_TEAM}-claude`, and `${ATMUX_TEAM}-gemini` (by default) deliberate substantively; collaborators are configurable via the team manifest.

The live discussion uses `atmux send --to ${ATMUX_TEAM} ...`. Durable output belongs in `${ATMUX_TEAM_DOC_DIR}`, especially `${ATMUX_TEAM_DOC_DIR}/final.md`.
