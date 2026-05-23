# Research: M4 Run, Event, and SSE

## Decision: Build M4 as a local deterministic run/event vertical slice

**Decision**: M4 will create real durable runs and events, but the execution source is a deterministic local simulated run. It does not call an LLM, execute tools, start a worker queue, or depend on desktop runtime permissions.

**Rationale**: The constitution requires a runnable vertical slice and observable agent execution before LLM/tool/worker complexity. A deterministic local run lets the project validate run lifecycle, persisted events, live stream delivery, stop behavior, UI wiring, and documentation without introducing model-provider or worker failure modes too early.

**Alternatives considered**:

- **LLM-backed run in M4**: Rejected because it would pull M5 LLM Gateway concerns forward.
- **Tool/worker-backed run in M4**: Rejected because worker ownership, leases, permissions, and recovery are M6 concerns.
- **Manual-only debug events**: Rejected because it would not demonstrate a user-triggered execution slice.

## Decision: Use one active run per thread, with parallelism across threads

**Decision**: The system may have active runs in different threads at the same time, but must not create more than one active run for the same thread.

**Rationale**: This avoids timeline ambiguity inside a single conversation while preserving the user's ability to work in multiple conversations. It also creates a clean path to later worker leases and cancellation semantics.

**Alternatives considered**:

- **One active run globally**: Rejected because it blocks unrelated conversations.
- **Multiple active runs per thread**: Rejected for M4 because the UI and stop semantics would need current-run selection and conflict rules before the worker/job model exists.

## Decision: Persist run events as the source of truth

**Decision**: Runs and run events are persisted product execution data. Live delivery is a projection of persisted event history, not the only source of truth.

**Rationale**: Refresh and reconnect must recover the timeline. Persisted events also establish the audit trail required for future model deltas, tool calls, worker state transitions, and debugging.

**Alternatives considered**:

- **In-memory events only**: Rejected because refresh/reconnect would lose the execution explanation.
- **Run snapshot only**: Rejected because a snapshot cannot explain how execution reached its current state.

## Decision: Event stream uses history-then-live delivery

**Decision**: When a client opens or reconnects to a run stream, it first receives existing persisted events for that run and then receives subsequent live events.

**Rationale**: This gives one recovery model for first load, refresh, and reconnect. The client can dedupe by event identity and does not need separate mental models for history and live updates.

**Alternatives considered**:

- **Live-only stream**: Rejected because missed events require a second catch-up path and increase UI state complexity.
- **Latest snapshot then live**: Rejected because M4's purpose is timeline observability, not only latest-state display.

## Decision: Initial event taxonomy is lifecycle, progress, message, error, final

**Decision**: M4 run events use five initial categories: `lifecycle`, `progress`, `message`, `error`, and `final`.

**Rationale**: These categories are sufficient to explain deterministic local execution, cancellation, success, and failure without introducing model/tool/worker-specific event types prematurely.

**Alternatives considered**:

- **Model/tool-oriented taxonomy**: Rejected because `model_delta`, `tool_call`, and `tool_result` belong to later LLM/tool milestones.
- **Generic-only event type**: Rejected because it would make the timeline hard to test and would push semantic classification into UI copy.

## Decision: Stop is best-effort cooperative in M4

**Decision**: Stop records a user-visible stop request and the deterministic local run cooperatively enters a stopped terminal state at a local step boundary. Terminal runs report their existing terminal state instead of creating conflicting lifecycle transitions.

**Rationale**: M4 has no worker scheduler or hard interruption primitive. Cooperative stop establishes the user-facing cancellation boundary truthfully while leaving stronger worker interruption for M6.

**Alternatives considered**:

- **Immediate hard stop**: Rejected because M4 cannot truthfully guarantee hard interruption without a worker execution model.
- **Mark-only stop request**: Rejected because it would not demonstrate a useful stopped run outcome.

## Decision: Keep M4 API and data local-development scoped

**Decision**: M4 continues M3's fixed local identity and local API boundary. It adds run/event persistence, start/stop operations, stream observation, structured errors, readiness validation, and documentation, but not production auth or hosted multi-user behavior.

**Rationale**: M4 should build directly on M3 without changing identity assumptions. Ownership fields remain necessary so later auth can replace the fixed local identity without changing every run/event boundary.

**Alternatives considered**:

- **Production auth in M4**: Rejected as out of roadmap order.
- **Anonymous runs without ownership**: Rejected because it creates future authorization debt.

## Decision: M4 readiness requires run/event schema availability

**Decision**: M4 readiness should require the schema version that introduces runs and run events, in addition to existing M2/M3 dependency readiness.

**Rationale**: The run/event endpoints depend on durable tables. Reporting ready before schema availability would create confusing runtime failures in the web shell and smoke path.

**Alternatives considered**:

- **Only health checks, no schema readiness**: Rejected because M3 already established schema-aware readiness.
- **Auto-migrate on startup**: Rejected because migrations remain explicit local operations.

## Decision: Treat event payload text as data, not instructions

**Decision**: Run event payloads may include user-controlled or diagnostic text, but UI and diagnostics treat that content as data. User-visible errors must be useful while redacting secrets and local configuration values.

**Rationale**: This preserves the constitution's safety/data boundary and avoids creating prompt-injection-like interpretation paths before LLM features exist.

**Alternatives considered**:

- **Free-form event payloads interpreted by UI logic**: Rejected because it blurs data and control flow.
- **Full internal errors in events**: Rejected because it risks leaking secrets and local paths.
