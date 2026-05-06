# Collab Team

Create a multi-agent deliberation team:

```sh
atmux team create planning --role collab --time-limit 45m
atmux send --to planning "Goal: produce a roadmap for ..."
```

The team uses state-backed membership and does not create a dashboard by default. Use `atmux team view planning` when you want a tmux overview.

Collaborators are configurable at creation time:

```sh
atmux team create planning --role collab \
  --set collaborators='${ATMUX_TEAM}-codex --role collaborator --intelligence 85 --adapter codex --shared-worktree' \
  --set collaborators='${ATMUX_TEAM}-claude --role collaborator --intelligence 90 --adapter claude-code --shared-worktree' \
  --set collaborators='${ATMUX_TEAM}-gemini --role collaborator --intelligence 90 --adapter gemini --shared-worktree' \
  --set recorder.model=composer-2-fast
```

The recorder writes the final artifact to `docs/atmux/<team>/final.md`.
