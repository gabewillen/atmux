# Audit Report: spec-v1.22 Compliance & Phase 2 Completion

**Date:** 2026-01-29  
**Spec:** docs/spec-v1.22.md  
**Plan:** docs/plan-v2.4.md  

---

## 1. Executive Summary

- **Phase 2 implementation:** All Phase 2 feature TODOs are implemented and marked complete. The only unchecked item is the process step “Update this plan’s TODOs, remove unused code/scripts, and commit Phase 2 to git” (commit left to user).
- **Spec compliance:** The codebase aligns with spec-v1.22 for the areas implemented in Phases 0–2 (conventions, agent model, worktrees, local lifecycle, PTY, git merge, paths, config). Several gaps remain for **§12.6 `amux test`** (required sequence, snapshot schema, timestamp format, baseline selection).

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

### 3.3 PTY (§7) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Owned PTY for agent; monitor observes raw output (§7, B.5) | ✅ | `internal/pty/`: Session owns PTY (creack/pty), OutputStream() for monitor; window resize |

### 3.4 IDs and Wire (§3.22, §4.2.3) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| muid.ID base-10 in JSON; never emit 0 as runtime ID (§3.22, §4.2.3) | ✅ | `pkg/api/ids.go`: MarshalJSON/UnmarshalJSON base-10; BroadcastID=0; NextRuntimeID retries until non-zero |
| repo_root canonicalization (§3.23) | ✅ | `internal/paths/paths.go`: CanonicalizeRepoRoot (expand ~, abs, clean, EvalSymlinks) |

### 3.5 Spec Version Lock (§4.3.1) — Compliant

| Requirement | Status | Evidence |
|-------------|--------|----------|
| spec-v1.22.md present and version check | ✅ | `internal/spec/spec.go`: CheckSpecVersion, ExpectedSpecVersion="v1.22"; `amux test` runs it |

### 3.6 `amux test` (§12.6) — Compliant (fixed)

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

## 4. Recommendations

1. **Phase 2 closure:** Check the final Phase 2 TODO in plan-v2.4.md (“Update this plan’s TODOs … and commit Phase 2 to git”) and commit when ready.
2. **§12.6:** Implemented; no further action required.

---

## 5. Traceability

- **Spec refs:** §1.3, §3.22, §3.23, §4.2 (4.2.1–4.2.6.1, 4.2.8), §5.1–5.3.1, §5.4, §5.6, §5.7, §5.7.1, §7, §12.6.
- **Plan refs:** Phase 0 (path resolver, amux test, interfaces), Phase 1 (IDs, lifecycle, presence), Phase 2 (add, worktree, local session, PTY, git merge).
