## Packages

- [cmd/amux](cmd/amux/README.md) — Package amux is the Agent Multiplexer implementation.
- [cmd/amux-node](cmd/amux-node/README.md) — Package amuxnode is the unified daemon binary for Agent Multiplexer.
- [internal/adapter](internal/adapter/README.md) — Package adapter provides stable interfaces for adapters (to be fully implemented in Phase 8).
- [internal/config](internal/config/README.md) — Package config provides configuration management with live updates.
- [internal/conformance](internal/conformance/README.md) — Package conformance provides the conformance harness and test runner for amux.
- [internal/demo](internal/demo/README.md) — main.go demonstrates core dependencies for Phase 0 completion.
- [internal/errors](internal/errors/README.md) — Package errors provides common error handling conventions and sentinel errors for amux.
- [internal/event](internal/event/README.md) — Package event provides stable interfaces for event system (to be fully implemented in Phase 7).
- [internal/hsm](internal/hsm/README.md) — Package hsm provides a minimal hierarchical state machine implementation for Phase 0.
- [internal/inference](internal/inference/README.md) — Package inference provides local inference engine integration for amux.
- [internal/paths](internal/paths/README.md) — Package paths provides centralized filesystem path resolution for amux.
- [internal/pty](internal/pty/README.md) — Package pty provides PTY management for amux.
- [internal/telemetry](internal/telemetry/README.md) — Package telemetry provides OpenTelemetry instrumentation for amux.
- [internal/test](internal/test/README.md) — Package test implements the 'amux test' CLI subcommand.
- [internal/wasm](internal/wasm/README.md) — Package wasm provides WASM runtime management for adapters.

