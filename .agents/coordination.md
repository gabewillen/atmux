# Multi-Agent Coordination Protocol

## Objective
Ship a fully functioning library and application while preventing divergence, unplanned overlap,
and merge churn.

## tmux Topology
- Coordinator session: `agent-0`
- Worker sessions: `agent-1`, `agent-2`, `agent-3`, `agent-4`
- Every agent works inside its own tmux session and worktree.

## Session Mapping
- `agent-0` -> repository root (`main` integration and coordination)
- `agent-1` -> `../machine-agent-1`
- `agent-2` -> `../machine-agent-2`
- `agent-3` -> `../machine-agent-3`
- `agent-4` -> `../machine-agent-4`

## Ownership Rules
- Agents only modify files in the owned-path list from their plan.
- Editing files owned by another agent is blocked unless the owner records approval in that
  owner's `agents/agent-N/notes.md`.
- `agent-0` is the only agent that updates:
  - `agents/contracts.md`
  - `agents/notes.md`
  - `agents/coordination.md`
  - merge commits into `main`

## Communication Rules
- Workers (`agent-1..agent-4`) communicate cross-team through `agent-0` only.
- Worker-to-worker direct coordination is not allowed.
- Questions, blockers, and merge requests are sent to `agent-0` via tmux messages.

## tmux Message Format
- Use this normalized payload:
  - `[MSG][from:agent-N][type:question|blocker|handoff|ready-to-merge|status][ts:ISO-8601] text`
- Worker send command:
```bash
./agents/tmux/send_to_agent0.sh agent-1 blocker "Need decision on shared API field"
```
- Coordinator reply command:
```bash
./agents/tmux/send_to_agent.sh agent-2 decision "Proceed with INFERENCE-GEN-V1 assumptions"
```

## Contract and Notes Layout
- Agent 0 contract: `agents/agent-0/contract.md`
- Agent 1 contract: `agents/agent-1/contract.md`
- Agent 2 contract: `agents/agent-2/contract.md`
- Agent 3 contract: `agents/agent-3/contract.md`
- Agent 4 contract: `agents/agent-4/contract.md`
- Agent notes: `agents/agent-N/notes.md`
- Contract registry: `agents/contracts.md` (owned by `agent-0`)
- Blocker registry: `agents/notes.md` (owned by `agent-0`)

## Contract Lifecycle
1. `Draft`: contract exists but may change.
2. `Frozen`: downstream agents may implement against it.
3. `Superseded`: replaced by a higher version.

Rules:
- Breaking changes after `Frozen` require a new version (`V2`, `V3`, ...).
- Do not rewrite older contract guarantees; mark them `Superseded`.

## Sync Cadence
- Rebase at start of day, before each push, and at least every 90 minutes during active work.
- Before requesting merge, run:
```bash
git fetch origin
git rebase origin/main
git diff --name-only origin/main...HEAD
```
- The file list must stay within owned paths.

## Merge Train (Coordinator-Managed)
1. `agent-1` freezes `ABI-LIFECYCLE-V1` and sends `ready-to-merge` to `agent-0`.
2. `agent-0` merges `agent-1` branch into `main`.
3. `agent-2` and `agent-3` run in parallel, freeze contracts, and request merge.
4. `agent-0` merges `agent-2` and `agent-3` after validation.
5. `agent-4` integrates app target after upstream contracts are frozen.
6. `agent-0` merges `agent-4` last and runs final quality gates.

## Conflict and Blocker Escalation
- If overlap is required, worker sends blocker to `agent-0` and logs details in
  `agents/agent-N/notes.md`.
- `agent-0` decides owner, sequencing, and required contract updates.
- No worker resolves cross-owner conflict without `agent-0` direction.
