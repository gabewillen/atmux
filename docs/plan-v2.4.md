# Agent Multiplexer (amux) Implementation Plan (plan-v2.4)

## Plan Header

- Version: v2.4
- Derived from: plan-v2.3.md
- Applied addendum: plan_addendum_1.md
- Revision summary:
  - The plan MUST be updated to target `spec-v1.22.md` as the authoritative specification, including version-locking checks.
  - The plan MUST require Go 1.25.6 (`go1.25.6`) for the core toolchain and dependency pinning.
  - Each phase MUST include documentation work per spec §4.2.6.1, including inline Go doc comments and `go-docmd`-generated per-package `README.md` files with an automated docs sync check.


## Overview

### Purpose
Deliver an implementation plan for **Agent Multiplexer (amux)** that is directly traceable to the authoritative specification **spec-v1.22.md**.

### How this plan maps to the spec
- Phases align to major spec sections (Conformance, Agent management, Presence, PTY monitoring, Process tracking, Event system, Adapter interface, LLM coordination, CLI control plane, CLI plugin system).
- Every TODO includes explicit **Spec reference(s)** to the authoritative spec.
- Acceptance criteria are written to be testable via unit tests, integration tests, and the conformance suite (see Spec §4.3).


### Plan execution conventions

#### Normative language and TODO semantics
- The keywords **MUST**, **SHOULD**, and **MAY** are normative.
- Every checkbox TODO item in this plan is a **MUST** unless the item explicitly says **SHOULD** or **MAY**.
- Each TODO’s “Acceptance criteria” is **MUST** and defines “done” for that item.

#### Spec references and conflict handling
- `spec-v1.22.md` **MUST** be present in the repository and treated as the normative source for any referenced requirement.
- If an implementer identifies a conflict between this plan and `spec-v1.22.md`, they **MUST NOT** silently guess or diverge; they **MUST** record the conflict and resolve it via a plan addendum (or tracked issue) before continuing.

#### Phase ordering and stubbing policy (dependency inversion)
- Phases are ordered by intended dependencies, but some phases reference later-phase implementations.
- To keep implementation sequential and the build green, Phase 0 **MUST** introduce stable interfaces and noop/local implementations for:
  - Event emission/dispatch (used by Phases 1–6; fully networked routing is completed in Phase 7).
  - Adapter-provided pattern matching/actions (used by Phase 5; WASM-backed runtime is completed in Phase 8).
- Later phases **MUST** swap in full implementations behind these interfaces without changing earlier-phase call sites.

### Roles and terminology (local glossary)

- **CLI client (`amux`)**: the user-facing CLI binary.
- **Daemon (`amux-node`)**: the long-running node/daemon binary that hosts agents and serves the JSON-RPC control plane.
- **Director role**: a daemon instance responsible for orchestrating remote hosts and sessions.
- **Manager role**: a daemon instance running on a remote host that connects to the director’s hub and runs sessions.
- **Hub / leaf**: the NATS hub (director side) and leaf server connection (manager side) used for remote orchestration.
- **Adapter**: a TinyGo-compiled WASM module loaded by the core runtime for pattern matching and actions.
- **Plugin**: a CLI extension module (WASM or remote) that talks to the daemon over JSON-RPC, gated by declared permissions.

### Filesystem & persistence layout

- The implementation **MUST** centralize all filesystem path resolution in a single “path resolver” package (e.g., `internal/paths`) fed by config/env and (where applicable) `repo_root` canonicalization.
- Subsystems **MUST NOT** hardcode paths; they **MUST** use the shared resolver.
- Repo-scoped state **MUST** be rooted in the repository’s `repo_root` and the existing `.amux/worktrees/{agent_slug}/` invariant (Phase 2) **MUST** remain true.
- Registry roots (plugins/adapters) and runtime/IPC paths (e.g., Unix sockets) **MUST** be resolved via config and the shared resolver, with deterministic layouts tested per Phase 11/Phase 10 requirements.

### Verification entrypoints (required)

The repository **MUST** provide documented, reproducible entrypoints for verification (either `make` targets or `scripts/*` commands). At minimum, the following capabilities **MUST** exist:

- Run unit tests (e.g., `go test ./...`).
- Run lint/static analysis (e.g., `staticcheck` + `go vet`).
- Run `go-docmd` documentation generation and a docs sync check that MUST fail on diffs (Spec §4.2.6.1).
- Run integration tests (may require provisioning dependencies such as NATS/JetStream).
- Run `amux test` to create a Go verification snapshot for the current module (Spec §12.6).
- Run `amux test --regression` to compare against the previous snapshot and fail on regressions (Spec §12.6.5).
- Run the conformance suite end-to-end.

### Conformance harness output contract (minimum)

- Conformance runs **MUST** emit machine-readable “structured results” as UTF-8 JSON at a deterministic path (default path MAY be configured).
- The results format **MUST** include at least:
  - `run_id`, `spec_version`, `started_at`, `finished_at`
  - a list of per-flow results including `name`, `status` (`pass|fail|skip`), and an error field/artifact references when failed.

### Spec coverage matrix

| Spec section | Primary plan phase(s) | Primary verification |
|---|---|---|
| §4.2 Conventions (Go, wazero/TinyGo, HSM/IDs, PTY, errors, structure, config, OTel, liquidgen) | Phase 0, Phase 1, Phase 5, Phase 6 | Unit tests + integration smoke tests |
| §4.3 Conformance harness and suite | Phase 0, Phase 12 | Conformance suite |
| §5 Agent management (local and remote) | Phase 1, Phase 2, Phase 3 | Integration tests + conformance status/control flows |
| §6 Presence and roster | Phase 1, Phase 4, Phase 5 | Unit tests (HSM) + conformance status flows |
| §7 PTY monitoring | Phase 5 | Integration tests + conformance menu flows |
| §8 Process tracking and notifications | Phase 6 | Integration tests + conformance notification flows |
| §9 Event system (hsmnet, dispatch, schemas) | Phase 7 | Unit tests + integration tests + conformance control plane flows |
| §10 Adapter interface | Phase 8 | Adapter fixtures + conformance adapter tests |
| §11 LLM coordination | Phase 9 | Integration tests (snapshots/actions) |
| §12 CLI client and daemon (JSON-RPC control plane) | Phase 10 | Conformance control plane flows |
| §12.6 `amux test` (snapshots + `--regression`) | Phase 0 | `amux test` / `amux test --regression` |
| §13 CLI plugin system | Phase 11 | Conformance control plane flows + permission tests |

