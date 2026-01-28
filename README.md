# amux - Agent Multiplexer

**Version:** v0.1.0-phase0  
**Status:** Phase 0 Complete

## Overview

amux is an agent-agnostic application that manages multiple pseudo-terminal (PTY) sessions to orchestrate agent collaboration. The system employs a 100% event-driven architecture where all state changes, commands, and agent interactions are modeled as events.

This is **Phase 0** of the implementation, establishing the foundational structure and interfaces.

## Phase 0 Completion

Phase 0 has implemented:

- ✅ Repository structure per spec §4.2.6, §1.5.1
- ✅ Go 1.25.6 module with core dependencies
- ✅ Error handling conventions (§4.2.5)
- ✅ Configuration subsystem (§4.2.8)
- ✅ OpenTelemetry scaffolding (§4.2.9)
- ✅ Liquidgen inference interface (§4.2.10)
- ✅ Path resolver and .amux/ invariants
- ✅ Event dispatch interfaces (noop Phase 0)
- ✅ Adapter interfaces (noop Phase 0)
- ✅ Conformance harness skeleton
- ✅ Verification infrastructure (Makefile)
- ✅ `amux test` command with snapshot/regression support (§12.6)

## Building

```bash
make build
```

## Testing

```bash
# Run unit tests
make test

# Run with race detector
make test-race

# Full verification
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

Full documentation will be generated via `go-docmd` in later phases. For now, see:
- [spec-v1.22.md](docs/spec-v1.22.md) - Authoritative specification
- [plan-v2.4.md](docs/plan-v2.4.md) - Implementation plan

## Next Steps

Phase 1 will implement:
- Core domain model and IDs
- HSM-driven lifecycle and presence state machines
- Agent and Session structures

## License

(TBD)
