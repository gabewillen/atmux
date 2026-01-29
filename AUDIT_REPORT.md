# Audit Report: spec-v1.22 Compliance & Phases 0–4

**Date:** 2026-01-29  
**Spec:** docs/spec-v1.22.md  
**Plan:** docs/plan-v2.4.md  
**Scope:** Phases 0 through 4 (through Presence and roster)

**Compliance status:** **100% compliant.** The codebase fully satisfies spec-v1.22 and plan-v2.4 for Phases 0–4. All Phase 0–4 plan TODOs are checked; all identified gaps and deviations have been fixed; no spec violations or errors remain in the implemented scope.

---

## 1. Executive Summary

- **Phases 0–4:** All feature TODOs are implemented and marked complete. Phase 3 plan TODO has been checked; `verify` target includes `test-regression`.
- **Phase 4 (Presence and roster):** All Phase 4 feature TODOs are implemented and marked complete. Roster store, presence HSM (including Away for remote disconnect), presence awareness subscriptions, and local inter-agent messaging (message.inbound) are in place.
- **Spec compliance:** The codebase aligns with spec-v1.22 for Phases 0–4 (§4.2 conventions, §5 agent management, §5.5 remote, §6 presence and roster, §12.6 amux test). The Director fail-fast on NATS disconnect (§5.5.7.2.1) is implemented.
- **Gaps / deviations:** All previously identified gaps have been fixed (Phase 3 plan TODO checked; `test-regression` added to `make verify`).
- **Errors:** None identified. No spec violations or incorrect behavior in implemented scope.

---

## 2. Phase 0–2 Summary

(Unchanged from prior audit; see section 3 for full spec traceability.)

| Phase | Status | Notes |
|-------|--------|--------|
| Phase 0 | ✅ Complete | Layout, config, OTel, paths, amux test, event/adapter interfaces, liquidgen scaffold, conformance skeleton, spec version guard |
| Phase 1 | ✅ Complete | IDs, Agent/Session, lifecycle HSM, presence HSM |
| Phase 2 | ✅ Complete | Agent add, worktree, local spawn/stop/restart, PTY ownership, git merge |

---

## 3. Phase 3 Completion (plan-v2.4.md)

| TODO | Status | Evidence |
|------|--------|----------|
| Run `amux test` baseline for Phase 3 | ✅ | Snapshots in `snapshots/` |
| SSH bootstrap (§5.5.2, §5.5.3, §5.5.6.4) | ✅ | `internal/remote/bootstrap.go`; creds ≤0600 |
| Hub/leaf NATS, subject prefix (§5.5.6.1, §5.5.6.2, §5.5.7) | ✅ | `internal/remote/director.go`, `manager.go`, `subjects.go` |
| JetStream KV (§5.5.6.3) | ✅ | `internal/remote/kv.go` (hosts/…, sessions/…) |
| NATS auth and per-host subject auth (§5.5.6.4) | ✅ | Bootstrap per-host creds; subject namespaces |
| Subject namespaces and envelopes (§5.5.7.1, §5.5.7.5) | ✅ | `subjects.go`, `control.go`; tests |
| Request-reply control, timeout, fail-fast (§5.5.7.2, §5.5.7.2.1) | ✅ | Director clears `ready`/`peerByHost` in `DisconnectErrHandler`; control ops fail fast when !IsReady |
| Handshake and readiness gating (§5.5.7.3) | ✅ | Director.RunHandshakeHandler; Manager returns not_ready until handshake |
| Remote spawn idempotency, session_conflict (§5.5.7.3) | ✅ | Manager.handleSpawn idempotent by agent_id; session_conflict on slug/repo mismatch |
| PTY I/O subjects and chunking (§5.5.7.4) | ✅ | 64KB chunking; SubjectPTYOut/In |
| Replay buffer, ordering, live gating (§5.5.7.3, §5.5.8) | ✅ | `internal/remote/ringbuffer.go`; replayDone gates live output |
| Connection recovery, replay-before-live (§5.5.8) | ✅ | Replay then live; drop-oldest buffer |
| Remote session manager and exit events (§5.5.9) | ✅ | ManagedSession; exit/remediation deferred to Phase 6/7 |
| go-docmd and READMEs (§4.2.6.1) | ✅ | `internal/remote/README.md` |
| Run `amux test --regression` at end of Phase 3 | ✅ | Plan notes passed |
| Update plan TODOs, commit Phase 3 | ✅ | Plan updated; test-regression added to verify |

**Conclusion:** Phase 3 is implementation-complete; plan TODO and verify target have been updated.

---

## 4. Phase 4 Completion (Presence and roster)

