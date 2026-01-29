## Packages

- [adapters/claude-code](adapters/claude-code/README.md) — Package main implements the Claude Code adapter for amux
- [adapters/cursor](adapters/cursor/README.md) — Package main implements the Cursor adapter for amux
- [adapters/windsurf](adapters/windsurf/README.md) — Package main implements the Windsurf adapter for amux
- [cmd/amux](cmd/amux/README.md) — Package main implements the amux CLI client
- [cmd/amux-node](cmd/amux-node/README.md) — Package main implements the unified amux node daemon (amuxd/amux-manager)
- [hooks](hooks/README.md) — Package hook implements the exec hook library with CGO
- [internal/adapter](internal/adapter/README.md) — Package adapter implements WASM adapter runtime (loads any adapter)
- [internal/adapteriface](internal/adapteriface/README.md) — Package adapteriface implements a stable interface for adapter-provided pattern matching and actions that can be used by other packages during Phase 0 before the full WASM implementation is complete in Phase 8.
- [internal/agent](internal/agent/README.md) — Package agent implements agent orchestration (lifecycle, presence, messaging)  Package agent implements agent orchestration (lifecycle, presence, messaging)  Package agent implements agent orchestration (lifecycle, presence, messaging)  Package agent implements agent orchestration (lifecycle, presence, messaging)  The agent package provides core functionality for managing agents including:   - Lifecycle management (Pending → Starting → Running → Terminated/Errored)   - Presence management (Online ↔ Busy ↔ Offline ↔ Away)   - Roster maintenance for tracking all agents and their states   - Inter-agent messaging capabilities
- [internal/config](internal/config/README.md) — Package config implements a configuration actor with live updates and subscriptions  Package config implements configuration management with hierarchy, environment mapping, and parsing conventions as specified in the amux specification.
- [internal/conformance](internal/conformance/README.md) — Package conformance implements the conformance harness for amux
- [internal/errors](internal/errors/README.md) — Package errors implements error handling conventions for the amux project
- [internal/event](internal/event/README.md) — Package event implements a basic event dispatcher that can be used by other packages This implementation will eventually be replaced with a full NATS-based implementation
- [internal/git](internal/git/README.md) — Package git implements git operations and merge strategies for the amux project
- [internal/ids](internal/ids/README.md) — Package ids implements identifier utilities and normalization functions for the amux project
- [internal/inference](internal/inference/README.md) — Package inference implements the local inference integration interface using liquidgen  Package inference implements local inference integration (liquidgen)  Package inference implements the local inference integration using liquidgen
- [internal/monitor](internal/monitor/README.md) — Package monitor implements PTY output monitoring (delegates to adapters)
- [internal/otel](internal/otel/README.md) — Package otel implements OpenTelemetry scaffolding for the amux project
- [internal/paths](internal/paths/README.md) — Package paths implements a centralized path resolution system for the amux project
- [internal/process](internal/process/README.md) — Package process implements process tracking and interception (generic)
- [internal/protocol](internal/protocol/README.md) — Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)  Package protocol implements remote communication protocol (transports events)
- [internal/pty](internal/pty/README.md) — Package pty implements PTY management (generic PTY operations)
- [internal/snapshot](internal/snapshot/README.md) — Package snapshot implements the snapshot functionality for amux test
- [internal/specchecker](internal/specchecker/README.md) — Package specchecker implements verification that spec-v1.22.md is present and version-locked
- [internal/tui](internal/tui/README.md) — Package tui implements terminal screen decoding and TUI XML encoding (agent-agnostic)
- [pkg/api](pkg/api/README.md) — Package api contains public API types (Agent.Adapter is a string)