### Planning constraints (instantiated)
- **System / project name:** Agent Multiplexer (amux)
- **Target audience:** Senior engineers and AI agents implementing amux core, adapters, and plugins
- **Implementation scope:** Implement amux core binaries (`amux`, `amux-node`), agent management (local and remote), presence/roster, PTY monitoring, process tracking, event system, WASM adapter runtime/interface, LLM coordination loop, JSON-RPC control plane, and CLI plugin system, plus the conformance harness/suite.
- **Out-of-scope items (explicit):** As listed in Spec §1.4 (CLI presentation details; external LLM provider integration; agent-specific prompting; general authn/z beyond NATS host authentication and local plugin permissions; general persistent storage beyond JetStream and the specified SQLite/sqlite-vec usage; network protocols beyond SSH bootstrap).
- **Hard constraints (mandatory):**
  - Core MUST be implemented in Go 1.25.6 (`go1.25.6`) (Spec §4.2.1)
  - WASM runtime is wazero; adapters compiled with TinyGo (Spec §4.2.2)
  - HSM via stateforward/hsm-go; IDs via stateforward/hsm-go/muid (Spec §4.2.3)
  - PTY management via creack/pty (Spec §4.2.4)
  - Explicit error handling and wrapping conventions (Spec §4.2.5)
  - Project structure and agent-agnostic core invariant (Spec §1.5.1, §4.2.6)
  - No built-in adapters in core; any default adapter must be a WASM asset loaded via discovery (Spec §1.5.4)
  - OpenTelemetry instrumentation (Spec §4.2.9)
  - Local inference engine `liquidgen` for features requiring local inference (Spec §4.2.10)
- **Quality bars / acceptance criteria (global):**
  - Conformance suite includes required E2E flows (Spec §4.3.2) and passes for implemented functionality (Spec §4.3.2 item 6).
  - Cross-compilation works per spec guidance (Spec §4.2.7, §8.3.9) and hook libraries build for supported platforms (Spec §4.2.7, §8.3).
  - Internal core remains agent-agnostic with no imports/references to adapters (Spec §1.5.1, §4.2.6).

---

## Phase 0: Prerequisites & setup

### Objective
Establish repository structure, build/toolchain, configuration, observability scaffolding, and conformance harness skeleton.

### Inputs
- spec-v1.22.md (authoritative)
- Go 1.25.6 (`go1.25.6`) toolchain, TinyGo, platform toolchains for hook libraries
- NATS + JetStream for remote features (deployment dependency)

### Outputs
- Go module and directory structure
- Config loader + config actor skeleton
- OTel scaffolding
- Conformance harness scaffolding

### TODO list
- [ ] Create repository layout and packages exactly per spec project structure
  - Spec reference(s): §4.2.6, §1.5.1
  - Acceptance criteria: `cmd/amux` and `cmd/amux-node` exist; `internal/*` packages compile; `internal/*` has no imports from `adapters/*` (enforced by lint or test).

- [ ] Pin core language/runtime dependencies (Go 1.25.6 (`go1.25.6`), wazero, stateforward/hsm-go (including /muid), creack/pty)
  - Spec reference(s): §4.2.1–§4.2.4
  - Acceptance criteria: `go.mod` pins compatible versions; minimal “smoke” programs demonstrate wazero instantiation, hsm-go event dispatch, and PTY allocation on supported OS targets.

- [ ] Implement error handling conventions and sentinel error strategy
  - Spec reference(s): §4.2.5
  - Acceptance criteria: package templates use `fmt.Errorf("context: %w", err)`; sentinel errors defined; no ignored errors in core paths (enforced by staticcheck configuration).

- [x] Implement configuration subsystem: format, hierarchy, env mapping, parsing conventions
  - Spec reference(s): §4.2.8.1–§4.2.8.4, §4.2.8.10
  - Acceptance criteria: config loads from default paths and supports overrides; env var mapping verified in unit tests; duration/bytes/bool parsing matches spec conventions.

- [ ] Implement adapter configuration handling and sensitive configuration support (opaque adapter config blocks)
  - Spec reference(s): §4.2.8.5–§4.2.8.6
  - Acceptance criteria: adapters receive only their scoped config; sensitive fields are redacted in logs and debug outputs per spec requirements.

- [x] Implement configuration actor, live updates, and subscriptions
  - Spec reference(s): §4.2.8.7–§4.2.8.9
  - Acceptance criteria: components can subscribe to config changes; hot reload updates dependent subsystems without restart in an integration test.

- [x] Implement OpenTelemetry scaffolding (traces/metrics/logs (OTel))
  - Spec reference(s): §4.2.9.1–§4.2.9.5
  - Acceptance criteria: OTel can be enabled by env vars or config; spans follow naming convention; required baseline metrics are emitted in a test exporter.

- [x] Scaffold `liquidgen` local inference integration interface and configuration
  - Spec reference(s): §4.2.10
  - Acceptance criteria: interface types compile; unknown model IDs return error; mapping of logical IDs to artifacts is observable via telemetry fields/events.

- [x] The implementation MUST integrate the pre-existing `liquidgen` inference engine from `third_party/liquidgen` (local) as a dependency and MUST wire it to the local inference interface
  - Spec reference(s): §4.2.10
  - Acceptance criteria: `go test ./...` succeeds on a development machine; the default build uses `liquidgen` behind the Phase 0 interface (no new inference engine implementation is introduced); build and runtime logs MUST include the `liquidgen` module version or commit identifier for traceability.

- [x] Create conformance harness skeleton and test runner wiring
  - Spec reference(s): §4.3.1
  - Acceptance criteria: `go test` can run a placeholder conformance suite that boots a daemon + CLI client fixture and records structured JSON results per the “Conformance harness output contract (minimum)” section above.

---
- [x] Ensure `spec-v1.22.md` is present and version-locked for this plan
  - Spec reference(s): §4.3.1, §4.2.6
  - Acceptance criteria: `spec-v1.22.md` exists in-repo; a guard test or startup check fails fast with a clear error if the file is missing or the expected version marker does not match.

- [x] Implement shared path resolver (`internal/paths`) and repo-scoped `.amux/` invariants
  - Spec reference(s): §4.2.6, §4.2.8
  - Acceptance criteria: a single package resolves all filesystem paths from config/env and `repo_root`; worktree paths use the resolver and still satisfy `.amux/worktrees/{agent_slug}/`; unit tests lock the invariants.

