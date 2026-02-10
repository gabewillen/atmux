# AGENT 1: C ABI and Runtime Worker

## Identity
- Session: `agent-1`
- Worktree: `../machine-agent-1`
- Coordinator: `agent-0`

## Communication Rule
- Send all questions, blockers, and merge readiness notices to `agent-0` only.
- Do not coordinate directly with `agent-2`, `agent-3`, or `agent-4`.

## Send to Coordinator
```bash
./agents/tmux/send_to_agent0.sh agent-1 question "Need clarification on lifecycle edge case"
./agents/tmux/send_to_agent0.sh agent-1 blocker "Cannot proceed without API decision"
./agents/tmux/send_to_agent0.sh agent-1 ready-to-merge "ABI-LIFECYCLE-V1 frozen and checks passed"
```

## Required Events
1. Send `status` at work start.
2. Send `blocker` immediately when blocked.
3. Send `handoff` when contract changes.
4. Send `ready-to-merge` only after all plan checks pass.

## Inherited Global Rules
The section below is copied verbatim from `AGENTS.md` and is mandatory for this agent.

# engine.cpp Development Rules

## Mission
Build a deterministic, production-grade C++ inference engine where Boost.SML state machines are
first-class orchestration, public APIs are C-compatible, and quality gates are enforced in CI.

## Rule Levels
- `MUST`: non-negotiable and expected to be enforced by code review, tests, or CI.
- `SHOULD`: strong guidance; deviations require rationale.
- `TARGET`: long-term direction; do not block delivery when not yet implemented.

## MUST: Project Layout
- Keep public API headers in `include/engine/`.
- Keep implementation in `src/`.
- Keep tests in `tests/`.
- Map directory layout to namespaces.
  Example: `src/inference/sampler/` -> `engine::inference::sampler`.
- For state machine components, colocate machine definition, `data`, `guards`, `actions`, and
  `events` under the same component directory.
- Keep state-machine orchestration logic out of data-only files.

## MUST: State Machine Architecture
- Use Boost.SML for orchestration state machines.
- Keep guards pure and deterministic.
- Keep side effects in actions only.
- Model orchestration decisions with transitions and guards, not ad-hoc control flow.
- Process events with run-to-completion semantics.
- Never silently drop unexpected events; define behavior explicitly.
- Keep machine coupling event-based; do not mutate another machine's context directly.
- Keep context mutation inside transition actions (including internal transitions when needed).

## MUST: Determinism and Error Handling
- Inject time, randomness, and external services.
- Avoid globals/singletons in transition logic.
- Represent errors as explicit machine states/transitions.
- Classify and handle recoverable vs permanent errors through events.
- Do not use exceptions for control flow in hot paths.

## MUST: Build and Toolchains
- Default development and production builds use Zig toolchain (`zig cc` / `zig c++`).
- Coverage builds use native `clang`/`gcc` only.
- Do not rely on compiler-specific behavior without explicit compatibility checks.

## MUST: Testing and Quality Gates
- Use `doctest` for unit tests.
- Use SML introspection for machine assertions (`sm.is(...)`, state visitors, testing policy).
- Name test files by machine or domain (for example `tests/inference/sampler_tests.cpp` or
  `tests/inference/inference_domain_tests.cpp`); avoid arbitrary or ad-hoc test file names.
- Monolithic test files are not allowed.
- Scope tests to one machine, one system, or one behavior per file (for example
  `tests/inference/sampler_machine_tests.cpp`, `tests/runtime/system_boot_tests.cpp`,
  `tests/inference/sampler_cancel_behavior_tests.cpp`).
- Keep snapshot baselines under `./snapshots`.
- Snapshot regressions must fail tests unless explicitly updated.
- Lint snapshot baseline lives under `./snapshots/lint`.
- Missing required tools (for example `clang-format`, `llvm-cov`, `llvm-profdata`, `gcovr`)
  must hard-fail the run.
- Coverage threshold is strict: line coverage must be greater than 90% (enforced as >= 91%).
- CI and local coverage runs must fail below threshold.

## MUST: Documentation Sync
- Reference `docs/sml.md` for SML patterns and testing semantics.
- Keep related `docs/puml/*.puml` diagrams synchronized when state machine structure changes.
- Document state purpose, key invariants, guard semantics, and action side effects.

## MUST: Public API and Interop
- Public API functions use C-compatible signatures with `extern "C"`.
- Use fixed-width integer types at API boundaries.
- Return error codes, not exceptions, across API boundaries.
- Do not expose C++ templates/classes or STL containers directly in public C ABI.

## MUST: Performance Baselines
- Prefer compile-time polymorphism in hot paths.
- Avoid dynamic dispatch in inference hot paths unless justified.
- Avoid allocations during token-generation hot paths.
- Avoid heap allocation by default; use stack storage, fixed-capacity containers, or preallocated
  buffers unless heap usage is absolutely necessary.
- If heap allocation is necessary, perform it outside hot paths, reuse the allocation, and document
  the rationale in code.
- New heap allocations in inference/sampling hot paths require explicit justification and a
  measurable performance rationale.
- Keep telemetry non-blocking and optional.

## SHOULD: Code Style
- Use snake_case for functions/variables/namespaces.
- Use PascalCase for types and state names.
- Use SCREAMING_SNAKE_CASE for constants/macros.
- Keep line length near 100 columns and use 2-space indentation.
- Avoid `using namespace` in headers.

## SHOULD: Cross-Platform
- Keep code portable across Linux, macOS, and Windows.
- Test on x86_64 and arm64 in CI as coverage permits.
- Avoid platform-specific APIs unless wrapped behind abstraction.

## TARGET: llama.cpp/GGML Parity
- Reuse proven mathematical kernels and layout patterns from ggml/llama.cpp where beneficial.
- Validate numerical behavior and performance against llama.cpp baselines.
- Maintain GGUF compatibility and versioned state schema migration paths.

## Enforcement Map
- Build (Zig): `scripts/build_with_zig.sh`
- Coverage + thresholds: `scripts/test_with_coverage.sh`
- Lint snapshot gate: `scripts/lint_snapshot.sh`
- Test runner: `ctest` targets `engine_tests` and `lint_snapshot`
