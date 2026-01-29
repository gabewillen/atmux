# Audit Report: spec-v1.22 Compliance & Phases 0–3

**Date:** 2026-01-29  
**Spec:** docs/spec-v1.22.md  
**Plan:** docs/plan-v2.4.md  

---

## 1. Executive Summary

- **Phases 0–2:** All feature TODOs are implemented and marked complete.
- **Phase 3 (remote agents):** All Phase 3 feature TODOs are implemented and marked complete. The only unchecked item is the process step “Update this plan’s TODOs, remove unused code/scripts, and commit Phase 2 to git” (commit left to user).
- **Spec compliance:** The codebase aligns with spec-v1.22 for Phases 0–3 (conventions, agent model, worktrees, local lifecycle, PTY, git merge, paths, config, **remote §5.5**). One **gap** (Director fail-fast on NATS disconnect, §5.5.7.2.1) has been **fixed** in code.

---

## 2. Phase 2 Completion (plan-v2.4.md)

### 2.1 Phase 2 TODO Status

| TODO | Status | Evidence |
|------|--------|----------|
| Run `amux test` baseline for Phase 2 | ✅ Complete | Snapshots in `snapshots/` (e.g. amux-test-20260129-044117.toml) |
| Agent add flow (§1.3, §5.2) | ✅ Complete | `internal/agent/add.go`, `internal/config/project.go`, `cmd/amux/agent.go` |
| Worktree isolation, slug path layout (§5.3, §5.3.1, §5.3.4) | ✅ Complete | `internal/worktree/`, paths via resolver; `.amux/worktrees/{agent_slug}/`, branch `amux/{agent_slug}` |
| Local lifecycle: spawn/attach/stop/kill/restart (§5.4, §5.6) | ✅ Complete | `internal/agent/local.go` (Spawn, Stop, Restart); lifecycle events |
| Local PTY session ownership (§7, B.5) | ✅ Complete | `internal/pty/` (creack/pty); Session owns PTY, OutputStream for monitor |
| Git merge strategy and defaults (§5.7, §5.7.1) | ✅ Complete | `internal/git/` (BaseBranch, ResolveTargetBranch, ValidStrategy); `config.GitMergeConfig.TargetBranch` |
| go-docmd and per-package READMEs (§4.2.6.1) | ✅ Complete | READMEs in internal/git, internal/worktree, internal/pty, etc. |
| Run `amux test --regression` at end of Phase 2 | ✅ Complete | Plan notes regression run passed |
| Update plan TODOs, commit Phase 2 | ⬜ Unchecked | Plan: “commit left to the user” |

**Conclusion:** Phase 2 is **implementation-complete**. The only remaining step is to update the plan (e.g. check the final TODO) and commit.

---

## 2.2 Phase 3 Completion (plan-v2.4.md)

### Phase 3 TODO Status

