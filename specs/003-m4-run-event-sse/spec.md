# Feature Specification: M4 Run, Event, and SSE

**Feature Branch**: `[003-m4-run-event-sse]`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "开始M4吧"

## Clarifications

### Session 2026-05-23

- Q: What execution behavior should the M4 runnable slice use before LLM Gateway, tools, and workers exist? → A: Deterministic local simulated run.
- Q: How should M4 handle concurrent active runs? → A: Allow active runs across different threads, but only one active run per thread.
- Q: What should the M4 event stream deliver when a client connects or reconnects? → A: Existing persisted event history first, then subsequent live events.
- Q: Which event categories should M4 support first? → A: Minimal observable set: lifecycle, progress, message, error, and final.
- Q: What stop behavior should M4 guarantee? → A: Best-effort cooperative stop.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start and observe a local run (Priority: P1)

As a local Loomi user, I need to start a run for an existing thread and watch its status change over time, so the product can move from durable conversations to observable agent execution without requiring a full LLM or worker platform yet.

**Why this priority**: M4's core value is making a run a first-class observable object. A runnable slice must demonstrate that a thread can have an execution attempt, that the attempt has a lifecycle, and that the lifecycle is visible to the user.

**Independent Test**: Can be tested by selecting a durable thread, starting a run, and confirming the run appears with an identifier, lifecycle status, timestamps, and visible timeline entries that update from start to completion or failure.

**Acceptance Scenarios**:

1. **Given** an active thread exists, **When** the user starts a run for that thread, **Then** the system creates one run associated with that thread and shows it as the current run.
2. **Given** a run has started, **When** the run changes state, **Then** the user can see the latest state without refreshing the page.
3. **Given** the run reaches a terminal state, **When** the user inspects the run, **Then** the timeline shows the state history and final outcome.
4. **Given** one thread has an active run and another thread is idle, **When** the user starts a run in the idle thread, **Then** both threads can have active runs without blocking each other.
5. **Given** a thread already has an active run, **When** the user tries to start another run in that same thread, **Then** the system reports that the thread already has an active run and does not create a second active run.

---

### User Story 2 - Persist a timeline of run events (Priority: P1)

As a Loomi user or developer, I need each run to leave a durable timeline of events, so execution can be explained after refresh and failures can be diagnosed without relying on transient UI state.

**Why this priority**: The constitution requires observable agent execution. Persisted events are the audit trail that later model, tool, worker, and cancellation features will depend on.

**Independent Test**: Can be tested by starting a run, allowing it to emit several events, refreshing the web shell, and confirming the same ordered event timeline is still visible.

**Acceptance Scenarios**:

1. **Given** a run emits lifecycle events, **When** the user opens the run timeline, **Then** events appear in stable chronological order.
2. **Given** a run completes or fails, **When** the page is refreshed, **Then** the terminal event and earlier timeline events remain visible.
3. **Given** a run emits an error event, **When** the user inspects it, **Then** the event explains the failure without exposing secrets.

---

### User Story 3 - Stream run updates to the web shell (Priority: P1)

As a Loomi user, I need the run timeline and debug rail to update while a run is active, so I can understand what the system is doing without manual refreshes.

**Why this priority**: M4 is the first streaming milestone. The web shell should demonstrate real event-driven UI behavior before later LLM, tool, or worker complexity is introduced.

**Independent Test**: Can be tested by starting a run while the web shell is open and confirming new timeline entries appear automatically until the run reaches a terminal state.

**Acceptance Scenarios**:

1. **Given** the web shell connects to a run event stream, **When** the stream opens, **Then** existing persisted events for that run are delivered before later live events.
2. **Given** the web shell is connected to a run event stream, **When** a new run event is produced, **Then** the corresponding timeline entry appears automatically.
3. **Given** the event stream disconnects while a run is active, **When** the user remains on the run view, **Then** the UI shows a recoverable stream error and does not pretend the run is still live.
4. **Given** the user refreshes the page after stream interruption, **When** the run is loaded again, **Then** the persisted timeline recovers the known event history before receiving newer events.

---

### User Story 4 - Stop a local run safely (Priority: P2)

As a Loomi user, I need to request cancellation for a running local run, so I can stop work that is no longer useful and see that the request was recorded.