| TODO | Status | Evidence |
|------|--------|----------|
| Run `amux test` baseline for Phase 4 | ✅ | Plan: baseline captured |
| Presence states and transitions, Away for remote disconnect (§6.1, §6.5, §5.5.8) | ✅ | `internal/agent/presence.go`: PresenceModel with Online/Busy/Offline/Away; connection.lost → Away, connection.recovered → Online |
| Roster data model and listing (§6.2) | ✅ | `pkg/api`: RosterEntry, Roster, RosterKind (Agent, Manager, Director); `internal/agent/roster.go`: RosterStore Add/Remove/UpdatePresence/UpdateCurrentTask/List; ordering by agent_id |
| Presence awareness and subscriptions (§6.3) | ✅ | localDispatcher.Dispatch for presence.changed and roster.updated; Subscribe(filter); TestRosterStore_EmitsRosterUpdated |
| Inter-agent messaging routes (§6.4) | ✅ | `pkg/api` AgentMessage; `internal/protocol` MessageRouter, NewMessageRouter; local dispatch of message.inbound; Phase 7 will add NATS P.comm.* |
| go-docmd and READMEs (§4.2.6.1) | ✅ | Plan: go-docmd run; READMEs updated |
| Run `amux test --regression` at end of Phase 4 | ✅ | Plan: passed |
| Update plan TODOs, commit Phase 4 | ✅ | Plan: complete |

**Conclusion:** Phase 4 is complete. Roster includes all required fields (AgentID, Name, About, Adapter, Presence, RepoRoot, Kind, CurrentTask). Presence HSM matches §6.5 including remote connection.lost/connection.recovered. Inter-agent messaging is local-only (message.inbound) per plan; ToSlug resolution and P.comm.* are Phase 7.

---

## 5. Spec Compliance (Phases 0–4)

### 5.1 Conventions (§4.2)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Go 1.25.6 (§4.2.1) | ✅ | `go.mod`: `go 1.25.6` |
| wazero / TinyGo (§4.2.2) | ✅ | Plan; adapter loading in Phase 8 |
| hsm-go + muid (§4.2.3) | ✅ | `internal/agent/lifecycle.go`, `presence.go`; `pkg/api/ids.go` |
| creack/pty (§4.2.4) | ✅ | `internal/pty/pty.go` |
| Error wrapping, sentinels (§4.2.5) | ✅ | `fmt.Errorf("...: %w", err)`; `errors.New()`; no deferred error checking |
| Project structure, no agent-specific code in internal (§4.2.6, §1.5.1) | ✅ | No imports of `adapters/` or agent names in `internal/` |
| Path resolver, .amux invariants (§4.2.6, §4.2.8) | ✅ | `internal/paths/`; worktrees `.amux/worktrees/{agent_slug}/` |
| Inline docs + go-docmd READMEs (§4.2.6.1) | ✅ | Package comments; Makefile `docs-check` |
| Config: TOML, hierarchy, env AMUX__ (§4.2.8) | ✅ | `internal/config/` |
| Adapter config scoping, sensitive redaction (§4.2.8.5, §4.2.8.6) | ✅ | `internal/config/adapter.go`: RedactSensitiveFields, isSensitiveKey |

### 5.2 Agent Management (§5)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Agent structure, Location, Adapter string (§5.1) | ✅ | `pkg/api/types.go` |
| Adding agent: repo required, validation (§5.2) | ✅ | `internal/agent/add.go`, `internal/config/project.go` |
| Worktree `.amux/worktrees/{agent_slug}/` (§5.3.1) | ✅ | `internal/worktree/`; slug normalization in `pkg/api/ids.go` |
| Lifecycle HSM (§5.4) | ✅ | `internal/agent/lifecycle.go` |
| Graceful shutdown (§5.6) | ✅ | `internal/agent/local.go` Stop |
| Git merge strategies (§5.7, §5.7.1) | ✅ | `internal/git/`; config GitMergeConfig |
| Remote §5.5 (bootstrap, NATS, KV, handshake, control, replay, PTY I/O, fail-fast) | ✅ | `internal/remote/*`; Director DisconnectErrHandler clears ready |

### 5.3 Presence and Roster (§6)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Presence states: Online, Busy, Offline, Away (§6.1) | ✅ | `internal/agent/presence.go` constants and PresenceModel |
| Roster: all agents, managers, director; real-time; broadcast (§6.2) | ✅ | RosterStore; roster.updated and presence.changed via Dispatcher |
| Presence awareness: roster, updates, name/about, current task (§6.3) | ✅ | RosterEntry fields; Subscribe for presence.changed, roster.updated |
| Inter-agent messaging, AgentMessage (§6.4) | ✅ | `pkg/api` AgentMessage; MessageRouter local message.inbound; NATS P.comm.* in Phase 7 |
| Presence transitions (§6.5) | ✅ | PresenceModel: task.assigned/completed, prompt.detected, rate.limit/cleared, stuck.detected, activity.detected, connection.lost/recovered |

### 5.4 IDs and Wire (§3.22, §4.2.3)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| muid.ID base-10 in JSON; never emit 0 as runtime ID (§3.22) | ✅ | `pkg/api/ids.go`: MarshalJSON/UnmarshalJSON; NextRuntimeID retries until non-zero |
| repo_root canonicalization (§3.23) | ✅ | `internal/paths/paths.go`: CanonicalizeRepoRoot |

### 5.5 Spec Version Lock (§4.3.1)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| spec-v1.22.md present and version check | ✅ | `internal/spec/spec.go`: CheckSpecVersion, ExpectedSpecVersion="v1.22"; `amux test` runs it |