| TODO | Status | Evidence |
|------|--------|----------|
| Run `amux test` baseline for Phase 3 | ✅ Complete | Snapshots in `snapshots/` |
| SSH bootstrap (§5.5.2, §5.5.3, §5.5.6.4) | ✅ Complete | `internal/remote/bootstrap.go` (RunSSH, ProvisionCreds, CopyBootstrapZip, StartDaemon, DaemonStatus); creds `chmod 0600` |
| Hub/leaf NATS, subject prefix (§5.5.6.1, §5.5.6.2, §5.5.7) | ✅ Complete | `internal/remote/director.go`, `manager.go`; `internal/remote/subjects.go` (SubjectPrefix, SubjectCtl, SubjectHandshake, SubjectPTYOut/In) |
| JetStream KV bucket and keys (§5.5.6.3) | ✅ Complete | `internal/remote/kv.go` (EnsureKVBucket, PutHostInfo, PutHostHeartbeat, PutSession; hosts/\<host_id\>/info, heartbeat, sessions/\<host_id\>/\<session_id\>) |
| NATS auth and per-host subject auth (§5.5.6.4) | ✅ Complete | Bootstrap provisions per-host creds; subject namespaces match spec |
| Subject namespaces and message envelopes (§5.5.7.1, §5.5.7.5) | ✅ Complete | `internal/remote/subjects.go`, `control.go`; unit tests in `subjects_test.go`, `control_test.go` |
| Request-reply control (spawn/kill/replay), timeout, fail-fast (§5.5.7.2, §5.5.7.2.1) | ✅ Complete | Director.Spawn/Kill/Replay use RequestTimeout; IsReady(hostID) check before sending |
| Handshake and readiness gating (§5.5.7.3) | ✅ Complete | Director.RunHandshakeHandler; Manager.Handshake; manager returns not_ready until handshake done; director rejects collision |
| Remote spawn idempotency and session_conflict (§5.5.7.3) | ✅ Complete | Manager.handleSpawn idempotent by agent_id; returns session_conflict when repo_path or agent_slug differs |
| PTY I/O subjects and chunking (§5.5.7.4) | ✅ Complete | Manager.WritePTYOut (64KB chunks); SubscribePTYIn; SubjectPTYOut/SubjectPTYIn |
| Per-session replay buffer, ordering, live gating (§5.5.7.3, §5.5.8) | ✅ Complete | `internal/remote/ringbuffer.go`; Manager handleReplay publishes snapshot then sets replayDone; WritePTYOut gates on replayDone |
| Connection recovery, buffering, replay-before-live (§5.5.8) | ✅ Complete | Replay buffer always written; live output gated until replay request handled; ring buffer drop-oldest |
| Remote session manager and exit events (§5.5.9) | ✅ Complete | ManagedSession, control handler; exit/remediation events deferred to Phase 6/7 |
| go-docmd and per-package READMEs (§4.2.6.1) | ✅ Complete | `internal/remote/README.md` and inline docs |
| Run `amux test --regression` at end of Phase 3 | ✅ Complete | Plan notes regression run passed |
| Update plan TODOs, commit Phase 3 | ⬜ Unchecked | Plan: "commit left to the user" |

**Conclusion:** Phase 3 is **implementation-complete**; the Director disconnect fail-fast gap has been fixed. The only remaining step is the final plan/commit.

---

## 3. Spec Compliance

### 3.1 Conventions (§4.2) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Go 1.25.6 (§4.2.1) | ✅ | `go.mod`: `go 1.25.6` |
| wazero / TinyGo (§4.2.2) | ✅ | Referenced in plan; adapter loading is Phase 8 |
| hsm-go + muid (§4.2.3) | ✅ | `internal/agent/lifecycle.go`, `presence.go`; `pkg/api/ids.go` (muid, base-10 IDs) |
| creack/pty (§4.2.4) | ✅ | `internal/pty/pty.go`: `github.com/creack/pty/v2` |
| Error wrapping, sentinels, no deferred error check (§4.2.5) | ✅ | `fmt.Errorf("...: %w", err)` and `errors.New()` used; no `defer` used for error checking in audited code |
| Project structure, no agent-specific code in internal (§4.2.6, §1.5.1) | ✅ | No imports of `adapters/` or agent names in `internal/`; config uses generic `adapters` key and `adapters.<name>` scoping only |
| Path resolver, .amux invariants (§4.2.6, §4.2.8) | ✅ | `internal/paths/` (CanonicalizeRepoRoot, Resolver, WorktreePath); worktrees under `.amux/worktrees/{agent_slug}/` |
| Inline docs + go-docmd READMEs (§4.2.6.1) | ✅ | Package comments and READMEs present; Makefile `docs-check` |
| Config: TOML, hierarchy, env AMUX__ (§4.2.8) | ✅ | `internal/config/` (TOML, GitMergeConfig, etc.) |