**Why this priority**: Cancellation is part of observable execution, but a minimal start-and-stream slice can be useful before cancellation is added. M4 should still establish the cancellation boundary before workers arrive.

**Independent Test**: Can be tested by starting a run, issuing a stop request, and confirming the run enters a terminal stopped state with a timeline event that records the user-visible cancellation outcome.

**Acceptance Scenarios**:

1. **Given** a run is active, **When** the user requests stop, **Then** the system records the stop request and the deterministic local run cooperatively enters a stopped terminal state as soon as its local step boundary allows.
2. **Given** a run has already completed, failed, or stopped, **When** the user requests stop, **Then** the system reports that the run is already terminal and does not create a conflicting lifecycle.
3. **Given** a stop request succeeds, **When** the user refreshes the run view, **Then** the stopped outcome remains visible in the timeline.

---

### User Story 5 - Keep later platform capabilities deferred (Priority: P2)

As a maintainer, I need M4 to define clear boundaries around what the run/event layer does and does not do, so later LLM gateway, tool calling, worker queue, desktop runtime, and permission systems can build on it without being pulled forward prematurely.

**Why this priority**: M4 sits between product data and execution platform complexity. It must create durable observability primitives while avoiding broad agent-platform scope creep.

**Independent Test**: Can be tested by reviewing the M4 documentation and verifying it explains implemented run/event/SSE behavior, explicitly deferred capabilities, and how errors and user-controlled data are represented.

**Acceptance Scenarios**:

1. **Given** a contributor reads the M4 documentation, **When** they look for LLM or tool execution support, **Then** they see those capabilities marked as deferred beyond M4.
2. **Given** run events include user-controlled text or external data, **When** the timeline renders those events, **Then** the content is treated as data rather than instructions.
3. **Given** an event contains diagnostic details, **When** it is shown in the UI or logs, **Then** secrets and sensitive local configuration are not exposed.

### Edge Cases

- A run is started for a missing, archived, or inaccessible thread: the system must reject the request with a clear user-facing error.
- A user starts runs in different threads: each thread may have its own active run, and the UI must keep their timelines separate.
- A user starts another run in the same thread while one is active: the system must reject the request or report the existing active run instead of creating a second active run for that thread.
- A run emits duplicate or retried events: the timeline must avoid duplicate entries for the same event identity.
- A run reaches a terminal state while a stop request is in flight: the final state must remain consistent and explainable, and M4 must not claim hard interruption when the run already completed.
- The event stream disconnects or the browser refreshes mid-run: the user must be able to reload persisted run and event history.
- A client reconnects after missing several events: the stream must deliver persisted history first and then continue with live events without duplicating timeline entries.
- A run error includes internal diagnostics: the user-facing event must be useful without leaking secrets, database URLs, tokens, or local file contents.
- A later mocked or deferred surface remains visible in the UI: M4 must not label mock model/tool/worker behavior as real execution.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST allow a user to start a run for an active thread owned by the current local identity.
- **FR-002**: Each run MUST be associated with exactly one thread and one local owner context.
- **FR-002a**: The system MUST allow active runs in different threads at the same time, but MUST NOT create more than one active run for the same thread.
- **FR-003**: Each run MUST expose a lifecycle status that distinguishes at least pending or active work, successful completion, failure, and user-stopped termination.
- **FR-004**: The system MUST record durable run timestamps for creation, latest update, and terminal completion when applicable.
- **FR-005**: The system MUST persist an ordered event timeline for each run.
- **FR-006**: Run events MUST use the initial M4 event categories `lifecycle`, `progress`, `message`, `error`, and `final` to explain lifecycle changes, user-visible progress, simulated output, errors, cancellation, and final outcome.
- **FR-007**: Run event ordering MUST remain stable after refresh and across stream reconnection.
- **FR-008**: The system MUST prevent events from one run from appearing in another run's timeline.
- **FR-009**: The web shell MUST show the current run and event timeline from real run/event data when M4 is configured.
- **FR-010**: The web shell MUST receive existing persisted events before live events when opening or reconnecting to a run event stream.
- **FR-010a**: The web shell MUST receive new run events automatically while viewing an active run.
- **FR-011**: If live event streaming is unavailable or interrupted, the web shell MUST show a recoverable stream state and allow persisted run/event history to be reloaded.
- **FR-012**: The user MUST be able to request a best-effort cooperative stop for an active run and see the resulting lifecycle state and timeline event.
- **FR-013**: Stop requests for terminal runs MUST be handled without creating conflicting lifecycle states.
- **FR-014**: Run and event errors MUST use stable user-facing codes or categories, human-readable messages, and request or event identifiers without exposing secrets.
- **FR-015**: M4 MUST preserve M3 thread/message behavior and MUST NOT require LLM gateway, real model output, tool execution, worker queue, desktop runtime, attachment processing, RAG, plugin runtime, or production authentication.
- **FR-016**: M4 MUST use a deterministic local simulated run as the first runnable execution behavior and MUST make it clearly distinguishable from future real LLM/tool/worker execution.
- **FR-017**: M4 readiness or validation MUST include evidence that run/event persistence and live updates are available when the required local dependencies are usable.
- **FR-018**: Documentation MUST describe run lifecycle, event categories, stream behavior, stop behavior, error/safety boundaries, validation steps, and deferred capabilities.
- **FR-019**: User-controlled or external event payload content MUST be treated as data in UI and diagnostics, not as instructions.
- **FR-020**: The feature MUST provide a runnable local smoke path that demonstrates start, live update, persistence after refresh, and terminal outcome for at least one run.