- [ ] Create reproducible verification entrypoints (Makefile or `scripts/*`) and CI-friendly “one command” verification
  - Spec reference(s): §4.3.1–§4.3.2, §4.2.5
  - Acceptance criteria: documented commands exist for unit, lint, integration, and conformance runs; commands return non-zero on failure; they are suitable for CI automation.

- [x] Implement `amux test` CLI subcommand (snapshot + `--regression`) and wire it into verification entrypoints
  - Spec reference(s): §12.6–§12.6.5
  - Acceptance criteria: `amux test` runs the required command sequence and writes a TOML snapshot to `<module_root>/snapshots/`; `amux test --regression` compares to the previous snapshot and exits non-zero on regressions; `--no-snapshot` writes the snapshot to stdout and writes all human-readable logs/regression reports to stderr.

- [ ] Run `amux test` to capture the baseline snapshot for Phase 0
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 0 regression checking.

- [ ] Introduce stable interfaces + noop implementations to unblock phased work (event dispatch + adapter hooks)
  - Spec reference(s): §1.5.1, §4.2.6, §9.1, §10.4
  - Acceptance criteria: Phase 4–6 code can emit/subscribe to events via the interface in local/noop mode; Phase 5 can call a pattern/action interface that returns no matches by default; the repo compiles/tests without implementing Phase 7/Phase 8 yet.

- [ ] The implementation MUST maintain inline Go documentation and MUST generate per-package `README.md` files via `go-docmd`, enforced by automated docs-check
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: all packages and exported identifiers implemented in Phase 0 MUST have `go doc`-suitable comments; `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` MUST run successfully at the module root and MUST update per-package `README.md` files in place; generated `README.md` files MUST be committed; a CI job or verification entrypoint MUST run the canonical command and MUST fail if it produces any uncommitted changes.

- [ ] Run `amux test --regression` at the end of Phase 0 to verify no regressions relative to the Phase 0 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 0 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 0, remove unused code/scripts, and commit Phase 0 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 0 TODOs are updated; `git status` is clean; the Phase 0 baseline + latest snapshots are retained; a Phase 0 commit exists in git history.



## Phase 1: Core domain model, IDs, and state machines

### Objective
Implement authoritative types, identifiers, and HSM-driven lifecycle and presence state machines.

### Inputs
- Config subsystem (Phase 0)
- stateforward/hsm-go and /muid

### Outputs
- `pkg/api` public types
- `internal/agent` actor model and HSMs
- Stable identifiers and normalization utilities

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 1
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 1 regression checking.

- [ ] Implement identifiers and normalization rules (agent_id, peer_id, host_id, agent_slug, repo_root canonicalization)
  - Spec reference(s): §3 (Definitions), §4.2.3, §5.3.1, §3.23
  - Acceptance criteria: ID encoding/decoding matches spec (base-10 strings where required); `agent_slug` normalization matches spec examples; `repo_root` canonicalization passes unit tests (including `~/` expansion semantics for remote).

- [ ] Implement Agent and Session core data structures and invariants
  - Spec reference(s): §5.1, §5.5.9, §4.2.3
  - Acceptance criteria: structures contain all required fields; invariants enforced via constructors and validation tests.

- [ ] Implement Agent lifecycle HSM (Pending → Starting → Running → Terminated/Errored) and dispatch integration
  - Spec reference(s): §4.2.3, §5.4
  - Acceptance criteria: lifecycle transitions only via defined events; tests cover normal and error paths; transitions emit required events.

- [ ] Implement Presence HSM (Online ↔ Busy ↔ Offline ↔ Away) and transition triggers
  - Spec reference(s): §4.2.3, §6.1, §6.5
  - Acceptance criteria: presence transitions follow spec rules; PTY and process events can trigger presence changes through `hsm.Dispatch()`.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 1 to verify no regressions relative to the Phase 1 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 1 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 1, remove unused code/scripts, and commit Phase 1 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 1 TODOs are updated; `git status` is clean; the Phase 1 baseline + latest snapshots are retained; a Phase 1 commit exists in git history.
---

## Phase 2: Local agent management (repo/worktree), lifecycle operations, and merge strategies

### Objective
Support adding/removing local agents, worktree isolation, local spawn/attach, graceful shutdown, and git merge strategies.

### Inputs
- Core model + HSMs (Phase 1)
- Config and filesystem access

### Outputs
- Local agent directory/worktree management
- Local session spawning and PTY ownership
- Git strategy implementation hooks

### TODO list
- [x] Run `amux test` to capture the baseline snapshot for Phase 2
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 2 regression checking.

- [x] Implement agent add flow (validation, repo required, config persistence)
  - Spec reference(s): §1.3, §5.2, §5.1
  - Acceptance criteria: adding an agent outside a git repo fails; adding within a repo creates agent entry with required fields; `amux agent add` uses same validation as daemon API.

- [x] Implement worktree isolation, slug-based path layout, and normalization rules
  - Spec reference(s): §5.3, §5.3.1, §5.3.4
  - Acceptance criteria: worktrees created under `.amux/worktrees/{agent_slug}/`; idempotent reuse behavior; cleanup on remove as specified; tests validate behavior on existing worktrees.

- [x] Implement local agent lifecycle operations: spawn/start, attach, stop/kill, restart semantics
  - Spec reference(s): §5.4, §5.6
  - Acceptance criteria: lifecycle HSM transitions align to operations; PTY processes are started in correct workdir; shutdown drains and closes resources per graceful shutdown requirements.

- [x] Implement local PTY session ownership model (owned PTY)
  - Spec reference(s): §7 (monitor relies on owned PTY), §B.5
  - Acceptance criteria: amux owns PTY for agent; monitor can observe raw output; no dependency on external terminal multiplexers unless implemented as optional backend.

- [x] Implement git merge strategy selection and defaults (base_branch/target_branch)
  - Spec reference(s): §5.7, §5.7.1
  - Acceptance criteria: selected strategy produces expected git operations in dry-run tests; defaulting rules validated against config examples.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [x] Run `amux test --regression` at the end of Phase 2 to verify no regressions relative to the Phase 2 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 2 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [x] Update this plan’s TODOs for Phase 2, remove unused code/scripts, and commit Phase 2 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 2 TODOs are updated; `git status` is clean; the Phase 2 baseline + latest snapshots are retained; a Phase 2 commit exists in git history.
