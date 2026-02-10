# Agent 3 Plan: System Reliability Machines (Backend/IO/State/Error)

## Goal
Make non-inference orchestration production-safe: explicit error paths, deterministic state/IO
behavior, and reliable recovery semantics.

## Required Context
- `AGENTS.md`
- `agents/coordination.md`
- `agents/agent-3/AGENT.md`
- `agents/agent-1/contract.md`

## tmux Session
- Session: `agent-3`
- Coordinator session: `agent-0`

## Worktree
- Branch: `agent-3/system-reliability-machines`
- Worktree: `../machine-agent-3`
- Setup:
```bash
git fetch origin
git worktree add ../machine-agent-3 -b agent-3/system-reliability-machines origin/main
```

## Owned Paths (Strict)
- `src/backend/**`
- `src/io/**`
- `src/state/**`
- `src/error/**`
- `src/telemetry/**`
- `tests/machines/backend_*`
- `tests/machines/io_*`
- `tests/machines/state_*`
- `tests/machines/error_*`
- `tests/machines/telemetry_*`
- `docs/puml/engine::backend*.puml`
- `docs/puml/engine::io*.puml`
- `docs/puml/engine::state*.puml`
- `docs/puml/engine::error*.puml`
- `agents/agent-3/**`

## Forbidden Paths
- `include/engine/*.h`
- `src/engine/**`
- `src/runtime/**`
- `src/inference/**`
- `CMakeLists.txt`
- `docs/AUDIT.md`
- `docs/CHANGELOG.md`
- `docs/INDEX.md`
- `agents/agent-0/**`
- `agents/agent-1/**`
- `agents/agent-2/**`
- `agents/agent-4/**`

## Deliverables
1. Deterministic backend/io/state/error behavior with explicit unwind/recovery transitions.
2. Focused tests for failure and recovery semantics.
3. `STATE-IO-RECOVERY-V1` published in `agents/agent-3/contract.md` and set to `Frozen`.
4. Cross-cutting caveats/blockers logged in `agents/agent-3/notes.md`.

## Execution Milestones
1. Wait for `ABI-LIFECYCLE-V1` to be `Frozen`.
2. Implement reliability behavior and tests.
3. Publish `STATE-IO-RECOVERY-V1` as `Draft`.
4. Rebase and finalize.
5. Switch `STATE-IO-RECOVERY-V1` to `Frozen`.
6. Send `ready-to-merge` to `agent-0` via tmux.

## Blocking Points
- If status/handle assumptions are unclear, send blocker to `agent-0` and pause.
- If unowned files are required, send blocker to `agent-0` and pause.
- Agent 4 waits for `STATE-IO-RECOVERY-V1` if app uses checkpoint/load.
- Breaking changes after freeze require `STATE-IO-RECOVERY-V2`.

## Sync Cadence
- Rebase at start, before push, and at least every 90 minutes.

## Validation
```bash
./scripts/build_with_zig.sh build-agent-3-zig
./scripts/test_with_coverage.sh build-agent-3-cov
./scripts/lint_snapshot.sh
```

## Merge Checklist
1. `git fetch origin && git rebase origin/main`
2. `git diff --name-only origin/main...HEAD` is limited to owned paths.
3. Validation commands pass.
4. Contract shows `STATE-IO-RECOVERY-V1` as `Frozen`.
5. `agent-0` acknowledges `ready-to-merge` and performs merge.