### Key Entities

- **Run**: A durable execution attempt for one thread and local owner; carries lifecycle status, timestamps, current outcome, and relationships to its event timeline.
- **Run Event**: A durable timeline entry for one run; uses one of the initial categories `lifecycle`, `progress`, `message`, `error`, or `final` to represent lifecycle changes, user-visible progress, simulated output, errors, cancellation, or final outcome in stable order.
- **Event Stream**: A live connection that delivers new run events to the web shell while a run is active and supports recovery through persisted history.
- **Stop Request**: A user-visible best-effort cooperative request to end an active run; results in a recorded lifecycle outcome and event timeline entry without claiming worker-level hard interruption.
- **Stream State**: The web shell's user-visible status for connected, reconnecting, failed, or recovered event streaming.
- **Execution Boundary**: The explicit M4 line between real run/event observability and deferred LLM, tool, worker, desktop, and permission systems.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A local user can start a run from an existing thread and see the first visible run status or event within 2 seconds.
- **SC-002**: At least 95% of live run events produced during a local smoke test appear in the web timeline without manual refresh.
- **SC-003**: After refreshing the web shell, the full timeline for a completed, failed, or stopped run is visible in the same order as before refresh.
- **SC-004**: A best-effort cooperative stop request for an active run reaches a terminal user-visible state within 3 seconds in the local smoke path.
- **SC-005**: A stream interruption is shown as a recoverable state within 2 seconds, and reconnection delivers persisted history before live events without duplicate timeline entries.
- **SC-006**: Run/event error cases return or display stable error information with no secrets in 100% of documented smoke scenarios, and every smoke-test event belongs to one of the five M4 event categories.
- **SC-007**: A contributor can read M4 documentation and identify implemented run/event/SSE behavior and deferred LLM/tool/worker/runtime capabilities within 10 minutes.
- **SC-008**: Existing M3 thread/message creation, listing, message persistence, and real/mock frontend switching remain demonstrably usable after M4 is enabled.

## Assumptions

- M3 local identity, threads, messages, real/mock frontend switching, seed behavior, and structured errors are available as the foundation for M4.
- M4 is a local-development milestone focused on observable run/event primitives, not production multi-user execution.
- The first M4 runnable slice uses deterministic local simulated execution behavior to produce real runs and events without calling an LLM, executing tools, or starting a background worker queue.
- The web shell already has timeline/debug surfaces that can be connected to real run/event data while preserving clear labels for any deferred or simulated behavior.
- Event streaming should be recoverable through durable history rather than relying on the live connection as the only source of truth.
- Production authentication, hosted operations, cross-device synchronization, desktop runtime permissions, model provider accounts, tool permissions, worker lease ownership, attachments, RAG, and plugin/catalog behavior remain deferred.
- Documentation may be detailed because it is also used as the project owner's learning record.