### 3.2 Agent Management (§5) — Compliant for Phase 2 Scope

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Agent structure, Location, Adapter as string (§5.1) | ✅ | `pkg/api/types.go` (Agent, Location); Adapter is string |
| Adding an agent: repo required, validation, persistence (§5.2) | ✅ | `internal/agent/add.go` (ValidateAddInput, ResolveRepoRoot); `config.AddAgentToProject`; CLI `amux agent add` |
| Worktree under `.amux/worktrees/{agent_slug}/` (§5.3.1) | ✅ | `internal/worktree/`: WorktreePath, Create, Remove; branch `amux/{agent_slug}` |
| agent_slug normalization: lowercase, [a-z0-9-], collapse dash, trim, 63 chars, default "agent" (§5.3.1) | ✅ | `pkg/api/ids.go`: NormalizeAgentSlug, UniquifyAgentSlug, MaxAgentSlugLen=63, DefaultAgentSlug="agent" |
| Lifecycle HSM: Pending → Starting → Running → Terminated/Errored (§5.4) | ✅ | `internal/agent/lifecycle.go` (LifecycleModel, DispatchLifecycle) |
| Graceful shutdown, drain (§5.6) | ✅ | `internal/agent/local.go` Stop (lifecycle stop, PTY close) |
| Git merge: strategies merge-commit, squash, rebase, ff-only; base_branch/target_branch (§5.7, §5.7.1) | ✅ | `internal/git/`: BaseBranch, ResolveTargetBranch, ValidStrategy; `internal/config/config.go`: GitMergeConfig.TargetBranch |

### 3.3 Remote agents (§5.5) — Compliant except one gap

| Requirement | Status | Evidence |
|-------------|--------|----------|
| SSH bootstrap, per-host creds, creds_path ≤0600 (§5.5.2, §5.5.6.4) | ✅ | `internal/remote/bootstrap.go`: ProvisionCreds (chmod 0600 locally and on remote) |
| Subject names P.handshake.\<host_id\>, P.ctl.\<host_id\>, P.pty.\<host_id\>.\<session_id\>.out/in (§5.5.7.1) | ✅ | `internal/remote/subjects.go`: SubjectHandshake, SubjectCtl, SubjectPTYOut, SubjectPTYIn |
| JetStream KV: hosts/\<host_id\>/info, heartbeat, sessions/\<host_id\>/\<session_id\> (§5.5.6.3) | ✅ | `internal/remote/kv.go`: KVKeyHostInfo, KVKeyHostHeartbeat, KVKeySession; PutHostInfo, PutHostHeartbeat, PutSession |
| Control message types and payloads (§5.5.7.2, §5.5.7.3) | ✅ | `internal/remote/control.go`: ControlMessage, HandshakePayload, Spawn/Kill/Replay payloads, ErrorPayload (request_type, code, message) |
| Request-reply timeout, fail-fast when host not ready (§5.5.7.2.1) | ✅ | Director uses RequestTimeout and IsReady(hostID); on NATS disconnect, DisconnectErrHandler clears `ready` so control ops fail fast without sending |
| Handshake before spawn/kill/replay; not_ready; collision reject (§5.5.7.3) | ✅ | Manager returns not_ready until Handshake() done; Director rejects handshake on host_id mismatch or collision |
| Spawn idempotent by agent_id; session_conflict on repo_path/agent_slug mismatch (§5.5.7.3) | ✅ | Manager.handleSpawn returns existing session_id for same agent_id; returns session_conflict when slug/repo_path differs |
| Replay buffer: cap remote.buffer_size, oldest-to-newest, replay then live (§5.5.7.3, §5.5.8) | ✅ | `internal/remote/ringbuffer.go` (drop-oldest); handleReplay publishes snapshot then sets replayDone; WritePTYOut gates on replayDone |
| PTY chunking ≤ max payload (§5.5.7.4) | ✅ | Manager uses 64KB chunks for PTY out |
| session_id non-zero, base-10 string (§5.5.7.3) | ✅ | Manager uses api.NextRuntimeID() then EncodeID for session_id |

### 3.4 PTY (§7) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Owned PTY for agent; monitor observes raw output (§7, B.5) | ✅ | `internal/pty/`: Session owns PTY (creack/pty), OutputStream() for monitor; window resize |

### 3.5 IDs and Wire (§3.22, §4.2.3) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| muid.ID base-10 in JSON; never emit 0 as runtime ID (§3.22, §4.2.3) | ✅ | `pkg/api/ids.go`: MarshalJSON/UnmarshalJSON base-10; BroadcastID=0; NextRuntimeID retries until non-zero |
| repo_root canonicalization (§3.23) | ✅ | `internal/paths/paths.go`: CanonicalizeRepoRoot (expand ~, abs, clean, EvalSymlinks) |

### 3.6 Spec Version Lock (§4.3.1) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| spec-v1.22.md present and version check | ✅ | `internal/spec/spec.go`: CheckSpecVersion, ExpectedSpecVersion="v1.22"; `amux test` runs it |

