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
- Contract ID: STATE-IO-RECOVERY-V1
- Version: V1
- Status: Frozen
- Owner Agent: agent-3
- Date: 2026-02-09
- Scope: Deterministic unwind/recovery semantics for backend/error orchestration state machines
  used by non-inference reliability flows.
- Files:
  - src/backend/graph_scheduler/machine.hpp
  - src/backend/graph_scheduler/events/events.hpp
  - src/backend/graph_scheduler/actions/actions.hpp
  - src/error/error_classifier/machine.hpp
  - src/error/error_classifier/events/events.hpp
  - src/error/error_classifier/actions/actions.hpp
  - src/error/retry_planner/machine.hpp
  - src/error/retry_planner/events/events.hpp
  - src/error/retry_planner/actions/actions.hpp
  - tests/machines/backend_graph_scheduler_machine_tests.cpp
  - tests/machines/backend_graph_scheduler_behavior_tests.cpp
  - tests/machines/error_error_classifier_machine_tests.cpp
  - tests/machines/error_error_classifier_behavior_tests.cpp
  - tests/machines/error_retry_planner_machine_tests.cpp
  - tests/machines/error_retry_planner_behavior_tests.cpp
  - docs/puml/engine::backend::graph_scheduler.puml
  - docs/puml/engine::error::error_classifier.puml
  - docs/puml/engine::error::retry_planner.puml
  - snapshots/sml/audited/backend_graph_scheduler.snap
  - snapshots/sml/audited/error_error_classifier.snap
  - snapshots/sml/audited/error_retry_planner.snap
- Guarantees:
  - `engine::backend::graph_scheduler` no longer transitions directly from phase states to
    `failed`; all `error` events enter `unwind_pending`, and only `cmd_unwind` transitions to
    `failed`.
  - `engine::error::error_classifier` no longer transitions directly from `inspecting`/`mapping`/
    `emitting` on `step_failed`; those failures now transition through `unwind_pending` and require
    `cmd_unwind` before `failed`.
  - `engine::error::retry_planner` no longer transitions directly from `planning`/`canceling` on
    `error`; those failures now transition through `unwind_pending` and require `cmd_unwind` before
    `failed`.
  - Machine coverage tests, behavior tests, snapshot baselines, and PUML diagrams are synchronized
    with these unwind semantics.
  - Guards remain deterministic and side effects remain action-only for modified transitions.
- Non-goals:
  - No public C ABI signature or status-code changes.
  - No changes to inference/runtime/session APIs or owned paths outside agent-3 scope.
  - No new retry policy math or scheduler ordering heuristics.
- Breaking Changes Since Prior Version:
  - N/A (initial published version).
- Downstream Agents:
  - agent-4
