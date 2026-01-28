# Agent Multiplexer (amux)

**Version:** v0.1.0-phase0  
**Status:** Phase 0 Complete  
**Spec:** [spec-v1.22.md](docs/spec-v1.22.md)

An agent-agnostic multiplexer for orchestrating multiple coding agents through managed PTY sessions.

## Overview

This is **Phase 0** of the implementation, establishing the foundational structure and interfaces per [plan-v2.4.md](docs/plan-v2.4.md).

## Phase 0 Completion

Phase 0 has implemented all required components:

### Core Infrastructure
- ✅ Repository structure per spec §4.2.6
- ✅ Go 1.25.6 toolchain (§4.2.1)
- ✅ Core dependencies: wazero, hsm-go, muid, creack/pty, ssh_config (§4.2.2-§4.2.4)
- ✅ Error handling conventions with sentinel errors (§4.2.5)
- ✅ Smoke tests for all core dependencies

### Configuration & Observability
- ✅ TOML configuration system with hierarchy (§4.2.8)
- ✅ Duration and byte size parsing
- ✅ OpenTelemetry scaffolding (§4.2.9)
- ✅ Path resolver and .amux/ invariants

### Interfaces & Stubs
- ✅ Event dispatch interfaces (noop Phase 0) (§9.1)
- ✅ Adapter interfaces (noop Phase 0) (§10.4)
- ✅ Liquidgen inference interface with traceability (§4.2.10)

### Testing & Verification
- ✅ Conformance harness skeleton (§4.3)
- ✅ Verification infrastructure (Makefile)
- ✅ `amux test` command with snapshot/regression support (§12.6)
- ✅ Per-package READMEs via go-docmd (§4.2.6.1)
- ✅ Automated docs-check enforcement

## Building

```bash
make build
```

This creates:
- `bin/amux` - CLI client
- `bin/amux-node` - Unified daemon binary

## Testing

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run verification suite (tidy, vet, lint, test, test-race)
make verify
```

## Snapshot Testing

```bash
# Create a baseline snapshot
./bin/amux test

# Check for regressions
./bin/amux test --regression
```

## Documentation

Per-package documentation is automatically generated and synchronized:

```bash
# Check documentation sync (fails if out of sync)
make docs-check

# Regenerate documentation
go run github.com/agentflare-ai/go-docmd@latest -cmd -all -inplace ./...
```

See [PACKAGES.md](PACKAGES.md) for the complete package index.

## Project Structure

```
amux/
├── cmd/
│   ├── amux/           # CLI client
│   └── amux-node/      # Unified daemon (director/manager roles)
├── internal/           # Agent-agnostic core (no agent-specific code)
│   ├── adapter/        # WASM adapter runtime interface
│   ├── config/         # Configuration management
│   ├── errors/         # Error handling
│   ├── event/          # Event dispatch
│   ├── inference/      # Local inference (liquidgen)
│   ├── paths/          # Path resolution
│   ├── snapshot/       # Test snapshot system
│   └── telemetry/      # OpenTelemetry
├── conformance/        # Conformance test harness
├── snapshots/          # Test snapshots
├── third_party/
│   └── liquidgen/      # Local inference engine (git submodule)
└── docs/
    ├── spec-v1.22.md   # Authoritative specification
    └── plan-v2.4.md    # Implementation plan
```

## Development

All verification checks must pass before committing:

```bash
make verify      # Run all checks
make docs-check  # Verify docs are in sync
```

## Next Steps

Phase 0 is complete. Phase 1 will implement:
- Core domain model and identifiers
- HSM-driven lifecycle and presence state machines
- Agent data structures
- Full muid integration

See [docs/plan-v2.4.md](docs/plan-v2.4.md) for the complete implementation roadmap.

## License

See project license files.
