# Agent 2 Plan: Inference Path to Functional Generation

## Goal
Deliver deterministic inference/session generation behavior that downstream application code can use
without relying on hidden control flow.

## Required Context
- `AGENTS.md`
- `agents/coordination.md`
- `agents/agent-2/AGENT.md`
- `agents/agent-1/contract.md`

## tmux Session
- Session: `agent-2`
- Coordinator session: `agent-0`

## Worktree
- Branch: `agent-2/inference-session-generation`
- Worktree: `../machine-agent-2`
- Setup:
```bash
git fetch origin
git worktree add ../machine-agent-2 -b agent-2/inference-session-generation origin/main
```

## Owned Paths (Strict)
- `src/inference/**`
- `tests/machines/inference_*`
- `docs/puml/engine::inference*.puml` (only affected diagrams)
- `agents/agent-2/**`

## Forbidden Paths
- `include/engine/*.h`
- `src/engine/**`
- `src/runtime/**`
- `CMakeLists.txt`
- `docs/AUDIT.md`
- `docs/CHANGELOG.md`
- `docs/INDEX.md`
- `agents/agent-0/**`
- `agents/agent-1/**`
- `agents/agent-3/**`
- `agents/agent-4/**`

## Deliverables
1. Deterministic inference/session generation sequencing.
2. Updated machine/behavior tests for inference paths.
3. `INFERENCE-GEN-V1` published in `agents/agent-2/contract.md` and set to `Frozen`.
4. Assumptions/blockers logged in `agents/agent-2/notes.md`.

## Execution Milestones
1. Wait for `ABI-LIFECYCLE-V1` to be `Frozen`.
2. Implement inference behavior and tests.
3. Publish `INFERENCE-GEN-V1` as `Draft`.
4. Rebase and finalize.
5. Switch `INFERENCE-GEN-V1` to `Frozen`.
6. Send `ready-to-merge` to `agent-0` via tmux.

## Blocking Points
- If Agent 1 contract is not frozen, do not finalize.
- If API/header changes are required, send blocker to `agent-0` and pause.
- Agent 4 waits for `INFERENCE-GEN-V1` status `Frozen`.
- Breaking changes after freeze require `INFERENCE-GEN-V2`.

## Sync Cadence
- Rebase at start, before push, and at least every 90 minutes.

## Validation
```bash
./scripts/build_with_zig.sh build-agent-2-zig
./scripts/test_with_coverage.sh build-agent-2-cov
./scripts/lint_snapshot.sh
```

## Merge Checklist
1. `git fetch origin && git rebase origin/main`
2. `git diff --name-only origin/main...HEAD` is limited to owned paths.
3. Validation commands pass.
4. Contract shows `INFERENCE-GEN-V1` as `Frozen`.
5. `agent-0` acknowledges `ready-to-merge` and performs merge.
