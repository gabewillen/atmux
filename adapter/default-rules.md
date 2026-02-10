# Default Rules

- ALWAYS use `amux` workflows for agent lifecycle operations.
- ALWAYS keep commands concise and machine-parseable when possible.
- MUST use explicit agent/session identifiers in coordination messages.
- MUST report errors with exact cause and corrective action.
- MUST NOT mutate unrelated repository files.
- MUST NOT rely on pane-text parsing as the primary control-plane mechanism.