---

## Phase 3: Remote agents (SSH bootstrap, NATS + JetStream runtime orchestration)

### Objective
Implement remote host manager/director roles, NATS subjects, handshake, request-reply control plane, and remote session I/O with replay and reconnection semantics.

### Inputs
- NATS + JetStream deployment
- Core IDs and lifecycle HSMs
- Config for remote

### Outputs
- Director side remote orchestration
- Manager-role daemon behaviors for remote sessions
- Robust reconnection with buffering and replay ordering

### TODO list
- [x] Run `amux test` to capture the baseline snapshot for Phase 3
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 3 regression checking.

- [x] Implement SSH bootstrap for remote hosts (daemon install/config, per-host NATS creds provisioning)
  - Spec reference(s): §5.5.2, §5.5.3, §5.5.6.4
  - Acceptance criteria: director provisions a unique credential per `host_id`; credential copied to `remote.nats.creds_path` with permissions <= `0600`; remote daemon starts and connects using the credential.

- [x] Implement hub NATS server configuration (director role) and manager leaf server configuration
  - Spec reference(s): §5.5.6.1, §5.5.6.2, §5.5.7
  - Acceptance criteria: hub accepts leaf connections; `remote.nats.subject_prefix` is honored for all subjects; integration test validates a director and manager-role daemon can connect and exchange handshake.

- [x] Implement JetStream KV bucket provisioning and required durable state keys
  - Spec reference(s): §5.5.6.3
  - Acceptance criteria: director creates KV bucket (default `AMUX_KV`) if missing; `hosts/<host_id>/info`, `hosts/<host_id>/heartbeat`, and `sessions/<host_id>/<session_id>` keys are written and read back as UTF-8 JSON; reconnect uses KV session metadata for recovery tests.

- [x] Implement NATS authentication and per-host subject authorization rules
  - Spec reference(s): §5.5.6.4
  - Acceptance criteria: for a given `host_id`, publish/subscribe permissions are restricted to the exact subject sets in the spec; unauthorized publish/subscribe attempts are denied by the NATS server in tests; credentials are unique per `host_id` and never reused.

- [x] Implement NATS subject namespaces and message envelopes for remote protocol
  - Spec reference(s): §5.5.7.1, §5.5.7.5, §9.1.3
  - Acceptance criteria: subject strings match spec; director subscribes/publishes correct subjects; message schema validated via JSON fixtures.

- [x] Implement request-reply control operations (spawn/kill/replay) with timeout and fail-fast semantics
  - Spec reference(s): §5.5.7.2, §5.5.7.2.1
  - Acceptance criteria: director uses `remote.request_timeout`; disconnected hosts are rejected without sending NATS requests; `not_ready` errors block retries until `connection.established` observed.

- [x] Implement handshake exchange and readiness gating
  - Spec reference(s): §5.5.7.3
  - Acceptance criteria: daemon sends handshake request on connect; director validates host_id; daemon rejects pre-handshake spawn/kill/replay with `error` `code="not_ready"`; collision handling works.

- [x] Implement remote spawn idempotency by agent_id and session_conflict behavior
  - Spec reference(s): §5.5.7.3 (spawn)
  - Acceptance criteria: second spawn for same agent_id returns existing session_id; conflicting repo_path or agent_slug returns `session_conflict`.

- [x] Implement PTY I/O subjects and payload chunking behavior
  - Spec reference(s): §5.5.7.4
  - Acceptance criteria: PTY out published to `P.pty.<host_id>.<session_id>.out`; PTY in subscribed on `.in`; chunking never exceeds configured NATS max payload.

- [x] Implement per-session replay buffer (ring buffer) with ordering and live-output gating
  - Spec reference(s): §5.5.7.3 (Replay buffer and ordering)
  - Acceptance criteria: buffer capped at `remote.buffer_size`; disabled when 0; replay publishes snapshot oldest-to-newest; replay bytes always published before live bytes after replay request.

- [x] Implement connection recovery, hub disconnection buffering, and replay-before-live semantics after reconnect
  - Spec reference(s): §5.5.8
  - Acceptance criteria: during hub disconnection PTY output continues and is retained for replay; cross-host pubs buffer up to `remote.buffer_size` total, drop-oldest, FIFO per subject; after reconnect, no live PTY out subject publishes until replay request handled.

- [x] Implement remote session manager behaviors and director-visible exit events
  - Spec reference(s): §5.5.9
  - Acceptance criteria: unexpected agent exit emits event; optional remediation actions (if enabled) emit events describing actions taken.


- [x] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [x] Run `amux test --regression` at the end of Phase 3 to verify no regressions relative to the Phase 3 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 3 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [x] Update this plan's TODOs for Phase 3, remove unused code/scripts, and commit Phase 3 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 3 TODOs are updated; `git status` is clean; the Phase 3 baseline + latest snapshots are retained; a Phase 3 commit exists in git history.
---

## Phase 4: Presence and roster

### Objective
Provide presence state model, roster listing, presence awareness, and inter-agent messaging routes.

### Inputs
- Presence HSM (Phase 1)
- Event dispatch interface + local/noop implementation (introduced in Phase 0; Phase 7 finalizes hsmnet + remote distribution)

### Outputs
- Roster store and query APIs
- Presence transitions and derived status

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 4
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 4 regression checking.

- [ ] Implement presence states and transitions, including Away semantics for remote disconnection
  - Spec reference(s): §6.1, §6.5, §5.5.8
  - Acceptance criteria: state machine matches allowed transitions; remote disconnect moves to Away; reconnect and replay moves Away → Running and presence updates accordingly.

- [ ] Implement roster data model and listing outputs
  - Spec reference(s): §6.2
  - Acceptance criteria: roster includes all required fields; ordering and filtering match spec; CLI and JSON-RPC surfaces expose roster entries.

- [ ] Implement presence awareness and subscriptions
  - Spec reference(s): §6.3
  - Acceptance criteria: components can subscribe to roster/presence changes; updates delivered reliably in tests.

- [ ] Implement inter-agent messaging routes
  - Spec reference(s): §6.4
  - Acceptance criteria: messages are addressed and delivered per spec; notification gating and batching integrate with process notification pipelines where applicable.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 4 to verify no regressions relative to the Phase 4 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 4 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 4, remove unused code/scripts, and commit Phase 4 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 4 TODOs are updated; `git status` is clean; the Phase 4 baseline + latest snapshots are retained; a Phase 4 commit exists in git history.
