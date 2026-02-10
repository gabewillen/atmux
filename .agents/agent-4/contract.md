# Contract

## Template
- Contract ID:
- Version:
- Status: Draft | Frozen | Superseded
- Owner Agent:
- Date:
- Scope:
- Files:
- Guarantees:
- Non-goals:
- Breaking Changes Since Prior Version:
- Downstream Agents:

## Published
- Contract ID: APP-INTEGRATION-V1
- Version: V1
- Status: Frozen
- Owner Agent: agent-4
- Date: 2026-02-09
- Scope:
  - Provide an application executable that exercises the public C ABI lifecycle.
  - Provide end-to-end smoke validation for runtime boot, session open, generation, and shutdown.
  - Keep app flow constrained to public API calls from `include/engine/*.h`.
- Files:
  - CMakeLists.txt
  - src/app/main.cpp
  - src/app/smoke_flow.cpp
  - src/app/smoke_flow.hpp
  - tests/engine/app_smoke_flow_tests.cpp
  - tests/engine/lifecycle_integration_smoke_tests.cpp
- Guarantees:
  - `engine_app` target is built by default and links against `engine`.
  - `engine_app --smoke` performs runtime create/boot, session create/open, generation steps,
    optional checkpoint save/load, session close/destroy, and runtime shutdown/destroy.
  - App integration tests run in `engine_tests` (C API lifecycle smoke + app smoke-flow tests).
  - A dedicated CTest `engine_app_cli_smoke` validates the CLI binary flow.
- Non-goals:
  - No changes to C ABI surface in `include/engine/*.h`.
  - No changes to machine internals in non-owned directories.
  - No direct dependency on internal C++ runtime/session structs from app code.
- Breaking Changes Since Prior Version:
  - N/A (first version).
- Downstream Agents:
  - agent-0 (merge/integration coordinator).
- Upstream Dependencies (Explicit):
  - `ABI-LIFECYCLE-V1` (agent-1): assumed lifecycle status/semantics of runtime/session C ABI.
  - `INFERENCE-GEN-V1` (agent-2): assumed generation start/step/cancel semantics.
  - `STATE-IO-RECOVERY-V1` (agent-3): assumed checkpoint save/load behavior used by app
    checkpoint smoke path.
