# Collab Team

Create a multi-agent deliberation team:

```sh
atmux team create planning --role collab --time-limit 45m
atmux send --to planning "Goal: produce a roadmap for ..."
```

The team uses state-backed membership and does not create a dashboard by default. Use `atmux team view planning` when you want a tmux overview.

Configure the facilitator with `--set facilitator.<field>=…` (matches the facilitator member’s `--adapter`, `--model`, `--intelligence`, `--reasoning`) or with undotted role keys such as `--set facilitator_adapter=codex`; the team manifest exposes only **`ATMUX_TEAM_SET_FACILITATOR_ADAPTER`**, **`…INTELLIGENCE`**, **`…MODEL`**, **`…REASONING`**. Overrides for retired `leader` / `arbiter` / `recorder` members are ignored.

Collaborators are configurable at creation time:

```sh
atmux team create planning --role collab \
  --set collaborators='${ATMUX_TEAM}-codex --role collaborator --intelligence 85 --adapter codex --shared-worktree' \
  --set collaborators='${ATMUX_TEAM}-claude --role collaborator --intelligence 90 --adapter claude-code --shared-worktree' \
  --set collaborators='${ATMUX_TEAM}-gemini --role collaborator --intelligence 90 --adapter gemini --shared-worktree' \
  --set facilitator.model=composer-2-fast
```

The facilitator writes the final artifact to `docs/atmux/<team>/final.md` (via `${ATMUX_TEAM_DOC_DIR}/final.md`).
