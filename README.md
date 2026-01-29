## Packages

- [cmd/amux](cmd/amux/README.md)
- [cmd/amux-node](cmd/amux-node/README.md)
- [hooks](hooks/README.md) — Package hooks implements the exec hook library used for process tracking.
- [internal/adapter](internal/adapter/README.md) — Package adapter defines the WASM adapter runtime interface.
- [internal/agent](internal/agent/README.md) — Package agent manages agent lifecycle and presence state machines.
- [internal/config](internal/config/README.md) — Package config implements configuration loading, parsing, and live updates.
- [internal/conformance](internal/conformance/README.md) — Package conformance provides the conformance harness skeleton.
- [internal/daemon](internal/daemon/README.md) — Package daemon hosts the JSON-RPC control plane.
- [internal/git](internal/git/README.md) — Package git provides worktree and merge helpers for local repositories.
- [internal/inference](internal/inference/README.md) — Package inference provides local inference engine integration.
- [internal/manager](internal/manager/README.md) — Package manager manages local agents, worktrees, and sessions.
- [internal/monitor](internal/monitor/README.md) — Package monitor observes PTY output and detects adapter patterns.
- [internal/paths](internal/paths/README.md) — Package paths centralizes filesystem path resolution for amux.
- [internal/process](internal/process/README.md) — Package process tracks child processes launched within agent sessions.
- [internal/protocol](internal/protocol/README.md) — Package protocol defines the event transport interfaces for amux.
- [internal/pty](internal/pty/README.md) — Package pty provides PTY creation and I/O helpers.
- [internal/rpc](internal/rpc/README.md) — Package rpc implements JSON-RPC 2.0 transport over Unix sockets.
- [internal/session](internal/session/README.md) — Package session manages owned PTY sessions for local agents.
- [internal/telemetry](internal/telemetry/README.md) — Package telemetry provides OpenTelemetry scaffolding for amux.
- [internal/tui](internal/tui/README.md) — Package tui handles terminal screen decoding and TUI encoding.
- [pkg/api](pkg/api/README.md) — Package api defines public types shared between amux clients and the daemon.

