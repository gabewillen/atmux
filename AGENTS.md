# Core Invariants

- ALWAYS make breaking changes; NEVER leave unused or deprecated code behind
- NEVER place agent-specific code in `internal/`; all agent knowledge lives in `adapters/`
- ALWAYS reference adapters by string name and load via WASM registry
- ALWAYS use hsm-go for state machines, muid for IDs, wazero for WASM, creack/pty for PTYs
- ALWAYS use NATS as event transport even in local-only deployments
- ALWAYS search the codebase for similar solutions before generating new code
- NEVER generate new code when suitable existing solutions exist
- ALWAYS explain why existing solutions weren't suitable when creating new code
- ALWAYS follow existing code patterns and conventions found in the codebase
- NEVER use time.Sleep; always use hsm.After methods or ErrorGroup or channels
- NEVER use any external dependencies without explicit user approval that aren't in the docs/plan-v2.4.md file or docs/spec-v1.22.md file
- NEVER use any external testing library only use the built-in testing library
- ALWAYS try to write real integration tests over mocking

# Project Layout

- ALWAYS put agent-agnostic core code in `internal/` (agent, adapter, pty, monitor, process, config, protocol, tui, inference)
- ALWAYS put agent-specific WASM sources in `adapters/` (claude-code, cursor, windsurf)
- ALWAYS put CLI entry point in `cmd/amux/` and unified daemon binary in `cmd/amux-node/`
- ALWAYS put public types in `pkg/api/`, exec hook library in `hooks/`, ONNX models in `models/`

# Error Handling

- ALWAYS wrap errors: `fmt.Errorf("context: %w", err)`
- ALWAYS define sentinels as package-level `errors.New()`
- NEVER defer error checking

# IDs and Wire Format

- ALWAYS encode `muid.ID` as base-10 strings in JSON
- NEVER assign or emit `0` as a runtime ID (reserved sentinel)
- ALWAYS encode timestamps as RFC 3339 UTC, durations as Go duration strings, binary data as base64

# Configuration

- ALWAYS use TOML; load order: built-in < adapter < user < project < env (`AMUX__` prefix)
- NEVER store secrets in config files; use env vars only

# Agents and Worktrees

- ALWAYS require a git repo; worktrees at `.amux/worktrees/{agent_slug}/`, branches `amux/{agent_slug}`
- ALWAYS derive `agent_slug`: lowercase, non-`[a-z0-9-]` → `-`, collapse, trim, max 63 chars

# State Machines

- ALWAYS model lifecycle as HSM: Pending → Starting → Running → Terminated/Errored
- ALWAYS model presence as HSM: Online ↔ Busy ↔ Offline ↔ Away
- NEVER rely on agent self-reporting; infer state from PTY output patterns

# Remote (NATS)

- ALWAYS complete handshake before accepting spawn/kill/replay
- ALWAYS fail fast on control ops when host is disconnected; never enqueue for later
- ALWAYS keep PTY sessions alive during hub disconnection; buffer output in replay ring
- NEVER publish live PTY output after reconnect until replay is handled
- NEVER depend on SSH after bootstrap; runtime uses NATS leaf→hub only

# Process Tracking

- ALWAYS intercept via `LD_PRELOAD`/`DYLD_INSERT_LIBRARIES`; fall back to polling
- ALWAYS forward captured child stdout/stderr to PTY slave (output stream)
- NEVER write captured output to PTY master (would be interpreted as input)

# Adapter WASM ABI

- ALWAYS export: `amux_alloc`, `amux_free`, `manifest`, `on_output`, `format_input`, `on_event`
- ALWAYS return packed `(ptr << 32 | len)` uint64; `0` = failure
- ALWAYS one WASM instance per agent; serialize calls; 256MB memory cap
- NEVER share adapter instances across agents

# Event Dispatch

- ALWAYS route all events through NATS subjects; never dispatch locally bypassing NATS
- NEVER coalesce `process.spawned`/`process.completed` events

# Git Merge

- NEVER attempt automatic conflict resolution
- ALWAYS support: merge-commit, squash, rebase, ff-only

# CLI and Daemon

- ALWAYS communicate `amux` → `amuxd` via JSON-RPC 2.0 over Unix socket (`~/.amux/amuxd.sock`)
- ALWAYS treat `amuxd` and `amux-manager` as the same binary; role set by config/flags

# Plugins

- ALWAYS enforce permission checks; reject with JSON-RPC error `-32001`
- ALWAYS bridge WASM plugins via FD 3 (newline-delimited JSON-RPC)

# amux test

- ALWAYS run: tidy → vet → lint → test -race → test → coverage → bench (continue on failure)
- ALWAYS write TOML snapshot to `snapshots/`; `--no-snapshot` → stdout; `--regression` compares to previous
- ALWAYS require minimum total coverage of 80% for `amux test`
- ALWAYS require > 80% coverage for any new files