---

## Phase 5: PTY management and PTY monitoring

### Objective
Own PTYs for agents, monitor output for activity and patterns, and decode TUI screens for snapshotting.

### Inputs
- creack/pty integration (Phase 0)
- Adapter pattern/action interfaces + noop implementation (introduced in Phase 0; Phase 8 provides the WASM-backed runtime)

### Outputs
- `internal/pty` PTY lifecycle
- `internal/monitor` PTY observation pipeline
- `internal/tui` decoding and XML encoding

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 5
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 5 regression checking.

- [ ] Implement PTY creation, window sizing, and lifecycle management
  - Spec reference(s): §4.2.4
  - Acceptance criteria: PTY can be created and attached; size changes apply; PTY closes cleanly on shutdown.

- [ ] Implement PTY monitor responsibilities and event emission pipeline
  - Spec reference(s): §7.2, §9.3
  - Acceptance criteria: monitor reads PTY output continuously; emits activity and pattern match events; integrates with HSM dispatch.

- [ ] Implement activity detection, pattern matching hooks, and timeout configuration
  - Spec reference(s): §7.3, §7.4, §7.5
  - Acceptance criteria: activity triggers are detected; pattern matching delegates to adapter-provided patterns; timeouts are configurable and unit tested.

- [ ] Implement presence inference from PTY signals
  - Spec reference(s): §7.6, §6.5
  - Acceptance criteria: presence inference rules map observed activity/idle to presence changes in tests.

- [ ] Implement TUI decoding and XML capture format for snapshots
  - Spec reference(s): §7.7, §11.2.1
  - Acceptance criteria: full-screen TUIs decode to screen model; XML output matches schema expectations; conformance suite “menu flows” can verify decoding.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 5 to verify no regressions relative to the Phase 5 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 5 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 5, remove unused code/scripts, and commit Phase 5 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 5 TODOs are updated; `git status` is clean; the Phase 5 baseline + latest snapshots are retained; a Phase 5 commit exists in git history.
---

## Phase 6: Process tracking, interception, and notifications

### Objective
Track spawned processes, intercept exec where required, emit process and I/O events, and implement batching, LLM-gated notifications, and MCP subscriptions.

### Inputs
- Hook library build pipeline (platform-specific)
- Event dispatch interface + shared event types (introduced in Phase 0; Phase 7 provides network routing)
- liquidgen engine (Phase 0)

### Outputs
- `internal/process` tracker and subscriptions
- Hook shared libraries and protocol
- Notification gating pipeline + storage for subscriptions

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 6
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 6 regression checking.

- [ ] Implement process model, process tree tracking, and events
  - Spec reference(s): §8.2, §8.6, §8.4
  - Acceptance criteria: process start/exit events emitted; parent-child tree maintained; tests cover common shells and subprocess patterns.

- [ ] Implement process interception via hook libraries (LD_PRELOAD/DYLD_INSERT_LIBRARIES) and fallback polling as specified
  - Spec reference(s): §8.3, §B.8, §B.9
  - Acceptance criteria: hook builds on supported platforms; intercepted exec events reach tracker; polling fallback works when hooks unavailable.

- [ ] Implement hook library compilation pipeline and artifact packaging
  - Spec reference(s): §8.3.9, §4.2.6, §4.2.7
  - Acceptance criteria: per-platform c-shared hook libraries are built (Linux/macOS; amd64/arm64) and placed under `hooks/bin/`; the main binaries can locate/load the correct hook artifact; build verification is covered by an automated build script.

- [ ] Implement hook protocol: wire framing, FD passing, and I/O pipe architecture
  - Spec reference(s): §8.3.2, §8.3.2.1, §8.3.3, §8.3.4
  - Acceptance criteria: hook-to-tracker protocol frames are parsed correctly; SCM_RIGHTS FD passing works in integration tests; stdin/stdout/stderr pipes are attributed to the correct process and agent.

- [ ] Implement environment inheritance, I/O attribution, and process-tree tracking semantics
  - Spec reference(s): §8.3.5, §8.3.7, §8.3.8
  - Acceptance criteria: child processes inherit required env vars; I/O attribution rules match the spec; process tree remains accurate across nested execs.

- [ ] Implement subscriptions and filters for process events and I/O streams
  - Spec reference(s): §8.5, §8.4.2
  - Acceptance criteria: subscriber receives only requested events; backpressure handled per spec batching strategy.

- [ ] Implement batching/coalescing pipeline and batcher HSM
  - Spec reference(s): §8.4.3.1–§8.4.3.5
  - Acceptance criteria: batch sizes and timing follow spec; coalescing rules match; deterministic tests for coalescing.

- [ ] Implement LLM-gated notifications with liquidgen and observability for throughput
  - Spec reference(s): §8.4.3.6, §4.2.10, §4.2.9
  - Acceptance criteria: gating prompt uses required model IDs; token streaming supported; tokens/sec and queue latency metrics emitted; failure modes return errors and fall back per spec.

- [ ] Implement Notification MCP server transport, framing, and concurrency
  - Spec reference(s): §8.4.3.7
  - Acceptance criteria: when `events.subscriptions.enabled = true`, the director listens on a Unix domain stream socket at `events.subscriptions.socket_path`; transport uses newline-delimited UTF-8 JSON-RPC 2.0 (one JSON object per line); multiple concurrent clients are supported; server-to-client notifications are delivered as JSON-RPC notifications.

- [ ] Implement embedding runtime and model asset packaging for semantic subscriptions
  - Spec reference(s): §8.4.3.7, §4.2.6
  - Acceptance criteria: supported ONNX embedding models can be loaded from `models/` (default all-MiniLM-L6-v2 or implementation-chosen default among the spec list); required ONNX Runtime shared libraries are packaged under `models/onnxruntime/` and load successfully on supported platforms; embeddings are computed on CPU; resolved model choice and embedding latency are observable via OTel metrics/logs.

- [ ] Implement MCP notification subscriptions storage using SQLite + sqlite-vec embeddings
  - Spec reference(s): §8.4.3.7
  - Acceptance criteria: sqlite + sqlite-vec are available and the vector extension loads in supported environments; embeddings are stored and queried via vector similarity search; subscription matching returns stable results; migrations and schema tests exist; performance meets spec expectations for typical subscription sets.

