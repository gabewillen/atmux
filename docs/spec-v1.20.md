# Agent Multiplexer (amux) Specification

**Version:** v1.20
**Derived from:** spec-v1.19.md
**Applied addendum:** addendum_1.md
**Change summary:**
- Specify remote disconnection buffering limits, drop policy, and request timeout behavior.
- Define replay ordering semantics and require replay after hub reconnection to avoid out-of-order PTY output.
- Clarify handshake readiness error handling (`not_ready`) and make `spawn` idempotent per `agent_id`.
- Define `repo_root` canonicalization (including symlink resolution) for deterministic `repo_key` computation.
- Clarify `liquidgen` model identifiers as logical aliases and specify error behavior for unknown/unavailable models.
- Impose a per-adapter-instance WASM linear memory cap and specify host recovery behavior.
- Fix an incorrect bootstrap step cross-reference (step 7 → step 8).
- Add RFC 3339 to normative references and require TUI XML decoders to reject unsupported schema versions.
**Status:** Draft
**Date:** 2026-01-26

**Abstract**—This specification defines the Agent Multiplexer (amux), an agent-agnostic application that enables a user or agent to orchestrate and drive a group of coding agents through multiplexed pseudo-terminal (PTY) sessions. The core system has no knowledge of specific agents; all agent-specific behavior is isolated in pluggable WASM adapters.

**Keywords**—agent, multiplexer, PTY, session management, event-driven architecture, WebAssembly, adapter, Go, agent-agnostic

---

## 1. Scope

### 1.1 Overview
This specification defines the Agent Multiplexer (amux), an application that manages multiple pseudo-terminal (PTY) sessions to orchestrate agent collaboration. The system employs a 100% event-driven architecture where all state changes, commands, and agent interactions are modeled as events.

**amux is fundamentally agent-agnostic.** The core system has no knowledge of specific coding agents (Claude Code, Cursor, Windsurf, etc.). All agent-specific behavior is isolated in pluggable WASM adapters, ensuring the core remains a generic orchestration platform.

amux operates through a **managed** architecture: the user never communicates directly with subordinate agents. All interaction flows through the amux director, which mediates communication, distributes tasks, and aggregates status information.

```
┌──────────┐         ┌──────────────┐         ┌─────────────────────┐
│   User   │◀───────▶│ amux director │◀───────▶│  Subordinate Agents │
│          │         │   (generic)  │         │ (via WASM adapters) │
└──────────┘         └──────────────┘         └─────────────────────┘
                            │
                    ┌───────┴───────┐
                    │ Agent-Agnostic │
                    │     Core       │
                    └───────────────┘
```

Each agent type is implemented as a WebAssembly (WASM) adapter, providing a pluggable interface for different agent implementations. The core application is implemented in Go and contains **zero agent-specific code**.

amux SHOULD be operated through a docker-like CLI client (`amux`) that communicates with a local daemon (commonly invoked as `amuxd`) over JSON-RPC (see §12). The `amuxd` and `amux-manager` command names MUST refer to the same binary; the active role is selected by configuration and/or flags (see §5.5 and §12).

### 1.2 Purpose
To provide a standardized, **agent-agnostic** mechanism for multi-agent orchestration using owned PTY sessions, enabling parallel agent execution, coordination, and communication through asynchronous event processing.

The primary motivation is to leverage existing subscription-based coding agents (Claude Code, Cursor, Windsurf, or any future CLI-based agent) in a collaborative fashion, allowing multiple agent instances to work together on a shared codebase without requiring custom API integrations.

**Key design principle:** The core amux system treats all agents uniformly through the adapter interface. Adding support for a new coding agent requires only creating a new WASM adapter—no modifications to the core system.

### 1.3 Prerequisites
- A git repository is required for each agent; amux MUST NOT operate an agent outside of a git repository.
- The director (including when hosted by `amuxd`) MUST either (a) be provided a request working directory that is within a git repository, or (b) be configured with an explicit repository root for each local agent (see §5.1 and §5.3.4).

### 1.4 Non-goals
This specification explicitly does not define:

- **CLI presentation details** — Visual layout, styling, and interactive affordances are implementation details; however the required command surfaces for the CLI client, daemon API, and CLI plugins are defined in §12–§13
- **External LLM provider integration** — API details for specific hosted LLM providers (OpenAI, Anthropic, etc.)
- **Agent-specific prompting** — System prompts or instructions for coordinating LLM behavior
- **Authentication/authorization** — User identity and multi-tenancy are out of scope; however NATS-based host authentication and per-host subject authorization for remote daemons are specified in §5.5.6.4, and per-plugin local permission gating is defined in §13.6
- **Persistent storage** — General-purpose database schemas and full session persistence across restarts are out of scope; however the NATS JetStream persistence and JetStream KV state required for non-local agents is specified in §5.5 and §9.1, and the SQLite + sqlite-vec storage for notification subscription embeddings is specified in §8.4.3.7.
- **Network protocols beyond SSH for agent orchestration** — SSH remains the bootstrap mechanism, but runtime orchestration for non-local agents uses NATS with JetStream as specified in §5.5; the core event system uses NATS for local and remote event distribution as specified in §9.1; HTTP(S) MAY still be used for self-update and CLI plugin distribution as specified in §12.4.7 and §13.4.
### 1.5 Agent-agnostic design principles

The amux architecture enforces strict separation between agent-agnostic core functionality and agent-specific adapter implementations.

#### 1.5.1 Core system requirements

The following packages shall contain **zero agent-specific code**:

| Package | Responsibility | Agent Knowledge |
|---------|---------------|-----------------|
| `internal/agent/` | Lifecycle, presence, messaging | None—treats all agents uniformly |
| `internal/adapter/` | WASM runtime, adapter loading | None—loads any conforming adapter |
| `internal/monitor/` | PTY observation, timeout detection | None—delegates pattern matching to adapters |
| `internal/process/` | Child process tracking | None—observes processes generically |
| `internal/pty/` | PTY creation and I/O | None—raw PTY operations |
| `internal/config/` | Configuration management | None—adapter configs are opaque |
| `internal/protocol/` | Remote communication | None—transports events generically |
| `pkg/api/` | Public types | None—`Agent.Adapter` is a string reference |

#### 1.5.2 Adapter isolation

All agent-specific code shall reside in adapter implementations:

```
adapters/
├── claude-code/      # Claude Code-specific patterns and commands
│   └── main.go       # Compiled to claude-code.wasm
├── cursor/           # Cursor-specific patterns and commands
│   └── main.go       # Compiled to cursor.wasm
└── windsurf/         # Windsurf-specific patterns and commands
    └── main.go       # Compiled to windsurf.wasm
```

#### 1.5.3 Adding a new agent type

To add support for a new coding agent:

1. Create `adapters/{agent-name}/main.go` with agent-specific patterns
2. Implement the WASM interface exports (`manifest`, `on_output`, `format_input`, `on_event`)
3. Compile to WASM: `tinygo build -o {agent-name}.wasm -target=wasi ./adapters/{agent-name}`
4. Place in discovery path (`~/.config/amux/adapters/` or `.amux/adapters/`)

**No modifications to core packages are required.**

#### 1.5.4 Built-in adapter prohibition

The `internal/` packages shall not contain built-in or fallback adapter implementations with hardcoded agent-specific patterns. If development convenience requires a default adapter, it shall be:

1. Compiled as a separate WASM module
2. Embedded as a binary asset (not Go source)
3. Loaded through the standard adapter discovery mechanism

This ensures the core remains truly agent-agnostic and testable without agent-specific dependencies.

## 2. Normative references
The following normative references are required to interpret or implement this specification:

- RFC 2119: Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", BCP 14, RFC 2119, March 1997. (https://www.rfc-editor.org/rfc/rfc2119)
- RFC 8259: Bray, T., "The JavaScript Object Notation (JSON) Data Interchange Format", RFC 8259, December 2017. (https://www.rfc-editor.org/rfc/rfc8259)
- RFC 4648: Josefsson, S., "The Base16, Base32, and Base64 Data Encodings", RFC 4648, October 2006. (https://www.rfc-editor.org/rfc/rfc4648)
- RFC 3339: Klyne, G., Newman, C., "Date and Time on the Internet: Timestamps", RFC 3339, July 2002. (https://www.rfc-editor.org/rfc/rfc3339)
- Semantic Versioning 2.0.0: "Semantic Versioning 2.0.0" (SemVer), for interpreting version constraints in `CLI.Constraint`. (https://semver.org/spec/v2.0.0.html)
- TOML v1.0.0: “TOML v1.0.0” specification, for interpreting configuration files and TOML value parsing. (https://toml.io/en/v1.0.0)
- JSON-RPC 2.0: “The JSON-RPC 2.0 Specification”, for MCP request/response and notification envelopes. (https://www.jsonrpc.org/specification)
- ECMA-48 (5th edition): Standard ECMA-48, "Control Functions for Coded Character Sets", for interpreting ANSI/ECMA-48 control functions used in terminal/TUI decoding. (https://ecma-international.org/wp-content/uploads/ECMA-48_5th_edition_june_1991.pdf)
- Xterm Control Sequences: "Xterm Control Sequences", for interpreting xterm-compatible escape sequences (including alternate screen buffers) used by common TUI libraries. (https://invisible-island.net/xterm/ctlseqs/ctlseqs.pdf)

## 3. Definitions
For the purposes of this document, the following terms and definitions apply.

**3.1 event:** An immutable record representing a state change or action within the system.

**3.2 amux director:** The intermediary component that mediates all communication between subordinate agents and the user. The amux director handles task distribution, aggregates status, and provides a unified interface. It may be driven by an LLM for autonomous operation.

**3.3 controller:** The user or LLM that directs subordinate agents through the amux director.

**3.4 subordinate agent:** An agent instance running within a PTY, managed by the controller.

**3.5 session:** An amux session containing one or more agent PTYs.

**3.6 event queue:** The hsm-go event queue mechanism for dispatching events to state machines. Supports broadcast (all machines) or targeted dispatch (by ID pattern).

**3.7 adapter:** A WebAssembly module that implements the agent interface, translating between amux's generic events and a specific agent implementation. Adapters are the sole location for agent-specific code.

**3.8 adapter interface:** The standardized, agent-agnostic contract that all WASM adapters shall implement to integrate with amux. This interface enables the core system to work with any coding agent without modification.

**3.9 agent:** An active agent instance with a name, description, assigned adapter, and dedicated worktree.

**3.10 worktree:** A git worktree providing an isolated working directory for an agent.

**3.11 repository:** The git repository from which amux operates; required for worktree creation.

**3.12 presence:** The availability state of an agent, indicating whether it can accept new tasks.

**3.13 roster:** The list of all agents in a session, including their presence and status information.

**3.14 PTY monitor:** A component that observes PTY output to detect activity, inactivity, and state changes.

**3.15 inactivity timeout:** The duration of no PTY output after which an agent is considered idle or stuck.

**3.16 child process:** A process spawned by an agent during task execution (e.g., `cargo build`, `npm test`).

**3.17 process tracker:** A component that monitors child processes spawned within agent PTYs.

**3.18 hsmnet:** A thin routing layer that bridges hsm-go's event dispatch across NATS subjects (optionally backed by JetStream) for non-local peers. Uses hsm's existing ID system; only maps IDs to peers.

**3.19 yamux:** A multiplexing library that enables multiple logical streams over a single connection. In this specification version, yamux is a legacy transport for remote agents and is only relevant if `remote.transport = "ssh_yamux"` is explicitly enabled.


**3.20 agent_slug:** A stable, filesystem-safe identifier derived from an agent’s configured `name`. Used for worktree directory names and git branch names. This is not the same as `Agent.ID`.

**3.21 agent runtime ID (Agent.ID):** A globally unique identifier of type `muid.ID` assigned to an agent at runtime. Used for HSM identity, event routing, and the remote protocol field `agent_id`.
**3.22 reserved ID value (0):** The value `0` of type `muid.ID` is reserved for sentinel use (for example `BroadcastID`). Implementations SHALL NOT assign `0` as a runtime ID for any agent, process, session, peer, or message. If an ID generator produces `0`, the implementation SHALL generate a new ID. A value of `0` MAY be used as a zero-value placeholder (unset) within in-memory structs or internal events prior to assignment, but MUST NOT be emitted in any externally observable ID field (roster entries, registries, or wire encodings) except where explicitly defined as a sentinel.

**3.23 repo_root:** The canonical absolute path (local) or canonical expanded path (remote) that identifies the root directory of a git repository on a given host. Each agent SHALL be associated with exactly one `repo_root`.

- Canonicalization MUST: (a) expand `~/` to the target host’s home directory, (b) convert to an absolute path, (c) clean `.`/`..` segments, and (d) resolve symbolic links to their target path where the underlying OS/filesystem provides a mechanism (for example `realpath`-style resolution).
- If symbolic link resolution is not possible (for example insufficient permissions or missing OS support), implementations MUST still apply (a)–(c) and MUST treat the result as canonical for the purpose of this specification.

**3.24 repo_key:** A stable, session-scoped identifier for a repository, derived from the tuple `(location.type, location.host, repo_root)`. Implementations MUST treat repositories with different tuples as distinct even if their contents are identical.

- The director MUST compute `repo_key` using the canonicalized `repo_root` per §3.23 (including symlink resolution where possible).
- If two configured agents would produce different `repo_root` strings before canonicalization but the same canonicalized `repo_root`, the director MUST treat them as the same repository for `repo_key` purposes.

**3.25 TUI (text user interface):** A full-screen terminal user interface that uses cursor-addressing and screen-control sequences (e.g., ANSI/ECMA-48) to render a 2D screen in a PTY.

**3.26 terminal screen model:** An in-memory representation of the current visual state of a terminal (cells, attributes, cursor state) obtained by decoding the PTY output byte stream.

**3.27 TUI decoder:** A component that incrementally decodes a PTY output byte stream into a terminal screen model and serializes that model into a compact XML representation.

**3.28 TUI XML:** A compact XML document produced by the TUI decoder that represents the visible terminal screen state and is intended for LLM ingestion.

**3.29 conformance harness:** A runnable test driver that executes the conformance suite against an amux implementation and one or more adapters.

**3.30 conformance suite:** A collection of normative tests that verify implementation and adapter behavior against this specification, including required end-to-end flows.

**3.31 adapter package:** A directory or Go module containing an adapter WASM module and any sidecar assets required for installation, setup, and conformance testing (see §10.8).

**3.32 adapter install spec (install.toml):** A sidecar TOML file located in an adapter package that declares automated setup steps (including auth) and conformance fixture commands (see §10.8.2–§10.8.4).

**3.33 adapter setup flow:** The sequence of checks and actions performed to make an agent CLI usable on a target host, including installing prerequisites, copying configuration, and completing authentication (see §5.5.10).


**3.34 amux CLI client (`amux`):** A docker-like command-line client that issues JSON-RPC requests to the local amux daemon and renders results to the user.

**3.35 amux daemon (`amuxd`):** An invocation of the unified amux node binary that serves a JSON-RPC control plane to clients (see §12). When configured in the director role, it hosts the amux director and MAY also run local host-manager functions.

**3.36 CLI plugin:** An extension that adds a top-level `amux <plugin>` command. A CLI plugin MAY be implemented as a WASM module or as a remote plugin endpoint (see §13).

**3.37 plugin registry:** A directory structure on the daemon host where installed CLI plugins are stored and discovered (see §13.3 and §13.5).

**3.38 plugin permission:** A named capability granted to a CLI plugin that constrains which daemon APIs and host resources the plugin may access. Permission prompts and enforcement are specified in §13.6.

**3.39 built-in CLI plugin:** A CLI plugin that is distributed with amux and available by default without any explicit installation step. This specification defines the built-in plugins `amux agent` and `amux chat` (see §13.8).

**3.40 NATS:** A message broker used as the runtime transport for local and non-local orchestration. In this specification, NATS is the required inter-host communication layer between the director (hub) and managers (leaf), including PTY I/O, control/events, and participant communication channels (see §5.5, §6.4, and §9.1).

**3.41 JetStream:** NATS’s persistence subsystem providing streams and a Key-Value (KV) store. In this specification, JetStream MUST be enabled on the amux host NATS server and MUST be used for durable remote control-plane state (see §5.5.6).

**3.42 host_id:** A stable identifier for a host running the unified amux node binary in manager role. A `host_id` MUST be unique among concurrently connected hosts and SHOULD be derived from `location.host` for SSH locations (see §5.5.6.3).

**3.43 host manager:** The long-running amux node process on a host, configured in manager role, that owns and monitors multiple agent PTY sessions. Each host that owns agent PTY sessions MUST run exactly one host manager. The host manager MUST also act as an agent (a "manager agent"): it MUST be able to receive tasks and messages, it MAY execute tasks directly, it MAY delegate work to agents on the same host, and it MUST report results back to the director (see §5.5.5, §6.4, and §5.5.9).

**3.44 unified amux node binary:** A single executable that MUST be installable and runnable under the command names `amuxd` and `amux-manager`. The role (director vs manager) MUST be determined by configuration and/or flags, not by the executable name.

**3.45 director role:** A node operating mode that MUST combine (a) the amux director logic and (b) a hub-mode NATS server with JetStream enabled. A director-role node MAY also run local host-manager functions to own PTYs on the same host.

**3.46 manager role:** A node operating mode that MUST run host-manager logic and MUST run (or connect to) a leaf-mode NATS server that connects to the hub. The manager role MUST be able to continue host-local operations when disconnected from the hub.


## 4. Conformance

### 4.1 Conformance language
The keywords **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** are to be interpreted as described in RFC 2119. In this document, **shall** and **shall not** are used as synonyms for **MUST** and **MUST NOT**.

### 4.2 Conventions

#### 4.2.1 Implementation language
The core application shall be implemented in Go (1.22+).

#### 4.2.2 WASM runtime
The application shall use wazero as the WebAssembly runtime. wazero is a pure Go implementation requiring no CGO, enabling straightforward cross-compilation.

Adapters shall be compiled to WASM using TinyGo for minimal binary size.

#### 4.2.3 State machine and identifiers
The application shall use [hsm-go](https://github.com/stateforward/hsm-go) for hierarchical state machine (HSM) implementation. Agents shall be modeled as actors with HSM-driven lifecycle and presence states.

The application shall use [`muid`](https://github.com/stateforward/hsm-go/tree/main/muid) (bundled with hsm-go) for entity identifiers. muid provides 64-bit snowflake-style IDs that are time-sortable and optimized for distributed systems.

Key state machines:
- **Agent lifecycle:** Pending → Starting → Running → Terminated/Errored
- **Agent presence:** Online ↔ Busy ↔ Offline ↔ Away

Events from the PTY monitor (e.g., `ActivityDetected`, `RateLimitDetected`) shall trigger state transitions via `hsm.Dispatch()`.

#### 4.2.4 PTY management
The application shall use `creack/pty` for pseudo-terminal management. This library provides:

- Cross-platform PTY creation (Linux, macOS)
- Non-blocking I/O via standard Go interfaces
- Window size (pty.Winsize) control

#### 4.2.5 Error handling
All errors shall be handled explicitly. Functions that can fail shall return an `error` value. The following patterns are required:

- Errors shall be wrapped with context using `fmt.Errorf("context: %w", err)`
- Sentinel errors shall be defined as package-level variables using `errors.New()`
- Error checking shall not be deferred; handle errors at the point of occurrence

#### 4.2.6 Project structure
The codebase shall be organized as follows, with a clear separation between agent-agnostic core and agent-specific adapters:

```
amux/
├── cmd/
│   ├── amux/            # Main CLI binary
│   └── amux-node/        # Unified daemon binary (director/manager roles; installed as amuxd and/or amux-manager)
├── internal/           # AGENT-AGNOSTIC CORE (no agent-specific code)
│   ├── agent/          # Agent orchestration (lifecycle, presence, messaging)
│   ├── adapter/        # WASM adapter runtime (loads any adapter)
│   ├── pty/            # PTY management (generic PTY operations)
│   ├── monitor/        # PTY output monitoring (delegates to adapters)
│   ├── tui/            # Terminal screen decoding and TUI XML encoding (agent-agnostic)
│   ├── process/        # Process tracking and interception (generic)
│   ├── config/         # Configuration management (adapter configs opaque)
│   ├── inference/      # Local inference integration (liquidgen)
│   └── protocol/       # Remote communication protocol (transports events)
├── pkg/
│   └── api/            # Public API types (Agent.Adapter is a string)
├── hooks/              # Exec hook library (Go c-shared)
│   ├── hook.go         # Main hook implementation with CGO
│   ├── protocol.go     # Hook ↔ tracker protocol types
│   ├── fd.go           # SCM_RIGHTS FD passing helpers
│   └── bin/            # Compiled shared libraries (embedded)
├── models/             # ONNX models for embeddings
│   ├── all-minilm-l6/  # Default embedding model
│   │   ├── model.onnx  # Quantized ONNX model
│   │   └── tokenizer.json
│   └── onnxruntime/    # ONNX Runtime libraries (per-platform)
└── adapters/           # AGENT-SPECIFIC CODE (all agent knowledge here)
    ├── claude-code/    # Claude Code adapter source
    ├── cursor/         # Cursor adapter source (example)
    └── windsurf/       # Windsurf adapter source (example)
```

**Key invariant:** The `internal/` directory shall never import from or reference specific adapters. All agent-specific patterns, commands, and behaviors must reside in `adapters/`.

#### 4.2.7 Build configuration
Cross-compilation shall be performed using Go's native toolchain:

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o amux-linux-amd64 ./cmd/amux

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o amux-linux-arm64 ./cmd/amux

# macOS x86_64
GOOS=darwin GOARCH=amd64 go build -o amux-darwin-amd64 ./cmd/amux

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o amux-darwin-arm64 ./cmd/amux
```

Adapter compilation using TinyGo:

```bash
tinygo build -o adapter.wasm -target=wasi ./adapters/claude-code
```

Hook library compilation (see §8.3.9 for details):

```bash
# Build hook library for current platform
CGO_ENABLED=1 go build -buildmode=c-shared -o hooks/bin/amux-hook.so ./hooks

# Cross-compile requires platform-specific toolchain
# See §8.3.9 for full cross-compilation commands
```

The compiled hook libraries shall be placed in `hooks/bin/` for embedding via `//go:embed`. Note that c-shared builds require CGO; cross-compilation requires the target platform's C toolchain.

#### 4.2.8 Configuration

The application shall follow 12-factor app best practices, particularly:

- **III. Config:** Store config in the environment
- **X. Dev/prod parity:** Keep development, staging, and production as similar as possible
- **XI. Logs:** Treat logs as event streams

The `amux` CLI client, the `amuxd` daemon, and all CLI plugins MUST adhere to these configuration and logging principles, including supporting environment-variable overrides as defined in §4.2.8.3 and avoiding mandatory interactive configuration for normal operation.

##### 4.2.8.1 Configuration format
All configuration files shall use TOML format.

##### 4.2.8.2 Configuration hierarchy
Configuration shall be loaded in the following order (later overrides earlier):

1. **Built-in defaults** — Compiled into the binary
2. **Adapter defaults** — Provided by the selected adapter (see §10.6) via `config_default` and/or `config.default.toml` (see below)
3. **User config** — `~/.config/amux/config.toml`
4. **User adapter config** — `~/.config/amux/adapters/{name}/config.toml`
5. **Project config** — `.amux/config.toml`
6. **Project adapter config** — `.amux/adapters/{name}/config.toml`
7. **Environment variables** — `AMUX__*` prefix (double underscore), as specified in §4.2.8.3

Adapter default configuration SHALL be loaded per adapter and SHALL be scoped to the adapter's configuration namespace:

- For each discovered adapter name, the director SHALL select exactly one module implementation according to the discovery priority in §10.6. Only the selected module's defaults apply.
- If the selected adapter module exports `config_default` (see §10.4.2), the director MUST call it and interpret the returned bytes as a UTF-8 TOML document.
- Otherwise, the director MUST attempt to read a UTF-8 TOML file named `config.default.toml` from the same directory as the adapter `.wasm` file.
- If neither source exists, the adapter contributes no defaults.
- Adapter default TOML MUST only set keys under `[adapters.<adapter_name>]` where `<adapter_name>` is the adapter `manifest.name`. Adapter defaults MUST NOT set or override any keys outside that subtree.
- If the adapter default TOML fails to parse or violates the scoping requirement above, the director MUST fail configuration loading and surface an error that identifies the adapter name and the source (`config_default` export vs file path).

##### 4.2.8.3 Environment variable mapping
Environment variables shall override configuration file values. Environment variables shall map to configuration key paths using the following convention:

- Only variables with prefix `AMUX__` (note the trailing double underscore) shall be considered.
- After the prefix, the remainder of the variable name shall be split into **path segments** using the delimiter `__` (double underscore).
- Each segment shall be lowercased to form the corresponding TOML path component.
- The TOML key path shall be formed by joining segments with `.` (dot).

**Adapter name normalization:**
- If the first segment is `adapters`, the *next* segment denotes the adapter name. In that adapter-name segment only, single underscores `_` shall be converted to hyphens `-` (for example `CLAUDE_CODE → claude-code`).
- In all other segments, single underscores `_` shall remain underscores.

**Examples:**

```
AMUX__GENERAL__LOG_LEVEL=info
→ [general] log_level = "info"

AMUX__EVENTS__COALESCE__IO_STREAMS=true
→ [events.coalesce] io_streams = true

AMUX__ADAPTERS__CLAUDE_CODE__CLI__CONSTRAINT=">=1.0.0 <2.0.0"
→ [adapters.claude-code] cli.constraint = ">=1.0.0 <2.0.0"
```

**Value parsing (environment variables):**
- The raw environment variable value is a string.
- Implementations shall attempt to parse the value as a TOML value by embedding it into a temporary TOML document of the form `v = <value>` (TOML v1.0.0) and reading `v`.
- If TOML value parsing fails, the value shall be treated as a string exactly as provided (equivalent to the TOML value `"<raw>"` with appropriate escaping).
- After value parsing, the key-specific parsing rules in §4.2.8.10 (durations, byte sizes, paths, SemVer constraints) shall be applied where relevant.

##### 4.2.8.4 Configuration structure

```toml
# ~/.config/amux/config.toml

[general]
log_level = "info"           # debug, info, warn, error
log_format = "text"          # text, json

[timeouts]
idle = "30s"
stuck = "5m"

[process]
capture_mode = "all"          # none, stdout, stderr, stdin, all
stream_buffer_size = "1MB"    # Ring buffer size per stream
hook_mode = "auto"            # auto, preload, polling, disabled
poll_interval = "100ms"       # Polling interval when using polling mode
hook_socket_dir = "/tmp"      # Directory for hook Unix sockets

[git.merge]
strategy = "squash"          # merge-commit, squash, rebase, ff-only
allow_dirty = false          # Refuse merges with uncommitted changes when false
# target_branch defaults to the repository base_branch (see §5.7.1)

[events]
batch_window = "50ms"         # Coalesce events within this window
batch_max_events = 100        # Maximum events per batch
batch_max_bytes = "64KB"      # Maximum bytes for I/O batches
batch_idle_flush = "10ms"     # Flush if no new events for this duration

[events.coalesce]
io_streams = true             # Coalesce stdout/stderr/stdin per process
presence = true               # Keep only latest presence per agent
activity = true               # Deduplicate activity events

[remote]
transport = "nats"            # nats, ssh_yamux (legacy)
buffer_size = "10MB"          # Per-session PTY replay buffer size; also caps buffered cross-host publications while hub-disconnected
request_timeout = "5s"        # Timeout for NATS request-reply control operations (spawn/kill/replay)
reconnect_max_attempts = 10
reconnect_backoff_base = "1s"
reconnect_backoff_max = "30s"

[remote.nats]
url = "nats://amux-host:4222"             # NATS server on the amux host (director host)
creds_path = "~/.config/amux/nats.creds"  # Per-host NATS credential file (provisioned during SSH bootstrap; scoped to this host, see §5.5.6.4)
subject_prefix = "amux"                   # Root subject namespace for all amux traffic
kv_bucket = "AMUX_KV"                     # JetStream KV bucket for remote state
stream_events = "AMUX_EVENTS"             # JetStream stream for EventMessage envelopes
stream_pty = "AMUX_PTY"                   # JetStream stream for PTY byte chunks
heartbeat_interval = "5s"

[remote.manager]
enabled = true
model = "lfm2.5-thinking"

[nats]
mode = "embedded"              # embedded, external
topology = "hub"               # hub, leaf
hub_url = "nats://amux-host:4222"  # Required when topology="leaf"
listen = "0.0.0.0:4222"
advertise_url = "nats://amux-host:4222"
jetstream_dir = "~/.amux/nats"

[node]
role = "director"              # director, manager

[daemon]
socket_path = "~/.amux/amuxd.sock"
autostart = true

[plugins]
dir = "~/.config/amux/plugins"
allow_remote = true

# Adapter-specific configuration overrides (example uses "claude-code")
# Replace with any installed adapter name: "cursor", "windsurf", etc.
[adapters.claude-code]
cli.constraint = ">=1.0.0 <2.0.0"
patterns.prompt = "^>\\s*$"

# Agent definitions - adapter field is a string reference to any installed adapter
[[agents]]
name = "frontend-dev"
about = "Works on React components and UI styling"
adapter = "claude-code"    # String reference - no compile-time dependency
location.type = "local"

[[agents]]
name = "docs-dev"
about = "Updates documentation in a separate repository"
adapter = "claude-code"
location.type = "local"
location.repo_path = "~/projects/docs-repo"

[[agents]]
name = "backend-dev"
about = "Handles API endpoints and database migrations"
adapter = "claude-code"    # Could also be "cursor", "windsurf", etc.
location.type = "ssh"
location.host = "devbox"   # Resolves via ~/.ssh/config
location.repo_path = "~/projects/my-repo"

[[agents]]
name = "test-runner"
about = "Runs tests and reports failures"
adapter = "cursor"         # Different adapter - system handles uniformly
location.type = "ssh"
location.host = "ci-server"
location.repo_path = "/srv/ci/my-repo"
```

##### 4.2.8.5 Adapter configuration
Each adapter may provide default configuration embedded in the WASM module. Users may override any adapter setting at the user or project level:

```toml
# .amux/adapters/claude-code/config.toml
# Project-specific overrides for claude-code adapter

[cli]
constraint = ">=1.0.23"  # Pin to specific version for this project

[patterns]
rate_limit = "rate limit exceeded|quota exceeded"
```

##### 4.2.8.6 Sensitive configuration
Sensitive values (API keys, credentials) shall be provided via environment variables only and shall not be stored in config files:

```bash
export AMUX__ADAPTERS__CLAUDE_CODE__API_KEY="sk-..."
```

The application shall warn if sensitive values appear in config files.

##### 4.2.8.7 Configuration actor
Configuration shall be managed by an HSM actor that supports live updates.

```go
var ConfigModel = hsm.Define("config",
    hsm.State("loading",
        hsm.Entry(func(c *ConfigActor) { c.loadAll() }),
    ),
    hsm.State("ready",
        hsm.Entry(func(c *ConfigActor) { c.startWatching() }),
        hsm.Exit(func(c *ConfigActor) { c.stopWatching() }),
    ),
    hsm.State("reloading",
        hsm.Entry(func(c *ConfigActor) { c.reload() }),
    ),

    hsm.Transition(hsm.On("config.loaded"), hsm.Source("loading"), hsm.Target("ready")),
    hsm.Transition(hsm.On("config.file_changed"), hsm.Source("ready"), hsm.Target("reloading")),
    hsm.Transition(hsm.On("config.reloaded"), hsm.Source("reloading"), hsm.Target("ready")),
    hsm.Transition(hsm.On("config.reload_failed"), hsm.Source("reloading"), hsm.Target("ready")),

    hsm.Initial(hsm.Target("loading")),
)
```

##### 4.2.8.8 Live config updates
The config actor shall watch for file changes and dispatch update events:

```go
type ConfigChange struct {
    Path     string      // Config key path: "coordination.interval"
    OldValue any
    NewValue any
}

// Config events
const (
    ConfigFileChanged = "config.file_changed"  // File modified on disk
    ConfigReloaded    = "config.reloaded"      // Reload complete
    ConfigUpdated     = "config.updated"       // Specific value changed
)
```

When config changes:
1. File watcher detects change → `config.file_changed` event
2. Config actor transitions to `reloading`, parses new config
3. For each changed value, dispatch `config.updated` with `ConfigChange`
4. Transition to `ready`, dispatch `config.reloaded`

##### 4.2.8.9 Subscribing to config changes
Actors may react to config updates:

```go
hsm.Transition(
    hsm.On("config.updated"),
    hsm.Source("*"),
    hsm.Effect(func(a *Agent, e hsm.Event) {
        change := e.Data.(ConfigChange)
        if change.Path == "timeouts.stuck" {
            a.stuckTimeout = change.NewValue.(time.Duration)
        }
    }),
)
```

Hot-reloadable config keys:
- `timeouts.*` — Idle/stuck timeouts
- `coordination.*` — Snapshot interval, buffer lines
- `adapters.*.patterns.*` — Detection patterns
- `telemetry.*` — Observability settings

Non-reloadable (require restart):
- `remote.transport` — Remote transport selection (`nats` or legacy `ssh_yamux`)
- `remote.nats.*` / `nats.*` — NATS connection and server settings
- Agent definitions (`[[agents]]`) — Agents must be added/removed explicitly



##### 4.2.8.10 Value parsing conventions

Configuration values shall be parsed using the following conventions:

- **Durations:** All values documented as durations (for example: `timeouts.idle`, `timeouts.stuck`, `events.batch_window`, `remote.request_timeout`, `remote.reconnect_backoff_base`) shall be strings that conform to Go's `time.ParseDuration` grammar (a sequence of decimal numbers with unit suffixes). Supported unit suffixes shall include: `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. Examples: `"500ms"`, `"30s"`, `"5m"`, `"1h30m"`.
- **Byte sizes:** All values documented as byte sizes (for example: `process.stream_buffer_size`, `events.batch_max_bytes`, `remote.buffer_size`) shall be either:
  - an integer number of bytes, or
  - a string of the form `<integer><unit>`, where `<unit>` is one of `B`, `KB`, `MB`, `GB`.

  The units `KB`, `MB`, and `GB` shall be interpreted as binary multiples: `1KB = 1024` bytes, `1MB = 1024^2` bytes, `1GB = 1024^3` bytes.
- **Paths:** A path value that begins with `~/` shall be expanded to the current user's home directory on the host where the configuration is loaded. Expansion of `~user/` is not required.
- **SemVer constraints:** `CLI.Constraint` shall support a conjunction (logical AND) of one or more comparisons separated by whitespace. Each comparison shall be an operator in `{=, ==, !=, >, >=, <, <=}` followed immediately by a SemVer 2.0.0 version (for example: `">=1.0.0"`, `"<2.0.0"`). If a constraint expression cannot be parsed, agent startup shall fail with a clear error identifying the adapter and the invalid constraint.

If any value fails to parse, configuration load or reload shall fail and shall emit `config.reload_failed` with an error message that includes the configuration key path.
#### 4.2.9 Observability

The application shall use OpenTelemetry for observability, providing traces, metrics, and logs.

##### 4.2.9.1 Instrumentation scope
The following operations shall be instrumented:

| Component | Traces | Metrics | Logs |
|-----------|--------|---------|------|
| Agent lifecycle | State transitions | Active count, by state | State changes |
| PTY monitor | Pattern match spans | Output bytes/sec, match counts | Detected events |
| Process tracker | Process lifecycle spans, I/O event spans | Process count, duration histogram, I/O bytes/sec by stream | Spawn/exit events, I/O events (debug) |
| Adapter | WASM call spans | Call duration, event counts | Adapter events |
| Remote agent | NATS connection spans (runtime) and SSH bootstrap spans | Connection count, reconnects | Connection events |
| HSM event queue | Event dispatch spans | Events/sec by type | Event payloads (debug) |

##### 4.2.9.2 Configuration
OpenTelemetry shall be configured via environment variables following the OTel specification, or via config file:

```bash
# Environment variables (OTel standard)
OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4317"
OTEL_EXPORTER_OTLP_PROTOCOL="grpc"
OTEL_SERVICE_NAME="amux"
OTEL_RESOURCE_ATTRIBUTES="deployment.environment=production"

# Sampling
OTEL_TRACES_SAMPLER="parentbased_traceidratio"
OTEL_TRACES_SAMPLER_ARG="0.1"

# Disable specific signals
OTEL_METRICS_EXPORTER="none"
OTEL_LOGS_EXPORTER="none"
```

```toml
# ~/.config/amux/config.toml
[telemetry]
enabled = true
service_name = "amux"

[telemetry.exporter]
endpoint = "http://localhost:4317"
protocol = "grpc"  # grpc, http/protobuf, http/json

[telemetry.traces]
enabled = true
sampler = "parentbased_traceidratio"
sampler_arg = 0.1

[telemetry.metrics]
enabled = true
interval = "60s"

[telemetry.logs]
enabled = true
level = "info"
```

##### 4.2.9.3 Span naming convention
Spans shall follow the convention: `{component}.{operation}`

```
agent.start
agent.stop
pty.monitor.scan
pty.monitor.pattern_match
process.spawn
process.wait
process.io.stdout
process.io.stderr
process.io.stdin
adapter.on_output
adapter.format_input
remote.connect
remote.bootstrap
event.dispatch
```

##### 4.2.9.4 Metrics
The following metrics shall be exported:

```go
// Counters
amux_agents_total{adapter, location_type, status}
amux_events_total{type}
amux_processes_total{agent_id, exit_status}
amux_remote_reconnects_total{host}
amux_process_io_events_total{agent_id, pid, stream}  // I/O event count by stream

// Gauges
amux_agents_active{adapter, presence}
amux_processes_running{agent_id}
amux_remote_connections{host, state}

// Histograms
amux_pty_output_bytes{agent_id}
amux_process_duration_seconds{agent_id, command}
amux_process_io_bytes{agent_id, pid, stream}         // I/O bytes per event by stream
amux_adapter_call_duration_seconds{adapter, function}
amux_event_dispatch_duration_seconds{type}
```

##### 4.2.9.5 Context propagation
Trace context shall be propagated:

- Between director and remote host daemons via the remote transport envelope (NATS subjects and message headers) defined in §5.5.7
- Through hsm event dispatch to correlate event chains
- To adapter WASM calls via host functions


#### 4.2.10 Local inference engine (liquidgen)

The implementation MUST support a local inference engine named `liquidgen` for any feature in this specification that requires local model inference (for example §8.4.3.6 and §11).

- `liquidgen` MUST support the following models:
  - `lfm2.5-thinking` (text-only reasoning)
  - `lfm2.5-VL` (vision-language)

**Model identifier semantics (normative):**
- The model identifiers `lfm2.5-thinking` and `lfm2.5-VL` are logical IDs required by this specification.
- An implementation MAY map each logical ID to a specific local quantized model artifact/version (implementation-defined), but it MUST:
  - use CPU-capable inference for the mapped model, and
  - expose the resolved mapping (logical ID → concrete artifact identifier) via the observability surfaces required in §4.2.9.
- If `Generate` is invoked with an unknown `Model` string, it MUST return an error and MUST NOT silently substitute a different model.
- If `Generate` is invoked with a known required logical model identifier but the mapped concrete model artifact is unavailable (missing files, load failure, incompatible runtime), it MUST return an error describing the failure.
- Implementations MUST use quantized variants of these models and MUST support CPU-only inference.
- When configured to stream tokens, the `liquidgen` integration SHOULD sustain 50–100 tokens per second on a CPU for quantized `lfm2.5-*` models under typical notification-gating workloads.
- If the implementation cannot meet the target throughput, it MUST expose runtime metrics for realized tokens-per-second and queue latency via the observability requirements in §4.2.9.

A minimal engine interface is:

```go
type LiquidgenEngine interface {
    // Generate produces a completion; implementations SHOULD stream tokens.
    Generate(ctx context.Context, req LiquidgenRequest) (LiquidgenStream, error)
}

type LiquidgenRequest struct {
    Model       string   // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
    Prompt      string
    MaxTokens   int
    Temperature float64
}

type LiquidgenStream interface {
    // Next returns the next token chunk; io.EOF indicates end of stream.
    Next() (token string, err error)
    Close() error
}
```

## 4.3 Conformance harness and conformance suite

### 4.3.1 General

The amux project MUST provide a fully functioning conformance harness and conformance suite.

- The conformance harness MUST be able to execute the conformance suite against:
  - the amux core implementation, and
  - any WASM adapter that claims conformance to this specification.
- All adapters intended for use with amux MUST be evaluated with the conformance suite and MUST pass all applicable tests.
- The conformance suite MUST include end-to-end (E2E) tests that exercise real PTY I/O, including remote PTY I/O when the implementation supports remote agents (see §5.5).

### 4.3.2 Required E2E flows

The conformance suite MUST include E2E tests that cover, at minimum:

1. **Auth flows**: unauthenticated detection, credential/config propagation where applicable, and interactive authentication completion.
2. **Menu flows**: full-screen TUI and interactive menu navigation using keystrokes, including verification of TUI decoding when enabled (see §7.7).
3. **Status flows**: presence and roster transitions, lifecycle events, and (where supported) remote connection recovery (see §5.4, §6.5, §5.5.8).
4. **Notification flows**: message routing, notification gating/batching, and subscription-driven notifications (see §6.4, §8.4.3.6, §8.4.3.7).
5. **CLI control plane flows**: JSON-RPC request/response, event subscription delivery, and permission enforcement for CLI plugins (see §12–§13).
6. **All supported functionality** defined by this specification that is implemented by the system under test.

For the purposes of conformance, “supported functionality” means:

- every behavior described with **MUST** or **MUST NOT** requirements in this specification that is applicable to the system under test, and
- for adapters, every behavior described with **MUST** or **MUST NOT** requirements in Section 10 and in any referenced sections that the adapter exercises (for example message formatting, prompt detection, and event emission).

### 4.3.3 Adapter conformance fixtures

To make E2E testing repeatable, each adapter package MUST provide a deterministic conformance fixture definition in its `install.toml` (see §10.8.2–§10.8.4).

- The conformance harness MUST be able to start the adapter’s fixture command in a PTY using the `conformance.fixture_start` command declared in `install.toml`.
- The fixture command MUST simulate, at minimum, the required E2E flows in §4.3.2, including a simulated auth flow and at least one interactive menu.
- The conformance harness MUST drive the fixture through these flows using only:
  - PTY input (keystrokes and text),
  - the amux tool surface (see §8.4.3.7), and
  - observation of emitted events (see §9).

### 4.3.4 Remote conformance runs

If the implementation supports remote agents (see §5.5), the conformance harness MUST support executing the conformance suite with at least one agent located on an SSH host.

Remote conformance runs MUST include:

- automated remote setup and authentication per §5.5.10, and
- validation that the remote agent PTY session can be established, driven, and recovered as specified.

## 5. Agent management

### 5.1 Agent structure
An agent shall consist of the following properties:

```go
type Agent struct {
    ID        muid.ID
    Name      string
    About     string
    Adapter   string    // String reference to adapter name (agent-agnostic)
    RepoRoot  string    // Canonical repository root path for this agent (see §5.3.4)
    Worktree  string    // Absolute path to the agent's working directory within RepoRoot
    Location  Location
    // Lifecycle and Presence managed by HSMs (see §5.4, §6.4)
    // Query via: agent.Lifecycle.State(), agent.Presence.State()
}
```

**Note:** The `Adapter` field is a **string reference**, not a typed dependency. The agent structure has no knowledge of specific adapter implementations. The adapter is loaded dynamically by name through the WASM registry:

```go
// Agent-agnostic usage - works with any adapter
adapter, err := registry.Load(agent.Adapter)  // "claude-code", "cursor", etc.
events := adapter.OnOutput(ptyOutput)
```

```go
type Location struct {
    Type     LocationType  // Local or SSH
    Host     string        // SSH host or alias from ~/.ssh/config
    User     string        // SSH user (optional if in ssh config)
    Port     int           // SSH port (optional if in ssh config)
    RepoPath string        // Path to git repository root on target host (required for SSH agents; optional for local agents to select a non-default repo)
}

type LocationType int

const (
    LocationLocal LocationType = iota
    LocationSSH
)
```

When parsing configuration, `location.type` SHALL be treated as a case-insensitive string with the following mappings:

- `"local"` → `LocationLocal`
- `"ssh"` → `LocationSSH`

Any other value SHALL be rejected during configuration validation.

For SSH locations, the system shall resolve host configuration using `kevinburke/ssh_config`:

```go
import sshconfig "github.com/kevinburke/ssh_config"

// Resolve SSH config for a host alias
hostname := sshconfig.Get(loc.Host, "HostName")      // Actual hostname
user := sshconfig.Get(loc.Host, "User")              // Username
port := sshconfig.Get(loc.Host, "Port")              // Port
identityFile := sshconfig.Get(loc.Host, "IdentityFile") // Key file
```

This allows users to reference hosts defined in `~/.ssh/config`:

```
# ~/.ssh/config
Host devbox
    HostName 192.168.1.100
    User deploy
    Port 2222
    IdentityFile ~/.ssh/devbox_key
```

### 5.2 Adding an agent
To add an agent, the controller shall:

1. Select an adapter from the available WASM adapters
2. Provide a `name` for the agent
3. Provide an `about` description for the agent
4. Specify a `location` (local or SSH). For SSH locations, `location.repo_path` MUST be provided and MUST point to a git repository on the remote host. For local locations, `location.repo_path` MAY be provided to select a specific local repository; if omitted, the director MUST use the repository that contains the request working directory (for example the `amux` client’s current working directory) as the agent’s `repo_root`.
5. The system shall ensure a dedicated git worktree exists for the agent on the host where the agent will run:
   - **Local:** Create or reuse the worktree under the resolved local `repo_root` (see §5.3.4).
   - **SSH:** After bootstrapping the remote daemon and connecting (step 6), the director MUST request worktree creation by sending a `spawn` control message that includes `repo_path` and `agent_slug` (see §5.5.7.3). The daemon MUST create or reuse the worktree under `location.repo_path` using the naming rules in §5.3.1.
6. The system shall initialize the agent based on location:
   - **Local:** Spawn a new PTY directly
   - **SSH:** Bootstrap the remote daemon (see 5.5), then connect and validate `location.repo_path`, then perform adapter remote setup (see §5.5.10) before starting the agent CLI PTY session
7. The system shall initialize the adapter within the PTY
8. The system shall emit an `agent.added` event

### 5.3 Worktree isolation
Each agent shall operate within its own git worktree to ensure:

- Isolated file system changes between agents
- Independent branch operations
- Conflict-free parallel work

#### 5.3.1 Worktree naming
Worktrees shall be created in `.amux/worktrees/{agent_slug}/` under the agent’s `repo_root` on the host where the agent runs (see §3.23 and §5.3.4).

`agent_slug` shall be a stable, filesystem-safe identifier derived from the configured agent `name` using the following normalization rules:

- Convert to lowercase.
- Replace any character not in `[a-z0-9-]` with `-`.
- Collapse consecutive `-` characters to a single `-`.
- Trim leading and trailing `-`.
- Truncate to at most 63 characters.
- If the result is empty, use `agent`.

If the normalized `agent_slug` collides with an existing agent’s `agent_slug`, the system shall append a numeric suffix `-2`, `-3`, ... until the slug is unique.

Example directory structure (where `agent_slug == name` after normalization):

```
.amux/
└── worktrees/
    ├── frontend-dev/
    ├── backend-dev/
    └── test-runner/
```

The worktree shall be created from the current repository HEAD unless otherwise specified. Each worktree shall be on its own branch named `amux/{agent_slug}`.

#### 5.3.2 Worktree cleanup
When an agent is removed, the system shall:

1. Terminate any running processes in the PTY
2. Remove the git worktree: `git worktree remove .amux/worktrees/{agent_slug}`
3. Optionally delete the associated branch (configurable, default: preserve)

#### 5.3.3 Merge conflicts
When agents merge their worktree branches, conflicts may arise. The system shall **not** automatically resolve conflicts. Instead:

- Agents are responsible for resolving merge conflicts in their own worktrees
- The director may assign merge conflict resolution as a task to an agent

#### 5.3.4 Multi-repository sessions

A single amux session MAY include agents operating on different git repositories (not limited to multiple worktrees of a single repository).

- Each agent MUST be associated with exactly one `repo_root` (see §3.23).
- For `location.type = "local"`:
  - If `location.repo_path` is set, the director MUST resolve and validate it as a git repository root and MUST use it as `repo_root`.
  - If `location.repo_path` is unset, the director MUST use the git repository root that contains the request working directory (for example the `amux` client’s current working directory) as `repo_root`.
  - If the session includes more than one distinct local `repo_root`, then `location.repo_path` MUST be set for every local agent whose `repo_root` is not the director’s repository.
- For `location.type = "ssh"`:
  - `location.repo_path` MUST be set and MUST resolve to a git repository root on the remote host; the daemon MUST reject `spawn` if `repo_path` is not a git repository.

The director MUST surface `repo_root` in any snapshot or status surface intended for coordination (see §11.1–§11.2). Cross-repository change integration (for example merging changes from one repository into another) is out of scope; git merge strategies in §5.7 apply only within a single repository.

Standard git merge/rebase workflows apply (see §5.7).

### 5.4 Agent lifecycle

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌────────────┐
│ Pending │───▶│ Starting│───▶│ Running │───▶│ Terminated │
└─────────┘    └─────────┘    └─────────┘    └────────────┘
                                   │
                                   ▼
                              ┌─────────┐
                              │ Errored │
                              └─────────┘
```

The lifecycle shall be implemented as an HSM:

```go
var LifecycleModel = hsm.Define("agent.lifecycle",
    hsm.State("pending"),
    hsm.State("starting",
        hsm.Entry(func(a *Agent) { a.bootstrap() }),
    ),
    hsm.State("running",
        hsm.Entry(func(a *Agent) { a.startMonitoring() }),
        hsm.Exit(func(a *Agent) { a.stopMonitoring() }),
    ),
    hsm.State("terminated", hsm.Final()),
    hsm.State("errored", hsm.Final()),

    hsm.Transition(hsm.On("start"), hsm.Source("pending"), hsm.Target("starting")),
    hsm.Transition(hsm.On("ready"), hsm.Source("starting"), hsm.Target("running")),
    hsm.Transition(hsm.On("stop"), hsm.Source("running"), hsm.Target("terminated")),
    hsm.Transition(hsm.On("error"), hsm.Source("*"), hsm.Target("errored")),

    hsm.Initial(hsm.Target("pending")),
)
```

### 5.5 Remote agents

Remote agents run on a separate host.

- SSH MAY be used for **bootstrapping** (installing or updating `amux-manager`, performing adapter setup, and starting the daemon).
- Runtime orchestration for non-local agents MUST NOT depend on an active SSH session. Once started, the remote node in manager role MUST persist beyond the bootstrap SSH session and MUST run (or connect to) a leaf-mode NATS server that connects to the hub NATS server running on the director-role node.

Each remote host runs exactly one node in manager role (commonly invoked as `amux-manager`) that acts as a **host manager** and can manage multiple agent PTY sessions (one per configured agent on that host).

#### 5.5.1 Agent architecture

```
Bootstrap phase (SSH):
┌────────────┐      SSH       ┌───────────────────────────────────────────────┐
│ amux client │───────────────▶│ 1. Copy bootstrap zip                         │
└────────────┘                │ 2. Unpack: binary + required adapter WASMs     │
                              │ 3. Provision leaf→hub connection material      │
                              │ 4. Start node in manager role if not running   │
                              └───────────────────────────────────────────────┘

Runtime phase (NATS hub + leaf):

  (director role: hub NATS + director logic)
┌───────────────────────────┐
│ amux node (director role)  │
│ - hub-mode NATS (JetStream)│
│ - director logic           │
└──────────────┬─────────────┘
               │ (leafnode links)
               ▼
      ┌───────────────────────────┐
      │ amux node (manager role)  │
      │ - leaf-mode NATS server    │
      │ - host manager + PTYs      │
      └──────────────┬────────────┘
                     │
              ┌──────▼───────┐
              │ Agent PTYs    │
              │ (N per host)  │
              └───────────────┘

Host-local agent↔agent communication MUST use the host's leaf-mode NATS server so that, when sender and receiver are on the same host, the message path remains local. NATS leaf/hub routing MUST carry cross-host messages when required.
```

#### 5.5.2 Daemon bootstrap
When adding a remote agent, the director MUST:

1. Resolve the SSH target host using `location.host` and the user’s SSH configuration.
2. Construct a bootstrap payload as a single ZIP file that contains:
   - the unified amux node binary for the remote host’s OS/arch, and
   - the required adapter WASM modules for agents that will run on that host.
3. Copy the bootstrap ZIP to the remote host (for example to `~/.amux/bootstrap/amux-bootstrap.zip`).
4. On the remote host, unpack the ZIP and install/replace:
   - the daemon executable at `~/.local/bin/amux-manager` (or other PATH location), and
   - the adapter modules under `~/.config/amux/adapters/` using either flat or package layout (see §10.6).
5. Provision leaf→hub connection material for the remote host:
   - The director MUST determine the hub NATS URL to advertise to the remote host.
   - The director MUST generate or retrieve unique per-host credentials for this `host_id` sufficient to establish the leaf-mode connection to the hub and to enforce per-host subject authorization as specified in §5.5.6.4.
   - The director MUST copy the credential material to the remote host prior to starting the daemon and MUST write it to `remote.nats.creds_path` with file permissions no more permissive than `0600`.
6. Check if the daemon is running: `ssh user@host "amux-manager status"`.
7. If not running, start the daemon in a way that survives the SSH session (daemonize): `ssh user@host "amux-manager daemon --role manager ..."`.
8. Verify the node has connected to the hub by observing either:
   - `amux-manager status` reporting `hub_connected=true`, or
   - an emitted `connection.established` event on the NATS event subject for that host (see §9.1.3.2).

The director MUST NOT require a persistent SSH port forward or SSH multiplexed connection after step 8 completes successfully.

#### 5.5.3 Agent binary selection
The system shall select the correct binary based on the remote host's architecture:

```go
type Architecture struct {
    OS   string  // "linux", "darwin"
    Arch string  // "amd64", "arm64"
}
```

The agent binary shall be embedded in the main amux binary or fetched from a known location.

#### 5.5.4 Agent daemon
The unified amux node binary SHALL support running as a daemon in both roles. When used on remote hosts, it is typically invoked as `amux-manager` and run in manager role.

```bash
amux-manager daemon --role manager   # Start daemon (forks to background) in manager role
amux-manager status                  # Check if daemon is running, print status
amux-manager stop                    # Graceful shutdown
amux-manager version                 # Print semantic version and exit 0
```

- `amux-manager version` SHALL print a single line to stdout of the form `amux-manager <semver>` (for example `amux-manager 1.5.0`).
- `<semver>` SHALL conform to Semantic Versioning 2.0.0.

The daemon MUST:
- Detach from the controlling terminal (daemonize).
- Maintain a stable `host_id` and `peer_id` identity used for routing (see §5.5.7.3 and §9.1.5).
- Run (or connect to) a leaf-mode NATS server that connects to the hub NATS server when `--role manager` is selected.
- Create and manage multiple PTY sessions (one per agent on this host).
- Persist across SSH connection interruptions and across director restarts, provided the hub remains available.

The daemon SHOULD accept NATS configuration via flags and/or environment variables. At minimum, it MUST support:
- `--role <director|manager>`: select the active role.
- `--host-id <string>`: the host identity to use on the bus.
- `--nats-url <url>`: hub NATS URL (required in manager role).
- `--nats-creds <path>` (optional): credential file path.

#### 5.5.5 Daemon responsibilities
The host manager (node in manager role) MUST:

- Own PTYs on the host (one per agent).
- Load the required WASM adapters from standard discovery paths, which MUST include `~/.config/amux/adapters/` (see §10.6).
- Perform adapter-based outbound message detection and inbound message formatting for agents on this host (see §6.4).
- Subscribe to participant communication subjects needed to receive messages for its local agents and host manager channel, and inject formatted inbound messages into local agent PTYs.
- Stream PTY output to the director over NATS subjects (§5.5.7.4) for observation and coordination.
- Receive PTY input from the director over NATS subjects (§5.5.7.4).
- Track child processes locally and emit process lifecycle and I/O events (§8.4 and §9.1.3.2).
- Participate in the hsm-go actor system locally; all host-local components that own state MUST be modeled as HSM actors.
- Continue local operations if the hub connection is lost:
  - PTY sessions MUST remain active.
  - Host-local agent↔agent communication MUST remain available via the local leaf NATS server.
  - Cross-host publications (for example `P.events.<host_id>` and `P.comm.*`) SHOULD be buffered up to configured limits and flushed when the hub connection is restored.

**Host-local management (optional):**
- When `remote.manager.enabled = true`, the host manager SHOULD run a local supervisor loop that uses the `liquidgen` engine (§4.2.10) to monitor its managed agents for liveness and failure conditions.
- If a managed agent process exits unexpectedly, the host manager MUST emit an event visible to the director (for example `process.failed` and/or a host-scoped diagnostic event) and MAY notify other agents on the same host by publishing `AgentMessage` objects to their communication channels (§6.4).

#### 5.5.6 NATS connectivity and durable state

Runtime orchestration for non-local agents MUST use NATS. JetStream MUST be enabled on the hub (director-role) NATS server.

##### 5.5.6.1 Hub NATS server (director role)

A director-role node MUST provide (or connect to) a hub-mode NATS server reachable by all configured manager-role hosts. JetStream MUST be enabled on the hub.

The NATS server MAY be:
- managed externally by the operator, or
- started and supervised by the unified amux node binary running in director role (implementation-defined).

If the NATS server listens on a non-loopback interface, the operator SHOULD enable transport security (for example TLS) and SHOULD restrict access to authorized hosts. User identity and multi-tenant authorization are out of scope (§1.4), but deployments MUST enforce host authentication and per-host subject authorization for remote daemons as specified in §5.5.6.4.

##### 5.5.6.2 Manager leaf server configuration

A manager-role node MUST be able to connect to the hub using the configured `remote.nats.url`.

The director MUST provision per-host credentials during SSH bootstrap, and the manager-role node MUST use them to establish the leaf-mode connection to the hub and for any hub-scoped NATS client connections it makes.

##### 5.5.6.3 JetStream KV state (required)

The NATS server MUST provide a JetStream Key-Value (KV) bucket (default name `AMUX_KV`, configurable via `remote.nats.kv_bucket`) for durable remote control-plane state.

The director MUST create the bucket if it does not exist.

The director and host managers MUST use the KV bucket for at least the following keys (all values are UTF-8 JSON objects):

- `hosts/<host_id>/info`: host metadata (version, os/arch, `peer_id`, and startup timestamp).
- `hosts/<host_id>/heartbeat`: last-seen heartbeat timestamp (RFC 3339).
- `sessions/<host_id>/<session_id>`: session metadata sufficient for reconnection (at minimum: `agent_id`, `agent_slug`, `repo_path`, and current session state).

##### 5.5.6.4 NATS authentication and per-host authorization (required)

Manager-role nodes MUST authenticate their leaf-mode connection to the hub NATS server using per-host credentials that are provisioned over SSH during bootstrap.

- For each `host_id`, the director MUST create a unique NATS credential (for example a NATS `.creds` file containing an NKey + JWT) and MUST associate it with exactly one `host_id`.
- The director MUST copy the credential to the remote host at `remote.nats.creds_path` during bootstrap (§5.5.2) and MUST ensure file permissions are no more permissive than `0600`.

**Host-bound subject permissions (normative):**

Let `P = remote.nats.subject_prefix` (default `amux`). The hub MUST enforce that traffic attributable to a given `host_id` is restricted to the following subject permissions when transiting the leaf→hub link:

Publish:
- `P.handshake.<host_id>` (handshake request)
- `P.events.<host_id>` (host events)
- `P.pty.<host_id>.*.out` (PTY output from daemon to director)
- `P.comm.director` (messages to the director channel)
- `P.comm.manager.*` (messages to any manager channel)
- `P.comm.agent.*.>` (messages to any agent channel)
- `P.comm.broadcast` (broadcast messages)

Subscribe:
- `P.ctl.<host_id>` (control requests)
- `P.pty.<host_id>.*.in` (PTY input from director to daemon)
- `P.comm.manager.<host_id>` (this host's manager channel)
- `P.comm.agent.<host_id>.>` (channels for agents on this host)
- `P.comm.broadcast` (broadcast messages)
- `_INBOX.>` (required for NATS request-reply replies)

A remote host credential MUST NOT be authorized for any other host's `P.ctl.*`, `P.events.*`, or `P.pty.*` subjects. (The `P.comm.*` communication subjects are authorized separately as listed above.)

If a daemon attempts to publish or subscribe outside its authorized subject set, the NATS server MUST deny the operation.

**Credential scope and host binding:**

- Credentials MUST be unique per `host_id`. The director MUST NOT reuse a credential across different `host_id` values.
- The director MUST treat the `<host_id>` token in the handshake subject (`P.handshake.<host_id>`) as the canonical host identity for the connecting daemon (see §5.5.7.3).
- Implementations MAY further harden host binding (for example by requiring TLS client certificates) but MUST NOT weaken the per-host subject authorization rules above.

#### 5.5.7 Communication protocol (NATS + JetStream)

All subjects in this section use a configurable prefix `P = remote.nats.subject_prefix` (default `amux`). Implementations MUST treat `P` as a literal NATS subject prefix and MUST append dot-delimited tokens to it as described below.

##### 5.5.7.1 Subject namespaces (normative)

For a given `host_id` and `session_id`, the following subject forms are defined:

- Handshake (daemon → director, request-reply): `P.handshake.<host_id>`
- Control requests (director → daemon, request-reply): `P.ctl.<host_id>`
- Host events (daemon → director): `P.events.<host_id>`
- PTY output (daemon → director): `P.pty.<host_id>.<session_id>.out`
- PTY input (director → daemon): `P.pty.<host_id>.<session_id>.in`

Additionally, the following subject forms are defined for participant communication channels (gossipsub-style pub-sub) (see §6.4):

- Director channel: `P.comm.director`
- Host manager channel: `P.comm.manager.<host_id>`
- Agent channel: `P.comm.agent.<host_id>.<agent_id>`
- Broadcast channel: `P.comm.broadcast`

Payloads published on `P.comm.*` subjects MUST be UTF-8 JSON encodings of `AgentMessage` (§6.4) and MUST follow the standard ID and timestamp encodings (§9.1.3.1). For unicast messages, the publishing component MUST publish the same `AgentMessage` (same `AgentMessage.ID`) to the sender channel and the recipient channel to support channel listeners.

If JetStream streams are configured (recommended), they SHOULD be configured to capture:
- `P.events.>` (event stream)
- `P.pty.>` (PTY stream)

##### 5.5.7.2 Control messages

Control requests and responses MUST be encoded as UTF-8 JSON objects whose top-level shape matches:

```go
type ControlMessage struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}
```

The following control message types are REQUIRED:
- `handshake`
- `ping` / `pong`
- `spawn`
- `kill`
- `replay`
- `error`

###### 5.5.7.2.1 Request-reply semantics (normative)

- The director MUST send `spawn`, `kill`, and `replay` as NATS request messages to `P.ctl.<host_id>` and MUST wait for a single response message.
- The daemon MUST respond to each request with either:
  - a success response whose `type` matches the request `type`, or
  - a single `error` message.
- Request correlation MUST be performed using NATS request-reply (the NATS `reply` subject). Implementations MUST NOT require an additional explicit request identifier field.

- The director MUST apply a timeout of `remote.request_timeout` when waiting for a response. If the timeout elapses, the director MUST treat the operation as failed and MUST surface an error to the caller.
- If the director considers the target host disconnected (for example the agent lifecycle is `Away`), it MUST fail fast by rejecting remote control operations (`spawn`, `kill`, `replay`) without issuing a NATS request.
- If the daemon replies with an `error` whose `code` is `"not_ready"`, the director MUST treat the daemon as not yet ready (handshake incomplete) and MUST NOT retry the operation until after it has observed `connection.established` for that `host_id` (§9.1.3.2).


##### 5.5.7.3 Required control payloads

Handshake (daemon request to `P.handshake.<host_id>`, director reply):

```json
{"type":"handshake","payload":{"protocol":1,"peer_id":"5678","role":"daemon","host_id":"devbox"}}
{"type":"handshake","payload":{"protocol":1,"peer_id":"1234","role":"director","host_id":"amux-host"}}
```

- The daemon MUST send a `handshake` request after establishing a NATS connection and MUST NOT accept `spawn`, `kill`, or `replay` requests until the handshake exchange is complete.
- If the daemon receives a `spawn`, `kill`, or `replay` request before the handshake exchange is complete (for example due to out-of-order delivery or implementation error), it MUST reply with an `error` whose payload has `request_type` set to the received request type, `code` set to `"not_ready"`, and a non-empty `message`.
- The daemon MUST send the `handshake` request to `P.handshake.<host_id>` where `<host_id>` is the daemon's configured `host_id`.
- The director MUST treat the `<host_id>` token in the request subject as canonical. If the handshake payload contains a different `host_id`, the director MUST reject the handshake.

- `peer_id` SHALL be encoded as a base-10 unsigned integer string in JSON and SHALL be unique among concurrently connected peers.
- `host_id` MUST be a non-empty string and MUST be unique among concurrently connected hosts.
- If `protocol` is unsupported or `peer_id`/`host_id` collides with an already-connected peer/host, the director MUST reject the handshake (by replying with `error`) and MUST treat the daemon as disconnected.

Error response:

```json
{"type":"error","payload":{"request_type":"spawn","code":"invalid_repo","message":"repo_path is not a git repository"}}
```

- `request_type` SHALL be one of: `"handshake"`, `"spawn"`, `"kill"`, `"replay"`, `"unknown"`.
- `code` SHALL be a short machine-readable string.
- `message` SHALL be a human-readable diagnostic.

Spawn request/response:

```json
{"type":"spawn","payload":{"agent_id":"42","agent_slug":"backend-dev","repo_path":"~/projects/my-repo","command":["claude-code"],"env":{"TERM":"xterm-256color"}}}
{"type":"spawn","payload":{"agent_id":"42","session_id":"9001"}}
```

- `agent_id` SHALL be encoded as a base-10 unsigned integer string in JSON.
- `agent_slug` MUST be provided. It MUST match the `agent_slug` computed by the director using the normalization rules in §5.3.1.
- `repo_path` is required for SSH agents and shall identify the git repository root on the remote host. It may begin with `~/` and shall be expanded by the remote daemon (see §4.2.8.10).
- On `spawn`, the daemon MUST create or reuse the agent worktree at `.amux/worktrees/{agent_slug}/` under `repo_path` and MUST set the PTY session working directory to that worktree path.
- `command` is the argv vector used to start the agent CLI in the PTY (first element is the executable).
- The response payload MUST include `agent_id` (echoed from the request) and `session_id` (base-10 unsigned integer string).
- `session_id` MUST be a non-zero value representable as an unsigned 64-bit integer.
- `spawn` MUST be idempotent for a given `agent_id` on a given daemon: if a session already exists for `agent_id`, the daemon MUST NOT start a second agent process. Instead it MUST return a `spawn` response containing the existing `session_id`.
  - If the existing session’s `agent_slug` or `repo_path` differs from the request, the daemon MUST reply with `error` using `code = "session_conflict"`.

Kill request/response:

```json
{"type":"kill","payload":{"session_id":"9001"}}
{"type":"kill","payload":{"session_id":"9001","killed":true}}
```

Replay request/response:

```json
{"type":"replay","payload":{"session_id":"9001"}}
{"type":"replay","payload":{"session_id":"9001","accepted":true}}
```

- For each `kill` request, the daemon SHALL respond with a `kill` response whose payload includes:
  - `session_id` (echoed)
  - `killed` (boolean): `true` if a session was found and termination was initiated, otherwise `false`.
- For each `replay` request, the daemon SHALL respond with a `replay` response whose payload includes:
  - `session_id` (echoed)
  - `accepted` (boolean): `true` if the daemon will replay buffered PTY output on the PTY output subject, otherwise `false`.

**Replay buffer and ordering (normative):**

- The daemon MUST maintain a per-session replay buffer of raw PTY output bytes, capped at `remote.buffer_size` bytes (ring-buffer semantics: when the cap is exceeded, the oldest bytes are dropped).
- The replay buffer MUST be updated for all PTY output bytes regardless of hub connectivity.
- If `remote.buffer_size` is `0`, replay buffering is disabled and the daemon MUST reply with `accepted = false`.
- If the requested `session_id` does not exist, the daemon MUST reply with `accepted = false`.
- If `accepted = true`, the daemon MUST publish the replay buffer contents to `P.pty.<host_id>.<session_id>.out` in oldest-to-newest byte order.
  - The replayed bytes MUST correspond to a snapshot of the replay buffer taken at the moment the daemon receives the `replay` request.
  - The daemon MUST chunk replay bytes such that no single NATS message exceeds the deployment’s maximum payload size (§5.5.7.4).
  - The daemon MUST publish all replay bytes for that session before publishing any subsequently produced live PTY output bytes for that session (bytes produced after the daemon receives the `replay` request). If necessary, the daemon MUST temporarily buffer live output while replay is in progress.


Ping/pong:

```json
{"type":"ping","payload":{"ts_unix_ms":1700000000000}}
{"type":"pong","payload":{"ts_unix_ms":1700000000000}}
```

##### 5.5.7.4 PTY I/O subjects

PTY I/O is transported as raw bytes on NATS subjects.

- PTY output bytes MUST be published by the daemon to `P.pty.<host_id>.<session_id>.out`.
- PTY input bytes MUST be subscribed to by the daemon on `P.pty.<host_id>.<session_id>.in`.

The director MUST:
- subscribe to the PTY output subject for each active session, and
- publish user or system input bytes to the PTY input subject for that session.

Implementations MUST chunk PTY bytes such that no single NATS message payload exceeds the maximum supported NATS payload size for the deployment.

##### 5.5.7.5 Host events

The daemon MUST publish `EventMessage` envelopes (as defined in §9.1.3) to `P.events.<host_id>`.

In this protocol version, the director SHOULD NOT publish `EventMessage` objects to `P.events.<host_id>`, and the daemon MUST ignore any `EventMessage` it receives.

##### 5.5.7.6 Required sequencing (director ↔ daemon)

For a new daemon connection, the daemon MUST:

1. Connect to NATS.
2. Perform the request-reply `handshake` exchange on `P.handshake.<host_id>`.
3. Start listening for:
   - control requests on `P.ctl.<host_id>`, and
   - PTY input on `P.pty.<host_id>.*.in` (implementation MAY use subject wildcards).

For a new remote agent session, the director MUST:

1. Ensure the daemon has completed handshake (e.g., by tracking `connection.established` events).
2. Send a `spawn` request to `P.ctl.<host_id>` and receive `session_id`.
3. Subscribe to `P.pty.<host_id>.<session_id>.out`.
4. Begin publishing input to `P.pty.<host_id>.<session_id>.in`.
5. If output replay is desired, send `replay` for that `session_id` after step 3 and handle replayed bytes arriving on the PTY output subject.

If any required step fails, the director MUST treat the affected agent as disconnected.

#### 5.5.8 Connection recovery
If the hub connection drops (for example the leaf→hub link is interrupted):

1. The manager-role node continues running.
2. PTY sessions remain active.
3. The manager-role node buffers PTY output (up to configurable limits) and SHOULD buffer cross-host publications (up to configurable limits).
4. The director detects disconnect and transitions the affected agent(s): `Running → Away`.
5. The director attempts reconnection with exponential backoff.
6. Upon daemon reconnection and handshake, the director MUST send `replay` for each active session and MUST handle replayed bytes on the PTY output subjects.
7. After replay (if requested), the agent transitions: `Away → Running`.

**Buffering during hub disconnection (normative):**

- While the hub connection is down, the manager-role node MUST continue reading PTY output and MUST retain it in each session’s replay buffer (§5.5.7.3).
- For any session that was active during the disconnection, the daemon MUST NOT publish PTY output bytes to `P.pty.<host_id>.<session_id>.out` after hub reconnection until it has received and handled a `replay` request for that `session_id` (even if it responds with `accepted = false`).
- For cross-host publications that would normally transit the hub (for example `P.events.<host_id>` and `P.comm.*`), the manager-role node:
  - SHOULD buffer outbound publications while disconnected, up to a maximum queued payload size of `remote.buffer_size` bytes total across all buffered publications.
  - MUST account queued size as the sum of NATS message payload lengths in bytes (excluding subject names and headers).
  - MUST drop oldest queued publications first once the maximum queued size is exceeded.
  - MUST preserve per-subject publish order for any buffered publications that are eventually flushed.
  - If it does not buffer a publication, it MUST drop it (that is, it MUST NOT block the producer waiting for hub connectivity).

**Request-reply operations during disconnection (normative):**

- While a host is disconnected (agent lifecycle `Away`), the director MUST NOT enqueue `spawn`, `kill`, or `replay` requests for later delivery. It MUST instead fail fast (see §5.5.7.2.1).

**Flush behavior on reconnection (normative):**

- When hub connectivity is restored, the manager-role node SHOULD flush any buffered cross-host publications.
  - Flush MUST be FIFO per subject.
  - New publications generated while a flush is in progress MUST be appended after older buffered publications for that same subject.
  - Relative ordering across different subjects is unspecified.


#### 5.5.9 Session management
Each remote host runs one host manager daemon that manages multiple PTY sessions:

```go
type Session struct {
    ID       muid.ID
    AgentID  muid.ID
    PTY      *pty.Pty
    Buffer   *RingBuffer  // Buffered output for replay
}
```

The session replay `Buffer` MUST implement the replay buffering requirements in §5.5.7.3 and MUST be capped by `remote.buffer_size`.

The director sends `spawn` to create a new session and `kill` to terminate one. Each session’s I/O is carried on a pair of NATS subjects (see §5.5.7.4); no additional multiplexing prefix is required.

The host manager SHOULD monitor each session and MUST emit a director-visible event when an agent process exits unexpectedly. If `remote.manager.enabled = true`, the host manager MAY attempt local remediation (for example restarting the agent) but MUST report such actions to the director via events.

#### 5.5.10 Remote adapter setup and authentication

Adapters MUST be fully remote installable, including all setup required to make the underlying agent CLI usable on a remote host.

When adding an agent with `location.type = "ssh"`, the director MUST perform an adapter setup flow on the remote host prior to starting the agent CLI PTY session.

The setup flow MUST support both of the following auth strategies:

1. **Config/credential copy**: copying configuration and authentication material from the director host to the remote host.
2. **Interactive auth flow**: running a login or authorization command on the remote host in a managed PTY session so that prompts, device-code flows, and other interactive UI can be completed.

The director MUST determine setup behavior using the adapter package’s `install.toml` (see §10.8.2).

##### 5.5.10.1 Setup preflight

For remote setup, the director MUST:

1. Resolve the adapter package directory for the agent’s selected adapter (see §10.6 and §10.8).
2. Load `install.toml` from that package directory.
3. Execute the adapter’s CLI version command on the remote host as declared by the adapter manifest (`manifest().cli.version_cmd`) and evaluate the result against the adapter’s declared constraint (`manifest().cli.constraint`) (see §10.2 and §10.4.2).

If the CLI is missing or does not satisfy the constraint, the director MUST execute remote install steps as declared in `install.toml` (see §5.5.10.2).

##### 5.5.10.2 Remote install steps

`install.toml` MUST support declaring a list of remote install steps under `[setup.remote.install]`.

- Each step MUST declare either:
  - `sh` (a shell command string executed using `sh -lc`), or
  - `exec` (an argv array executed without a shell).
- Steps MAY be conditional on `os` and `arch` values.
- Steps MUST be treated as idempotent. If a step fails, the director MUST emit `adapter.setup.failed` and MUST abort agent startup.

##### 5.5.10.3 Remote auth steps

`install.toml` MUST support declaring remote authentication under `[setup.remote.auth]`.

- If `setup.remote.auth.check_sh` is provided, the director MUST execute it on the remote host. Exit code `0` SHALL be treated as “authenticated”; any other exit code SHALL be treated as “not authenticated”.
- If `setup.remote.auth.copy_paths` is provided, the director MUST copy each listed path from the director host to the remote host prior to attempting interactive login. Copy behavior MUST be recursive for directories and MUST preserve relative paths.
- If, after any configured copy step, authentication is still not satisfied, and `setup.remote.auth.login_sh` or `setup.remote.auth.login_exec` is provided, the director MUST execute the login command in a managed PTY session on the remote host.

The director MUST emit `adapter.auth.started` and `adapter.auth.completed` events for the auth flow, and MUST emit `adapter.auth.failed` if login does not result in authentication success.

##### 5.5.10.4 Completion criteria

Remote setup MUST be considered complete only when:

- the agent CLI binary satisfies the adapter’s version constraint, and
- the remote auth check (if declared) returns success.

On success, the director MUST emit `adapter.setup.completed` before starting the agent CLI PTY session.

#### 5.5.11 Future leader promotion (constraints)
The architecture MUST support future leader promotion where any manager-role node could become the director-role node.

- Any node that can run in manager role MUST be capable of being started in director role using the same binary, subject to receiving appropriate configuration (hub listen/advertise and JetStream storage) and credentials.
- Subject namespaces for participant communication MUST remain stable across leader changes. In particular, `P.comm.director` MUST remain the director channel subject regardless of which host is currently the director.
- Leader election and automatic promotion are out of scope for this specification version. Implementations MUST NOT require protocol changes (subject renames, message schema changes) to permit a manual role switch from manager to director.


### 5.6 Graceful shutdown

Shutdown is modeled as an event-driven process using HSM transitions.

#### 5.6.1 Shutdown HSM

```go
var ShutdownModel = hsm.Define("system.shutdown",
    hsm.State("running"),
    hsm.State("draining",
        hsm.Entry(func(s *System) {
            // Dispatch shutdown.initiated to all agents
            hsm.DispatchAll(ctx, hsm.Event{Name: "shutdown.initiated"})
        }),
    ),
    hsm.State("terminating",
        hsm.Entry(func(s *System) {
            // Force-terminate remaining agents
            hsm.DispatchAll(ctx, hsm.Event{Name: "shutdown.force"})
        }),
    ),
    hsm.State("stopped", hsm.Final()),

    hsm.Transition(hsm.On("shutdown.request"),
        hsm.Source("running"), hsm.Target("draining"),
    ),
    hsm.Transition(hsm.On("shutdown.force"),
        hsm.Source("running"), hsm.Target("terminating"),
    ),
    hsm.Transition(hsm.On("shutdown.force"),
        hsm.Source("draining"), hsm.Target("terminating"),
    ),
    hsm.Transition(hsm.On("drain.complete"),
        hsm.Source("draining"), hsm.Target("stopped"),
    ),
    hsm.Transition(hsm.On("drain.timeout"),
        hsm.Source("draining"), hsm.Target("terminating"),
    ),
    hsm.Transition(hsm.On("terminate.complete"),
        hsm.Source("terminating"), hsm.Target("stopped"),
    ),

    hsm.Initial(hsm.Target("running")),
)
```

#### 5.6.2 Shutdown events

| Signal | Event | Behavior |
|--------|-------|----------|
| SIGTERM | `shutdown.request` | Initiate graceful drain |
| SIGINT | `shutdown.request` | Initiate graceful drain |
| Second SIGINT | `shutdown.force` | Force immediate termination |

#### 5.6.3 Agent shutdown behavior

When an agent receives `shutdown.initiated`:

1. Cancel any running task: `task.cancel` event
2. Transition lifecycle: `Running → Terminated`
3. Close PTY gracefully (send EOF, wait for shell exit)
4. Clean up worktree (optional, configurable)
5. Emit `agent.stopped` event

When all agents reach `Terminated`, the system emits `drain.complete`.

#### 5.6.4 Drain timeout

If agents do not terminate within `shutdown.drain_timeout` (default: 30s):

1. System transitions to `terminating`
2. `shutdown.force` dispatched to all agents
3. PTYs are killed (SIGKILL to shell process)
4. System proceeds to `stopped`

```toml
[shutdown]
drain_timeout = "30s"
cleanup_worktrees = false  # Preserve worktrees by default
```

### 5.7 Git merge strategies

This section specifies how changes made in an agent worktree branch are integrated into a target branch within the same repository.

Merge execution is performed by the director by running local `git` commands in the corresponding `repo_root`. Therefore, this section applies only to repositories that are locally accessible to the director (`location.type = "local"`). For agents whose `location.type = "ssh"`, automatic git merge execution by the director is out of scope for protocol v1; such merges MUST be performed on the remote host via normal git workflows (manual or by assigning tasks to an agent).

#### 5.7.1 Branch roles

- Each agent worktree branch MUST be named `amux/{agent_slug}` (see §5.3.1).
- For each `repo_root`, the director MUST record a `base_branch` at the time the first agent for that repository is added. The director MUST determine `base_branch` by running `git symbolic-ref --quiet --short HEAD` in `repo_root` and using the command output (trimmed of trailing newline). If this command fails (e.g., detached HEAD or an unborn branch), `base_branch` MUST be set to the configured `git.merge.target_branch` value if provided; otherwise the director MUST fail the operation with an error instructing the user to set `git.merge.target_branch`.
- A merge operation integrates `amux/{agent_slug}` into a `target_branch`. If not specified, `target_branch` MUST default to `base_branch`.

#### 5.7.2 Supported strategies (normative)

Implementations MUST support the following merge strategies:

- `merge-commit`: Perform a non-fast-forward merge that creates a merge commit.
- `squash`: Squash all commits from `amux/{agent_slug}` into a single commit on `target_branch`.
- `rebase`: Rebase `amux/{agent_slug}` onto the current tip of `target_branch` and then fast-forward `target_branch`.
- `ff-only`: Fast-forward `target_branch` to `amux/{agent_slug}` only if it is a direct descendant; otherwise fail.

#### 5.7.3 Preconditions

Before attempting integration, the director MUST verify:

- The repository at `repo_root` is a valid git repository and `target_branch` exists.
- The agent worktree has no uncommitted changes. If the worktree is dirty, the director MUST refuse integration unless `git.merge.allow_dirty = true`.
- The agent branch `amux/{agent_slug}` exists locally (or is fetchable from the same repository).

#### 5.7.4 Conflict handling

- If the chosen strategy produces merge conflicts, the director MUST stop the integration attempt without committing partial results.
- The director MUST emit `git.merge.conflict` and MUST NOT attempt automatic conflict resolution.
- Conflict resolution MAY be assigned as a task to an agent, but the system MUST NOT claim the repository is integrated until the conflicts are resolved and the merge is completed successfully.

#### 5.7.5 Required events

The director MUST emit the following events for merge operations:

- `git.merge.requested` when an integration attempt begins
- `git.merge.completed` when the target branch is updated successfully
- `git.merge.conflict` when conflicts are detected
- `git.merge.failed` when integration fails for any other reason

Event payloads SHOULD include at least: `repo_root`, `agent_slug`, `strategy`, `target_branch`, and any available commit identifiers.

## 6. Presence and roster

### 6.1 Presence states
Each agent shall maintain a presence state visible to all other agents:

| Presence | Description |
|----------|-------------|
| `Online` | Agent is idle and available to accept tasks |
| `Busy` | Agent is actively working on a task |
| `Offline` | Agent is rate-limited or temporarily unavailable |
| `Away` | Agent is connected but not responsive |

### 6.2 Roster
The system MUST maintain a roster containing all agents, all host managers (manager agents), and the director, and their current state. The roster MUST be:

- Accessible to all agents and host managers via the adapter interface
- Updated in real-time as presence changes occur
- Broadcast via `presence.changed` events

### 6.3 Presence awareness
Each agent shall have access to:

- The full roster of other agents
- Real-time presence updates for all agents
- The `name` and `about` of each agent
- The current task (if any) of busy agents

This enables Slack-like collaboration where agents can:

- Know who else is available for work
- Avoid assigning tasks to busy or offline agents
- Coordinate handoffs when rate-limited
- Understand the roles of other agents

### 6.4 Inter-agent messaging
Agents, host managers (manager agents), and the director MAY communicate with each other. Communication MUST be distributed over NATS using gossip-style pub-sub participant channels (see §5.5.7.1). The director MUST be able to listen to any communication on any manager or agent channel, each host manager MUST be able to listen to any agent within the same host, and agents MUST be able to listen to other agents and the host manager.

```go
type AgentMessage struct {
    ID        muid.ID
    From      muid.ID   // Sender runtime ID (set by publishing component)
    To        muid.ID   // Recipient runtime ID (set by publishing component, or BroadcastID)
    ToSlug    string    // Recipient token captured from text (typically agent_slug); case-insensitive
    Content   string
    Timestamp time.Time
}

const (
    BroadcastID muid.ID = 0  // Special ID for broadcast to all participants
)
```

The director and each host manager MUST have stable runtime IDs that can appear in `AgentMessage.From` and `AgentMessage.To`. The director MUST be addressable by the reserved slug `director`. Each host manager MUST be addressable by `manager@<host_id>`; additionally, agents MAY refer to their local host manager using the reserved slug `manager`.

Channel listening MUST be implemented as follows:

- The director MUST subscribe to `P.comm.>` for observation and coordination, and it MAY publish messages as a participant. The director MUST NOT require parsing remote PTY output to route participant messages.
- Each host manager MUST subscribe to `P.comm.manager.<host_id>`, `P.comm.agent.<host_id>.>`, and `P.comm.broadcast`.
- The system MUST provide a mechanism for an agent to listen to other participant channels (agent and host manager channels). The mechanism MAY be configuration-driven; when enabled, the host manager SHOULD mirror listened messages into the listening agent’s PTY (formatted via the agent’s adapter), prefixed with the source channel.

#### 6.4.1 Message routing
When an agent sends a message:

1. Agent writes to its PTY.
2. The local host manager feeds PTY output to the agent’s adapter. If the adapter detects an outbound message, it emits `message.outbound` with an `AgentMessage` payload (§10.5.2) whose `ToSlug` and `Content` are populated.
3. The host manager enriches and resolves the message:
   - The host manager MUST set `From` to the sender runtime ID.
   - The host manager MUST set `ID` to a newly generated non-zero `muid.ID`.
   - The host manager MUST set `Timestamp` to the current time in UTC.
   - Resolution of `ToSlug` MUST follow the rules in §6.4.1.3 below.
4. The publishing host manager MUST publish the resolved `AgentMessage` to NATS participant channels (see §5.5.7.1):
   - For unicast, it MUST publish the message to both the sender and recipient channels using the same `AgentMessage.ID`.
   - For broadcast (`To == BroadcastID`), it MUST publish the message to `P.comm.broadcast` and MAY also mirror it to participant channels.
5. The recipient’s host manager MUST deliver the message to any local PTY targets by:
   - formatting the message using the recipient adapter’s `format_input`, and
   - writing the resulting bytes to the recipient PTY input stream.

The director MUST be able to observe participant communication by subscribing to `P.comm.>`, but it MUST NOT be required as a message router for agent↔agent delivery.

##### 6.4.1.3 ToSlug resolution (normative)
The publishing host manager MUST resolve `ToSlug` as follows (case-insensitive):
- If `ToSlug` is `"all"`, `"broadcast"`, or `"*"`, it MUST set `To = BroadcastID`.
- If `ToSlug` is `"director"`, it MUST set `To` to the director runtime ID.
- If `ToSlug` is `"manager"`, it MUST set `To` to the sender’s local host manager runtime ID.
- If `ToSlug` has the form `"manager@<host_id>"`, it MUST resolve the referenced host manager and set `To` to that manager runtime ID.
- Otherwise, it MUST resolve `ToSlug` against known `agent_slug` values and set `To` to the resolved agent runtime ID.
- If resolution fails, it MUST NOT route the message; it MUST log a warning and SHOULD emit a diagnostic message to the sender’s PTY.

#### 6.4.2 Message events

| Event | Direction | Description |
|-------|-----------|-------------|
| `message.outbound` | Agent → Host manager | Agent sent a message (detected from PTY output) |
| `message.inbound` | Host manager → Agent/Manager/Director | Message delivered to a participant |
| `message.broadcast` | Director → All | Broadcast to all participants |

#### 6.4.3 Adapter message patterns
Adapters shall define patterns for detecting outbound messages:

```go
// Regular expression strings using Go's RE2 syntax.
// The core SHALL compile these strings to regex objects before use.
type AdapterPatterns struct {
    Prompt     string `json:"prompt"`                // Prompt readiness
    RateLimit  string `json:"rate_limit"`            // Rate limiting
    Error      string `json:"error"`                 // Error detection
    Completion string `json:"completion"`            // Task completion
    Message    string `json:"message,omitempty"`     // Outbound message detection; optional
}
```

Example pattern for Claude Code: `@(\w+):\s*(.+)` to detect `@backend-dev: can you review this?`

#### 6.4.4 Remote agent messaging (leaf-routed)
Remote agents participate fully in the messaging system.

In this protocol version, all message detection and message formatting MUST occur on the host manager that owns the agent PTY, regardless of whether the agent is local to the director host or on a remote host. The director MUST NOT be required to interpret raw PTY output in order to route participant messages.

Message routing for remote agents:

1. **Outbound (remote → any):**
   - Remote agent writes a message to its PTY.
   - The remote host manager detects and routes the message locally (see §6.4.1) and publishes the resulting `AgentMessage` on `P.comm.*` subjects.

2. **Inbound (any → remote):**
   - The remote host manager receives the `AgentMessage` via `P.comm.*` subjects.
   - The remote host manager formats the message using the recipient adapter’s `format_input` and injects it into the remote agent’s PTY.

3. **Remote-to-remote:**
   - Messages between agents on different hosts MUST be routed via NATS leaf/hub routing using the `P.comm.*` subjects.
   - The director MAY observe all messages by subscribing to `P.comm.>`, but it MUST NOT be required for delivery.

PTY I/O subjects (§5.5.7.4) remain the mechanism for the director to send arbitrary PTY input (tooling, task injection) and to observe PTY output (monitoring, snapshots). Participant messaging MUST use `P.comm.*` subjects.

### 6.5 Presence transitions

```
                    ┌──────────────────┐
                    ▼                  │
┌────────┐    ┌─────────┐    ┌────────┐
│ Online │◀──▶│  Busy   │───▶│ Offline│
└────────┘    └─────────┘    └────────┘
     ▲              │              │
     │              ▼              │
     │         ┌────────┐          │
     └─────────│  Away  │◀─────────┘
               └────────┘
```

The presence shall be implemented as an HSM:

```go
var PresenceModel = hsm.Define("agent.presence",
    hsm.State("online"),
    hsm.State("busy"),
    hsm.State("offline"),
    hsm.State("away"),

    // Online ↔ Busy
    hsm.Transition(hsm.On("task.assigned"), hsm.Source("online"), hsm.Target("busy")),
    hsm.Transition(hsm.On("task.completed"), hsm.Source("busy"), hsm.Target("online")),
    hsm.Transition(hsm.On("prompt.detected"), hsm.Source("busy"), hsm.Target("online")),

    // → Offline (rate limited)
    hsm.Transition(hsm.On("rate.limit"), hsm.Source("busy"), hsm.Target("offline")),
    hsm.Transition(hsm.On("rate.limit"), hsm.Source("online"), hsm.Target("offline")),
    hsm.Transition(hsm.On("rate.cleared"), hsm.Source("offline"), hsm.Target("online")),

    // → Away (unresponsive)
    hsm.Transition(hsm.On("stuck.detected"), hsm.Source("*"), hsm.Target("away")),
    hsm.Transition(hsm.On("activity.detected"), hsm.Source("away"), hsm.Target("online")),

    hsm.Initial(hsm.Target("online")),
)
```

Transition triggers:
- `Online → Busy`: `task.assigned` event
- `Busy → Online`: `task.completed` or `prompt.detected` event
- `* → Offline`: `rate.limit` event
- `Offline → Online`: `rate.cleared` event
- `* → Away`: `stuck.detected` event (no output for `timeouts.stuck`)
- `Away → Online`: `activity.detected` event

Note: "agent" here refers to the subordinate coding agents (Claude Code, Cursor, etc.), not the amux director.

## 7. PTY monitoring

### 7.1 Rationale
Agents cannot be relied upon to self-report status changes. An agent may complete a task, become stuck, crash, or hit a rate limit without explicitly signaling. The PTY monitor provides external observation of agent activity.

### 7.2 Monitor responsibilities
The PTY monitor shall:

- Subscribe to PTY output streams for each agent
- Track time since last output activity
- Detect patterns indicating completion, errors, or rate limiting
- Emit events based on observed state changes
- Not rely on agent cooperation or explicit signals
- Maintain a terminal screen model for TUI decoding when enabled (see §7.7)

### 7.3 Activity detection

| Observation | Interpretation | Event |
|-------------|----------------|-------|
| Output activity | Agent is working | `activity.detected` |
| No output for `timeouts.idle` | Task may be complete | `inactivity.detected` |
| No output for `timeouts.stuck` | Agent may be stuck | `stuck.detected` |
| Rate limit pattern matched | Agent is rate-limited | `rate.limit` |
| Error pattern matched | Agent encountered error | `error.detected` |
| Prompt pattern matched | Agent awaiting input | `prompt.detected` |

### 7.4 Pattern matching
Adapters shall provide patterns for the PTY monitor to recognize:

```go
// Regular expression strings using Go's RE2 syntax.
// The core SHALL compile these strings to regex objects before use.
type AdapterPatterns struct {
    Prompt     string `json:"prompt"`         // Pattern indicating agent is ready for input
    RateLimit  string `json:"rate_limit"`     // Pattern indicating rate limiting
    Error      string `json:"error"`          // Pattern indicating error state
    Completion string `json:"completion"`     // Pattern indicating task completion
    Message    string `json:"message,omitempty"` // Optional: agent message detection
}
```

### 7.5 Timeout configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| `timeouts.idle` | 30s | Time before `inactivity.detected` |
| `timeouts.stuck` | 5m | Time before `stuck.detected` |

### 7.6 Presence inference
The PTY monitor shall update agent presence based on observations:

```
ActivityDetected   → Busy
PromptDetected     → Online (idle, awaiting input)
InactivityDetected → Online (task likely complete)
StuckDetected      → Away
RateLimitDetected  → Offline
ErrorDetected      → status = Errored
```

### 7.7 Terminal UI (TUI) decoding

#### 7.7.1 Purpose
Full-screen TUIs redraw the terminal using cursor addressing and screen-control sequences. Line-oriented PTY capture is often noisy for these applications and is token-inefficient due to large amounts of whitespace and border characters.

When enabled, the director MUST decode the PTY output byte stream into a terminal screen model and MUST make a compact XML representation of the current visible screen available for LLM ingestion (see §11.2.1).

#### 7.7.2 Decoder input and placement
- The director MUST feed the TUI decoder the same ordered PTY output byte stream that is used for adapter pattern matching and user display.
- The TUI decoder MUST be incremental and MUST correctly handle escape sequences that are split across read boundaries.
- For remote agents, the director MUST decode the raw bytes received from the remote daemon (see §6.4.4). The remote daemon MUST NOT filter, normalize, or remove terminal escape sequences.

#### 7.7.3 Required terminal semantics
The TUI decoder MUST implement a virtual terminal sufficient to render common TUI libraries that target `TERM=xterm-256color`.

At a minimum, the decoder MUST support:

- Printable UTF-8 text. Invalid UTF-8 byte sequences MUST be replaced with U+FFFD.
- C0 control characters: BS (0x08), HT (0x09), LF (0x0A), CR (0x0D). BEL (0x07) MAY be ignored.
- ANSI/ECMA-48 escape sequences required for cursor-addressing TUIs:
  - Cursor movement: CUU/CUD/CUF/CUB (`CSI A/B/C/D`), CUP/HVP (`CSI H` / `CSI f`)
  - Erase functions: ED (`CSI J`), EL (`CSI K`)
  - Insert/delete: ICH (`CSI @`), DCH (`CSI P`), IL (`CSI L`), DL (`CSI M`)
  - Scrolling: DECSTBM (`CSI r`), SU/SD (`CSI S` / `CSI T`)
  - SGR (`CSI m`) for text attributes and color

The decoder MUST support xterm/DEC private modes commonly used by TUIs:

- Alternate screen buffer: `CSI ? 1049 h` and `CSI ? 1049 l`
- Cursor visibility: `CSI ? 25 h` and `CSI ? 25 l`

Implementations MAY additionally support `CSI ? 47 h/l` and `CSI ? 1047 h/l`.

Unrecognized or unsupported escape sequences MUST be ignored without corrupting the terminal screen model.

#### 7.7.4 Screen model requirements
- The screen model MUST represent the terminal as a fixed grid of `rows x cols` cells.
- Each cell MUST track the displayed rune (or blank) and SHOULD track SGR-derived attributes (foreground, background, and flags such as bold/reverse).
- The decoder MUST maintain separate main and alternate screen buffers and MUST switch according to DEC private mode sequences.
- On PTY resize, the decoder MUST update its `rows` and `cols` and MUST preserve the overlapping region of existing content. Newly exposed cells MUST be blank with default attributes.

#### 7.7.5 XML serialization
The decoder MUST be able to serialize the current visible screen into TUI XML as defined in §11.2.1.

Serialization MUST be deterministic for a given terminal screen model (byte-for-byte identical output for identical model state), to support efficient downstream diffing and caching.

## 8. Process tracking

### 8.1 Rationale
Agents invoke external processes (compilers, test runners, linters, etc.) that may run for extended periods. Tracking these processes provides insight into what an agent is waiting on and enables subscribers to react to process completion.

### 8.2 Process structure

```go
type Process struct {
    PID       int
    AgentID   muid.ID
    Command   string
    Args      []string
    StartedAt time.Time
    EndedAt   *time.Time  // nil if still running
    ExitCode  *int        // nil if still running
    Status    ProcessStatus
}

type ProcessStatus int

const (
    ProcessRunning ProcessStatus = iota
    ProcessCompleted
    ProcessFailed
    ProcessKilled
)
```

### 8.3 Process interception

The process tracker shall intercept all processes spawned by agents to capture their lifecycle and I/O streams.

#### 8.3.1 Exec hooking via preload

The system shall inject an exec hook library into agent PTY sessions to intercept all process spawning:

```go
type ExecHook struct {
    SocketPath string    // Unix socket for hook → tracker communication
    LibPath    string    // Path to preload library
}
```

**Linux implementation:**

The system shall use `LD_PRELOAD` to inject a shared library that wraps `execve()` and related syscalls:

```bash
# Set in PTY environment before spawning agent
LD_PRELOAD=/path/to/amux-hook.so
AMUX_HOOK_SOCKET=/tmp/amux-{session}.sock
```

The hook library shall:
1. Intercept `execve()`, `execvp()`, `execvpe()`, `posix_spawn()`, and `forkexec()` variants
2. Before exec: notify tracker via Unix socket with **PID**, command, args, env, and working directory
3. Receive assigned `ProcessID` and optional file descriptors for I/O capture
4. Before calling the real exec: set up any requested I/O redirection through tracker-provided file descriptors

**macOS implementation:**

The system shall use `DYLD_INSERT_LIBRARIES` with equivalent functionality:

```bash
DYLD_INSERT_LIBRARIES=/path/to/amux-hook.dylib
AMUX_HOOK_SOCKET=/tmp/amux-{session}.sock
```

Note: macOS System Integrity Protection (SIP) may restrict injection for system binaries. The hook shall gracefully degrade to polling-based detection when injection fails.

#### 8.3.2 Hook protocol

Communication between the hook library and process tracker shall use a binary protocol over Unix socket:

```go
type HookMessageType uint8

const (
    HookExecPre   HookMessageType = 1  // Before exec (request FDs)
    HookExecPost  HookMessageType = 2  // After exec succeeded
    HookExecFail  HookMessageType = 3  // Exec failed
    HookFork      HookMessageType = 4  // Fork occurred
    HookExit      HookMessageType = 5  // Process exiting
)

type HookExecPreRequest struct {
    PID        int  // PID of the process invoking exec*
    ParentPID  int
    Command    string
    Args       []string
    Env        []string
    WorkDir    string
    Timestamp  int64  // Unix nanoseconds
}

type HookExecPreResponse struct {
    ProcessID  muid.ID     // Assigned process ID
    Capture    CaptureMode // Which streams to capture
    // Pipe FDs are passed via SCM_RIGHTS on the response frame (see 8.3.2.1)
}

type HookExitNotify struct {
    PID       int
    ProcessID muid.ID // MAY be 0 if unknown; tracker SHALL map via PID
    ExitCode  int
    Signal    int  // 0 if not signaled
    Timestamp int64
}
```

If `HookExitNotify.ProcessID == 0`, the tracker SHALL resolve the ProcessID using the most recently assigned ProcessID for the same `PID` (from `HookExecPreResponse`). If no mapping exists, the tracker SHALL ignore the exit notification and MUST log a warning.

##### 8.3.2.1 Wire framing and file descriptor passing

All hook messages shall be sent over a Unix domain **stream** socket. Message boundaries shall be defined by an explicit frame header:

- **Byte 0:** `HookMessageType` (u8)
- **Bytes 1..4:** `payload_len` (u32, little-endian)
- **Bytes 5..(5+payload_len-1):** `payload` bytes

`payload` shall be a UTF-8 JSON object whose schema depends on `HookMessageType`:
- `HookExecPre` payload is `HookExecPreRequest` (request) or `HookExecPreResponse` (response), distinguished by message direction.
- `HookExit` payload is `HookExitNotify`.

**JSON key naming (normative):**
- Hook-protocol JSON payloads MUST use object keys that match the Go struct field names shown in §8.3.2 when encoded with Go `encoding/json` default rules (for example: `PID`, `ParentPID`, `Command`, `Args`, `Env`, `WorkDir`, `Timestamp`, `ProcessID`, `Capture`, `ExitCode`, `Signal`).
- Receivers MUST ignore unknown keys.

**Request/response rule:**
- For each `HookExecPre` request sent by the hook library, the tracker shall respond with exactly one `HookExecPre` response frame.
- The hook library shall block waiting for the response up to **250ms**. If the timeout expires, the hook library shall proceed without interception (passthrough) and shall not attempt to use any file descriptors.

**FD passing (SCM_RIGHTS):**
- The tracker MAY attach 0–3 file descriptors to the `HookExecPre` response using `SCM_RIGHTS`.
- The response payload’s `Capture` bitmask MUST define which streams are intercepted. The number of attached file descriptors MUST equal the number of set bits in `Capture` across `{CaptureStdin, CaptureStdout, CaptureStderr}`.
- Attached descriptors MUST be ordered by stream in the fixed stream order: `stdin`, then `stdout`, then `stderr`, including only those streams whose corresponding capture bit is set.
- The response JSON payload MUST NOT attempt to encode the receiving-side numeric FD values. The hook library MUST map attached descriptors to streams by iterating the fixed stream order and consuming one descriptor for each stream whose capture bit is set.
- If the attached FD count does not match the number of set bits in `Capture`, the hook library MUST treat the response as invalid and MUST proceed in passthrough mode (no interception) for that exec.

**Exec-post note:**
- Because successful `exec*()` calls do not return, `HookExecPost` is optional and MAY be omitted by implementations that cannot emit it.

#### 8.3.3 I/O pipe architecture

For each intercepted process, the tracker shall create pipe pairs to capture I/O:

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────┐
│  Child Process  │         │  Process Tracker │         │  PTY/Agent  │
├─────────────────┤         ├──────────────────┤         ├─────────────┤
│ stdout → pipe[1]│────────▶│ pipe[0] → tee ───│────────▶│ PTY stdin   │
│ stderr → pipe[1]│────────▶│ pipe[0] → tee ───│────────▶│ PTY stdin   │
│ stdin ← pipe[0] │◀────────│ pipe[1] ← tee ───│◀────────│ PTY stdout  │
└─────────────────┘         │         │        │         └─────────────┘
                            │         ▼        │
                            │   Event Dispatch │
                            │   (process.*)    │
                            └──────────────────┘
```

The tracker acts as a transparent proxy:
- Child output flows through tracker to PTY (visible to agent)
- PTY input flows through tracker to child stdin
- All data is copied to event dispatch for subscribers

**PTY direction note (normative):** In a PTY pair, bytes written to the PTY master are delivered to the PTY slave as input, and bytes written to the PTY slave are observed on the PTY master as output. When forwarding captured child stdout/stderr to keep output visible on the PTY output stream, the tracker MUST write forwarded bytes to the PTY slave (or otherwise inject into the PTY output stream). The tracker MUST NOT forward captured stdout/stderr by writing to the PTY master in a way that would be interpreted as input by the agent.

```go
type ProcessPipes struct {
    ProcessID muid.ID

    // Child-side pipe ends (passed to hooked process via SCM_RIGHTS)
    ChildStdin  *os.File  // Read end of stdin pipe (child side)
    ChildStdout *os.File  // Write end of stdout pipe (child side)
    ChildStderr *os.File  // Write end of stderr pipe (child side)

    // Tracker-side pipe ends (owned by tracker)
    TrackerStdinWrite  *os.File  // Write end → child reads
    TrackerStdoutRead  *os.File  // Read end ← child writes
    TrackerStderrRead  *os.File  // Read end ← child writes
}

func (t *ProcessTracker) createPipes(processID muid.ID) (*ProcessPipes, error)
func (t *ProcessTracker) proxyIO(pipes *ProcessPipes, pty *os.File) error
```

#### 8.3.4 FD passing via SCM_RIGHTS

The hook library and process tracker exchange file descriptors using Unix domain sockets with `SCM_RIGHTS` ancillary messages. This allows the tracker to create pipes and pass the appropriate ends to the child process before exec.

**Sequence diagram:**

```
┌──────────────┐                    ┌──────────────────┐
│  Hook (child)│                    │  Process Tracker │
└──────┬───────┘                    └────────┬─────────┘
       │                                     │
       │  1. execve() intercepted            │
       │                                     │
       │  2. HookExecPre ──────────────────▶ │
       │     {cmd, args, env, workdir}       │
       │                                     │
       │                                     │ 3. Create 3 pipe pairs
       │                                     │    stdin:  [r0, w0]
       │                                     │    stdout: [r1, w1]
       │                                     │    stderr: [r2, w2]
       │                                     │
       │  4. HookExecPreResponse ◀────────── │
       │     + SCM_RIGHTS: [r0, w1, w2]      │
       │     (tracker keeps: [w0, r1, r2])   │
       │                                     │
       │  5. dup2(r0, STDIN_FILENO)          │
       │     dup2(w1, STDOUT_FILENO)         │
       │     dup2(w2, STDERR_FILENO)         │
       │     close(r0, w1, w2)               │
       │                                     │
       │  6. Call real execve()              │
       │                                     │
       │  ═══════════════════════════════════│═══ Process image replaced
       │                                     │
       │  7. Child runs with redirected FDs  │
       │                                     │ 8. Tracker reads from r1, r2
       │                                     │    Tracker writes to w0
       │                                     │    Emits process.stdout/stderr/stdin events
       │                                     │    Forwards to PTY
```

**Hook library (Go with c-shared buildmode):**

The hook library shall be implemented in Go using `-buildmode=c-shared` to produce a shared library that can be loaded via `LD_PRELOAD`:

```go
// hooks/hook.go
package main

/*
#include <unistd.h>
#include <dlfcn.h>
#include <stdlib.h>

// Function pointer to real execve
typedef int (*execve_func)(const char*, char* const[], char* const[]);
static execve_func real_execve = NULL;

static void init_real_execve() {
    if (real_execve == NULL) {
        real_execve = (execve_func)dlsym(RTLD_NEXT, "execve");
    }
}

static int call_real_execve(const char *path, char *const argv[], char *const envp[]) {
    init_real_execve();
    return real_execve(path, argv, envp);
}
*/
import "C"

import (
    "net"
    "os"
    "syscall"
    "unsafe"
)

//export execve
func execve(path *C.char, argv **C.char, envp **C.char) C.int {
    goPath := C.GoString(path)
    goArgs := goStringSlice(argv)
    goEnv := goStringSlice(envp)

    // Connect to tracker
    socketPath := os.Getenv("AMUX_HOOK_SOCKET")
    if socketPath == "" {
        return C.call_real_execve(path, argv, envp)
    }

    conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
    if err != nil {
        return C.call_real_execve(path, argv, envp)
    }
    defer conn.Close()

    // Send exec pre-notification
    req := HookExecPreRequest{
        ParentPID: os.Getppid(),
        Command:   goPath,
        Args:      goArgs,
        Env:       goEnv,
        WorkDir:   mustGetwd(),
    }
    if err := sendMessage(conn, HookExecPre, req); err != nil {
        return C.call_real_execve(path, argv, envp)
    }

    // Receive response with FDs via SCM_RIGHTS
    resp, fds, err := recvMessageWithFDs(conn)
    if err != nil || len(fds) < 3 {
        return C.call_real_execve(path, argv, envp)
    }

    // Redirect standard FDs to tracker pipes
    if resp.Capture&CaptureStdin != 0 {
        syscall.Dup2(int(fds[0].Fd()), syscall.Stdin)
    }
    if resp.Capture&CaptureStdout != 0 {
        syscall.Dup2(int(fds[1].Fd()), syscall.Stdout)
    }
    if resp.Capture&CaptureStderr != 0 {
        syscall.Dup2(int(fds[2].Fd()), syscall.Stderr)
    }

    // Close received FDs (now duplicated to 0,1,2)
    for _, f := range fds {
        f.Close()
    }

    // Call real execve - process image replaced
    return C.call_real_execve(path, argv, envp)
}

func goStringSlice(cArray **C.char) []string {
    var result []string
    ptr := uintptr(unsafe.Pointer(cArray))
    for {
        cStr := *(**C.char)(unsafe.Pointer(ptr))
        if cStr == nil {
            break
        }
        result = append(result, C.GoString(cStr))
        ptr += unsafe.Sizeof(cArray)
    }
    return result
}

func main() {} // Required for c-shared
```

**Process actor (HSM-driven):**

Each tracked process shall be modeled as an HSM actor:

```go
// internal/process/actor.go
package process

import (
    "github.com/stateforward/hsm-go"
    "github.com/stateforward/hsm-go/muid"
)

// ProcessActor represents a tracked process as an HSM actor
type ProcessActor struct {
    ID        muid.ID
    PID       int
    AgentID   muid.ID
    Command   string
    Args      []string
    Pipes     *ProcessPipes
    StartedAt time.Time
    PTY       *os.File
}

var ProcessModel = hsm.Define("process",
    hsm.State("starting",
        hsm.Entry(func(p *ProcessActor) {
            p.setupPipes()
        }),
    ),
    hsm.State("running",
        hsm.Entry(func(p *ProcessActor) {
            go p.proxyStdout()
            go p.proxyStderr()
            go p.proxyStdin()
        }),
        hsm.Exit(func(p *ProcessActor) {
            p.closePipes()
        }),
    ),
    hsm.State("completed", hsm.Final()),
    hsm.State("failed", hsm.Final()),
    hsm.State("killed", hsm.Final()),

    hsm.Transition(hsm.On("exec.success"), hsm.Source("starting"), hsm.Target("running")),
    hsm.Transition(hsm.On("exec.failed"), hsm.Source("starting"), hsm.Target("failed")),
    hsm.Transition(hsm.On("exit.success"), hsm.Source("running"), hsm.Target("completed"),
        hsm.Effect(func(p *ProcessActor, e hsm.Event) {
            exit := e.Data.(ProcessExitEvent)
            hsm.Dispatch(ctx, p.AgentID, hsm.Event{
                Name: "process.completed",
                Data: exit,
            })
        }),
    ),
    hsm.Transition(hsm.On("exit.failed"), hsm.Source("running"), hsm.Target("failed"),
        hsm.Effect(func(p *ProcessActor, e hsm.Event) {
            exit := e.Data.(ProcessExitEvent)
            hsm.Dispatch(ctx, p.AgentID, hsm.Event{
                Name: "process.failed",
                Data: exit,
            })
        }),
    ),
    hsm.Transition(hsm.On("signal.received"), hsm.Source("running"), hsm.Target("killed"),
        hsm.Effect(func(p *ProcessActor, e hsm.Event) {
            exit := e.Data.(ProcessExitEvent)
            hsm.Dispatch(ctx, p.AgentID, hsm.Event{
                Name: "process.killed",
                Data: exit,
            })
        }),
    ),

    hsm.Initial(hsm.Target("starting")),
)

func (p *ProcessActor) proxyStdout() {
    buf := make([]byte, 64*1024)
    for {
        n, err := p.Pipes.TrackerStdoutRead.Read(buf)
        if err != nil {
            return
        }
        if n > 0 {
            hsm.Dispatch(ctx, p.AgentID, hsm.Event{
                Name: "process.stdout",
                Data: ProcessIOEvent{
                    PID:       p.PID,
                    AgentID:   p.AgentID,
                    ProcessID: p.ID,
                    Stream:    StreamStdout,
                    Data:      buf[:n],
                    Timestamp: time.Now(),
                },
            })
            // Forward to PTY
            p.PTY.Write(buf[:n])
        }
    }
}
```

**Process tracker (HSM-driven):**

The tracker itself is an HSM actor managing all process actors:

```go
// internal/process/tracker.go
package process

var TrackerModel = hsm.Define("process.tracker",
    hsm.State("idle"),
    hsm.State("listening",
        hsm.Entry(func(t *Tracker) {
            go t.acceptConnections()
        }),
    ),

    hsm.Transition(hsm.On("start"), hsm.Source("idle"), hsm.Target("listening")),
    hsm.Transition(hsm.On("hook.exec_pre"), hsm.Source("listening"), hsm.Target("listening"),
        hsm.Effect(func(t *Tracker, e hsm.Event) {
            req := e.Data.(HookExecPreRequest)
            t.handleExecPre(req)
        }),
    ),
    hsm.Transition(hsm.On("hook.exit"), hsm.Source("listening"), hsm.Target("listening"),
        hsm.Effect(func(t *Tracker, e hsm.Event) {
            notify := e.Data.(HookExitNotify)
            t.handleExit(notify)
        }),
    ),

    hsm.Initial(hsm.Target("idle")),
)

type Tracker struct {
    ID        muid.ID
    AgentID   muid.ID
    Socket    *net.UnixListener
    Processes map[muid.ID]*ProcessActor
    PTY       *os.File
    Config    ProcessCaptureConfig
}

func (t *Tracker) handleExecPre(req HookExecPreRequest) {
    processID := muid.New()

    // Create pipe pairs
    stdinR, stdinW, _ := os.Pipe()
    stdoutR, stdoutW, _ := os.Pipe()
    stderrR, stderrW, _ := os.Pipe()

    // Create process actor
    actor := &ProcessActor{
        ID:      processID,
        PID:     req.PID,
        AgentID: t.AgentID,
        Command: req.Command,
        Args:    req.Args,
        Pipes: &ProcessPipes{
            ProcessID:         processID,
            TrackerStdinWrite: stdinW,
            TrackerStdoutRead: stdoutR,
            TrackerStderrRead: stderrR,
        },
        StartedAt: time.Now(),
        PTY:       t.PTY,
    }

    // Register and start process actor HSM
    hsm.Start(ctx, ProcessModel, actor)
    t.Processes[processID] = actor

    // Send response with child-side FDs via SCM_RIGHTS
    resp := HookExecPreResponse{
        ProcessID: processID,
        Capture:   t.Config.Mode,
    }
    childFDs := []*os.File{stdinR, stdoutW, stderrW}
    t.sendWithFDs(req.Conn, resp, childFDs)

    // Close child-side FDs (now owned by child)
    stdinR.Close()
    stdoutW.Close()
    stderrW.Close()

    // Dispatch exec success to process actor
    hsm.Dispatch(ctx, processID, hsm.Event{Name: "exec.success"})

    // Emit spawned event to agent
    hsm.Dispatch(ctx, t.AgentID, hsm.Event{
        Name: "process.spawned",
        Data: ProcessSpawnedEvent{
            PID:       req.PID,
            AgentID:   t.AgentID,
            ProcessID: processID,
            Command:   req.Command,
            Args:      req.Args,
            WorkDir:   req.WorkDir,
            StartedAt: actor.StartedAt,
        },
    })
}

func (t *Tracker) handleExit(notify HookExitNotify) {
    actor, ok := t.Processes[notify.ProcessID]
    if !ok {
        return
    }

    exitEvent := ProcessExitEvent{
        PID:       actor.PID,
        AgentID:   actor.AgentID,
        ProcessID: actor.ID,
        Command:   actor.Command,
        ExitCode:  notify.ExitCode,
        StartedAt: actor.StartedAt,
        EndedAt:   time.Now(),
        Duration:  time.Since(actor.StartedAt),
    }

    if notify.Signal != 0 {
        exitEvent.Signal = &notify.Signal
        hsm.Dispatch(ctx, notify.ProcessID, hsm.Event{
            Name: "signal.received",
            Data: exitEvent,
        })
    } else if notify.ExitCode == 0 {
        hsm.Dispatch(ctx, notify.ProcessID, hsm.Event{
            Name: "exit.success",
            Data: exitEvent,
        })
    } else {
        hsm.Dispatch(ctx, notify.ProcessID, hsm.Event{
            Name: "exit.failed",
            Data: exitEvent,
        })
    }

    delete(t.Processes, notify.ProcessID)
}
```

#### 8.3.5 Environment inheritance

The hook mechanism relies on environment variable inheritance:

```
amux spawns PTY shell with:
  LD_PRELOAD=/tmp/amux-hook.so
  AMUX_HOOK_SOCKET=/tmp/amux-session-{id}.sock
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│ PTY Shell (bash)                                        │
│   env: LD_PRELOAD=..., AMUX_HOOK_SOCKET=...            │
│                                                         │
│   └─▶ claude (agent process)                           │
│         inherits env, hook loaded                       │
│                                                         │
│         └─▶ cargo build                                │
│               inherits env, hook loaded                 │
│               FDs redirected to tracker pipes           │
│                                                         │
│               └─▶ rustc                                │
│                     inherits env, hook loaded           │
│                     FDs redirected to tracker pipes     │
└─────────────────────────────────────────────────────────┘
```

Every process in the tree:
1. Inherits `LD_PRELOAD` and `AMUX_HOOK_SOCKET` from parent
2. Dynamic linker loads hook library before `main()`
3. When process calls `execve()` for a child, hook intercepts
4. Hook contacts tracker, receives pipe FDs, redirects, then execs

This provides complete coverage of all dynamically-linked processes spawned by the agent.

#### 8.3.6 Fallback: polling-based detection

When exec hooking is unavailable (SIP, static binaries, restricted environments), the tracker shall fall back to polling:

**Linux (`/proc` filesystem):**

```go
func (t *ProcessTracker) pollLinux(rootPID int, interval time.Duration) {
    for {
        children := readProcChildren(rootPID)  // /proc/{pid}/task/{tid}/children
        for _, pid := range children {
            if !t.known[pid] {
                info := readProcInfo(pid)  // /proc/{pid}/comm, cmdline, stat
                t.emitSpawned(pid, info)
            }
        }
        // Detect exits via missing /proc/{pid}
        time.Sleep(interval)
    }
}
```

**macOS (`sysctl`/`libproc`):**

```go
func (t *ProcessTracker) pollDarwin(rootPID int, interval time.Duration) {
    for {
        children := listAllProcs()  // proc_listallpids + proc_pidinfo
        tree := buildTree(children)
        descendants := tree.descendants(rootPID)
        // Compare with known set, emit events
        time.Sleep(interval)
    }
}
```

Polling limitations:
- Cannot capture I/O streams (only PTY-level observation)
- May miss short-lived processes between polls
- Higher latency for event emission

#### 8.3.7 I/O attribution

When using polling fallback, I/O cannot be attributed to specific processes. The system shall use heuristics:

```go
type IOAttribution struct {
    Method     AttributionMethod
    Confidence float64  // 0.0 - 1.0
    ProcessID  *muid.ID // nil if unknown
}

type AttributionMethod int

const (
    AttrExact    AttributionMethod = iota  // From hooked pipes (confidence: 1.0)
    AttrTiming                              // Based on process activity timing
    AttrPattern                             // Based on output patterns
    AttrUnknown                             // Cannot attribute
)
```

Timing-based attribution:
1. Track when each process is running vs waiting
2. Correlate PTY output bursts with active processes
3. Assign confidence based on timing overlap

Pattern-based attribution:
1. Adapters may provide output patterns for known tools (e.g., compiler errors, test output)
2. Match PTY output against patterns to identify source process

#### 8.3.8 Process tree tracking

The tracker shall maintain the full process hierarchy:

```go
type ProcessNode struct {
    Process   *Process
    Parent    *ProcessNode
    Children  []*ProcessNode
    Depth     int  // Distance from PTY shell
}

type ProcessTree struct {
    Root      *ProcessNode        // PTY shell
    ByPID     map[int]*ProcessNode
    ByID      map[muid.ID]*ProcessNode
}

func (t *ProcessTree) Spawn(parent int, child *Process) *ProcessNode
func (t *ProcessTree) Exit(pid int, exitCode int, signal *int)
func (t *ProcessTree) Descendants(pid int) []*ProcessNode
func (t *ProcessTree) Ancestors(pid int) []*ProcessNode
```

The tree enables:
- Determining if an agent is blocked on a subprocess
- Aggregating resource usage per agent
- Killing process groups on agent termination
- Attributing nested process output

#### 8.3.9 Hook library compilation

The hook library shall be compiled as a Go shared library using `-buildmode=c-shared` for each target platform:

```bash
# Linux x86_64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -buildmode=c-shared -o hooks/bin/amux-hook-linux-amd64.so ./hooks

# Linux ARM64
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc \
  go build -buildmode=c-shared -o hooks/bin/amux-hook-linux-arm64.so ./hooks

# macOS x86_64
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
  go build -buildmode=c-shared -o hooks/bin/amux-hook-darwin-amd64.dylib ./hooks

# macOS ARM64
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
  go build -buildmode=c-shared -o hooks/bin/amux-hook-darwin-arm64.dylib ./hooks
```

Note: Cross-compilation of c-shared libraries requires CGO and appropriate cross-compilers. For CI/CD, use platform-native builds or Docker containers with the target toolchain.

The Go binary shall embed and extract the appropriate library at runtime:

```go
//go:embed hooks/bin/amux-hook-linux-amd64.so
var hookLinuxAMD64 []byte

//go:embed hooks/bin/amux-hook-linux-arm64.so
var hookLinuxARM64 []byte

//go:embed hooks/bin/amux-hook-darwin-amd64.dylib
var hookDarwinAMD64 []byte

//go:embed hooks/bin/amux-hook-darwin-arm64.dylib
var hookDarwinARM64 []byte

func extractHookLibrary(dir string) (string, error) {
    var data []byte
    var name string

    switch runtime.GOOS + "/" + runtime.GOARCH {
    case "linux/amd64":
        data, name = hookLinuxAMD64, "amux-hook.so"
    case "linux/arm64":
        data, name = hookLinuxARM64, "amux-hook.so"
    case "darwin/amd64":
        data, name = hookDarwinAMD64, "amux-hook.dylib"
    case "darwin/arm64":
        data, name = hookDarwinARM64, "amux-hook.dylib"
    default:
        return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
    }

    path := filepath.Join(dir, name)
    if err := os.WriteFile(path, data, 0755); err != nil {
        return "", err
    }
    return path, nil
}
```

### 8.4 Process events

#### 8.4.1 Lifecycle events

| Event | Payload | Description |
|-------|---------|-------------|
| `process.spawned` | ProcessSpawnedEvent | A new child process was detected |
| `process.completed` | ProcessExitEvent | Process exited successfully (exit code 0) |
| `process.failed` | ProcessExitEvent | Process exited with non-zero exit code |
| `process.killed` | ProcessExitEvent | Process was terminated by signal |

```go
type ProcessSpawnedEvent struct {
    PID       int
    AgentID   muid.ID
    ProcessID muid.ID
    Command   string
    Args      []string
    WorkDir   string
    Env       map[string]string  // Relevant environment variables
    ParentPID int
    StartedAt time.Time
}

type ProcessExitEvent struct {
    PID       int
    AgentID   muid.ID
    ProcessID muid.ID
    Command   string
    ExitCode  int
    Signal    *int       // nil if not killed by signal
    StartedAt time.Time
    EndedAt   time.Time
    Duration  time.Duration
}
```

#### 8.4.2 I/O stream events

The process tracker shall emit events for all I/O activity on tracked processes:

| Event | Payload | Description |
|-------|---------|-------------|
| `process.stdout` | ProcessIOEvent | Data written to stdout |
| `process.stderr` | ProcessIOEvent | Data written to stderr |
| `process.stdin` | ProcessIOEvent | Data written to stdin |

```go
type ProcessIOEvent struct {
    PID       int
    AgentID   muid.ID
    ProcessID muid.ID
    Command   string     // Command associated with this PID/ProcessID at time of emission
    Stream    IOStream
    Data      []byte
    Timestamp time.Time
}

type IOStream int

const (
    StreamStdin  IOStream = 0
    StreamStdout IOStream = 1
    StreamStderr IOStream = 2
)
```

#### 8.4.3 Event batching and coalescing

To prevent thrashing and avoid interrupting agents with high-frequency notifications, the system shall batch and coalesce events:

##### 8.4.3.1 Batching strategy

```go
type EventBatcher struct {
    Window      time.Duration  // Coalesce events within this window (default: 50ms)
    MaxBatch    int            // Maximum events per batch (default: 100)
    MaxBytes    int            // Maximum bytes per batch for I/O events (default: 64KB)
    FlushOnIdle time.Duration  // Flush if no new events for this duration (default: 10ms)
}

type BatchedEvents struct {
    Events    []hsm.Event
    StartTime time.Time
    EndTime   time.Time
    Count     int
}
```

##### 8.4.3.2 Coalescing rules

Events shall be coalesced according to these rules:

| Event Type | Coalescing Behavior |
|------------|---------------------|
| `process.stdout` | Concatenate data within window, emit single event |
| `process.stderr` | Concatenate data within window, emit single event |
| `process.stdin` | Concatenate data within window, emit single event |
| `process.spawned` | No coalescing (each spawn is distinct) |
| `process.completed` | No coalescing (each exit is distinct) |
| `presence.changed` | Keep only latest state per agent |
| `activity.detected` | Deduplicate within window |
| `config.updated` | No semantic coalescing (config actor emits per-key updates); events MAY be delivered together inside `events.batch` |

##### 8.4.3.3 I/O event coalescing

I/O events for the same process and stream shall be coalesced:

```go
type CoalescedIOEvent struct {
    ProcessIOEvent
    Chunks     int       // Number of original events coalesced
    FirstSeen  time.Time // Timestamp of first chunk
    LastSeen   time.Time // Timestamp of last chunk
}

func (b *EventBatcher) CoalesceIO(events []ProcessIOEvent) CoalescedIOEvent {
    var data []byte
    for _, e := range events {
        data = append(data, e.Data...)
    }
    return CoalescedIOEvent{
        ProcessIOEvent: ProcessIOEvent{
            PID:       events[0].PID,
            AgentID:   events[0].AgentID,
            ProcessID: events[0].ProcessID,
            Command:   events[0].Command,
            Stream:    events[0].Stream,
            Data:      data,
            Timestamp: events[len(events)-1].Timestamp,
        },
        Chunks:    len(events),
        FirstSeen: events[0].Timestamp,
        LastSeen:  events[len(events)-1].Timestamp,
    }
}
```

##### 8.4.3.4 Batcher HSM

The event batcher shall be modeled as an HSM actor:

```go
var BatcherModel = hsm.Define("event.batcher",
    hsm.State("collecting",
        hsm.Entry(func(b *EventBatcher) {
            b.startTimer()
        }),
    ),
    hsm.State("flushing",
        hsm.Entry(func(b *EventBatcher) {
            b.flushBatch()
        }),
    ),

    // Collect events until window expires or batch full
    hsm.Transition(hsm.On("event.received"), hsm.Source("collecting"), hsm.Target("collecting"),
        hsm.Effect(func(b *EventBatcher, e hsm.Event) {
            b.addToBatch(e)
        }),
        hsm.Guard(func(b *EventBatcher) bool {
            return !b.batchFull()
        }),
    ),
    // Batch full → flush immediately
    hsm.Transition(hsm.On("event.received"), hsm.Source("collecting"), hsm.Target("flushing"),
        hsm.Guard(func(b *EventBatcher) bool {
            return b.batchFull()
        }),
    ),
    // Window expired → flush
    hsm.Transition(hsm.On("window.expired"), hsm.Source("collecting"), hsm.Target("flushing")),
    // Idle timeout → flush partial batch
    hsm.Transition(hsm.On("idle.timeout"), hsm.Source("collecting"), hsm.Target("flushing")),
    // After flush → back to collecting
    hsm.Transition(hsm.On("flush.complete"), hsm.Source("flushing"), hsm.Target("collecting")),

    hsm.Initial(hsm.Target("collecting")),
)
```

##### 8.4.3.5 Agent-aware batching

When dispatching to agents, batched events shall be delivered atomically to avoid partial state:

```go
func (b *EventBatcher) dispatchToAgent(agentID muid.ID, batch BatchedEvents) {
    // Wrap batch in single event for atomic delivery
    hsm.Dispatch(ctx, agentID, hsm.Event{
        Name: "events.batch",
        Data: batch,
    })
}
```

Agents may handle batched events:

```go
hsm.Transition(hsm.On("events.batch"), hsm.Source("*"), hsm.Target("*"),
    hsm.Effect(func(a *Agent, e hsm.Event) {
        batch := e.Data.(BatchedEvents)
        for _, event := range batch.Events {
            a.processEvent(event)
        }
    }),
)
```

##### 8.4.3.6 LLM-gated notifications

The system may optionally use a local CPU-based LLM to intelligently gate notifications, filtering noise and prioritizing important events before they reach agents.

When `events.gate.enabled = true`, the implementation MUST use the `liquidgen` inference engine (see §4.2.10).

**Rationale:**
- Time-based batching is deterministic but unintelligent
- Many events (verbose build output, repetitive logs) are noise
- Important events (errors, completions, user messages) should interrupt immediately
- A small local LLM can make these decisions with low latency

**Architecture:**

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────┐
│ Event       │────▶│ LLM Gate        │────▶│ Agent        │
│ Batcher     │     │ (CPU inference) │     │              │
└─────────────┘     └─────────────────┘     └──────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Filtered/    │
                    │ Summarized   │
                    └──────────────┘
```

**Gate HSM:**

```go
var NotificationGateModel = hsm.Define("notification.gate",
    hsm.State("idle"),
    hsm.State("evaluating",
        hsm.Entry(func(g *NotificationGate) {
            go g.runInference()
        }),
    ),
    hsm.State("dispatching",
        hsm.Entry(func(g *NotificationGate) {
            g.dispatchFiltered()
        }),
    ),

    hsm.Transition(hsm.On("batch.received"), hsm.Source("idle"), hsm.Target("evaluating")),
    hsm.Transition(hsm.On("inference.complete"), hsm.Source("evaluating"), hsm.Target("dispatching")),
    hsm.Transition(hsm.On("dispatch.complete"), hsm.Source("dispatching"), hsm.Target("idle")),

    // Bypass gate for high-priority events
    hsm.Transition(hsm.On("priority.event"), hsm.Source("*"), hsm.Target("dispatching"),
        hsm.Effect(func(g *NotificationGate, e hsm.Event) {
            g.priorityQueue = append(g.priorityQueue, e)
        }),
    ),

    hsm.Initial(hsm.Target("idle")),
)

type NotificationGate struct {
    ID            muid.ID
    Model         LLMModel
    AgentContext  string          // Current agent task context
    PriorityQueue []hsm.Event     // Bypass queue for urgent events
    FilteredBatch BatchedEvents   // Result after LLM filtering
}
```

**LLM interface:**

```go
type LLMModel interface {
    // Evaluate batch and return filtered/prioritized events
    FilterEvents(ctx context.Context, req FilterRequest) (*FilterResponse, error)
}

type FilterRequest struct {
    AgentID      muid.ID
    AgentTask    string          // Current task description
    AgentState   string          // Current presence/lifecycle state
    Events       BatchedEvents   // Incoming event batch
    MaxTokens    int             // Limit response size
}

type FilterResponse struct {
    PassThrough  []hsm.Event     // Events to deliver immediately
    Summarized   *EventSummary   // Optional summary of filtered events
    Suppressed   int             // Count of suppressed events
    Reason       string          // LLM's reasoning (for debugging)
}

type EventSummary struct {
    Text       string            // Human-readable summary
    Highlights []string          // Key points extracted
    Sentiment  string            // error, warning, info, success
}
```

**Supported local models (liquidgen):**

| Model | Modality | CPU Throughput | Use Case |
|-------|----------|----------------|----------|
| lfm2.5-thinking (quantized) | Text | 50–100 tps | Fast filtering, prioritization, and summarization |
| lfm2.5-VL (quantized) | Vision-language | 50–100 tps | Filtering that incorporates visual context (for example screenshots) |

Implementations MUST support both models listed above and MUST run them via `liquidgen` on CPU.

**Integration with liquidgen:**

```go
type LiquidgenModel struct {
    engine LiquidgenEngine
    model  string // "lfm2.5-thinking" or "lfm2.5-VL" (quantized variant)
}

func (m *LiquidgenModel) FilterEvents(ctx context.Context, req FilterRequest) (*FilterResponse, error) {
    prompt := m.buildPrompt(req)

    stream, err := m.engine.Generate(ctx, LiquidgenRequest{
        Model:       m.model,
        Prompt:      prompt,
        MaxTokens:   req.MaxTokens,
        Temperature: 0.1,  // Low temperature for consistent filtering
    })
    if err != nil {
        return nil, err
    }
    defer stream.Close()

    // Collect the streamed response into a single UTF-8 string.
    var out strings.Builder
    for {
        tok, err := stream.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        out.WriteString(tok)
    }

    return m.parseResponse(out.String(), req.Events)
}
```


**Prompt template:**

```go
const filterPromptTemplate = `You are a notification filter for a coding agent.

Current agent: {{.AgentID}}
Current task: {{.AgentTask}}
Agent state: {{.AgentState}}

Incoming events ({{.EventCount}} total):
{{range .Events}}
- [{{.Name}}] {{.Summary}}
{{end}}

Decide which events should interrupt the agent:
- PASS: Errors, task completions, user messages, important state changes
- SUPPRESS: Verbose output, routine progress, redundant notifications
- SUMMARIZE: Multiple related events that can be condensed

Respond in JSON:
{
  "pass": [event indices to pass through],
  "summarize": {"indices": [...], "summary": "..."},
  "suppress": [event indices to suppress],
  "reason": "brief explanation"
}`
```

**Priority bypass:**

Certain events shall bypass the LLM gate entirely:

```go
var PriorityEvents = map[string]bool{
    "process.failed":      true,  // Errors always pass
    "rate.limit":          true,  // Rate limits are urgent
    "message.inbound":     true,  // User messages always pass
    "shutdown.request":    true,  // Shutdown is critical
    "task.assigned":       true,  // New tasks always pass
    "connection.lost":     true,  // Connection issues are urgent
}

func (g *NotificationGate) shouldBypass(event hsm.Event) bool {
    return PriorityEvents[event.Name]
}
```

**Fallback behavior:**

If LLM inference fails or times out:

```go
func (g *NotificationGate) runInference() {
    ctx, cancel := context.WithTimeout(context.Background(), g.config.InferenceTimeout)
    defer cancel()

    resp, err := g.Model.FilterEvents(ctx, g.buildRequest())
    if err != nil || ctx.Err() != nil {
        // Fallback: pass all events through unfiltered
        log.Warn("LLM gate failed, passing all events", "error", err)
        g.FilteredBatch = g.pendingBatch
        hsm.Dispatch(ctx, g.ID, hsm.Event{Name: "inference.complete"})
        return
    }

    g.FilteredBatch = g.applyFilter(resp)
    hsm.Dispatch(ctx, g.ID, hsm.Event{Name: "inference.complete"})
}
```

##### 8.4.3.7 MCP notification subscriptions

The system shall expose an MCP (Model Context Protocol) server that allows agents to subscribe to notifications based on semantic similarity using embeddings.

**Transport:**
- When `events.subscriptions.enabled = true`, the director shall run the notification MCP server and listen on a Unix domain stream socket at `events.subscriptions.socket_path`.
- The server shall accept multiple concurrent client connections.

**Framing:**
- The MCP server transport shall use newline-delimited UTF-8 JSON-RPC 2.0 messages (one JSON object per line).
- Requests and responses shall use standard JSON-RPC 2.0 envelopes (`jsonrpc`, `id`, `method`, `params`, `result`, `error`).
- Notifications sent from server to clients shall be JSON-RPC notifications (no `id`).

**Rationale:**
- Pattern matching is brittle and requires exact knowledge of output formats
- Semantic subscriptions let agents describe *what* they care about conceptually
- "notify me about test failures" works regardless of test framework output format
- Embeddings enable fuzzy matching across different phrasings of the same concept

**Architecture:**

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Process I/O │────▶│ Embedding Engine │────▶│ Subscription    │
│ Events      │     │ (Gemma/CPU)      │     │ Matcher         │
└─────────────┘     └──────────────────┘     └─────────────────┘
                                                     │
                    ┌────────────────────────────────┘
                    ▼
            ┌───────────────┐     MCP      ┌─────────────┐
            │ Notification  │◀────────────▶│ Agent       │
            │ MCP Server    │              │ (subscriber)│
            └───────────────┘              └─────────────┘
```

**MCP server definition:**

```go
type NotificationMCPServer struct {
    ID       muid.ID
    Embedder EmbeddingModel

    // Persistent storage for subscriptions and query embeddings.
    // The SQLite database is the source of truth (see "Storage backend" below).
    DB *sql.DB
}

type Subscription struct {
    ID            muid.ID
    AgentID        muid.ID
    Queries        []string          // Subscription queries
    Threshold      float32           // Similarity threshold (0.0-1.0)
    Streams        []IOStream        // Which streams to monitor
    ProcessFilter  *ProcessFilter    // Optional: filter by command pattern
    Cooldown       time.Duration     // Min time between notifications
    LastNotify     time.Time
}

type ProcessFilter struct {
    CommandPattern *regexp.Regexp  // Match process command
    ExcludePattern *regexp.Regexp  // Exclude certain processes
}
```

**MCP tools exposed:**

```json
{
  "name": "subscribe_notifications",
  "description": "Subscribe to process output notifications based on semantic similarity",
  "inputSchema": {
    "type": "object",
    "properties": {
      "queries": {
        "type": "array",
        "items": {"type": "string"},
        "description": "List of semantic queries to match against (e.g., ['error', 'test failed', 'compilation warning'])"
      },
      "threshold": {
        "type": "number",
        "minimum": 0,
        "maximum": 1,
        "default": 0.7,
        "description": "Similarity threshold (0.0-1.0). Higher = more selective"
      },
      "streams": {
        "type": "array",
        "items": {"enum": ["stdout", "stderr", "stdin"]},
        "default": ["stdout", "stderr"],
        "description": "Which I/O streams to monitor"
      },
      "command_pattern": {
        "type": "string",
        "description": "Optional regex to filter by process command (e.g., 'cargo|npm|go')"
      },
      "cooldown_ms": {
        "type": "integer",
        "default": 1000,
        "description": "Minimum milliseconds between notifications for this subscription"
      }
    },
    "required": ["queries"]
  }
}

{
  "name": "unsubscribe_notifications",
  "description": "Remove a notification subscription",
  "inputSchema": {
    "type": "object",
    "properties": {
      "subscription_id": {
        "type": "string",
        "description": "ID of the subscription to remove"
      }
    },
    "required": ["subscription_id"]
  }
}

{
  "name": "list_subscriptions",
  "description": "List active notification subscriptions",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

The MCP tools listed in this section shall be exposed as JSON-RPC methods whose names exactly match the tool `name` fields:

- `subscribe_notifications`
- `unsubscribe_notifications`
- `list_subscriptions`

**Identifier format:** `subscription_id` SHALL be a string of the form:

- `sub_` + base-10 encoding of the underlying `muid.ID`

The server SHALL accept only this form and SHALL reject any other value as invalid params.

**Stream mapping:** Values in the `streams` parameter SHALL map as follows:

- `"stdin"` → `StreamStdin`
- `"stdout"` → `StreamStdout`
- `"stderr"` → `StreamStderr`

**Cooldown units:** `cooldown_ms` SHALL be interpreted as milliseconds and converted to `time.Duration` via `time.Millisecond * cooldown_ms`.

**Tool results:** The JSON-RPC `result` payloads SHALL be:

- `subscribe_notifications`: `{ "subscription_id": "<id>" }`
- `unsubscribe_notifications`: `{ "removed": true }` if removed, `{ "removed": false }` if not found
- `list_subscriptions`: `{ "subscriptions": [ ... ] }` where each entry includes at least:
  - `subscription_id` (string, format above)
  - `queries` (array of strings)
  - `threshold` (number)
  - `streams` (array of strings as above)
  - `cooldown_ms` (integer)

A client invokes a tool by sending a JSON-RPC request with `method` set to the tool name and `params` validated against the tool’s `inputSchema`.

**Embedding model interface (ONNX Runtime):**

The system shall use ONNX Runtime for embedding inference, providing optimized CPU execution with minimal dependencies.

```go
import "github.com/yalue/onnxruntime_go"

type EmbeddingModel interface {
    // Embed text into vector space
    Embed(ctx context.Context, text string) ([]float32, error)

    // Batch embed for efficiency
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

    // Vector dimension
    Dimension() int

    // Close and release resources
    Close() error
}

// ONNX-based embedding model
type ONNXEmbedder struct {
    session   *onnxruntime.Session
    tokenizer *Tokenizer
    dimension int
    maxLength int
}

func NewONNXEmbedder(modelPath string, tokenizerPath string) (*ONNXEmbedder, error) {
    // Initialize ONNX Runtime
    onnxruntime.SetSharedLibraryPath(getONNXLibPath())
    if err := onnxruntime.InitializeEnvironment(); err != nil {
        return nil, fmt.Errorf("failed to initialize ONNX: %w", err)
    }

    // Create session with CPU execution provider
    opts, _ := onnxruntime.NewSessionOptions()
    opts.SetIntraOpNumThreads(runtime.NumCPU())
    opts.SetInterOpNumThreads(1)
    opts.SetGraphOptimizationLevel(onnxruntime.GraphOptLevelAll)

    session, err := onnxruntime.NewSession(modelPath, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to load model: %w", err)
    }

    tokenizer, err := LoadTokenizer(tokenizerPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load tokenizer: %w", err)
    }

    return &ONNXEmbedder{
        session:   session,
        tokenizer: tokenizer,
        dimension: 384,  // Model-specific
        maxLength: 512,
    }, nil
}

func (e *ONNXEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
    // Tokenize input
    tokens := e.tokenizer.Encode(text, e.maxLength)

    // Create input tensors
    inputIDs, _ := onnxruntime.NewTensor(tokens.InputIDs)
    attentionMask, _ := onnxruntime.NewTensor(tokens.AttentionMask)

    // Run inference
    outputs, err := e.session.Run([]onnxruntime.Value{inputIDs, attentionMask})
    if err != nil {
        return nil, err
    }
    defer outputs[0].Destroy()

    // Extract embeddings and mean pool
    embeddings := outputs[0].GetData().([]float32)
    pooled := meanPool(embeddings, tokens.AttentionMask, e.dimension)

    // L2 normalize
    return normalize(pooled), nil
}

func (e *ONNXEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
    results := make([][]float32, len(texts))
    // Batch tokenization and inference for efficiency
    batchTokens := e.tokenizer.EncodeBatch(texts, e.maxLength)

    inputIDs, _ := onnxruntime.NewTensor(batchTokens.InputIDs)
    attentionMask, _ := onnxruntime.NewTensor(batchTokens.AttentionMask)

    outputs, err := e.session.Run([]onnxruntime.Value{inputIDs, attentionMask})
    if err != nil {
        return nil, err
    }
    defer outputs[0].Destroy()

    // Extract and pool each embedding
    allEmbeddings := outputs[0].GetData().([]float32)
    for i := range texts {
        offset := i * e.dimension
        pooled := meanPool(allEmbeddings[offset:offset+e.dimension], batchTokens.AttentionMask[i], e.dimension)
        results[i] = normalize(pooled)
    }
    return results, nil
}
```

**Supported ONNX embedding models:**

| Model | Dimension | ONNX Size | CPU Latency | Use Case |
|-------|-----------|-----------|-------------|----------|
| all-MiniLM-L6-v2 | 384 | ~23MB | ~2ms | Fast, general purpose |
| bge-small-en-v1.5 | 384 | ~33MB | ~3ms | High quality, English |
| nomic-embed-text-v1 | 768 | ~65MB | ~5ms | Longer context |
| gte-small | 384 | ~33MB | ~3ms | Good multilingual |

Models shall be stored in ONNX format with INT8 quantization for optimal CPU performance:

```bash
# Convert HuggingFace model to ONNX with quantization
python -m optimum.exporters.onnx \
    --model sentence-transformers/all-MiniLM-L6-v2 \
    --optimize O3 \
    --quantize \
    ./models/all-minilm-l6/
```

**Tokenizer:**

```go
// Using a pure-Go tokenizer (no Python dependency)
import "github.com/pkoukk/tiktoken-go"

type Tokenizer struct {
    vocab    map[string]int
    merges   []string
    maxLen   int
}

func (t *Tokenizer) Encode(text string, maxLength int) TokenizedInput {
    // BPE tokenization matching the ONNX model's expected format
    tokens := t.tokenize(text)
    if len(tokens) > maxLength {
        tokens = tokens[:maxLength]
    }

    // Pad to maxLength
    inputIDs := make([]int64, maxLength)
    attentionMask := make([]int64, maxLength)
    for i, tok := range tokens {
        inputIDs[i] = int64(tok)
        attentionMask[i] = 1
    }

    return TokenizedInput{
        InputIDs:      inputIDs,
        AttentionMask: attentionMask,
    }
}
```

**ONNX Runtime library embedding:**

The ONNX Runtime shared library shall be embedded and extracted at runtime:

```go
//go:embed onnxruntime/libonnxruntime-linux-amd64.so
var onnxLibLinuxAMD64 []byte

//go:embed onnxruntime/libonnxruntime-linux-arm64.so
var onnxLibLinuxARM64 []byte

//go:embed onnxruntime/libonnxruntime-darwin-amd64.dylib
var onnxLibDarwinAMD64 []byte

//go:embed onnxruntime/libonnxruntime-darwin-arm64.dylib
var onnxLibDarwinARM64 []byte

func extractONNXRuntime(dir string) (string, error) {
    // Extract appropriate library for platform
    // Similar to hook library extraction
}
```

**Similarity matching:**

```go
func (s *NotificationMCPServer) matchSubscription(
    sub *Subscription,
    event ProcessIOEvent,
) (*NotificationMatch, bool) {
    // Skip if on cooldown
    if time.Since(sub.LastNotify) < sub.Cooldown {
        return nil, false
    }

    // Check stream filter
    if !slices.Contains(sub.Streams, event.Stream) {
        return nil, false
    }

    // Check process filter
    if sub.ProcessFilter != nil {
        if !sub.ProcessFilter.CommandPattern.MatchString(event.Command) {
            return nil, false
        }
    }

    // Embed the output
    outputEmbed, err := s.Embedder.Embed(ctx, string(event.Data))
    if err != nil {
        return nil, false
    }

    // Find best matching query using sqlite-vec.
    // Implementations SHOULD perform a single vector query across all stored query embeddings
    // and then map the winning query back to its subscription.
    bestQuery, bestScore, err := s.bestQueryFromDB(ctx, sub.ID, outputEmbed)
    if err != nil {
        return nil, false
    }

    // Check threshold
    if bestScore < sub.Threshold {
        return nil, false
    }

    return &NotificationMatch{
        SubscriptionID: sub.ID,
        Query:          bestQuery,
        Score:          bestScore,
        Event:          event,
    }, true
}

func (s *NotificationMCPServer) bestQueryFromDB(
    ctx context.Context,
    subID muid.ID,
    outputEmbed []float32,
) (bestQuery string, bestScore float32, err error) {
    // Pseudocode: query sqlite-vec for the nearest stored query embedding for this subscription.
    // Implementations MUST store and query embeddings in SQLite (see "Storage backend" below).
    //
    // Example query shape (illustrative):
    //
    // SELECT q.query_text, v.distance
    // FROM subscription_query_vec v
    // JOIN subscription_queries q ON q.query_id = v.query_id
    // WHERE q.subscription_id = ?
    // ORDER BY v.distance ASC
    // LIMIT 1;
    //
    // If v.distance is cosine distance, similarity score can be computed as: score = 1.0 - distance.
    return "", 0, nil
}
```

**MCP notification delivery:**

```go
type NotificationMatch struct {
    SubscriptionID muid.ID
    Query          string
    Score          float32
    Event          ProcessIOEvent
    Timestamp      time.Time
}

// Delivered as MCP notification
{
    "jsonrpc": "2.0",
    "method": "notifications/process_output",
    "params": {
        "subscription_id": "sub_123",
        "matched_query": "test failed",
        "similarity_score": 0.85,
        "process": {
            "pid": 12345,
            "command": "cargo test",
            "stream": "stderr"
        },
        "content": "test result: FAILED. 2 passed; 1 failed;",
        "timestamp": "2026-01-18T10:30:00Z"
    }
}
```

**Notification schema (normative):**
- The server MUST deliver matches as JSON-RPC 2.0 **notifications** (no `id`) with `method` exactly `"notifications/process_output"`.
- `params` MUST be an object containing the following required fields:
  - `subscription_id` (string): MUST have the form `"sub_<ID>"` where `<ID>` is a base-10 `muid.ID` string.
  - `matched_query` (string): the subscription query that produced the best match.
  - `similarity_score` (number): MUST be in the inclusive range `[0.0, 1.0]`.
  - `process` (object) with required fields:
    - `pid` (integer)
    - `command` (string): MUST be derived by joining the process argv with a single ASCII space (`0x20`).
    - `stream` (string): MUST be one of `"stdin"`, `"stdout"`, or `"stderr"`.
  - `content` (string): MUST be a UTF-8 string. If the underlying process I/O bytes are not valid UTF-8, the server MUST replace invalid byte sequences with the Unicode replacement character (U+FFFD).
  - `timestamp` (string): MUST be an RFC 3339 UTC timestamp using `time.RFC3339Nano` formatting (see §9.1.3.1).

**Subscription HSM:**

```go
var SubscriptionModel = hsm.Define("notification.subscription",
    hsm.State("active"),
    hsm.State("cooldown",
        hsm.Entry(func(s *Subscription) {
            s.startCooldownTimer()
        }),
    ),
    hsm.State("paused"),
    hsm.State("removed", hsm.Final()),

    hsm.Transition(hsm.On("io.event"), hsm.Source("active"), hsm.Target("active"),
        hsm.Effect(func(s *Subscription, e hsm.Event) {
            s.checkMatch(e.Data.(ProcessIOEvent))
        }),
    ),
    hsm.Transition(hsm.On("match.found"), hsm.Source("active"), hsm.Target("cooldown"),
        hsm.Effect(func(s *Subscription, e hsm.Event) {
            s.deliverNotification(e.Data.(NotificationMatch))
        }),
    ),
    hsm.Transition(hsm.On("cooldown.expired"), hsm.Source("cooldown"), hsm.Target("active")),
    hsm.Transition(hsm.On("pause"), hsm.Source("active"), hsm.Target("paused")),
    hsm.Transition(hsm.On("resume"), hsm.Source("paused"), hsm.Target("active")),
    hsm.Transition(hsm.On("remove"), hsm.Source("*"), hsm.Target("removed")),

    hsm.Initial(hsm.Target("active")),
)
```

**Example agent usage:**

```
Agent: I want to be notified about any test failures or compilation errors
       while I work on other tasks.

→ Calls MCP tool:
  subscribe_notifications({
    "queries": [
      "test failed",
      "assertion error",
      "compilation error",
      "build failed",
      "panic",
      "undefined reference"
    ],
    "threshold": 0.75,
    "streams": ["stdout", "stderr"],
    "command_pattern": "cargo|go|npm|pytest",
    "cooldown_ms": 5000
  })

← Returns: {"subscription_id": "sub_123"}

... later, cargo test runs and fails ...

← MCP notification:
  {
    "method": "notifications/process_output",
    "params": {
      "subscription_id": "sub_123",
      "matched_query": "test failed",
      "similarity_score": 0.89,
      "content": "test auth::tests::login_invalid ... FAILED",
      ...
    }
  }
```

**Storage backend (normative): SQLite + sqlite-vec**

The notification MCP server MUST use SQLite as its persistent subscription store and MUST use sqlite-vec for vector similarity search.

- The SQLite database file path MUST be configurable via `events.subscriptions.db_path`.
- The database MUST be treated as the source of truth. The implementation MAY cache subscriptions in memory, but it MUST be able to rebuild all in-memory state from the SQLite database on startup.
- Query embeddings MUST be stored as `float32` vectors. Implementations MUST store embeddings in a sqlite-vec vector table and MUST NOT rely on an in-memory-only embedding index.

A minimal schema is:

```sql
-- IDs are base-10 strings (see "Identifier format" above)
CREATE TABLE IF NOT EXISTS subscriptions (
  subscription_id TEXT PRIMARY KEY,
  agent_id        TEXT NOT NULL,
  threshold       REAL NOT NULL,
  cooldown_ms     INTEGER NOT NULL,
  streams_json    TEXT NOT NULL,
  command_pattern TEXT,
  exclude_pattern TEXT,
  created_at      TEXT NOT NULL,
  last_notify_at  TEXT
);

CREATE TABLE IF NOT EXISTS subscription_queries (
  query_id        INTEGER PRIMARY KEY AUTOINCREMENT,
  subscription_id TEXT NOT NULL REFERENCES subscriptions(subscription_id) ON DELETE CASCADE,
  query_text      TEXT NOT NULL
);

-- sqlite-vec virtual table storing embeddings for each query_id.
-- The embedding dimension MUST match the configured embedding model output dimension.
CREATE VIRTUAL TABLE IF NOT EXISTS subscription_query_vec
USING vec0(
  query_id INTEGER PRIMARY KEY,
  embedding FLOAT[EMBED_DIM]
);
```

Vector query behavior (normative):

- When a Process I/O event is observed, the server MUST compute an embedding for the candidate text and MUST query `subscription_query_vec` for the nearest neighbor query embeddings (implementation-defined `k`, but SHOULD be at least 10).
- The server MUST map matching `query_id` values back to `subscription_id` via `subscription_queries` and MUST apply:
  - per-subscription `threshold` filtering, and
  - per-subscription `cooldown` filtering,
  before emitting any MCP notification.


##### 8.4.3.8 Configuration

```toml
[events]
batch_window = "50ms"       # Coalesce events within this window
batch_max_events = 100      # Maximum events per batch
batch_max_bytes = "64KB"    # Maximum bytes for I/O batches
batch_idle_flush = "10ms"   # Flush if no new events for this duration

[events.coalesce]
io_streams = true           # Coalesce stdout/stderr/stdin
presence = true             # Keep only latest presence per agent
activity = true             # Deduplicate activity events

[events.gate]
enabled = false             # Enable LLM-gated notifications
engine = "liquidgen"        # liquidgen
model = "lfm2.5-thinking"   # Quantized variant (default for gating)
model_vl = "lfm2.5-VL"      # Quantized variant (optional; for visual context)
inference_timeout = "500ms" # Max time for LLM inference
max_tokens = 256            # Max response tokens
temperature = 0.1           # Low for consistent filtering

[events.gate.bypass]
# Events that always bypass the gate
events = [
    "process.failed",
    "rate.limit",
    "message.inbound",
    "shutdown.request",
    "task.assigned",
    "connection.lost",
]

[events.subscriptions]
enabled = true              # Enable MCP subscription server
socket_path = "~/.amux/mcp.sock"  # Unix socket path for the notification MCP server
embedding_model = "all-MiniLM-L6-v2"  # ONNX embedding model
embedding_model_path = ""   # Path to .onnx file (auto-download if empty)
tokenizer_path = ""         # Path to tokenizer.json (bundled with model)
default_threshold = 0.7     # Default similarity threshold
default_cooldown = "1s"     # Default cooldown between notifications
max_subscriptions = 100     # Max subscriptions per agent
db_path = "~/.amux/subscriptions.sqlite3"  # SQLite DB path for subscriptions (sqlite-vec)
# sqlite-vec is required; the implementation MUST load it (see §8.4.3.7)
onnx_threads = 0            # ONNX intra-op threads (0 = auto)
```

#### 8.4.4 Stream capture modes

The process tracker shall support different capture modes per process:

```go
type CaptureMode int

const (
    CaptureNone   CaptureMode = 0           // No I/O capture
    CaptureStdout CaptureMode = 1 << iota   // Capture stdout only
    CaptureStderr                            // Capture stderr only
    CaptureStdin                             // Capture stdin only
    CaptureAll    = CaptureStdout | CaptureStderr | CaptureStdin
)

type ProcessCaptureConfig struct {
    Mode       CaptureMode
    BufferSize int  // Ring buffer size for each stream (default: 1MB)
}
```

#### 8.4.5 Event dispatch

Process events shall be dispatched via hsm-go's event queue:

```go
// Lifecycle event dispatch
hsm.Dispatch(ctx, agentID, hsm.Event{
    Name: "process.spawned",
    Data: ProcessSpawnedEvent{
        PID:       pid,
        AgentID:   agentID,
        Command:   cmd,
        Args:      args,
        StartedAt: time.Now(),
    },
})

// I/O event dispatch
hsm.Dispatch(ctx, agentID, hsm.Event{
    Name: "process.stdout",
    Data: ProcessIOEvent{
        PID:       pid,
        AgentID:   agentID,
        Stream:    StreamStdout,
        Data:      output,
        Timestamp: time.Now(),
    },
})
```

For remote agents, process events shall be relayed through the NATS transport using hsmnet (see §9.1)

### 8.5 Subscriptions
Subscribers may register interest in process events:

```go
type ProcessSubscription struct {
    EventTypes []ProcessEventType
    AgentID    *muid.ID  // nil for all agents
    PID        *int      // nil for all processes
    Handler    ProcessEventHandler
}

type ProcessEventType int

const (
    EventProcessSpawned ProcessEventType = iota
    EventProcessCompleted
    EventProcessFailed
    EventProcessKilled
    EventProcessStdout
    EventProcessStderr
    EventProcessStdin
)

type ProcessEventHandler interface {
    OnLifecycle(event ProcessLifecycleEvent)
    OnIO(event ProcessIOEvent)
}
```

If `AgentID` is provided, only events for that agent are delivered. If `nil`, all process events are delivered. If `PID` is provided, only events for that specific process are delivered.

#### 8.5.1 I/O subscription filtering

For high-volume I/O events, subscribers may specify additional filters:

```go
type IOSubscriptionFilter struct {
    Streams     []IOStream       // Filter by stream type
    Pattern     *regexp.Regexp   // Only deliver data matching pattern
    MinBytes    int              // Minimum bytes before delivery
    MaxBytes    int              // Truncate data beyond this size
    Debounce    time.Duration    // Coalesce events within window
}

func (t *ProcessTracker) SubscribeIO(
    agentID *muid.ID,
    filter IOSubscriptionFilter,
    handler func(ProcessIOEvent),
) (unsubscribe func())
```

#### 8.5.2 Subscription example

```go
// Subscribe to all stderr output for a specific agent
unsubscribe := tracker.SubscribeIO(
    &agentID,
    IOSubscriptionFilter{
        Streams: []IOStream{StreamStderr},
    },
    func(event ProcessIOEvent) {
        log.Printf("Agent %s stderr: %s", event.AgentID, event.Data)
    },
)
defer unsubscribe()

// Subscribe to lifecycle events for all agents
tracker.Subscribe(ProcessSubscription{
    EventTypes: []ProcessEventType{
        EventProcessSpawned,
        EventProcessCompleted,
        EventProcessFailed,
    },
    Handler: &myHandler{},
})
```

### 8.6 Process tree tracking
The tracker shall maintain the full process tree for each agent:

```
PTY Shell (bash/zsh)
├── Agent Process (e.g., claude)
│   ├── cargo build
│   │   └── rustc (multiple)
│   └── npm test
│       └── node
```

This enables:
- Determining if an agent is blocked on a subprocess
- Aggregating resource usage per agent
- Detecting runaway or hung processes

## 9. Event system

The event system MUST use NATS as the default event transport for all deployments, including local-only runs. hsm-go remains the local state machine execution engine, but peers MUST exchange events over NATS subjects as defined in this section.

### 9.1 hsmnet: Network-aware dispatch

The `hsmnet` package defines a NATS-backed, gossip-style pub-sub event bus used for dispatch across peers (director and any host managers). It is also the default dispatch transport in local-only deployments.

#### 9.1.1 Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         hsmnet                                  │
│                                                                 │
│   Broadcast(event)     → all local + all remote peers          │
│   Multicast(ids, event) → specific targets (local or remote)   │
│   Unicast(id, event)   → single target (local or remote)       │
│                                                                 │
│   ┌──────────┐         NATS         ┌──────────┐               │
│   │ Director │◀────────────────────▶│ Host A   │               │
│   └──────────┘                      └──────────┘               │
│        │ NATS                                                  │
│   ┌────▼─────┐                                                  │
│   │ Host B   │                                                  │
│   └──────────┘                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### 9.1.2 hsmnet API

```go
package hsmnet

// IDs are encoded as base-10 strings on the wire (see 9.1.3.1).
func encodeID(id muid.ID) string { return fmt.Sprintf("%d", uint64(id)) }

func encodeIDs(ids []muid.ID) []string {
    out := make([]string, len(ids))
    for i, id := range ids {
        out[i] = encodeID(id)
    }
    return out
}

// subjectPrefix is remote.nats.subject_prefix (default "amux").
var subjectPrefix string

// localPeerID is the peer_id of this process.
var localPeerID uint64

func subj(parts ...string) string {
    // Join with '.'; no escaping is performed.
    return strings.Join(append([]string{subjectPrefix}, parts...), ".")
}

var (
    subjBroadcast = subj("hsmnet", "broadcast")
    // Per-peer subject: subj("hsmnet", "<peer_id>")
)

// Broadcast sends event to all state machines (local and remote) via NATS.
// Implementations MUST NOT directly dispatch locally from the sender; the sender receives its own
// publication through its NATS subscription (see 9.1.4).
func Broadcast(ctx context.Context, event hsm.Event) {
    nc.Publish(subjBroadcast, mustJSON(EventMessage{Type: MsgBroadcast, Event: toWire(event)}))
}

// Multicast sends event to specific targets via NATS.
// Local targets MUST be delivered by publishing to the local peer subject and handling it via the
// same NATS receive path as remote targets.
func Multicast(ctx context.Context, targets []muid.ID, event hsm.Event) {
    byPeer := make(map[uint64][]muid.ID)

    for _, id := range targets {
        if peerID, remote := routePeer(id); remote {
            byPeer[peerID] = append(byPeer[peerID], id)
        } else {
            byPeer[localPeerID] = append(byPeer[localPeerID], id)
        }
    }

    for peerID, ids := range byPeer {
        nc.Publish(subj("hsmnet", fmt.Sprintf("%d", peerID)),
            mustJSON(EventMessage{Type: MsgMulticast, Targets: encodeIDs(ids), Event: toWire(event)}))
    }
}

// Unicast sends event to a single target via NATS.
func Unicast(ctx context.Context, target muid.ID, event hsm.Event) {
    if peerID, remote := routePeer(target); remote {
        nc.Publish(subj("hsmnet", fmt.Sprintf("%d", peerID)),
            mustJSON(EventMessage{Type: MsgUnicast, Target: encodeID(target), Event: toWire(event)}))
    } else {
        nc.Publish(subj("hsmnet", fmt.Sprintf("%d", localPeerID)),
            mustJSON(EventMessage{Type: MsgUnicast, Target: encodeID(target), Event: toWire(event)}))
    }
}
```

Usage:
```go
hsmnet.Broadcast(ctx, event)
hsmnet.Multicast(ctx, []muid.ID{id1, id2, id3}, event)
hsmnet.Unicast(ctx, agentID, event)
```

Name-to-ID resolution uses hsm's registry. hsmnet doesn't duplicate it.

#### 9.1.3 Wire format

```go
type MessageType uint8

const (
    MsgBroadcast  MessageType = 1
    MsgMulticast  MessageType = 2
    MsgUnicast    MessageType = 3
)

type EventMessage struct {
    Type    MessageType `json:"type"`
    Target  string      `json:"target,omitempty"`   // For unicast; base-10 muid.ID string (see 9.1.3.1)
    Targets []string    `json:"targets,omitempty"`  // For multicast; base-10 muid.ID strings (see 9.1.3.1)
    Event   WireEvent   `json:"event"`
}

type WireEvent struct {
    Name string          `json:"name"`
    Data json.RawMessage `json:"data"`
}
```

#### 9.1.3.1 Wire scalar encodings (normative)

- **IDs (`muid.ID`, `peer_id`, `agent_id`, `session_id`, `process_id`, and any other ID-like fields):** When carried over any JSON transport defined by this spec, IDs SHALL be encoded as a JSON string containing the base-10 unsigned integer form.
- Unless explicitly specified for a given field, an encoded ID value MUST NOT be `"0"`.
- **Message types (`EventMessage.type`):** `type` MUST be encoded as a JSON number with value `1`, `2`, or `3`, corresponding to `MsgBroadcast`, `MsgMulticast`, and `MsgUnicast` respectively.
- **Timestamps:** All timestamps SHALL be encoded as RFC 3339 strings in UTC using the `time.RFC3339Nano` format (for example `"2026-01-18T10:30:00Z"`).
- **Durations:** All durations in event payloads SHALL be encoded as a string conforming to Go `time.ParseDuration` (for example `"250ms"`, `"5m"`).
- **Binary data:** Any field that represents raw bytes (for example process I/O data) SHALL be encoded as a base64 string (standard RFC 4648 encoding) in JSON.

Example `EventMessage` envelope (broadcast):

```json
{"type":1,"event":{"name":"process.spawned","data":{"pid":12345,"agent_id":"42","process_id":"9002","command":"cargo","args":["test"],"work_dir":"/repo","parent_pid":12000,"started_at":"2026-01-18T10:30:00Z"}}}
```

#### 9.1.3.2 Required event payload schemas on the host events subject (normative)

For any `EventMessage` sent on the remote host **events** subject (§5.5.7.5) whose `event.name` is one of the following, `event.data` SHALL be a JSON object matching the schema below.

1. `connection.established`:

```json
{"peer_id":"5678","timestamp":"2026-01-18T10:30:00Z"}
```

2. `connection.lost`:

```json
{"peer_id":"5678","timestamp":"2026-01-18T10:30:00Z","reason":"io_error"}
```

3. `connection.recovered`:

```json
{"peer_id":"5678","timestamp":"2026-01-18T10:30:00Z"}
```

4. `process.spawned`:

```json
{"pid":12345,"agent_id":"42","process_id":"9002","command":"cargo","args":["test"],"work_dir":"/repo","parent_pid":12000,"started_at":"2026-01-18T10:30:00Z"}
```

5. `process.completed` / `process.failed` / `process.killed`:

```json
{"pid":12345,"agent_id":"42","process_id":"9002","command":"cargo","exit_code":0,"signal":null,"started_at":"2026-01-18T10:30:00Z","ended_at":"2026-01-18T10:31:05Z","duration":"1m5s"}
```

- `signal` SHALL be either `null` or an integer signal number.

6. `process.stdout` / `process.stderr` / `process.stdin`:

```json
{"pid":12345,"agent_id":"42","process_id":"9002","command":"cargo","stream":"stderr","data_b64":"dGVzdCBmYWlsZWQK","timestamp":"2026-01-18T10:30:30Z"}
```

- `stream` SHALL be one of: `"stdin"`, `"stdout"`, `"stderr"`.

**Unknown events:** For any other `event.name`, receivers MAY ignore the event. Receivers that do not ignore the event SHALL treat `event.data` as opaque JSON and SHALL NOT assume any schema.

#### 9.1.4 Peer connection

Each peer (director or host manager) MUST subscribe to:
- `P.hsmnet.broadcast`, and
- `P.hsmnet.<peer_id>` for its own `peer_id`.

In local-only deployments, `amuxd` MUST still start (or connect to) a NATS server and MUST establish these subscriptions even when all peers run on the same host.

On receipt of an `EventMessage`, the peer MUST deserialize the embedded `WireEvent` and MUST dispatch it locally using hsm-go. Senders MUST NOT bypass this path by dispatching directly; all dispatch MUST flow through NATS subjects for consistent behavior.

Illustrative pseudocode:

```go
type Peer struct {
    ID      muid.ID     // peer_id
    HostID  string      // host_id (optional for director)
    NATS    *nats.Conn
    Sub1    *nats.Subscription // broadcast
    Sub2    *nats.Subscription // unicast/multicast to this peer
}

func (p *Peer) Send(ctx context.Context, subject string, msg EventMessage) error {
    return p.NATS.Publish(subject, mustJSON(msg))
}

func (p *Peer) handleMessage(ctx context.Context, msg EventMessage) {
    event := deserialize(msg.Event)
    switch msg.Type {
    case MsgBroadcast:
        hsm.DispatchAll(ctx, event)
    case MsgMulticast:
        for _, sid := range msg.Targets {
            id, err := decodeID(sid)
            if err != nil {
                continue
            }
            hsm.Dispatch(ctx, id, event)
        }
    case MsgUnicast:
        id, err := decodeID(msg.Target)
        if err != nil {
            return
        }
        hsm.Dispatch(ctx, id, event)
    }
}
```

#### 9.1.5 ID routing

hsm-go provides the ID system; hsmnet maps IDs to peers using the peer routing identity exchanged in the remote `handshake` (§5.5.7.3).

**Routing rule (normative):**
- Each connected peer is identified by a unique `peer_id`.
- `muid.ID` values that must be routable across peers shall embed an owning-peer node component that can be extracted deterministically.
- To route a target ID, hsmnet shall extract the owning peer’s node component from the `muid.ID` and forward the message to the peer whose `peer_id` matches that node component. If the extracted node component matches the local peer’s `peer_id`, the message is local and shall not be forwarded.

Illustrative pseudocode:

```go
// peer_id (from handshake) → connection metadata
var peersByPeerID = make(map[uint64]*Peer)

// ExtractNode is an accessor over muid.ID that returns the embedded node component.
// Implementations shall use the muid package's supported accessor for this.
func ExtractNode(id muid.ID) uint64 { /* ... */ }

func routePeer(id muid.ID, localPeerID uint64) (uint64, bool) {
    node := ExtractNode(id)
    if node == localPeerID {
        return 0, false // local
    }
    return node, true
}

### 9.2 Event dispatch

Within a peer, after a received `EventMessage` is validated, events SHALL be dispatched using hsm-go's dispatch mechanism:

```go
// Dispatch to a specific state machine
done := hsm.Dispatch(ctx, machine, hsm.Event{Name: "task.assigned", Data: task})
<-done  // Wait for processing

// Dispatch to all state machines
hsm.DispatchAll(ctx, hsm.Event{Name: "roster.updated", Data: roster})

// Dispatch to state machines matching ID pattern
hsm.DispatchTo(ctx, "agent.remote.*", hsm.Event{Name: "connection.lost"})
```

### 9.3 Event types

| Category | Events |
|----------|--------|
| Lifecycle | `agent.added`, `agent.removed`, `agent.started`, `agent.stopped` |
| Presence | `presence.changed`, `roster.updated` |
| Task | `task.assigned`, `task.completed`, `task.failed`, `task.cancel`, `task.unassign` |
| PTY | `activity.detected`, `inactivity.detected`, `stuck.detected`, `prompt.detected`, `pty.input` |
| Rate limit | `rate.limit`, `rate.cleared` |
| Process | `process.spawned`, `process.completed`, `process.failed`, `process.killed`, `process.stdout`, `process.stderr`, `process.stdin` |
| Remote | `connection.established`, `connection.lost`, `connection.recovered` |
| Message | `message.outbound`, `message.inbound`, `message.broadcast` |
| Config | `config.file_changed`, `config.reloaded`, `config.updated`, `config.reload_failed` |
| Adapter | `adapter.setup.started`, `adapter.setup.completed`, `adapter.setup.failed`, `adapter.auth.started`, `adapter.auth.completed`, `adapter.auth.failed` |
| Git | `git.merge.requested`, `git.merge.completed`, `git.merge.conflict`, `git.merge.failed` |
| Tool | `tool.invoke`, `tool.search`, `tool.scroll`, `tool.send_message`, `tool.result` |
| Batch | `events.batch`, `window.expired`, `idle.timeout`, `flush.complete` |
| Gate | `batch.received`, `inference.complete`, `dispatch.complete`, `priority.event` |
| Subscription | `io.event`, `match.found`, `cooldown.expired`, `subscription.created`, `subscription.removed` |
| Shutdown | `shutdown.request`, `shutdown.initiated`, `shutdown.force`, `drain.complete` |

### 9.4 Event structure

```go
// Events use hsm.Event with typed data
type TaskAssigned struct {
    AgentID    muid.ID
    Task       Task
    AssignedAt time.Time
}

// Dispatch with typed payload
hsm.Dispatch(ctx, agentID, hsm.Event{
    Name: "task.assigned",
    Data: TaskAssigned{
        AgentID:    id,
        Task:       task,
        AssignedAt: time.Now(),
    },
})
```

### 9.5 Event handlers

State machines define event handlers via transitions:

```go
hsm.Transition(
    hsm.On("task.assigned"),
    hsm.Source("online"),
    hsm.Target("busy"),
    hsm.Effect(func(a *Agent, e hsm.Event) {
        task := e.Data.(TaskAssigned)
        a.currentTask = task.Task
        a.sendToPTY(task.Task.Prompt)
    }),
)
```

### 9.6 Event deferral

hsm-go supports event deferral for events that should be processed after a state change:

```go
hsm.State("busy",
    hsm.Defer("task.assigned"),  // Queue new tasks while busy
)
```

Deferred events are automatically replayed when the machine exits the deferring state.

### 9.7 Task model

Tasks are event-driven actors with their own HSM lifecycle.

#### 9.7.1 Task structure

```go
type Task struct {
    ID           muid.ID
    Prompt       string
    Priority     int           // 0 = normal, higher = more urgent
    AssigneeID   *muid.ID      // Agent assigned to this task
    DependsOn    []muid.ID     // Tasks that must complete first
    CreatedAt    time.Time
    // State managed by HSM, not stored directly
}
```

#### 9.7.2 Task lifecycle HSM

```go
var TaskModel = hsm.Define("task",
    hsm.State("pending",
        hsm.Entry(func(t *Task) { t.emitEvent("task.created") }),
    ),
    hsm.State("blocked",
        hsm.Entry(func(t *Task) { t.watchDependencies() }),
    ),
    hsm.State("ready"),
    hsm.State("assigned",
        hsm.Entry(func(t *Task) { t.notifyAgent() }),
    ),
    hsm.State("running"),
    hsm.State("completed", hsm.Final()),
    hsm.State("failed", hsm.Final()),
    hsm.State("cancelled", hsm.Final()),

    // Pending → Blocked (has dependencies)
    hsm.Transition(hsm.On("task.queued"),
        hsm.Source("pending"), hsm.Target("blocked"),
        hsm.Guard(func(t *Task) bool { return len(t.DependsOn) > 0 }),
    ),
    // Pending → Ready (no dependencies)
    hsm.Transition(hsm.On("task.queued"),
        hsm.Source("pending"), hsm.Target("ready"),
        hsm.Guard(func(t *Task) bool { return len(t.DependsOn) == 0 }),
    ),
    // Blocked → Ready (all dependencies resolved)
    hsm.Transition(hsm.On("dependencies.resolved"),
        hsm.Source("blocked"), hsm.Target("ready"),
    ),
    // Ready → Assigned
    hsm.Transition(hsm.On("task.assign"),
        hsm.Source("ready"), hsm.Target("assigned"),
    ),
    // Assigned → Running (agent acknowledged)
    hsm.Transition(hsm.On("task.started"),
        hsm.Source("assigned"), hsm.Target("running"),
    ),
    // Running → Completed
    hsm.Transition(hsm.On("task.completed"),
        hsm.Source("running"), hsm.Target("completed"),
        hsm.Effect(func(t *Task) { hsm.DispatchTo(ctx, "task.blocked.*", hsm.Event{Name: "dependency.completed", Data: t.ID}) }),
    ),
    // Running → Failed
    hsm.Transition(hsm.On("task.failed"),
        hsm.Source("running"), hsm.Target("failed"),
    ),
    // Any → Cancelled
    hsm.Transition(hsm.On("task.cancel"),
        hsm.Source("*"), hsm.Target("cancelled"),
        hsm.Guard(func(t *Task, e hsm.Event) bool {
            // Cannot cancel completed/failed tasks
            return t.State() != "completed" && t.State() != "failed"
        }),
    ),
    // Reassignment: Running → Ready (agent became unavailable)
    hsm.Transition(hsm.On("task.unassign"),
        hsm.Source("assigned"), hsm.Target("ready"),
    ),
    hsm.Transition(hsm.On("task.unassign"),
        hsm.Source("running"), hsm.Target("ready"),
    ),

    hsm.Initial(hsm.Target("pending")),
)
```

#### 9.7.3 Task events

| Event | Trigger | Effect |
|-------|---------|--------|
| `task.created` | Task instantiated | Added to task queue |
| `task.queued` | Task submitted for execution | Evaluates dependencies |
| `task.assign` | Director/Manager assigns to agent or manager agent | Assignee notified |
| `task.started` | Agent begins work | Presence → Busy |
| `task.completed` | Agent finishes successfully | Notify dependents |
| `task.failed` | Agent reports failure | May trigger reassignment |
| `task.cancel` | Controller cancels task | Cleanup, notify agent |
| `task.unassign` | Agent unavailable (rate-limited, stuck) | Returns to ready queue |

#### 9.7.4 Dependency resolution

When a task completes, the system shall dispatch `dependency.completed` to all blocked tasks. Each blocked task tracks its pending dependencies; when all resolve, it transitions to `ready`.

```go
func (t *Task) watchDependencies() {
    t.pendingDeps = make(map[muid.ID]bool)
    for _, depID := range t.DependsOn {
        t.pendingDeps[depID] = true
    }
}

func (t *Task) onDependencyCompleted(depID muid.ID) {
    delete(t.pendingDeps, depID)
    if len(t.pendingDeps) == 0 {
        hsm.Dispatch(ctx, t.ID, hsm.Event{Name: "dependencies.resolved"})
    }
}
```

## 10. Adapter interface

### 10.1 Overview
Adapters are the **sole location for agent-specific code** in the amux architecture. They are WASM modules that translate between amux's generic event system and specific CLI agent implementations. Each adapter encapsulates the patterns, commands, and behaviors of a particular coding agent.

The adapter interface is the boundary that enables amux's agent-agnostic design:

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent-Agnostic Core                       │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │  Agent  │  │ Monitor │  │ Process │  │  PTY    │        │
│  │ Manager │  │         │  │ Tracker │  │ Manager │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │            │            │            │              │
│       └────────────┴─────┬──────┴────────────┘              │
│                          │                                  │
│              ┌───────────▼───────────┐                      │
│              │   Adapter Interface   │ ◀─── Generic contract│
│              │ (on_output, manifest, │                      │
│              │  format_input, etc.)  │                      │
│              └───────────┬───────────┘                      │
└──────────────────────────┼──────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
    ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
    │ Claude Code │ │   Cursor    │ │  Windsurf   │
    │   Adapter   │ │   Adapter   │ │   Adapter   │
    │   (WASM)    │ │   (WASM)    │ │   (WASM)    │
    └─────────────┘ └─────────────┘ └─────────────┘
         Agent-Specific Implementations
```

This design ensures:
- **Pluggability:** New agents can be supported without modifying core code
- **Isolation:** Agent-specific bugs cannot affect core stability
- **Testability:** Core can be tested with mock adapters
- **Maintainability:** Agent-specific changes are localized to adapters

### 10.2 Adapter manifest
Each adapter shall declare a manifest describing its capabilities and requirements:

```go
// JSON field names are normative for the WASM adapter ABI (see 10.4).
type AdapterManifest struct {
    Name        string          `json:"name"`                  // Adapter identifier: "claude-code"
    Version     string          `json:"version"`               // Adapter version: "1.2.0"
    Description string          `json:"description,omitempty"` // Human-readable description
    CLI         CLIRequirement  `json:"cli"`                   // CLI tool requirements
    Patterns    AdapterPatterns `json:"patterns"`              // Output patterns for monitoring
    Commands    AdapterCommands `json:"commands"`              // Commands to interact with the agent
}

type CLIRequirement struct {
    Binary     string `json:"binary"`      // Binary name: "claude"
    VersionCmd string `json:"version_cmd"` // Command to get version: "claude --version"
    VersionRe  string `json:"version_re"`  // Regex to extract version: `v(\d+\.\d+\.\d+)`
    Constraint string `json:"constraint"`  // SemVer constraint: ">=1.0.20 <2.0.0"
}

type AdapterCommands struct {
    Start       []string `json:"start"`        // argv to start the agent (first element is executable)
    SendMessage string   `json:"send_message"` // Template for sending a message: "{{message}}"
}
```

`CLIRequirement.VersionCmd` SHALL be executed as a POSIX shell command (Linux/macOS) using:

- `sh -lc <version_cmd>`

Stdout SHALL be captured and matched against `VersionRe`. A non-zero exit status or a non-matching stdout SHALL be treated as “version unavailable” and SHALL cause the CLI requirement check to fail.

### 10.3 CLI version pinning
Adapters shall specify CLI version constraints. At agent startup, the system shall:

1. Execute `CLI.VersionCmd` to detect installed version
2. Extract version using `CLI.VersionRe`
3. Validate against `CLI.Constraint` (SemVer)
4. Fail with clear error if incompatible:
   ```
   Error: adapter "claude-code@1.2.0" requires claude >=1.0.20 <2.0.0, found 0.9.5
   ```

### 10.4 WASM interface
Adapters shall be compiled to WASM (WASI) using TinyGo and loaded by the director.

The WASM boundary shall use a stable, language-agnostic ABI based on linear memory pointers and lengths. Go pointers, slices, and structs shall not be passed directly across the WASM boundary.

#### 10.4.1 ABI: amux-wasm-abi/1

- The module shall export a linear memory named `memory` (the default WASM memory export).
- All pointers are 32-bit unsigned integer offsets into `memory`.
- All strings are UTF-8 byte sequences.
- All JSON payloads shall conform to RFC 8259.

**Return value packing:** Any exported function that returns variable-length data shall return a single unsigned 64-bit integer where:
- the high 32 bits are `ptr` (u32) and
- the low 32 bits are `len` (u32)

A return value of `0` shall be interpreted as `ptr = 0, len = 0`.

**Memory management:** The module shall export:

```go
//export amux_alloc
func amux_alloc(size uint32) uint32

//export amux_free
func amux_free(ptr uint32, size uint32)
```

- The host shall call `amux_alloc` to allocate an input buffer in module memory, write input bytes into `memory[ptr:ptr+len]`, invoke the target function, and then call `amux_free` for the input buffer.
- For any non-zero `(ptr,len)` returned by an adapter function:
  - If `len > 0`, the host SHALL read `memory[ptr:ptr+len]` and then SHALL call `amux_free(ptr,len)` to release the output buffer.
  - If `len == 0`, the host SHALL NOT read any bytes and SHALL NOT call `amux_free`. The returned `ptr` SHALL be treated as a borrowed, stable pointer valid until the next adapter call.
- The adapter shall not retain pointers to host-provided input buffers after returning from a call.

For all required adapter exports that return a packed `(ptr,len)`:

- On success, the adapter MUST return a non-zero packed value (that is, `ptr != 0`). Output MAY be empty by returning a non-zero `ptr` with `len == 0`.
- On failure, the adapter MUST return `0`. If `amux_last_error` is exported, it MUST return a non-empty UTF-8 error message describing the failure.

Adapters that export `amux_last_error` SHALL implement the following semantics:
- On any successful call to a required export, the adapter SHALL clear the stored last error to an empty string.
- On any failure where the required export returns `0`, the adapter SHALL set the stored last error to a non-empty UTF-8 error message.
- The host MAY call `amux_last_error` only when the required export returned `0`.

**Error reporting:** The module may export:

```go
//export amux_last_error
func amux_last_error() uint64
```

If `amux_last_error` is exported, it shall return a packed `(ptr,len)` UTF-8 error message describing the last error that occurred in any adapter call. After the host reads the message, it shall call `amux_free(ptr,len)`.

#### 10.4.2 Required exports

Adapters shall export the following functions:

```go
//export manifest
func manifest() uint64  // packed (ptr,len) JSON AdapterManifest

//export on_output
func on_output(ptr uint32, len uint32) uint64  // packed (ptr,len) JSON []Event

//export format_input
func format_input(ptr uint32, len uint32) uint64  // packed (ptr,len) raw bytes to write to PTY

//export on_event
func on_event(ptr uint32, len uint32) uint64  // packed (ptr,len) JSON []Action
```

Adapters MAY also export the following function:

```go
//export config_default
func config_default() uint64  // packed (ptr,len) UTF-8 TOML adapter defaults document
```

If `config_default` is exported, the host MUST load its TOML content as adapter defaults per §4.2.8.2. The returned TOML document MUST only contain keys under `[adapters.<adapter_name>]` where `<adapter_name>` is the `manifest.name`.

The host shall interpret return values as described in 10.4.1.

| Export | Purpose |
|--------|---------|
| `amux_alloc` / `amux_free` | Allocate and free module memory for host<->adapter data exchange |
| `manifest` | Return adapter manifest as JSON |
| `on_output` | Process raw PTY output bytes, return detected events as JSON |
| `format_input` | Format a message (UTF-8 bytes) into the agent's input bytes |
| `on_event` | React to a system event (JSON), return actions as JSON |
| `config_default` | Optional: Return adapter default configuration as TOML |
| `amux_last_error` | Optional error string for debugging adapter failures |

#### 10.4.3 Adapter instance lifecycle and concurrency

- The host shall instantiate a separate WASM module instance per **agent instance** (one adapter instance per running agent PTY session). Adapter instances shall not be shared across multiple agents.
- Calls into a single adapter instance (`manifest`, `on_output`, `format_input`, `on_event`, and optional `amux_last_error`) shall be **serialized**. The host shall not invoke exported functions concurrently within the same instance.
- The host may execute calls concurrently across different adapter instances.
- Adapter instances may keep internal state across calls (for example partial-line buffering for `on_output`), but that state is scoped to the owning agent instance.
- The host MUST impose a per-adapter-instance WASM linear memory cap of 256MB (256 * 1024^2 bytes). If an adapter instance exceeds (or traps due to attempting to exceed) the cap, the host MUST treat it as an adapter failure and MAY terminate and recreate the adapter instance.

### 10.5 Event and action types

```go
type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Events emitted by adapter
const (
    EventPromptDetected    = "prompt.detected"
    EventRateLimitDetected = "rate.limit"
    EventErrorDetected     = "error.detected"
    EventTaskCompleted     = "task.completed"
    EventMessageOutbound   = "message.outbound"
)

type Action struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Actions requested by adapter
const (
    ActionSendInput     = "send.input"      // Send input to PTY
    ActionUpdatePresence = "update.presence" // Request presence change
    ActionEmitEvent     = "emit.event"       // Emit event to bus
)
```


#### 10.5.1 Adapter JSON encoding rules

- The `Event` and `Action` envelopes MUST be JSON objects with keys `type` (string) and `payload` (any JSON value). If `payload` is omitted, it SHALL be treated as JSON `null`.
- All JSON objects inside adapter `payload` values MUST use lower_snake_case keys unless a section explicitly specifies otherwise.
- Adapter payload scalar encodings MUST follow §9.1.3.1 (IDs as base-10 strings, timestamps as RFC3339Nano strings, durations as Go duration strings, raw bytes as base64).

#### 10.5.2 Standard adapter-emitted event payloads

The following event types are standardized for interoperability. Adapters MAY emit additional event types.

##### message.outbound

`message.outbound` MUST have a payload object with the following fields:

- `to_slug` (string, REQUIRED): Target agent slug or a broadcast alias (`"all"`, `"broadcast"`, `"*"`).
- `content` (string, REQUIRED): Message content.
- `id` (string, OPTIONAL): Message ID. Adapters MAY omit this; if present it MUST be `"0"` or a non-zero base-10 `muid.ID` string.
- `from` (string, OPTIONAL): Sender runtime agent ID. Adapters MAY omit this.
- `to` (string, OPTIONAL): Resolved recipient runtime agent ID. Adapters MAY omit this; broadcast MAY be `"0"` (see `BroadcastID` in §6.4).
- `timestamp` (string, OPTIONAL): RFC3339Nano timestamp. Adapters MAY omit this.

If an adapter cannot provide an `id`, `from`, `to`, or `timestamp`, it SHOULD omit the field rather than inventing values.

For `prompt.detected`, `rate.limit`, `error.detected`, and `task.completed`, the payload MAY be omitted or be an empty object. Hosts MUST accept both forms.

#### 10.5.3 Standard adapter-requested action payloads

`on_event` returns a JSON array of `Action` objects. The host MUST execute actions in array order.

##### send.input

Payload: `{ "data_b64": "<base64>" }` where `data_b64` is REQUIRED and is base64-encoded bytes to write to the agent PTY input stream.

##### update.presence

Payload: `{ "presence": "<state>" }` where `<state>` is REQUIRED and MUST be one of: `"online"`, `"busy"`, `"away"`, `"offline"`.

##### emit.event

Payload: `{ "event": { "type": "<event_type>", "payload": <json> } }` where `event.type` is REQUIRED. `event.payload` MAY be omitted (treated as `null`).

#### 10.5.4 on_event input schema

The host MUST call `on_event` with a UTF-8 JSON object matching the `Event` envelope:

```json
{ "type": "some.event", "payload": {} }
```

The meaning of `type` and `payload` is determined by the event bus. Adapters MUST ignore unknown event types.


### 10.6 Adapter discovery
Adapters shall be discovered from the following locations:

1. **Embedded adapters** — WASM binaries embedded in the amux binary via `//go:embed`
2. **User adapters** —
   - Flat layout (legacy): `~/.config/amux/adapters/*.wasm`
   - Package layout: `~/.config/amux/adapters/*/*.wasm`
3. **Project adapters** —
   - Flat layout (legacy): `.amux/adapters/*.wasm`
   - Package layout: `.amux/adapters/*/*.wasm`

Package layout SHOULD be used when an adapter includes sidecar assets such as `config.default.toml` or `install.toml` (see §10.8).

Priority: project > user > embedded (later overrides earlier).

**Important:** Embedded adapters are pre-compiled WASM binaries, not Go source code compiled into the core. This maintains the agent-agnostic boundary—the core loads adapters through the same WASM runtime interface regardless of their source location.

```go
// Correct: Embedded as pre-compiled WASM binary
//go:embed adapters/claude-code.wasm
var claudeCodeWasm []byte

// Incorrect: Would violate agent-agnostic design
// func NewClaudeCodeAdapter() *Adapter { /* Go code */ }
```

The adapter registry shall treat all adapters uniformly:

```go
type Registry struct {
    adapters map[string]*Adapter  // name → loaded adapter
}

func (r *Registry) Load(name string) (*Adapter, error) {
    // Check discovery paths in priority order
    // Load WASM module through wazero
    // No special cases for specific adapters
}
```

Note: Adapters provide the interface between amux and specific coding agent CLIs. Each subordinate agent uses one adapter, referenced by name as a string.

### 10.7 Example adapter implementation

The following example demonstrates how to implement an adapter for a hypothetical coding agent. This code resides **outside** the core `internal/` packages, in `adapters/{agent-name}/main.go`:

```go
// adapters/example-agent/main.go (compiled with TinyGo to WASM)
// This file contains ALL agent-specific code for this agent type.
package main

import (
    "encoding/json"
    "regexp"
    "unsafe"
)

// Agent-specific patterns (implementation detail). The manifest publishes these as strings.
var promptRe = regexp.MustCompile(`(?m)^>\s*$`)
var rateLimitRe = regexp.MustCompile(`rate limit|too many requests`)

// ---- Minimal allocator required by amux-wasm-abi/1 ----
//
// This example uses a map to keep allocated buffers alive until amux_free.
// Production adapters MAY use a more efficient allocator as long as the ABI
// requirements in Section 10.4 are met.
var allocs = map[uint32][]byte{}

var emptyByte byte

//export amux_alloc
func amux_alloc(size uint32) uint32 {
    if size == 0 {
        return 0
    }
    buf := make([]byte, size)
    ptr := uint32(uintptr(unsafe.Pointer(&buf[0])))
    allocs[ptr] = buf
    return ptr
}

//export amux_free
func amux_free(ptr uint32, size uint32) {
    delete(allocs, ptr)
}

func pack(ptr uint32, n uint32) uint64 {
    return (uint64(ptr) << 32) | uint64(n)
}

func bytesAt(ptr uint32, n uint32) []byte {
    if ptr == 0 || n == 0 {
        return nil
    }
    return (*[1 << 30]byte)(unsafe.Pointer(uintptr(ptr)))[:n:n]
}

func returnBytes(b []byte) uint64 {
    if len(b) == 0 {
        // Success with empty output: return a non-zero pointer with len==0.
        // Host must not read or free when len==0 (see §10.4.1).
        ptr := uint32(uintptr(unsafe.Pointer(&emptyByte)))
        return pack(ptr, 0)
    }
    ptr := amux_alloc(uint32(len(b)))
    copy(bytesAt(ptr, uint32(len(b))), b)
    return pack(ptr, uint32(len(b)))
}

// ---- Required adapter exports ----

//export manifest
func manifest() uint64 {
    m := AdapterManifest{
        Name:        "example-agent",
        Version:     "1.0.0",
        Description: "Adapter for Example Agent CLI",
        CLI: CLIRequirement{
            Binary:     "example-cli",
            VersionCmd: "example-cli --version",
            VersionRe:  `v(\d+\.\d+\.\d+)`,
            Constraint: ">=1.0.0 <2.0.0",
        },
        Patterns: AdapterPatterns{
            Prompt:    `(?m)^>\s*$`,
            RateLimit: `rate limit|too many requests`,
        },
        Commands: AdapterCommands{
            Start:       []string{"example-cli"},
            SendMessage: "{{message}}",
        },
    }

    out, err := json.Marshal(m)
    if err != nil {
        return 0
    }
    return returnBytes(out)
}

//export on_output
func on_output(ptr uint32, n uint32) uint64 {
    data := bytesAt(ptr, n)

    events := make([]Event, 0)
    if promptRe.Match(data) {
        events = append(events, Event{Type: EventPromptDetected})
    }
    if rateLimitRe.Match(data) {
        events = append(events, Event{Type: EventRateLimitDetected})
    }

    out, err := json.Marshal(events)
    if err != nil {
        return 0
    }
    return returnBytes(out)
}

//export format_input
func format_input(ptr uint32, n uint32) uint64 {
    msg := string(bytesAt(ptr, n))
    return returnBytes([]byte(msg + "\n"))
}

//export on_event
func on_event(ptr uint32, n uint32) uint64 {
    // For simplicity, this example does not emit actions.
    return returnBytes([]byte("[]"))
}
```

To build and install:

```bash
# Compile to WASM
tinygo build -o example-agent.wasm -target=wasi ./adapters/example-agent

# Install for user
cp example-agent.wasm ~/.config/amux/adapters/

# Or install for project
cp example-agent.wasm .amux/adapters/
```

The core amux system will automatically discover and load this adapter without any code changes.

### 10.8 Adapter packaging, installation, and setup

This specification defines a standard adapter package format to support:

- fully remote installable adapters (including auth and configuration setup), and
- installation by Go module reference (for example `github.com/foo/codex-adapter@v1.2.3`) that “pulls in” the adapter WASM.

#### 10.8.1 Adapter package directory layout

An adapter package is a directory in an adapter discovery root (see §10.6) that contains an adapter WASM module and optional sidecar assets.

In package layout, an adapter package directory MUST contain:

- `adapter.wasm` (required): the WASM module implementing the adapter ABI.
- `install.toml` (required): the adapter install spec (see §10.8.2).

An adapter package directory MAY contain:

- `config.default.toml` (optional): default adapter configuration (see §4.2.8.2 and §10.4.2).
- `README.md` (optional): human documentation.
- `conformance/` (optional): auxiliary conformance assets referenced by `install.toml`.

The adapter registry MUST treat `adapter.wasm` as the module entrypoint for package layout.

Flat layout (a single `.wasm` file in the discovery root) MAY be supported for backwards compatibility, but flat-layout adapters MUST be treated as non-conforming to v1.13 unless they also provide `install.toml` via package layout.

#### 10.8.2 Adapter install spec (install.toml)

Each adapter package MUST provide an `install.toml` file located in the same directory as `adapter.wasm`.

`install.toml` MUST be valid TOML and MUST use the following top-level structure:

- `[setup.local]` and `[setup.remote]` to declare automated setup flows for local and remote hosts.
- `[conformance]` to declare conformance fixture commands.

Unknown keys MUST be ignored to allow forward evolution.

##### 10.8.2.1 Setup table

`install.toml` MUST support the following keys:

- `[setup.remote.install]`:
  - `steps` (required): an array of install step objects.
- `[setup.remote.auth]`:
  - `check_sh` (optional): shell command string to verify authentication.
  - `copy_paths` (optional): array of paths to copy from the director host to the remote host.
  - `login_sh` (optional): shell command string to perform interactive login.
  - `login_exec` (optional): argv array to perform interactive login.

`install.toml` MAY also provide parallel `[setup.local.install]` and `[setup.local.auth]` tables with the same structure, which the director MAY use for local agent setup.

Install step objects in `steps` MUST support:

- `name` (optional): human-readable label.
- `sh` (optional): shell command string.
- `exec` (optional): argv array.
- `os` (optional): allowed OS values (`linux`, `darwin`, `windows`).
- `arch` (optional): allowed architecture values (`amd64`, `arm64`).
- `requires_tty` (optional, default `false`): whether the step must be executed in a PTY.

If both `sh` and `exec` are present in a step, the director MUST prefer `exec`.

##### 10.8.2.2 Conformance table

`install.toml` MUST include a `[conformance]` table with:

- `fixture_start` (required): an argv array used to start the deterministic conformance fixture in a PTY.
- `fixture_env` (optional): a map of environment variables to set when running the fixture.

The fixture started by `fixture_start` MUST implement the required E2E flows in §4.3.2.

#### 10.8.3 Go module style adapter installation

An amux implementation MUST support installing an adapter package from a Go module reference string of the form:

- `<module_path>@<version>`, or
- `<module_path>` (which MUST be treated as `<module_path>@latest`).

The Go module referenced MUST contain an adapter package directory at the module root with at least:

- `adapter.wasm`, and
- `install.toml`.

To install a module-referenced adapter, the implementation MUST:

1. Download the module sources for the requested module and version.
2. Read `adapter.wasm` from the module root.
3. Instantiate the WASM module and call `manifest()` to obtain the adapter `name` (see §10.2 and §10.4.2).
4. Install the adapter into the user adapter discovery root using package layout:
   - `~/.config/amux/adapters/{name}/adapter.wasm`
   - `~/.config/amux/adapters/{name}/install.toml`
   - and any optional sidecar files present at the module root (for example `config.default.toml` and `conformance/`).

If an adapter with the same `name` is already installed, the implementation MUST replace the existing package directory atomically.

#### 10.8.4 Relationship to remote setup

Remote setup behavior is defined in §5.5.10.

- The director MUST treat the installed adapter package directory as the source of truth for `install.toml` when performing setup.
- A conforming adapter package MUST declare remote setup steps sufficient to make the agent CLI usable on a remote host, including auth behavior.

## 11. LLM coordination

The amux director may be driven by an LLM that observes all agent PTY output and makes coordination decisions.

When `coordination.enabled = true`, the coordinating model MUST be executed via `liquidgen` (see §4.2.10).

Implementations SHOULD treat the local CPU-based coordinating model as the primary liaison for the system and SHOULD enable coordination by default when the required quantized models are available.

### 11.1 Observation loop

The amux director shall periodically capture snapshots of all agent PTY buffers and present them to the coordinating LLM:

```go
type Snapshot struct {
    Timestamp time.Time
    Agents    []AgentSnapshot
}

type AgentSnapshot struct {
    ID       muid.ID
    Name     string
    About    string
    RepoRoot string    // Canonical repository root path for this agent
    Presence Presence
    Buffer   string    // Last N lines from PTY ring buffer
    TUIXML   string    // Compact XML representation of the current terminal screen (see §11.2.1)
    Task     *Task     // Current task, if any
}

func (a *AmuxAgent) observationLoop(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            snapshot := a.captureSnapshot()
            actions := a.llm.Coordinate(ctx, snapshot)
            a.executeActions(ctx, actions)
        }
    }
}
```

### 11.2 Snapshot format

Snapshots shall be formatted as markdown for LLM consumption.

When `AgentSnapshot.TUIXML` is non-empty, the snapshot MUST include the XML under the corresponding agent section as a fenced code block tagged `xml`.

When an agent is rendering a full-screen TUI, implementations SHOULD prefer the decoded TUI XML view for operator/LLM consumption because line-oriented PTY buffers may not reflect the current visual state.


```markdown
# Amux Snapshot (2026-01-14T10:30:00Z)

## frontend-dev [BUSY]
**About:** Works on React components and UI styling
**Repo:** /repo/frontend
**Task:** Fix type errors in Button component

```
src/Button.tsx:12:5 - error TS2322: Type 'string' is not assignable...
> Analyzing the type error...
```

## backend-dev [ONLINE]
**About:** Handles API endpoints and database migrations
**Repo:** /repo/backend
**Task:** None (idle)

```
✓ All tests passed (42 passed, 0 failed)
>
```

## test-runner [OFFLINE]
**About:** Runs tests and reports failures
**Repo:** /repo/backend
**Status:** Rate limited, retry in 5m
```

### 11.2.1 TUI XML screen capture

When TUI decoding is enabled (see §7.7 and §11.4), the director MUST populate `AgentSnapshot.TUIXML` with a compact XML representation of the agent’s current visible terminal screen.

The goal of TUI XML is to provide an LLM-ingestable view of the current screen state while reducing token usage compared to raw terminal captures (which are often dominated by whitespace and box-drawing borders).

#### 11.2.1.1 XML requirements
The TUI XML document:

- MUST be well-formed XML 1.0 and MUST be encoded as UTF-8.
- MUST NOT include a DTD.
- MUST be deterministic for a given terminal screen model (see §7.7.5).
- MUST represent screen content in terminal cell coordinates (monospace grid semantics).

#### 11.2.1.2 Schema (v1)
The root element MUST be `<tui>` with the following attributes:

- `v` (required): XML schema version. MUST be `"1"`.
- Decoders MUST reject any TUI XML document whose `v` attribute is not exactly "1".
- `cols` (required): Terminal column count.
- `rows` (required): Terminal row count.
- `alt` (required): `"1"` if the alternate screen buffer is active, otherwise `"0"`.
- `cur_x` (optional): Cursor column (0-based).
- `cur_y` (optional): Cursor row (0-based).
- `cur_vis` (optional): `"1"` if the cursor is visible, otherwise `"0"`.

The root element MUST contain zero or more `<row>` elements.

Each `<row>` element:

- MUST have `y` (required): the row index (0-based).
- MAY be empty, in which case the row is implicitly filled with default-style spaces.
- MAY contain one or more `<r>` (run) elements.

Each `<r>` element represents a contiguous run of terminal cells:

- `x` (optional): starting column (0-based). If omitted, the run begins immediately after the previous run in the row (or at 0 for the first run).
- `c` (required): run width in terminal cells. MUST be ≥ 1.
- `ch` (optional): a single Unicode character. If present, the run represents `c` copies of that character and the element MUST have no text content.
- `fg` (optional): foreground color. If present, MUST be `"d"` (default), an integer `0-255` (ANSI 256-color index), or a hex color `#RRGGBB` (truecolor).
- `bg` (optional): background color. Same encoding rules as `fg`.
- `a` (optional): SGR attribute flags as a compact string. Supported flags:
  - `b` bold
  - `u` underline
  - `i` italic
  - `r` reverse/invert
  - `d` dim/faint
  - `s` strikethrough
  - `k` blink

If `ch` is omitted, the element’s text content MUST be the string to render in the run (with XML escaping applied). Implementations SHOULD group adjacent cells with identical attributes into as few `<r>` elements as practical.

#### 11.2.1.3 Token minimization rules
To reduce token usage, the encoder:

- MUST omit trailing default-style spaces in each row.
- SHOULD omit leading default-style spaces by using `x` offsets (implicit gaps) rather than emitting explicit space runs.
- SHOULD use `ch` with `c` for long runs of repeated characters (including whitespace and box-drawing borders).
- MUST emit explicit runs for spaces when they are styled (non-default `bg`, reverse video, etc.), because styled whitespace can be semantically meaningful in TUIs.

#### 11.2.1.4 Example
```xml
<tui v="1" cols="40" rows="10" alt="1" cur_x="0" cur_y="9" cur_vis="0">
  <row y="0"><r c="40" ch="─"/></row>
  <row y="1">
    <r x="2" c="12">Select target</r>
  </row>
  <row y="2">
    <r x="2" c="10" a="r" ch=" "/>
    <r c="12" a="r">backend-dev</r>
  </row>
  <row y="9"><r c="40" fg="2">Press Enter to confirm</r></row>
</tui>
```

### 11.3 Coordination actions

The LLM may return coordination actions:

```go
type CoordinationAction struct {
    Type    ActionType
    Target  muid.ID           // Target agent
    Payload json.RawMessage
}

type ActionType string

const (
    ActionAssignTask    ActionType = "assign_task"
    ActionSendMessage   ActionType = "send_message"
    ActionReassignTask  ActionType = "reassign_task"
    ActionPauseAgent    ActionType = "pause_agent"
    ActionResumeAgent   ActionType = "resume_agent"
)

type AssignTaskPayload struct {
    Prompt      string
    Priority    int
    DependsOn   []muid.ID  // Tasks that must complete first
}

type SendMessagePayload struct {
    Message string
}
```

### 11.4 Configuration

```toml
[coordination]
enabled = true
interval = "30s"           # Snapshot interval
buffer_lines = 50          # Lines to capture per agent
engine = "liquidgen"       # liquidgen
model = "lfm2.5-thinking"  # Quantized variant
model_vl = "lfm2.5-VL"     # Quantized variant; used only when visual context is supplied

[coordination.tui]
enabled = true             # Enable PTY TUI decoding and XML snapshots
mode = "auto"              # "auto", "always", or "never"
include_sgr = true         # Include SGR-derived styling attributes (fg/bg/a)
rle_min = 4                # Minimum run length to prefer <r ch="…"/> encoding

[coordination.prompt]
system = """
You are coordinating multiple coding agents working on a shared codebase.
Review the current state of all agents and decide on actions.
Available actions: assign_task, send_message, reassign_task, pause_agent, resume_agent
"""
```

The coordination TUI settings:

- `coordination.tui.enabled`: When `true`, the director MUST maintain a terminal screen model per agent (see §7.7) and MUST populate `AgentSnapshot.TUIXML` per the selected mode.
- `coordination.tui.mode`:
  - `"auto"` (default): The director MUST populate `TUIXML` when the alternate screen buffer is active (`alt=1`) OR when cursor-addressing/erase control sequences have been observed since the prior snapshot.
  - `"always"`: The director MUST populate `TUIXML` for all agents on every snapshot.
  - `"never"`: The director MUST NOT populate `TUIXML`.
- `coordination.tui.include_sgr`: When `false`, the encoder MUST omit `fg`, `bg`, and `a` attributes from `<r>` elements.
- `coordination.tui.rle_min`: The encoder SHOULD use the `ch` attribute only for runs with `c >= rle_min`. For shorter runs, the encoder SHOULD use text content.

### 11.5 Manual mode

When `coordination.enabled = false`, the amux agent operates in manual mode. The observation loop still runs for status display, but no automatic LLM coordination occurs. The user interacts directly to assign tasks and send messages.

### 11.6 Coordination tools

The coordinating LLM (or user in manual mode) may invoke tools to interact with agents. Tools are modeled as events dispatched through the HSM.

#### 11.6.1 Tool definitions

```go
// Tool represents a callable function for the director
type Tool struct {
    Name        string
    Description string
    Parameters  []ToolParameter
    Handler     func(ctx context.Context, args ToolArgs) (ToolResult, error)
}

type ToolParameter struct {
    Name        string
    Type        string  // "string", "integer", "boolean", "array"
    Description string
    Required    bool
}

type ToolArgs map[string]any
type ToolResult struct {
    Success bool
    Data    any
    Error   string
}

type ToolInvocation struct {
    Tool string
    Args ToolArgs
}
```

#### 11.6.2 search

Fuzzy search across agent PTY buffers.

```go
var SearchTool = Tool{
    Name:        "search",
    Description: "Search for text across agent PTY output using fuzzy matching",
    Parameters: []ToolParameter{
        {Name: "query", Type: "string", Description: "Search query (fuzzy matched)", Required: true},
        {Name: "agent", Type: "string", Description: "Agent name to search (omit for all)", Required: false},
        {Name: "limit", Type: "integer", Description: "Max results to return (default: 10)", Required: false},
        {Name: "context_lines", Type: "integer", Description: "Lines of context around matches (default: 2)", Required: false},
    },
}

type SearchResult struct {
    Matches []SearchMatch
}

type SearchMatch struct {
    Agent    string
    Line     int
    Content  string
    Context  []string  // Lines before/after
    Score    float64   // Fuzzy match score (0-1)
}
```

**Event:** `tool.invoke` dispatched to the tools actor with `ToolInvocation.Tool = "search"`

**Use cases:**
- Find error messages across all agents
- Locate specific file references
- Search for patterns in output

#### 11.6.3 scroll

Scroll an agent's PTY buffer view.

```go
var ScrollTool = Tool{
    Name:        "scroll",
    Description: "Scroll an agent's PTY buffer to view historical output",
    Parameters: []ToolParameter{
        {Name: "agent", Type: "string", Description: "Agent name", Required: true},
        {Name: "direction", Type: "string", Description: "Scroll direction: 'up', 'down', 'top', 'bottom'", Required: true},
        {Name: "lines", Type: "integer", Description: "Number of lines to scroll (for up/down)", Required: false},
        {Name: "to_line", Type: "integer", Description: "Scroll to specific line number", Required: false},
    },
}

type ScrollResult struct {
    Agent      string
    StartLine  int
    EndLine    int
    TotalLines int
    Content    string  // Visible content after scroll
}
```

**Event:** `tool.invoke` dispatched to the tools actor with `ToolInvocation.Tool = "scroll"`

**Use cases:**
- Review earlier output that scrolled off screen
- Jump to start of a build/test run
- Navigate to specific line from search result

#### 11.6.4 send_message

Send input to an agent's PTY.

```go
var SendMessageTool = Tool{
    Name:        "send_message",
    Description: "Send a message/command to an agent's PTY input",
    Parameters: []ToolParameter{
        {Name: "agent", Type: "string", Description: "Agent name", Required: true},
        {Name: "message", Type: "string", Description: "Message to send", Required: true},
        {Name: "newline", Type: "boolean", Description: "Append newline (default: true)", Required: false},
    },
}

type SendMessageResult struct {
    Agent     string
    BytesSent int
}
```

**Event:** `tool.invoke` dispatched to the tools actor with `ToolInvocation.Tool = "send_message"`; the handler dispatches `pty.input` to the target agent

**Use cases:**
- Send prompts/instructions to an agent
- Answer agent questions
- Send keyboard shortcuts (Ctrl+C, etc.)

#### 11.6.5 Tool invocation

Tools are invoked via events and return results asynchronously:

```go
// Dispatch tool invocation
result := hsm.Dispatch(ctx, "tools", hsm.Event{
    Name: "tool.invoke",
    Data: ToolInvocation{
        Tool: "search",
        Args: ToolArgs{
            "query": "error",
            "agent": "backend-dev",
            "limit": 5,
        },
    },
})

// Result returned via callback or channel
<-result  // Blocks until tool completes
```

If `ToolInvocation.Tool` does not match a registered tool name, the implementation SHALL return a `ToolResult{Success:false, Error:"unknown tool"}`.

#### 11.6.6 Tool registry

Available tools shall be registered at startup:

```go
var DefaultTools = []Tool{
    SearchTool,
    ScrollTool,
    SendMessageTool,
}

func (a *AmuxAgent) RegisterTool(tool Tool) {
    a.tools[tool.Name] = tool
}
```

Tools are exposed to the LLM via function calling schema:

```json
{
  "name": "search",
  "description": "Search for text across agent PTY output using fuzzy matching",
  "parameters": {
    "type": "object",
    "properties": {
      "query": {"type": "string", "description": "Search query (fuzzy matched)"},
      "agent": {"type": "string", "description": "Agent name to search (omit for all)"},
      "limit": {"type": "integer", "description": "Max results to return (default: 10)"},
      "context_lines": {"type": "integer", "description": "Lines of context around matches (default: 2)"}
    },
    "required": ["query"]
  }
}
```



## 12. CLI client and daemon (JSON-RPC control plane)

### 12.1 Overview

amux MUST provide:

- a docker-like CLI client named `amux`, and
- a local daemon process commonly invoked as `amuxd` that hosts the director, owns local PTYs when configured to do so, and persists session state in memory while running. The `amuxd` and `amux-manager` command names MUST refer to the same binary; role selection is configuration/flag-driven.

The `amux` CLI client MUST communicate with `amuxd` using JSON-RPC 2.0 (see Normative references) over a local transport as specified in §12.2. A local-only deployment MUST be supported by running a single node in director role (hub NATS + director logic) with local host-manager functions enabled.

### 12.2 Transport and framing

#### 12.2.1 Default socket

By default, `amuxd` MUST listen on a Unix domain stream socket at:

- `~/.amux/amuxd.sock`

The socket path MUST be configurable via `daemon.socket_path` (see §4.2.8.4) and via environment variable mapping (`AMUX__DAEMON__SOCKET_PATH`, see §4.2.8.3).

#### 12.2.2 Message framing

The `amux` client and `amuxd` daemon MUST exchange UTF-8 JSON-RPC 2.0 request/response objects, one object per line (newline-delimited JSON). Each line MUST contain exactly one JSON object. Receivers MUST ignore blank lines.

The daemon MUST support both JSON-RPC requests and JSON-RPC notifications from clients. The daemon MUST send JSON-RPC notifications to clients only after an explicit subscription is established (see §12.5).

### 12.3 Client/daemon lifecycle and turnkey defaults

- If `amux` is invoked and `amuxd` is not reachable at the configured socket path, `amux` SHOULD start `amuxd` automatically when `daemon.autostart = true` (default). Autostart MUST NOT require any configuration file to exist.
- If no configuration file is present, `amuxd` MUST start with built-in defaults and MUST permit the following to work without additional setup:
  - `amux agent` for creating and managing agents using CLI arguments, and
  - `amux chat` for connecting to the director and exchanging messages.

### 12.4 Required JSON-RPC methods

All method names in this section are normative.

#### 12.4.1 daemon.ping

Request:

```json
{"jsonrpc":"2.0","id":1,"method":"daemon.ping","params":{}}
```

Response:

```json
{"jsonrpc":"2.0","id":1,"result":{"ok":true}}
```

#### 12.4.2 daemon.version

The daemon MUST report its own version and the spec version it conforms to.

Response:

```json
{"jsonrpc":"2.0","id":1,"result":{"amux_version":"1.14.0","spec_version":"1.14"}}
```

#### 12.4.3 events.subscribe

The daemon MUST support server-to-client event delivery as JSON-RPC notifications.

Request params:

- `filters` (optional): an array of event name glob patterns (for example `"process.*"`). If omitted, the subscription receives all events.

Response:

- `subscription_id`: a string identifier for the subscription.

Example:

```json
{"jsonrpc":"2.0","id":1,"method":"events.subscribe","params":{"filters":["message.*","task.*","process.*"]}}
{"jsonrpc":"2.0","id":1,"result":{"subscription_id":"sub_42"}}
```

After a successful subscription, the daemon MUST deliver events as JSON-RPC notifications with:

- `method` exactly `"notifications.event"`, and
- `params` containing a single field `event` whose value is an `EventMessage` JSON object as defined in §9.1.3.

Example notification:

```json
{"jsonrpc":"2.0","method":"notifications.event","params":{"event":{"type":1,"event":{"name":"process.spawned","data":{"pid":12345,"agent_id":"42","process_id":"9002","command":"cargo","args":["test"],"work_dir":"/repo","parent_pid":12000,"started_at":"2026-01-18T10:30:00Z"}}}}}
```

#### 12.4.4 agent.add

The daemon MUST support creating an agent according to §5.2. The `params` object MUST match:

- `name` (string, required)
- `about` (string, optional)
- `adapter` (string, required)
- `location` (object, required; see §5.1)
- `cwd` (string, optional): the request working directory on the daemon host. If `location.type = "local"` and `location.repo_path` is unset, the daemon MUST resolve `repo_root` from `cwd` as described in §5.2.

The result MUST include the created `agent_id` encoded per §9.1.3.1.

#### 12.4.5 agent.list

The daemon MUST return the current roster. The result MUST include, at minimum, each agent’s `agent_id`, `name`, `adapter`, `presence`, and `repo_root`.

#### 12.4.6 agent.remove

The daemon MUST remove an agent and perform cleanup as described in §5.3.2.

#### 12.4.7 system.update

The daemon MUST support updating amux itself.

- `amux` MUST expose a user-facing command that triggers `system.update`.
- `system.update` MUST perform an atomic update of the `amux` and `amuxd` executables on the daemon host.
- If the update requires restarting `amuxd`, the daemon MUST return success only after it has either:
  - completed an in-place restart, or
  - emitted an error describing why restart could not be completed safely.

The update source (for example a release channel URL or local artifact path) is implementation-defined, but MUST be configurable via environment variables and/or configuration files per §4.2.8.

### 12.5 Permissions context for CLI plugins

The daemon MUST treat CLI plugin-initiated operations as distinct from normal CLI client operations.

- The daemon MUST support permission-scoped operations for CLI plugins as defined in §13.6.
- The daemon MUST reject any plugin-initiated operation that is not permitted, returning a JSON-RPC error.

## 13. CLI plugin system (WASM and remote)

### 13.1 Overview

amux MUST support installing and running CLI plugins that extend the CLI command surface.

- A plugin installed with `amux plugin install <ref>` MUST become invokable as the second argument to `amux` in the form: `amux <plugin_name> [args...]`.
- When a plugin is invoked, `amux` MUST pass all remaining arguments to the plugin unchanged.
- `amux <plugin_name> --help` MUST display the plugin’s help output.

A CLI plugin MAY be:

- a WASM plugin (a WASI module loaded by `amux`), or
- a remote plugin (an endpoint contacted over a protocol defined in §13.4.3).

### 13.2 Plugin management commands

The `amux` CLI client MUST provide a `plugin` management namespace with at least:

- `amux plugin install <ref>`
- `amux plugin list`
- `amux plugin remove <name>`

### 13.3 Plugin manifest (plugin.toml)

Each plugin MUST provide a manifest file named `plugin.toml` with the following required keys:

- `name` (string): the command name used for invocation (for example `bar`).
- `version` (string): semantic version of the plugin.
- `kind` (string): MUST be either `"wasm"` or `"remote"`.
- `permissions` (array of strings): the permissions the plugin requests (see §13.6).

For `kind = "wasm"`, the package MUST also include `plugin.wasm`.

For `kind = "remote"`, the manifest MUST include:

- `remote.rpc_url` (string): the remote plugin RPC endpoint URL.

Example:

```toml
name = "bar"
version = "1.0.0"
kind = "wasm"
permissions = ["events.subscribe", "agent.list"]
```

### 13.4 Plugin installation sources

#### 13.4.1 Go module installation

`amux plugin install` MUST accept a Go module reference of the form:

- `<module_path>@<version>`, or
- `<module_path>` (treated as `<module_path>@latest`).

The referenced module MUST contain a plugin package directory at the module root with at least:

- `plugin.toml`, and
- either `plugin.wasm` (for `kind="wasm"`) or remote configuration fields (for `kind="remote"`).

Installation MUST place the plugin into the user plugin registry directory (see §13.5) under a directory named after `plugin.toml:name`.

The installer SHOULD warn if the plugin name differs from the last path segment of the module reference, because users commonly expect `amux plugin install github.com/foo/bar` to install a plugin invokable as `amux bar`.

#### 13.4.2 URL installation

`amux plugin install` MUST accept an `http://` or `https://` URL reference.

- The installer MUST perform an HTTP GET to the URL.
- If the response content type is `application/vnd.amux.plugin+toml`, the response body MUST be treated as `plugin.toml`.
- If the response content type is `application/vnd.amux.plugin+json`, the response body MUST be treated as a JSON descriptor with, at minimum:
  - `name` (string),
  - `version` (string),
  - `kind` (`"wasm"` or `"remote"`),
  - `plugin_toml_url` (string),
  - and for `"wasm"` plugins, `plugin_wasm_url` (string) and `sha256` (string, lowercase hex).

For `"wasm"` URL installs, the installer MUST download `plugin.wasm` from `plugin_wasm_url`, verify the SHA-256 digest, and then install the plugin.

#### 13.4.3 Remote plugin protocol (protocol 1)

A remote plugin endpoint specified by `remote.rpc_url` MUST implement JSON-RPC 2.0 over HTTP(S) with the following constraints:

- Requests MUST be HTTP POST with a JSON body containing one JSON-RPC request object.
- Responses MUST be a JSON-RPC response object.
- The endpoint MUST support the method `plugin.invoke`.

`plugin.invoke` params MUST include:

- `args` (array of strings): argv after the plugin name.
- `env` (object, optional): environment variables (string to string) provided by the host.
- `cwd` (string, optional): the request working directory on the host.

`plugin.invoke` result MUST include:

- `exit_code` (integer)
- `stdout` (string)
- `stderr` (string)

### 13.5 Plugin registry layout

The user plugin registry root MUST default to:

- `~/.config/amux/plugins/`

Each installed plugin MUST be stored under:

- `~/.config/amux/plugins/<name>/plugin.toml`, and
- if `kind="wasm"`, `~/.config/amux/plugins/<name>/plugin.wasm`.

### 13.6 Plugin permissions and daemon access

Plugins MAY require access to daemon APIs (for example listing agents, subscribing to events). This access MUST be permission-scoped.

Permission identifiers MUST be stable strings. A permission string MUST be either an exact JSON-RPC method name (for example `agent.list`) or a wildcard prefix ending in `.*` (for example `agent.*`). An operation is permitted if its method name matches an approved exact permission or is covered by an approved wildcard prefix.

#### 13.6.1 Permission grant

- At install time, `amux` MUST present the plugin’s requested permissions to the user.
- The user MUST be able to approve or deny each requested permission.
- The granted permission set MUST be persisted alongside the plugin in the plugin registry.

#### 13.6.2 Enforcement

When a plugin attempts to invoke a daemon operation through `amux`, the host MUST enforce permissions.

- If a plugin is not granted a permission required for an operation, the operation MUST be rejected.
- Rejections MUST be surfaced as JSON-RPC errors with `code = -32001` and `message = "permission denied"`.

#### 13.6.3 WASM plugin to host bridge

For `kind="wasm"` plugins, the `amux` client MUST provide a JSON-RPC bridge transport over an inherited file descriptor `3` (FD 3) using newline-delimited UTF-8 JSON objects.

- The plugin MAY send JSON-RPC requests on FD 3 with `method` names that match daemon method names in §12.4.
- The `amux` client MUST validate the request against the plugin’s granted permissions and MUST forward permitted requests to `amuxd`.
- The `amux` client MUST return the daemon’s JSON-RPC response on FD 3.

### 13.7 Built-in plugins

amux MUST ship with the following built-in CLI plugins implemented as local WASM plugins:

- `amux agent`: a plugin for managing agents.
- `amux chat`: a plugin for chatting with the director.

Built-in plugins MUST be discoverable without requiring `amux plugin install`.

### 13.8 Built-in plugin: amux agent

The `amux agent` plugin MUST provide agent management operations that correspond to the daemon methods in §12.4 and the agent model in §5.

- `amux agent` MUST support defining agents via TOML configuration (`[[agents]]`, see §4.2.8.4) and via CLI arguments.
- When both TOML configuration and CLI arguments are provided for the same property, CLI arguments MUST take precedence.
- `amux agent` MUST follow 12-factor app configuration principles (see §4.2.8), including environment variable overrides and avoiding mandatory interactive configuration.

### 13.9 Built-in plugin: amux chat

The `amux chat` plugin MUST provide an interactive terminal UI for exchanging messages with the director.

- The UI MUST be implemented using Bubble Tea.
- The plugin MUST establish an event subscription via `events.subscribe` (see §12.4.3) and MUST render inbound `message.*` events.
- The plugin MUST send user-entered messages to the director by invoking daemon methods; the method names and message routing semantics are implementation-defined, but MUST use JSON-RPC and MUST result in `message.*` events being emitted on success.


<!--
TEMPLATE NOTES:
- Clause numbers: Arabic (1, 2, 3)
- Subclauses: Decimal (5.1, 5.2)
- Sub-subclauses: (5.2.1, 5.2.2)
- Cross-references: "see Clause 5" or "see 5.2.1"
- Requirements: "shall" (mandatory), "shall not" (prohibited)
- Recommendations: "should", "should not"
- Options: "may"
- Keep titles sentence case
- Tables and figures: "Table 1", "Figure 1"
- Protocol field: include only when protocol/wire version differs from document version
- Conformance subclauses: 4.2 and 4.3 are optional; delete if not applicable
- Versioning relation (1.4): delete if document version = protocol version
-->

---

## Annex A
(informative)

**Revision history**

### A.1 Changes in version 1.0.0

Initial specification release.

---

## Annex B
(informative)

**Design rationale**

### B.1 Subscription-based agent collaboration

#### B.1.1 Issue
Modern coding agents (Claude Code, Cursor, Windsurf, etc.) are offered through subscription models. Users may have access to multiple agent subscriptions but no standardized way to orchestrate them collaboratively on a single codebase.

#### B.1.2 Risk
Without a multiplexing solution, users must manually coordinate agents, losing the potential for parallel work, shared context, and efficient task distribution.

#### B.1.3 Resolution
amux treats each coding tool as a subordinate agent running in an owned PTY session. WASM adapters abstract the differences between agent interfaces, allowing heterogeneous agents to collaborate. Git worktrees provide isolation so each agent can work independently without conflicts.

#### B.1.4 Remaining uncertainty
None.

### B.2 WASM adapter architecture

#### B.2.1 Issue
Different coding agents have different CLI interfaces, output formats, and interaction patterns.

#### B.2.2 Risk
Hard-coding agent-specific logic into amux would create tight coupling and maintenance burden.

#### B.2.3 Resolution
Each agent type is implemented as a WASM adapter, providing a pluggable architecture. New agents can be supported by developing new adapters without modifying the core.

#### B.2.4 Remaining uncertainty
None.

### B.3 Git worktree isolation

#### B.3.1 Issue
Multiple agents working on the same files simultaneously would create conflicts.

#### B.3.2 Risk
Agents could overwrite each other's changes, leading to lost work or corrupted state.

#### B.3.3 Resolution
Each agent operates in its own git worktree, providing complete file system isolation while sharing the same repository history. Changes can be merged through standard git workflows.

#### B.3.4 Remaining uncertainty
None.

### B.4 PTY monitoring over agent self-reporting

#### B.4.1 Issue
Agents cannot be relied upon to signal when they complete tasks, encounter errors, or become rate-limited.

#### B.4.2 Risk
Without external monitoring, the system would have no reliable way to detect agent state changes, leading to stale presence information and poor coordination.

#### B.4.3 Resolution
The PTY monitor subscribes directly to PTY output streams, detecting activity patterns, inactivity timeouts, and adapter-specific patterns (prompts, errors, rate limits). This provides reliable state detection without requiring agent cooperation.

#### B.4.4 Remaining uncertainty
None.

### B.5 Owned PTY over external multiplexer

#### B.5.1 Issue
Agent orchestration requires continuous monitoring of terminal output, pattern matching, and process tracking.

#### B.5.2 Risk
Using external multiplexers (tmux, Zellij) requires shelling out for commands and polling for pane content. This introduces latency, parsing complexity, and external dependencies.

#### B.5.3 Resolution
amux owns the PTY layer directly. This provides:
- Event-driven output streams (no polling)
- Direct process tree access
- No external dependencies
- Self-contained binary distribution

#### B.5.4 Remaining uncertainty
None.

### B.6 Amux agent architecture

#### B.6.1 Issue
Direct communication between users and multiple agents creates complexity, inconsistency, and coordination challenges.

#### B.6.2 Risk
Without a mediating layer, users would need to manage multiple agent interfaces, manually track status across agents, and handle task distribution themselves.

#### B.6.3 Resolution
All communication flows through the amux director. The user interacts only with the amux director, which handles task distribution, status aggregation, and agent coordination. This provides a single point of control and a unified view of all agents.

#### B.6.4 Remaining uncertainty
None.

### B.7 Go over Rust

#### B.7.1 Issue
The implementation language affects cross-compilation, binary distribution, and the self-deploying remote agent capability.

#### B.7.2 Risk
Rust requires additional tooling (Docker, cross, musl targets) for cross-compilation. This complicates CI/CD and the self-deploying agent feature where binaries must be copied to remote hosts of varying architectures.

#### B.7.3 Resolution
Go provides native cross-compilation via `GOOS` and `GOARCH` environment variables with no additional tooling. The wazero WASM runtime is pure Go (no CGO), preserving cross-compilation simplicity. TinyGo compiles adapters to WASM.

This enables:
- Trivial cross-compilation for all target platforms
- Self-contained binaries with no external dependencies
- Simple CI/CD pipelines
- Reliable self-deploying remote agent

#### B.7.4 Remaining uncertainty
None.

### B.8 Exec hooking via LD_PRELOAD/DYLD_INSERT_LIBRARIES

#### B.8.1 Issue
To capture per-process I/O streams, the system must intercept process spawning and redirect stdin/stdout/stderr through trackable pipes.

#### B.8.2 Risk
Without process-level I/O capture, all output is multiplexed through the PTY with no way to attribute specific output to specific processes. This limits observability and makes it difficult to track what individual commands are doing.

#### B.8.3 Resolution
The system injects a hook library via `LD_PRELOAD` (Linux) or `DYLD_INSERT_LIBRARIES` (macOS) that intercepts `execve()` and related syscalls. Before each exec, the hook notifies the process tracker, which can set up pipe pairs to capture I/O streams.

This approach provides:
- Per-process I/O capture with exact attribution
- Lifecycle events at the moment of spawn/exit (not polling-delayed)
- Minimal overhead (only at exec boundaries)
- Works with any executable (not just specific shells)

#### B.8.4 Remaining uncertainty
- macOS SIP restricts injection for system binaries; graceful fallback to polling is required
- Statically linked binaries ignore `LD_PRELOAD`; fallback required
- Go c-shared libraries include the Go runtime, increasing hook library size (~2-5MB)
- Cross-compilation of c-shared requires CGO and platform-specific toolchains

### B.9 Polling fallback for process detection

#### B.9.1 Issue
Exec hooking may be unavailable in certain environments (macOS SIP, containerized environments, static binaries).

#### B.9.2 Risk
Without a fallback mechanism, process tracking would fail entirely in restricted environments.

#### B.9.3 Resolution
The system provides a polling-based fallback that scans `/proc` (Linux) or uses `libproc` (macOS) to detect process tree changes. While this cannot capture per-process I/O, it still provides lifecycle events and process tree tracking.

The polling fallback:
- Requires no special permissions or injection
- Works universally across all process types
- Trades latency and I/O attribution for reliability

#### B.9.4 Remaining uncertainty
None.

### B.10 LLM-gated notifications

#### B.10.1 Issue
High-frequency events (verbose build output, streaming logs, process I/O) can overwhelm agents and cause context thrashing, reducing effectiveness.

#### B.10.2 Risk
Time-based batching alone is unintelligent—it treats all events equally. Important events (errors, completions) may be delayed while noise fills the batch. Agents waste context processing irrelevant notifications.

#### B.10.3 Resolution
An optional local CPU-based LLM executed via `liquidgen` using quantized `lfm2.5-thinking` (and optionally `lfm2.5-VL` when visual context is available) evaluates batched events and decides:
- **PASS**: Important events that should interrupt the agent
- **SUPPRESS**: Noise that can be dropped
- **SUMMARIZE**: Related events that can be condensed

Benefits:
- Context-aware filtering based on agent's current task
- Sub-100ms latency with small models on modern CPUs
- Graceful fallback to pass-through if inference fails
- Priority bypass for critical events (errors, user messages)

#### B.10.4 Remaining uncertainty
- Model selection and tuning for optimal filtering quality vs latency
- Memory footprint of local LLM (~500MB-4GB depending on model)
- Prompt engineering for consistent JSON responses

### B.11 MCP semantic subscriptions

#### B.11.1 Issue
Agents need to monitor for specific conditions (errors, test failures, warnings) but output formats vary across tools, languages, and frameworks. Pattern matching requires brittle regex for each case.

#### B.11.2 Risk
- Agents miss important events due to format mismatches
- Regex maintenance becomes unsustainable across different build tools
- No way for agents to express high-level intent ("notify me about failures")

#### B.11.3 Resolution
Expose an MCP server that allows agents to subscribe to notifications using semantic queries. ONNX Runtime with INT8-quantized embedding models (all-MiniLM-L6-v2, BGE-small, GTE-small) computes vector similarity on CPU.

Benefits:
- Agents describe intent conceptually ("test failed", "compilation error")
- Works across any output format that semantically matches
- Sub-3ms embedding latency with quantized ONNX models
- ONNX Runtime is lightweight (~15MB) vs full ML frameworks
- Pure Go bindings via onnxruntime_go, no Python dependency
- SQLite + sqlite-vec vector search for efficient multi-subscription matching with persistence across restarts
- Cooldown prevents notification spam

#### B.11.4 Remaining uncertainty
- Optimal similarity thresholds for different use cases
- Handling multi-language output (may need multilingual embeddings like gte-small)
- ONNX Runtime shared library adds ~15MB to binary size per platform

---

## Annex C
(informative)

**Considered alternatives**

### C.1 tmux as terminal backend

**Considered:** Using tmux to manage terminal sessions, with amux controlling tmux via CLI commands or control mode.

**Rejected:** tmux has no library API or proper IPC mechanism. Control mode is a text-based protocol requiring parsing. Pane monitoring would require polling `capture-pane` instead of receiving output events. This adds latency, complexity, and an external dependency.

**Resolution:** amux owns the PTY layer directly, providing event-driven output streams and no external dependencies.

### C.2 Zellij as terminal backend

**Considered:** Using Zellij, a Rust-native terminal multiplexer with WASM plugin support.

**Rejected:** Despite being Rust-native, Zellij's programmatic interface is still CLI-based. Its WASM plugins extend the multiplexer UI rather than orchestrate external processes. The WASM alignment with amux adapters is superficial—different use cases.

**Resolution:** Same as E.1—owned PTY layer.

### C.3 Rust as implementation language

**Considered:** Implementing amux in Rust with wasmtime for WASM runtime and portable-pty for PTY management.

**Rejected:** Rust cross-compilation requires Docker and the `cross` tool, adding complexity to CI/CD and the self-deploying remote agent feature. While Rust's ecosystem (wasmtime, portable-pty) is mature, the tooling overhead outweighs the benefits for this use case.

**Resolution:** Go with wazero (pure Go WASM runtime) and native cross-compilation. TinyGo compiles adapters to WASM.

### C.4 ptrace for process interception

**Considered:** Using `ptrace` syscall to intercept all child process syscalls, capturing exec and I/O operations.

**Rejected:** ptrace requires attaching to the target process and causes significant performance overhead (every syscall traps to the tracer). It also complicates process tree management and has portability issues between Linux and macOS.

**Resolution:** LD_PRELOAD-based hooking intercepts only at exec boundaries with minimal overhead, and gracefully falls back to polling when unavailable.

### C.5 eBPF for process monitoring

**Considered:** Using eBPF (Linux) to attach probes to exec and I/O syscalls for kernel-level observability.

**Rejected:** eBPF requires Linux kernel 4.4+ with specific config options, root privileges or CAP_BPF, and typically CGO bindings. It does not work on macOS. This violates the cross-platform and pure-Go constraints.

**Resolution:** LD_PRELOAD provides similar functionality in userspace without kernel dependencies.

### C.6 Wrapper scripts for process interception

**Considered:** Injecting PATH modifications so common commands (gcc, npm, cargo, etc.) are replaced with wrapper scripts that report to amux before executing the real command.

**Rejected:** Fragile approach that only works for known commands. Agents or users could bypass wrappers, and maintaining wrappers for all possible commands is impractical.

**Resolution:** LD_PRELOAD intercepts at the syscall level, capturing all exec calls regardless of how they're invoked.

### C.7 Shell integration for process tracking

**Considered:** Using shell-specific hooks (bash PROMPT_COMMAND, zsh preexec/precmd) to track command execution.

**Rejected:** Only tracks commands entered at the shell prompt, not programmatic process spawning. Different shells have different hook mechanisms, and agents may not use interactive shells.

**Resolution:** LD_PRELOAD works at the process level, independent of shell type or interactivity.

### C.8 C implementation for hook library

**Considered:** Implementing the hook library in pure C for minimal size and no runtime dependencies.

**Rejected:**
- Requires maintaining separate C codebase alongside Go
- Cannot share protocol types and serialization code with the main application
- Loses type safety and memory safety guarantees
- Complicates the build process with separate C toolchain requirements

**Resolution:** Go with `-buildmode=c-shared` produces a shared library loadable via `LD_PRELOAD`. While larger (~2-5MB due to Go runtime), the benefits outweigh the size cost:
- Single language codebase
- Shared protocol types between hook and tracker
- Full Go standard library available (networking, JSON, etc.)
- Memory safety and garbage collection
- Easier maintenance and debugging

---

**References**

[1] Tetrate Labs, "wazero," GitHub, 2025. https://github.com/tetratelabs/wazero

[2] TinyGo Authors, "TinyGo," GitHub, 2025. https://github.com/tinygo-org/tinygo

[3] creack, "pty," GitHub, 2025. https://github.com/creack/pty

[4] StateForward, "hsm-go," GitHub, 2025. https://github.com/stateforward/hsm-go

[5] Kevin Burke, "ssh_config," GitHub, 2025. https://github.com/kevinburke/ssh_config

[6] OpenTelemetry Authors, "opentelemetry-go," GitHub, 2025. https://github.com/open-telemetry/opentelemetry-go

[7] HashiCorp, "yamux," GitHub, 2025. https://github.com/hashicorp/yamux

[8] NATS Authors, "NATS and JetStream," 2025. https://nats.io