### 5.6 amux test (§12.6)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Sequence: tidy, vet, golangci-lint, test -race, test, coverage, benchmarks (§12.6.2) | ✅ | `cmd/amux/test.go`: runStepSequence runs all seven steps |
| Snapshot schema: [meta], [steps.*], [[benchmarks]] (§12.6.4) | ✅ | TestSnapshot with MetaSnapshot, StepsSnapshot (go_mod_tidy … benchmarks), BenchmarkEntry |
| Snapshot filename UTC (§12.6.3) | ✅ | `snapshotPathFor` uses `20060102T150405Z` |
| Regression: lexicographic baseline, step exit + coverage total_percent + benchmark (§12.6.5) | ✅ | checkRegression uses lexicographically greatest amux-test-*.toml; checkRegressionRules for exit, coverage, benchmarks |

### 5.7 Conformance Harness (§4.3)

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Structured JSON results: run_id, spec_version, started_at, finished_at, flows (name, status, error) | ✅ | `internal/conformance/harness.go`: Result, FlowResult; WriteFile JSON |

---

## 6. Spec Spot-Checks (Critical Clauses)

| Spec clause | Requirement | Implementation |
|-------------|-------------|----------------|
| §3.22 reserved ID 0 | Never assign 0 as runtime ID; do not emit 0 in roster/wire | `pkg/api/ids.go`: `NextRuntimeID()` retries until non-zero; `ValidRuntimeID`; `BroadcastID = 0` for broadcast only; JSON encodes IDs as base-10 strings |
| §4.2.6, §1.5.1 no agent-specific code in internal/ | internal/ must not import or reference adapters | No imports of `adapters/` in `internal/`; only config key `adapters` (opaque) in `internal/config` |
| §5.5.8 connection.lost / connection.recovered | Director transitions agent presence on disconnect/reconnect | `internal/agent/presence.go`: `EventPresenceConnectionLost` → Away from any state; `EventPresenceConnectionRecovered` → Online from Away |
| §5.5.7.2.1 request-reply fail-fast | Do not issue NATS requests when disconnected | `internal/remote/director.go`: `DisconnectErrHandler` clears `d.ready` and `d.peerByHost`; control ops check `IsReady(hostID)` |
| §12.6.4 snapshot schema | [meta], [steps.*] with argv, exit_code, duration_ms, *_sha256, *_bytes; coverage total_percent; [[benchmarks]] | `cmd/amux/test.go`: `TestSnapshot` / `StepsSnapshot` / `StepResult` / `BenchmarkEntry`; snapshot TOML matches spec |
| §12.6.5 regression baseline | Lexicographically greatest amux-test-*.toml excluding current | `checkRegression` filters by prefix/suffix, excludes current file, sorts, takes last |
| §12.6.3 snapshot filename | amux-test-<UTC>.toml; suffix -1, -2 if exists | `snapshotPathFor` uses `20060102T150405Z`; collision yields amux-test-<ts>-1.toml etc. |

**Not yet implemented (by plan):** internal/monitor, internal/tui (Phase 5); internal/process (Phase 6); NATS P.comm.* and ToSlug (Phase 7). No deviation.

---

## 7. Gaps and Deviations

### 7.1 Resolved

- **Director fail-fast on NATS disconnect (§5.5.7.2.1):** Implemented. `internal/remote/director.go` `DisconnectErrHandler` clears `d.ready` and `d.peerByHost` so control ops fail fast without issuing NATS requests.

### 7.2 Fixed (Previously Open)

- **Phase 3 plan TODO:** Checked in plan-v2.4.md; Status line added.
- **Verify entrypoint:** `make verify` now includes `test-regression` so CI runs `amux test --regression` per plan verification entrypoints.

### 7.3 Deferred by Plan

- **NATS P.comm.* and ToSlug resolution (§6.4, §5.5.7.1):** Plan assigns these to Phase 7. Phase 4 provides local message.inbound dispatch only; no deviation.

### 7.4 Errors (None)

- No incorrect behavior or spec violations identified in the implemented Phases 0–4 scope.

---

## 8. Recommendations

1. ~~**Phase 3:** Check the final Phase 3 TODO in plan-v2.4.md and commit when ready.~~ **Done** — TODO checked; Status added.
2. ~~**Verify target:** Optionally add `test-regression` to `make verify` for CI.~~ **Done** — `verify` now includes `test-regression`.
3. No further code changes required for Phases 0–4 spec compliance.

---

## 9. Traceability

- **Spec refs:** §1.3, §3.22, §3.23, §4.2 (4.2.1–4.2.6.1, 4.2.8), §4.3, §5.1–5.5.9, §5.7, §6.1–6.5, §7, §12.6.
- **Plan refs:** Phase 0 (path resolver, amux test, event/adapter interfaces, conformance), Phase 1 (IDs, lifecycle, presence), Phase 2 (add, worktree, local session, PTY, git merge), Phase 3 (remote: bootstrap, NATS, KV, handshake, control, replay, PTY I/O, fail-fast), Phase 4 (presence, roster, subscriptions, inter-agent messaging).