- [ ] Implement stream capture modes and dispatch behavior
  - Spec reference(s): §8.4.4–§8.4.5
  - Acceptance criteria: capture modes selectable by config; events dispatched correctly; conformance suite “notification flows” pass.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 6 to verify no regressions relative to the Phase 6 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 6 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 6, remove unused code/scripts, and commit Phase 6 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 6 TODOs are updated; `git status` is clean; the Phase 6 baseline + latest snapshots are retained; a Phase 6 commit exists in git history.
---

## Phase 7: Event system (hsmnet, local and remote dispatch)

### Objective
Implement event types, dispatch, handlers, deferral, and network-aware routing across local and remote nodes.

### Inputs
- NATS connectivity for remote (Phase 3)
- HSM event dispatch usage from other phases

### Outputs
- `internal/event` (or `internal/protocol`) dispatcher with hsmnet
- Event envelopes and routing over NATS subjects

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 7
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 7 regression checking.

- [ ] Implement hsmnet network-aware dispatch foundation
  - Spec reference(s): §9.1
  - Acceptance criteria: local and remote event buses can be created; events route with correct scoping; unit tests cover local-only mode.

- [ ] Implement hsmnet peer connection lifecycle and ID routing
  - Spec reference(s): §9.1.4, §9.1.5
  - Acceptance criteria: peers are tracked by `peer_id`; unicast/multicast/broadcast routes to the correct local or remote targets; connection lost/recovered updates routing tables.

- [ ] Implement hsmnet wire format and required host event payload schemas
  - Spec reference(s): §9.1.3, §9.1.3.1, §9.1.3.2, §5.5.7.5
  - Acceptance criteria: EventMessage envelopes serialize/deserialize; scalar encodings use base-10 ID strings; `connection.*` and `process.*` event payloads conform to the required schemas; `data_b64` is base64 for I/O events; unknown events are treated as opaque JSON.

- [ ] Implement event dispatch pipeline, event types, and event structure schemas
  - Spec reference(s): §9.2–§9.4
  - Acceptance criteria: events serialize/deserialize; required fields present; schema fixtures validated.

- [ ] Implement event handlers, deferral, and task model integration
  - Spec reference(s): §9.5–§9.7
  - Acceptance criteria: deferred events resume; tasks model can run and complete; tests cover handler ordering and retries.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 7 to verify no regressions relative to the Phase 7 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 7 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 7, remove unused code/scripts, and commit Phase 7 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 7 TODOs are updated; `git status` is clean; the Phase 7 baseline + latest snapshots are retained; a Phase 7 commit exists in git history.
---

## Phase 8: Adapter interface (WASM runtime, discovery, and packaging)

### Objective
Load adapters as WASM modules, expose host functions, and integrate adapters for pattern matching and actions.

### Inputs
- wazero runtime (Phase 0)
- Config actor (Phase 0)
- Event system types (Phase 7)

### Outputs
- `internal/adapter` runtime
- Adapter manifest parser and validation
- Adapter discovery and install flows

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 8
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 8 regression checking.

- [ ] Implement adapter manifest parsing and validation
  - Spec reference(s): §10.2
  - Acceptance criteria: invalid manifests rejected with actionable errors; required fields enforced; tests cover manifest examples.

- [ ] Implement CLI version pinning logic for adapters
  - Spec reference(s): §10.3
  - Acceptance criteria: incompatible CLI versions fail with clear error; pinning rules unit tested.

- [ ] Implement WASM interface (host functions, memory, ABI) and TinyGo adapter build expectations
  - Spec reference(s): §10.4, §4.2.2
  - Acceptance criteria: sample TinyGo adapter can be loaded; host calls succeed; call spans/metrics emitted per OTel requirements.

- [ ] Add adapter ABI contract tests and fixtures (memory management, return packing, error reporting)
  - Spec reference(s): §10.4.1, §10.4.2, §4.3.3
  - Acceptance criteria: automated tests validate `memory` export presence, `amux_alloc/amux_free` semantics, packed `(ptr,len)` decoding, and `amux_last_error` behavior; failing adapters are rejected with clear diagnostics.

- [ ] Create a “new adapter bring-up” checklist and conformance fixture proving core remains agent-agnostic
  - Spec reference(s): §1.5.2, §1.5.3, §1.5.4, §10.6, §10.8, §4.3.3
  - Acceptance criteria: a newly added adapter (new folder under `adapters/`) is built to WASM, installed via standard discovery, and passes the adapter conformance fixture without any changes to `internal/` packages.

- [ ] Implement event and action types exchanged with adapters
  - Spec reference(s): §10.5, §7.4, §11.3
  - Acceptance criteria: adapter receives PTY/process snapshots; returns actions; actions executed by core; conformance fixtures validate.

- [ ] Implement adapter discovery, packaging, installation, and setup flows
  - Spec reference(s): §10.6, §10.8, §1.5.4
  - Acceptance criteria: adapters installed into registry layout; embedded default adapter (if used) is shipped as WASM asset, not Go code; discovery works across configured paths.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 8 to verify no regressions relative to the Phase 8 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 8 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 8, remove unused code/scripts, and commit Phase 8 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 8 TODOs are updated; `git status` is clean; the Phase 8 baseline + latest snapshots are retained; a Phase 8 commit exists in git history.
---

## Phase 9: LLM coordination loop

### Objective
Run the observation loop, produce snapshots, and perform coordination actions using adapters and liquidgen.

### Inputs
- PTY/TUI snapshots (Phase 5)
- Process and notification events (Phase 6)
- Adapter actions (Phase 8)
- liquidgen engine (Phase 0)

### Outputs
- `internal/coordination` (or equivalent) implementing §11
- Snapshot generation and action execution

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 9
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 9 regression checking.

- [ ] Implement observation loop orchestration and scheduling
  - Spec reference(s): §11.1
  - Acceptance criteria: loop runs at configured cadence; backpressure handled; loop emits spans/metrics.

- [ ] Implement snapshot format and serialization, including TUI XML capture
  - Spec reference(s): §11.2, §11.2.1
  - Acceptance criteria: snapshots include required fields; TUI capture included when enabled; fixtures validated.

