# Codebase Audit Report

**Date:** 2026-01-29  
**Audit Scope:** Compliance with `docs/spec-v1.22.md` and completion of Phase 1 in `docs/plan-v2.4.md`  
**Remediation (2026-01-29):** Phase 1 findings fixed: spec version guard invoked at start of `amux test`; `internal/spec/spec_test.go` added for CheckSpecVersion; Phase 1 go-docmd TODO marked complete in plan.

---

## Executive Summary

The codebase is **compliant** with the specification sections applicable to Phase 0 and Phase 1 and **Phase 1 is complete**. All Phase 1 findings have been remediated: spec version guard runs at `amux test` startup, CheckSpecVersion is unit-tested, and the Phase 1 go-docmd TODO is marked complete in the plan.

### Overall Status

| Area | Status |
|------|--------|
| **Spec compliance (Phase 0+1 scope)** | Compliant |
| **Phase 0 completion** | Complete (per plan checkboxes) |
| **Phase 1 completion** | One open TODO (docs); implementation complete |
| **Agent-agnostic invariant** | Enforced (no `adapters/` imports in `internal/`) |

---

## 1. Spec Compliance

### 1.1 Conventions (§4.2)

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **§4.2.1** Go 1.25.6 | `go.mod`: `go 1.25.6` | OK |
| **§4.2.2** wazero / TinyGo | Not yet used in code (Phase 8); no wazero/creack/pty in `go.mod` | Deferred |
| **§4.2.3** hsm-go, muid, lifecycle/presence HSMs | `internal/agent`: LifecycleModel, PresenceModel; `pkg/api`: ID (muid-compatible), base-10 JSON | OK |
| **§4.2.4** creack/pty | Not in use yet (Phase 5) | Deferred |
| **§4.2.5** Error handling | `fmt.Errorf("...: %w", err)` and `errors.New()` used; no deferred error checks in core paths | OK |
| **§4.2.6** Project structure | `cmd/amux`, `cmd/amux-node`, `internal/*`, `pkg/api`; no agent-specific code in `internal/` | OK |
| **§4.2.6.1** Inline docs + go-docmd READMEs | Package comments and exported doc comments present; per-package `README.md`; `make docs-check` runs go-docmd and fails on uncommitted changes | OK |
| **§4.2.8** Config (TOML, hierarchy, env) | `internal/config`: TOML, load order, `AMUX__` env mapping, adapter scoping | OK |

### 1.2 Definitions and IDs (§3, §4.2.3)

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **§3.20** agent_slug | `pkg/api/ids.go`: NormalizeAgentSlug (lowercase, non-[a-z0-9-]→`-`, collapse, trim, max 63, default `"agent"`); UniquifyAgentSlug for collisions | OK |
| **§3.21** Agent.ID (muid) | `pkg/api`: ID type, NextRuntimeID (retries until non-zero), ValidRuntimeID | OK |
| **§3.22** Reserved ID 0 | BroadcastID = 0; never assigned as runtime ID; EncodeID used for wire; ValidRuntimeID(0)==false | OK |
| **§3.23** repo_root canonicalization | `internal/paths`: CanonicalizeRepoRoot (expand `~/`, absolute, clean, EvalSymlinks); expandHome for `~/` | OK |
| **Wire:** IDs base-10 in JSON | `pkg/api/ids.go`: MarshalJSON/UnmarshalJSON for ID; lifecycle/presence events use `api.EncodeID(agentID)` | OK |

### 1.3 Agent and Session (§5.1, §5.3.1, §5.4, §5.5.9)

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **§5.1** Agent struct | `pkg/api/types.go`: ID, Name, About, Adapter (string), RepoRoot, Worktree, Location | OK |
| **§5.1** Location | LocationType (Local/SSH), Host, User, Port, RepoPath | OK |
| **§5.3.1** Worktree path | `internal/paths`: WorktreePath → `.amux/worktrees/{agent_slug}/` under repo root | OK |
| **§5.4** Lifecycle HSM | `internal/agent/lifecycle.go`: pending→starting→running→terminated/errored; events start, ready, stop, error; lifecycle.changed emitted | OK |
| **§5.5.9** Session | `pkg/api/types.go`: Session{ID, AgentID} | OK |

