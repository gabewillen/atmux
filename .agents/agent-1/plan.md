# Agent 1 Plan: Library Contract Baseline (C ABI + Runtime)

## Goal
Create the baseline contract for a functioning library: stable C ABI behavior, deterministic
runtime/session lifecycle, and complete API lifecycle tests.

## Required Context
- `AGENTS.md`
- `agents/coordination.md`
- `agents/agent-1/AGENT.md`

## tmux Session
- Session: `agent-1`
- Coordinator session: `agent-0`

## Worktree
- Branch: `agent-1/core-api-runtime`
- Worktree: `../machine-agent-1`
- Setup:
```bash
git fetch origin
git worktree add ../machine-agent-1 -b agent-1/core-api-runtime origin/main
```

## Owned Paths (Strict)
- `include/engine/*.h`
- `src/engine/api.cpp`
- `src/runtime/**`
- `tests/engine/**`
- `tests/support/**` (only when needed by API/runtime tests)
- `agents/agent-1/**`

## Forbidden Paths
- `CMakeLists.txt`
- `src/inference/**`
- `src/backend/**`
- `src/io/**`
- `src/state/**`
- `src/error/**`
- `agents/agent-0/**`
- `agents/agent-2/**`
- `agents/agent-3/**`
- `agents/agent-4/**`

## Deliverables
1. Deterministic runtime/session lifecycle behavior through C ABI.
2. Doctest coverage for valid and invalid call ordering.
3. `ABI-LIFECYCLE-V1` published in `agents/agent-1/contract.md` and set to `Frozen`.
4. Caveats/blockers recorded in `agents/agent-1/notes.md`.

## Execution Milestones
1. Implement API/runtime fixes and tests.
2. Publish `ABI-LIFECYCLE-V1` as `Draft`.
3. Rebase and finalize implementation.
4. Switch `ABI-LIFECYCLE-V1` to `Frozen`.
5. Send `ready-to-merge` to `agent-0` via tmux.

## Blocking Points
- If changes are required outside owned paths, send blocker to `agent-0` and pause.
- Agents 2 and 4 wait for `ABI-LIFECYCLE-V1` status `Frozen`.
- Breaking changes after freeze require `ABI-LIFECYCLE-V2`.

## Sync Cadence
- Rebase at start, before push, and at least every 90 minutes.

## Validation
```bash
./scripts/build_with_zig.sh build-agent-1-zig
./scripts/test_with_coverage.sh build-agent-1-cov
./scripts/lint_snapshot.sh
```

## Merge Checklist
1. `git fetch origin && git rebase origin/main`
2. `git diff --name-only origin/main...HEAD` is limited to owned paths.
3. Validation commands pass.
4. Contract shows `ABI-LIFECYCLE-V1` as `Frozen`.
5. `agent-0` acknowledges `ready-to-merge` and performs merge.
