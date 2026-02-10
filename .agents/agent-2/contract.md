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
- Contract ID: INFERENCE-GEN-V1
- Version: V1
- Status: Frozen
- Owner Agent: agent-2
- Date: 2026-02-09
- Scope: Deterministic inference/session generation orchestration semantics for run loop completion.
- Files:
  - src/inference/generator/actions/actions.hpp
  - src/inference/generator/guards/guards.hpp
  - src/inference/generator/machine.hpp
  - tests/machines/inference_generator_behavior_tests.cpp
  - tests/machines/inference_generation_sequence_behavior_tests.cpp
  - docs/puml/engine::inference::generator.puml
- Guarantees:
  - `cmd_decode_step` in `awaiting_decode` deterministically transitions to `stopping` when no decode budget remains via explicit guard/action, instead of being dropped.
  - Budget exhaustion sets `stop_reason=1` when generation had a positive token budget and remaining budget reaches zero.
  - Session+generator sequencing is covered by deterministic behavior tests that verify run start, decode/sample/append/emit loop, budget exhaustion stop, finalize, and session run completion accounting.
  - Existing explicit `cmd_stop(reason)` behavior remains supported and unchanged for non-zero caller-provided stop reasons.
- Non-goals:
  - No C ABI/header changes.
  - No runtime/engine orchestration redesign outside inference machines.
  - No sampler policy/math behavior changes.
- Breaking Changes Since Prior Version:
  - None (V1 initial publication).
- Downstream Agents:
  - agent-4
