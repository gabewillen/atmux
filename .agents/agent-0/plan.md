# Agent 0 Plan: Coordinator, Blocker Resolver, and Merge Integrator

## Goal
Operate as tmux coordinator for all workers: answer questions, solve blockers, merge validated
changes into `main`, and keep all agents aligned.

## Required Context
- `AGENTS.md`
- `agents/coordination.md`
- `agents/agent-0/AGENT.md`

## tmux Session
- Session: `agent-0`
- Worktree: repository root

## Owned Paths (Strict)
- `agents/coordination.md`
- `agents/contracts.md`
- `agents/notes.md`
- `agents/tmux/**`
- `agents/agent-0/**`

## Forbidden Paths
- `src/**`
- `include/**`
- `tests/**`
- `docs/**`
- `agents/agent-1/**`
- `agents/agent-2/**`
- `agents/agent-3/**`
- `agents/agent-4/**`

## Deliverables
1. tmux communication and escalation protocol is enforced.
2. Blockers are triaged and routed with explicit decisions.
3. Merge train is executed in order: agent-1 -> agent-2/3 -> agent-4.
4. `COORDINATION-OPS-V1` published and `Frozen` in `agents/agent-0/contract.md`.

## Execution Milestones
1. Start/verify tmux sessions with `./agents/tmux/start_sessions.sh`.
2. Publish `COORDINATION-OPS-V1` as `Draft`.
3. Coordinate blocker resolution in tmux and registry files.
4. Freeze `COORDINATION-OPS-V1`.
5. Merge validated worker branches to `main`.
6. Run full quality gates on merged state.

## Validation
```bash
./scripts/build_with_zig.sh build-agent-0-zig
./scripts/test_with_coverage.sh build-agent-0-cov
./scripts/lint_snapshot.sh
```

## Merge Checklist (per worker branch)
1. `git fetch origin && git checkout main && git pull --ff-only`
2. Contract required by worker is `Frozen`.
3. Worker diff is within owned paths.
4. Worker validation evidence is available.
5. Merge branch into `main`.
6. Update `agents/contracts.md` and `agents/notes.md` statuses.
