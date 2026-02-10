# Agent 4 Plan: Application Target and End-to-End Integration

## Goal
Deliver a functioning application target that links the library and proves end-to-end behavior:
runtime boot -> session open -> generation -> shutdown.

## Required Context
- `AGENTS.md`
- `agents/coordination.md`
- `agents/agent-4/AGENT.md`
- `agents/agent-1/contract.md`
- `agents/agent-2/contract.md`
- `agents/agent-3/contract.md` (if checkpoint/load is in app flow)

## tmux Session
- Session: `agent-4`
- Coordinator session: `agent-0`

## Worktree
- Branch: `agent-4/app-integration`
- Worktree: `../machine-agent-4`
- Setup:
```bash
git fetch origin
git worktree add ../machine-agent-4 -b agent-4/app-integration origin/main
```

## Owned Paths (Strict)
- `CMakeLists.txt`
- `src/app/**`
- `tests/engine/**` (integration and app smoke tests)
- `docs/AUDIT.md`
- `docs/CHANGELOG.md`
- `docs/INDEX.md`
- `agents/agent-4/**`

## Forbidden Paths
- `src/inference/**`
- `src/backend/**`
- `src/io/**`
- `src/state/**`
- `src/error/**`
- `src/runtime/**`
- `include/engine/*.h`
- `agents/agent-0/**`
- `agents/agent-1/**`
- `agents/agent-2/**`
- `agents/agent-3/**`
- `agents/contracts.md`
- `agents/notes.md`

## Deliverables
1. New application executable target wired in `CMakeLists.txt`.
2. Minimal app entrypoint under `src/app/**` using public C API only.
3. End-to-end smoke tests in `tests/engine/**`.
4. `APP-INTEGRATION-V1` published in `agents/agent-4/contract.md` and set to `Frozen`.
5. Integration caveats/blockers logged in `agents/agent-4/notes.md`.

## Execution Milestones
1. Wait for `ABI-LIFECYCLE-V1` to be `Frozen`.
2. Wait for `INFERENCE-GEN-V1` to be `Frozen`.
3. If checkpoint/load is in app flow, wait for `STATE-IO-RECOVERY-V1` to be `Frozen`.
4. Implement app target and integration tests.
5. Rebase on merged upstream branches.
6. Freeze `APP-INTEGRATION-V1`.
7. Send `ready-to-merge` to `agent-0` via tmux.

## Blocking Points
- No final app wiring without frozen upstream contracts.
- If upstream contract changes after freeze, require version bump and revalidation.
- Any required edits to unowned files must be escalated to `agent-0`.

## Sync Cadence
- Rebase at start, before push, and at least every 90 minutes.

## Validation
```bash
./scripts/build_with_zig.sh build-agent-4-zig
./scripts/test_with_coverage.sh build-agent-4-cov
./scripts/lint_snapshot.sh
ctest --test-dir build-agent-4-cov --output-on-failure
```

## Merge Checklist
1. `git fetch origin && git rebase origin/main`
2. Confirm required upstream contracts are `Frozen`.
3. `git diff --name-only origin/main...HEAD` is limited to owned paths.
4. Validation commands pass.
5. `agent-0` acknowledges `ready-to-merge` and performs merge.