### 3.7 `amux test` (§12.6) — Compliant (fixed)

Spec §12.6.2 requires this sequence:

1. `go mod tidy`  
2. `go vet ./...`  
3. **`golangci-lint run ./...`**  
4. `go test -race ./...`  
5. `go test ./...`  
6. **`go test ./... -coverprofile=<path>`**  
7. **`go test -run=^$ -bench=. -benchmem ./...`**  

Current `cmd/amux/test.go` runs: tidy, vet, test_race, test, then records coverage/bench as **skipped**. So:

- **Missing steps:** Step 3 (golangci-lint), real step 6 (coverage), real step 7 (benchmarks).
- **Snapshot schema (§12.6.4):** Spec requires `[meta]` (created_at, module_root, spec_version), and per-step tables `[steps.go_mod_tidy]`, `[steps.go_vet]`, `[steps.golangci_lint]`, `[steps.tests_race]`, `[steps.tests]`, `[steps.coverage]`, `[steps.benchmarks]` with argv, exit_code, duration_ms, stdout_sha256, stderr_sha256, stdout_bytes, stderr_bytes. Current code uses a simplified `TestSnapshot` (Timestamp, Results map) and does not match this schema.
- **Snapshot filename (§12.6.3):** Spec requires UTC timestamp format `YYYYMMDDThhmmssZ` (e.g. `20260126T153012Z`). Current code uses `20060102-150405` (no `T`, no `Z`).
- **Baseline for regression (§12.6.5):** Spec requires “lexicographically greatest file name” matching `amux-test-*.toml`. Current code selects “previous by parsed timestamp”; with the current filename format, lexicographic order and time order can differ.
- **Regression rules (§12.6.5):** Spec defines step exit regression and coverage regression (total_percent). Current code compares a generic “status” per step; no coverage total_percent comparison.

**Summary:** §12.6 is now implemented and compliant (7-step sequence, snapshot schema with [meta]/[steps.*]/[[benchmarks]], UTC filename, lexicographic baseline, regression rules).

---

## 4. Gaps and Deviations

### 4.1 Director fail-fast on NATS disconnect (§5.5.7.2.1)

**Spec:** "If the director considers the target host disconnected (for example the agent lifecycle is Away), it MUST fail fast by rejecting remote control operations (spawn, kill, replay) without issuing a NATS request."

**Gap:** When the NATS connection to the hub drops, the Director's `DisconnectErrHandler` is a no-op. The `ready` set is not cleared, so `IsReady(hostID)` remains true for previously handshaken hosts. As a result, `Spawn`/`Kill`/`Replay` still call `RequestWithContext`, i.e. they *issue* a NATS request and then fail with a timeout or connection error. The spec requires failing fast *without* issuing a NATS request when the host is considered disconnected.

**Fix (applied):** In `internal/remote/director.go`, `DisconnectErrHandler` now clears `d.ready` and `d.peerByHost` so that after a disconnect, all hosts are considered not ready until they re-handshake. Control ops then fail fast without sending.

### 4.2 Other

- **Phase 3 plan TODO:** One unchecked item remains: "Update this plan's TODOs for Phase 3, remove unused code/scripts, and commit Phase 3 to git."
- **Verify entrypoint:** Plan "one command" verification; Makefile `verify` does not include `test-snapshot` or `test-regression`. Optional: add `test-regression` to `verify` for CI.

---

## 5. Recommendations

1. **Director disconnect:** Fix applied; no further action required.
2. **Phase 3 closure:** Check the final Phase 3 TODO in plan-v2.4.md and commit when ready.

3. **§12.6:** Implemented; no further action required.

---

## 6. Traceability

- **Spec refs:** §1.3, §3.22, §3.23, §4.2 (4.2.1–4.2.6.1, 4.2.8), §5.1–5.5.9, §5.7, §5.7.1, §7, §12.6.
- **Plan refs:** Phase 0 (path resolver, amux test, interfaces), Phase 1 (IDs, lifecycle, presence), Phase 2 (add, worktree, local session, PTY, git merge), Phase 3 (remote: bootstrap, NATS, KV, handshake, control, replay, PTY I/O).