- [ ] Implement coordination actions and tool invocation paths
  - Spec reference(s): §11.3, §11.6
  - Acceptance criteria: actions map to concrete operations (input injection, commands, messaging); errors are surfaced; actions audited via logs/metrics.

- [ ] Implement configuration surfaces and manual mode for coordination
  - Spec reference(s): §11.4–§11.5
  - Acceptance criteria: manual mode disables automated actions; configuration changes take effect via config actor without restart.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 9 to verify no regressions relative to the Phase 9 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 9 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 9, remove unused code/scripts, and commit Phase 9 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 9 TODOs are updated; `git status` is clean; the Phase 9 baseline + latest snapshots are retained; a Phase 9 commit exists in git history.
---

## Phase 10: CLI client and daemon (JSON-RPC control plane)

### Objective
Expose a JSON-RPC control plane for CLI clients and plugins, including required methods and permission enforcement.

### Inputs
- Core agent management (Phases 2–4)
- Event system (Phase 7)
- Plugin system (Phase 11)

### Outputs
- Daemon JSON-RPC server and client
- Stable request/response types and streaming subscription delivery

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 10
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 10 regression checking.

- [ ] Implement transport and framing for JSON-RPC between CLI and daemon
  - Spec reference(s): §12.2
  - Acceptance criteria: client can connect; request/response framing correct; malformed frames rejected safely.

- [ ] Implement daemon lifecycle and turnkey defaults
  - Spec reference(s): §12.3
  - Acceptance criteria: daemon starts with sensible defaults; auto-start behavior matches spec; state persisted as required.

- [ ] Implement required JSON-RPC methods: ping, version, events.subscribe, agent.add/list/remove, system.update
  - Spec reference(s): §12.4.1–§12.4.7
  - Acceptance criteria: methods conform to schemas; event subscription delivers events; system.update handles update sources per spec.

- [ ] Implement permissions context for CLI plugins and enforce gating
  - Spec reference(s): §12.5, §13.6
  - Acceptance criteria: plugin calls are restricted by declared permissions; denied calls return explicit errors; tests cover least-privilege defaults.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 10 to verify no regressions relative to the Phase 10 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 10 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 10, remove unused code/scripts, and commit Phase 10 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 10 TODOs are updated; `git status` is clean; the Phase 10 baseline + latest snapshots are retained; a Phase 10 commit exists in git history.
---

## Phase 11: CLI plugin system (WASM and remote)

### Objective
Implement plugin management commands, plugin registry and installation sources, and built-in plugins.

### Inputs
- JSON-RPC control plane (Phase 10)
- wazero runtime (Phase 0)

### Outputs
- Plugin manager
- Plugin manifest and permissions model
- Built-in plugins defined by the spec

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 11
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 11 regression checking.

- [ ] Implement plugin management commands and UX-independent surfaces
  - Spec reference(s): §13.2
  - Acceptance criteria: install/list/remove/enable/disable commands work; outputs match required fields; CLI presentation is flexible per non-goals.

- [ ] Implement plugin manifest (plugin.toml) parsing and validation
  - Spec reference(s): §13.3
  - Acceptance criteria: manifests validated; permissions declared and enforced; unit tests cover examples.

- [ ] Implement plugin installation sources (local, registry, git, HTTP(S))
  - Spec reference(s): §13.4, §12.4.7
  - Acceptance criteria: installs succeed from each supported source type; integrity checks and version selection behave as specified.

- [ ] Implement plugin registry layout and resolution
  - Spec reference(s): §13.5
  - Acceptance criteria: registry directory structure matches spec; plugin lookup order deterministic and tested.

- [ ] Implement daemon access mediation and local permission gating for plugins
  - Spec reference(s): §13.6, §12.5
  - Acceptance criteria: plugins operate with least privilege; sensitive operations require explicit permissions; audit logs show permission grants/denials.

- [ ] Implement required built-in plugins
  - Spec reference(s): §13.7–§13.9
  - Acceptance criteria: built-in plugin behaviors match spec; `amux agent` and `amux chat` plugin commands operate end to end using JSON-RPC.


- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 11 to verify no regressions relative to the Phase 11 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 11 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 11, remove unused code/scripts, and commit Phase 11 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 11 TODOs are updated; `git status` is clean; the Phase 11 baseline + latest snapshots are retained; a Phase 11 commit exists in git history.
---

## Phase 12: Conformance suite completion and release readiness

### Objective
Complete the conformance suite, validate required E2E flows, and ensure cross-platform build/release artifacts are correct.

### Inputs
- Implemented system components from Phases 0–11

### Outputs
- Conformance suite with required coverage
- Release build artifacts and verification docs

### TODO list
- [ ] Run `amux test` to capture the baseline snapshot for Phase 12
  - Spec reference(s): §12.6.1–§12.6.3
  - Acceptance criteria: a new `snapshots/amux-test-*.toml` exists under `<module_root>/snapshots/` and is retained as the baseline for Phase 12 regression checking.

- [ ] Implement required conformance E2E flows (auth, menu, status, notification, control plane, plus all implemented MUST/MUST NOT behaviors)
  - Spec reference(s): §4.3.2, §7.7, §5.4, §6.5, §5.5.8, §8.4.3.6–§8.4.3.7, §12–§13
  - Acceptance criteria: suite includes at least the 5 enumerated flow categories; each flow is reproducible and outputs structured results; failures include actionable diagnostics.

- [ ] Implement adapter conformance fixtures and example adapter tests
  - Spec reference(s): §4.3.3, §10.7
  - Acceptance criteria: fixtures validate adapter ABI; example adapter passes; Automation/CI MUST run these fixtures on each change via the repository’s verification entrypoints (see “Verification entrypoints (required)”).

- [ ] Implement remote conformance runs and multi-host scenarios
  - Spec reference(s): §4.3.4, §5.5
  - Acceptance criteria: suite can run against remote daemons; reconnect/replay ordering semantics verified; buffering drop policy verified under injected disconnects.

- [ ] Verify cross-compilation and hook build commands on supported platforms
  - Spec reference(s): §4.2.7, §8.3.9
  - Acceptance criteria: builds succeed for linux/darwin amd64/arm64; hook libraries compile per platform; artifact naming and packaging documented.


