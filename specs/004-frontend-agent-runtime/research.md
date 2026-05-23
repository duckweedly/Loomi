# Research: Frontend Agent Runtime Skeleton

## Decision: Use a frontend runtime adapter boundary alongside the API client

**Rationale**: M3 `ApiClient` represents durable user/thread/message operations, while M4 adds real run/event/SSE and M5+ will add LLM/tool execution. A small execution adapter boundary lets the UI consume runtime semantics consistently in mock mode without hiding the fact that configured real API mode should use M4's real run/event/SSE path.

**Alternatives considered**:

- Put all runtime semantics only in `mockApiClient`. Rejected because mock and real would diverge and later require UI rewrites.
- Replace `realApiClient` with the execution adapter immediately. Rejected because M4 already has a working real API contract and a broader adapter migration would add churn.
- Add a global state-management library. Rejected because the scope is small enough for focused modules and existing React state.

## Decision: Derive Chat Canvas state with a pure function

**Rationale**: The feature requires many visible states: no selected thread, empty thread, loading, load failure, history, waiting run, running, completed, failed, and backend unavailable. A pure derivation function can be tested exhaustively and keeps UI rendering from becoming a pile of conditional branches.

**Alternatives considered**:

- Inline all state checks in `ChatCanvas.tsx`. Rejected because it would be hard to test and easy to regress.
- Store a manually controlled enum only. Rejected because it can drift from messages/run/backend capability unless carefully derived.
- Model every UI state as a backend state. Rejected because several states are frontend presentation states.

## Decision: Use deterministic mock scripts for success, failure, and stopped flows

**Rationale**: The goal is to validate user experience independently from LLM/tool/worker execution. Deterministic scripts make the experience reproducible for tests, screenshots, docs, and learning. Scripts can still use short visible delays in the browser while exposing synchronous or controllable progression in tests.

**Alternatives considered**:

- Static mock messages only. Rejected because it does not test the execution loop or Timeline linkage.
- Randomized fake events. Rejected because unstable tests and screenshots would undermine the learning workflow.
- Full simulator with queues/workers. Rejected as premature M6 platform complexity.

## Decision: Real API mode uses M4 run/event/SSE without mock fallback

**Rationale**: The constitution requires visible failure and honest boundaries. If a real API base is configured, silently falling back to mock run behavior would hide real backend behavior and create misleading product feedback. After the M4 merge, real mode should create messages, start local simulated runs, subscribe to SSE, and stop runs through the real API.

**Alternatives considered**:

- Always fallback to mock execution. Rejected because it violates the no-silent-fallback boundary and makes real-mode smoke misleading.
- Disable Composer entirely in real mode. Rejected because M3 real user messages and M4 real runs both work.
- Show a generic backend-unavailable state for all real mode execution. Rejected after M4 because run/event/SSE is now implemented; backend-unavailable remains useful only for future missing capabilities.

## Decision: First implementation focuses on Chat mode while preserving Work mode separation

**Rationale**: The user explicitly identified Chat and Work as separate business logic. Chat mode is the correct place to validate message-to-run interaction. Work mode already has separate recent threads and can adopt runtime behavior later after work-specific project/task flows are designed.

**Alternatives considered**:

- Apply runtime scripts to both Chat and Work immediately. Rejected because Work mode has different product semantics and would require extra design.
- Delay Chat runtime until Work is also designed. Rejected because that blocks frontend learning and M4 readiness.
