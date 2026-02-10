# Contract

## Template
- Contract ID:
- Version:
- Status: Draft | Frozen | Superseded
- Owner Agent:
- Date:
- Scope:
- Files:
- Guarantees:
- Non-goals:
- Breaking Changes Since Prior Version:
- Downstream Agents:

## Published
- Contract ID: `COORDINATION-OPS`
- Version: `V1`
- Status: `Draft`
- Owner Agent: `agent-0`
- Date: `2026-02-09`
- Scope: tmux coordination, blocker routing, and merge control.
- Files: `agents/coordination.md`, `agents/contracts.md`, `agents/notes.md`, `agents/tmux/**`
- Guarantees:
  - workers communicate through `agent-0`
  - merges into `main` are coordinator-managed
  - blocker and contract registries are centralized
- Non-goals:
  - implementing feature code in worker-owned source trees
- Breaking Changes Since Prior Version: none
- Downstream Agents: `agent-1`, `agent-2`, `agent-3`, `agent-4`