### 1.4 Presence (§6.1, §6.5)

| Requirement | Implementation | Status |
|-------------|----------------|--------|
| **§6.1** Presence states | Online, Busy, Offline, Away (constants and HSM state names) | OK |
| **§6.5** Presence HSM | `internal/agent/presence.go`: transitions per spec (task.assigned, task.completed, prompt.detected, rate.limit, rate.cleared, stuck.detected, activity.detected); presence.changed emitted | OK |

### 1.5 Agent-Agnostic Invariant (§1.5.1, §4.2.6)

- **No imports from `adapters/` in `internal/`.** Only reference to "adapters" in `internal/` is the config key prefix `adapters.<name>` in `internal/config/adapter.go` (spec §4.2.8.2).  
- **Compliant.**

### 1.6 Spec Version Guard (Plan Phase 0)

- **Requirement:** “spec-v1.22.md exists in-repo; a guard test or startup check fails fast with a clear error if the file is missing or the expected version marker does not match.”
- **Implementation:** `internal/spec/spec.go` provides `CheckSpecVersion(repoRoot)` (file existence + `**Version:** v1.22` in content).
- **Remediation:** `amux test` now calls `spec.CheckSpecVersion(moduleRoot)` at startup (before running the test sequence). `internal/spec/spec_test.go` adds unit tests for missing file, wrong version marker, and valid marker.

---

## 2. Phase 1 Completion (plan-v2.4.md)

### 2.1 Completed Items

- **Baseline snapshot for Phase 1** — Snapshots under `snapshots/amux-test-*.toml` exist; regression run passes.
- **Identifiers and normalization** — `pkg/api/ids.go` (EncodeID, DecodeID, NormalizeAgentSlug, UniquifyAgentSlug, NextRuntimeID, ValidRuntimeID); `internal/paths` CanonicalizeRepoRoot; tests in `pkg/api/ids_test.go` and `internal/paths/paths_test.go`.
- **Agent and Session structures** — `pkg/api/types.go` (Agent, Location, Session); `internal/agent` NewActor validates non-zero ID.
- **Lifecycle HSM** — LifecycleModel, Actor.DispatchLifecycle, lifecycle.changed events; tests in `internal/agent/lifecycle_test.go`.
- **Presence HSM** — PresenceModel, Actor.DispatchPresence, presence.changed events; tests in `internal/agent/presence_test.go`.
- **Regression and plan** — `amux test --regression` exits 0; Phase 1 TODOs in plan updated and committed.

### 2.2 Phase 1 Docs TODO

- **Item:** “The implementation MUST maintain inline Go documentation and MUST regenerate per-package README.md files via go-docmd” (plan §4.2.6.1).
- **Remediation:** Marked complete in plan-v2.4.md; Phase 1 packages already had docs and generated READMEs; docs-check passes.

---

## 3. Verification Commands

| Command | Result |
|---------|--------|
| `go test ./...` | Pass |
| `go run ./cmd/amux test --regression` | Pass (no regressions) |
| `make docs-check` | Pass (no uncommitted doc changes) |
| `make verify` | Not re-run; assumes tidy, vet, lint, test, docs-check |

---

## 4. Recommendations

1. **Spec version guard:** Call `spec.CheckSpecVersion(moduleRoot)` at the beginning of `amux test` (after resolving module root) and/or add a test in `internal/spec` that runs from the repo root so the “guard test or startup check” requirement is satisfied.
2. **Phase 1 plan:** Mark the go-docmd/docs Phase 1 TODO as complete in `docs/plan-v2.4.md` (lines 241–244).
3. **Optional:** Add a test that ensures `CheckSpecVersion` fails when the spec file is missing or the version marker is wrong (e.g. in a temp dir with a modified or absent spec file).

---

## 5. Summary

- **Spec:** Compliant with spec-v1.22 for Phase 0 and Phase 1 scope; IDs, Agent/Session, lifecycle and presence HSMs, paths, and config align with the spec. Agent-agnostic invariant is maintained.
- **Phase 1:** All functional work is done; tests and regression pass; docs and go-docmd are in place and in sync. Phase 1 is complete: spec version guard runs at `amux test` startup, CheckSpecVersion is unit-tested, and the Phase 1 go-docmd TODO is marked complete in the plan.