- [ ] The repository SHOULD include GitHub Actions workflows that build release artifacts and publish them to Cloudflare R2 for curl and PowerShell installation
  - Spec reference(s): N/A (release engineering requirement)
  - Acceptance criteria: a workflow triggers on version tags and builds `amux` and `amux-node` for linux/darwin amd64/arm64; artifacts are packaged with a deterministic naming convention and SHA256 checksums; the workflow uploads artifacts and checksums to an R2 bucket via the S3 API; the workflow configures access to the private `gabewillen/liquidgen` module (for example via `GOPRIVATE` and a GitHub token or deploy key) so release builds do not depend on a developer machine.

- [ ] The release workflow SHOULD publish install scripts to R2 so users can install and upgrade via `curl` and PowerShell
  - Spec reference(s): N/A (release engineering requirement)
  - Acceptance criteria: `install.sh` and `install.ps1` are uploaded to R2 and documented; scripts install the latest version by default and MUST support pinning a specific version via an environment variable; scripts MUST be idempotent so re-running them upgrades an existing installation; scripts MUST verify downloaded artifacts using the published SHA256 checksums.

- [ ] The repository SHOULD add CI smoke tests that validate the R2 install and upgrade path on supported platforms
  - Spec reference(s): N/A (release engineering requirement)
  - Acceptance criteria: a CI job runs `install.sh` on linux and macOS runners and runs `install.ps1` via PowerShell; each job verifies `amux --version` matches the expected tag; each job re-runs the installer to confirm the upgrade path remains functional.

- [ ] The implementation MUST maintain inline Go documentation and MUST regenerate per-package `README.md` files via `go-docmd`
  - Spec reference(s): §4.2.6.1
  - Acceptance criteria: every package and exported identifier added or modified in this phase MUST include `go doc`-suitable comments; running `go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...` at the module root MUST produce no uncommitted changes; generated per-package `README.md` files MUST be committed.

- [ ] Run `amux test --regression` at the end of Phase 12 to verify no regressions relative to the Phase 12 baseline snapshot
  - Spec reference(s): §12.6.5
  - Acceptance criteria: `amux test --regression` exits 0; any regressions are fixed before Phase 12 is considered complete; the new snapshot is written to `<module_root>/snapshots/`.

- [ ] Update this plan’s TODOs for Phase 12, remove unused code/scripts, and commit Phase 12 to git
  - Spec reference(s): N/A (plan process requirement)
  - Acceptance criteria: Phase 12 TODOs are updated; `git status` is clean; the Phase 12 baseline + latest snapshots are retained; a Phase 12 commit exists in git history.
---

## Cross-cutting concerns

### Testing
- Unit tests per package for parsing/validation logic (config, manifests, schemas) and HSM transitions.
- Integration tests for PTY + monitor + adapter pattern matching.
- Conformance suite for E2E flows and normative behaviors (Spec §4.3).

### Observability
- OpenTelemetry across lifecycle, monitor, process tracker, adapter calls, remote protocol operations (Spec §4.2.9).
- Emit tokens/sec and queue latency metrics for liquidgen workloads (Spec §4.2.10).

### Performance
- Replay buffer and hub buffering bounded by `remote.buffer_size` and must drop-oldest (Spec §5.5.7.3, §5.5.8).
- Chunking for NATS payload limits and PTY throughput (Spec §5.5.7.4).

### Security
- Plugin permission enforcement and least privilege (Spec §12.5, §13.6).
- Respect sensitive config redaction (Spec §4.2.8.6).

---

## Explicit assumptions & risks

### Assumptions (explicit)
- **Authoritative spec selection:** `spec-v1.22.md` is present in-repo and is treated as authoritative for major version v1.
- NATS + JetStream infrastructure is available for remote agent features (Spec §5.5, §9.1).
- OS support targets are Linux and macOS on amd64 and arm64, per explicit cross-compilation commands (Spec §4.2.7, §8.3.9).
- The private `gabewillen/liquidgen` repository is accessible from development machines and is treated as the authoritative implementation of the local inference engine used by `amux`.
- A Cloudflare R2 bucket and publish credentials are available to GitHub Actions for hosting versioned release artifacts and installer scripts.

### Risks
- Cross-platform exec interception hooks require platform-specific toolchains and may differ in behavior (Spec §8.3, §B.8).
- TUI decoding robustness across diverse terminal apps may require iterative fixture expansion (Spec §7.7, §4.3.2).
- liquidgen performance targets on CPU may be difficult on low-end machines; metrics and backpressure must be correct (Spec §4.2.10).
- Release distribution via Cloudflare R2 introduces new credential and availability dependencies; CI MUST fail fast and produce actionable diagnostics if publish steps are misconfigured.
- Because `liquidgen` is a private dependency, CI and release workflows MUST ensure authenticated module access or use a vendored/mirrored copy to avoid build breaks.


## Plan Stability Declaration

### Locked assumptions
- `spec-v1.22.md` is the authoritative specification for major version v1 (see plan Overview assumptions).
- Inline Go documentation and `go-docmd`-generated per-package `README.md` files are required and MUST remain in sync via an automated docs-check (Spec §4.2.6.1).
- Core implementation is Go 1.25.6 (`go1.25.6`) and uses the mandated libraries and conventions: wazero + TinyGo, stateforward/hsm-go (including /muid), creack/pty, explicit error handling, and agent-agnostic `internal/` boundaries (Spec §4.2.1–§4.2.6, §1.5).
- Remote orchestration uses the NATS + JetStream design, including per-host credentials, per-host subject authorization, JetStream KV durable state, and the normative request-reply/replay/buffering semantics (Spec §5.5.6–§5.5.8).
- Observability uses OpenTelemetry per spec, and local inference uses `liquidgen` with the required logical model IDs (Spec §4.2.9–§4.2.10).
- Build targets are the platforms explicitly enumerated by the spec’s cross-compilation guidance (Spec §4.2.7, §8.3.9).

### Changes that require regenerating this plan
- Any new spec version that changes or adds MUST/MUST NOT requirements, especially in §4.2–§13, or any new applied addendum to the authoritative spec.
- Any change to the remote protocol surface: subject naming, handshake/control payload schemas, replay ordering/buffering rules, or required host event schemas (Spec §5.5.7–§5.5.8, §9.1.3.2).
- Any change to required model identifiers or inference engine requirements for `liquidgen`, or changes to the semantic subscription embedding model requirements and storage/framing (Spec §4.2.10, §8.4.3.7).
- Any expansion of supported OS/arch targets or a change to exec hook compilation strategy that alters the build matrix and artifact packaging (Spec §4.2.7, §8.3.9).
