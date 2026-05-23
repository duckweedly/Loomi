# Research: Frontend Agent Runtime Skeleton

## Decision: Use a frontend runtime adapter boundary separate from M3 thread/message clients

**Rationale**: M3 `ApiClient` currently represents durable user/thread/message operations. Run/event/SSE belongs to M4 and LLM/tool execution belongs to M5+. Adding future runtime operations directly into the M3 data client would blur milestones and make mock behavior harder to replace cleanly. A separate execution adapter lets the UI consume runtime semantics now while real backend support can be swapped in later.

**Alternatives considered**:

- Extend `ApiClient` with run/event methods now. Rejected because it mixes M3 durable data with M4/M5 execution semantics and encourages fake real-mode behavior.
- Keep runtime behavior inside `mockApiClient`. Rejected because mock and real would diverge and later require UI rewrites.
- Add a global state-management library. Rejected because the scope is small enough for focused modules and existing React state.

## Decision: Derive Chat Canvas state with a pure function

**Rationale**: The feature requires many visible states: no selected thread, empty thread, loading, load failure, history, waiting run, running, completed, failed, and backend unavailable. A pure derivation function can be tested exhaustively and keeps UI rendering from becoming a pile of conditional branches.

**Alternatives considered**:

- Inline all state checks in `ChatCanvas.tsx`. Rejected because it would be hard to test and easy to regress.
- Store a manually controlled enum only. Rejected because it can drift from messages/run/backend capability unless carefully derived.
- Model every UI state as a backend state. Rejected because several states are frontend presentation states and backend M4 is not implemented.

## Decision: Use deterministic mock scripts for success, failure, and stopped flows

**Rationale**: The goal is to validate user experience before backend run/event/SSE exists. Deterministic scripts make the experience reproducible for tests, screenshots, docs, and learning. Scripts can still use short visible delays in the browser while exposing synchronous or controllable progression in tests.

**Alternatives considered**:

- Static mock messages only. Rejected because it does not test the execution loop or Timeline linkage.
- Randomized fake events. Rejected because unstable tests and screenshots would undermine the learning workflow.
- Full simulator with queues/workers. Rejected as premature M6 platform complexity.

## Decision: Real API mode must expose backend capability unavailable until M4/M5 exist

**Rationale**: The constitution requires visible failure and honest boundaries. If a real API base is configured, silently falling back to mock run behavior would hide missing backend capability and create misleading product feedback. The UI should clearly show that thread/message persistence exists while run/event execution is not yet connected.

**Alternatives considered**:

- Always fallback to mock execution. Rejected because it violates the no-silent-fallback boundary and makes real-mode smoke misleading.
- Disable Composer entirely in real mode. Rejected because M3 real user messages still work; only run/event execution is unavailable.
- Show a generic error. Rejected because missing capability is expected staged-roadmap state, not an unexpected failure.

## Decision: First implementation focuses on Chat mode while preserving Work mode separation

**Rationale**: The user explicitly identified Chat and Work as separate business logic. Chat mode is the correct place to validate message-to-run interaction. Work mode already has separate recent threads and can adopt runtime behavior later after work-specific project/task flows are designed.

**Alternatives considered**:

- Apply runtime scripts to both Chat and Work immediately. Rejected because Work mode has different product semantics and would require extra design.
- Delay Chat runtime until Work is also designed. Rejected because that blocks frontend learning and M4 readiness.
