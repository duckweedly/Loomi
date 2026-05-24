# Feature Specification: M6 Worker Job Pipeline

**Feature Branch**: `[008-worker-job-pipeline]`

**Created**: 2026-05-24

**Status**: Draft

**Input**: User description: "M6 不要和现在的007混了。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Continues Outside the Request (Priority: P1)

As a Loomi user, I want a submitted agent run to continue as background work after the initial request returns, so that long-running model or future pipeline activity is not tied to a single open browser request.

**Why this priority**: This is the core M6 transition from request-scoped execution to recoverable background execution and unlocks the rest of the worker model.

**Independent Test**: Start a run from an existing thread, observe that the user receives a run acknowledgement quickly, and verify the same run progresses to a terminal state through persisted timeline events without requiring the original request to stay open.

**Acceptance Scenarios**:

1. **Given** an existing thread with a valid user message, **When** the user starts an agent run, **Then** Loomi acknowledges the run within 2 seconds and shows that execution is queued or running.
2. **Given** the browser disconnects after a run is acknowledged, **When** the user reconnects to the thread or run timeline, **Then** Loomi shows the persisted events that occurred while the browser was disconnected.
3. **Given** a worker completes the background run, **When** the user views the thread, **Then** the final assistant outcome appears in the message history and the run timeline explains how it was produced.

---

### User Story 2 - Recover Work After Worker Interruption (Priority: P2)

As a developer operating Loomi locally, I want interrupted work to be visible and recoverable, so that a stalled or crashed worker does not leave active runs permanently stuck or unexplained.

**Why this priority**: M6 must introduce leases, ownership, retry, and failure visibility before adding more complex pipelines or desktop execution.

**Independent Test**: Start a run, interrupt the worker before completion, restart worker processing, and verify the run either resumes safely or reaches a clear retry-exhausted terminal state with observable events.

**Acceptance Scenarios**:

1. **Given** a worker owns a queued job, **When** that worker stops updating ownership before the job completes, **Then** Loomi makes the job eligible for recovery after the ownership window expires.
2. **Given** a recoverable job is picked up again, **When** another worker processes it, **Then** the run timeline records recovery and continues without duplicating final messages.
3. **Given** a job exceeds its allowed recovery attempts, **When** Loomi stops retrying it, **Then** the associated run reaches a failed terminal state with a user-visible explanation.

---

### User Story 3 - Cancel Background Execution (Priority: P3)

As a Loomi user, I want to stop a background run and see the result reflected consistently, so that I can interrupt unwanted execution without creating conflicting timeline states.

**Why this priority**: M4 introduced cooperative stop; M6 must make cancellation truthful across queued and worker-owned background work.

**Independent Test**: Start a run, request stop while it is queued or running, and verify the timeline reaches a stopped terminal state without later producing an assistant final message.

**Acceptance Scenarios**:

1. **Given** a run is queued but not yet owned by a worker, **When** the user requests stop, **Then** Loomi marks the run stopped and prevents worker execution from starting.
2. **Given** a run is already owned by a worker, **When** the user requests stop, **Then** Loomi records the stop request and the worker cooperatively transitions the run to stopped at the next safe boundary.
3. **Given** a run is already terminal, **When** the user requests stop, **Then** Loomi returns the existing terminal state without creating a new conflicting event.

---

### User Story 4 - Inspect Queue and Worker Health (Priority: P4)

As a developer validating Loomi, I want enough queue and worker status to understand why a run is pending, running, recovered, stopped, or failed, so that M6 remains observable rather than a hidden background process.

**Why this priority**: Background execution adds operational ambiguity; minimal health and timeline visibility keeps the vertical slice explainable.

**Independent Test**: Create queued, running, recovered, failed, and stopped runs, then verify each state is visible through the run timeline and local validation output.

**Acceptance Scenarios**:

1. **Given** jobs exist in different lifecycle states, **When** a developer checks local validation output, **Then** the output distinguishes queued, owned, stale, retrying, stopped, completed, and failed work.
2. **Given** a run changes queue or ownership state, **When** the user views its timeline, **Then** the timeline includes enough events to explain the state transition.

### Edge Cases

- A run is queued while another active run already exists for the same thread.
- A worker claims a job at the same time another worker attempts to claim it.
- A worker completes a job after the user has requested stop.
- A worker loses ownership after producing partial events but before writing a terminal state.
- Recovery attempts produce repeated partial output for the same run.
- The queue contains jobs whose referenced run or thread can no longer be found.
- Local configuration starts zero workers or pauses worker processing.
- The browser reconnects while a job is between queued, owned, recovering, and terminal states.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST create background execution work for a newly started agent run without relying on the initial user request to perform the full run.
- **FR-002**: Loomi MUST preserve the existing user rule that a single thread cannot have more than one active run at a time.
- **FR-003**: Loomi MUST expose user-visible run states for queued, running, recovering, stopped, failed, and completed outcomes.
- **FR-004**: Loomi MUST persist enough job ownership information to identify which worker currently owns work and when that ownership becomes stale.
- **FR-005**: Loomi MUST ensure that only one worker can own and advance a job at a time.
- **FR-006**: Loomi MUST recover stale non-terminal jobs after their ownership window expires.
- **FR-007**: Loomi MUST limit recovery attempts and transition exhausted jobs to a failed terminal run state.
- **FR-008**: Loomi MUST make job processing idempotent so recovery or duplicate delivery cannot create duplicate terminal events or duplicate final assistant messages.
- **FR-009**: Loomi MUST allow users to request cancellation for queued and running background runs.
- **FR-010**: Loomi MUST prevent queued jobs with a recorded stop request from starting normal execution.
- **FR-011**: Loomi MUST require running jobs to observe stop requests at safe execution boundaries and produce a stopped terminal run state.
- **FR-012**: Loomi MUST record timeline events for queueing, ownership, recovery, retry, cancellation, failure, completion, and final output transitions.
- **FR-013**: Loomi MUST keep existing history-first timeline recovery behavior so reconnecting users can see background events that happened while disconnected.
- **FR-014**: Loomi MUST surface local worker and queue readiness clearly enough for a developer to distinguish ready, paused, unhealthy, and degraded execution states.
- **FR-015**: Loomi MUST avoid logging secrets or provider credentials in job, worker, retry, failure, or diagnostic output.
- **FR-016**: Loomi MUST keep tool execution and desktop activity capture out of this feature unless they are represented only as non-executed, observable boundaries.

### Key Entities *(include if feature involves data)*

- **Background Job**: A durable unit of work associated with a run, including lifecycle state, ownership, retry count, scheduling time, and terminal outcome.
- **Worker**: A local execution owner that claims jobs, renews ownership while active, observes cancellation, and records progress or terminal outcomes.
- **Worker Lease**: The time-bounded ownership record that prevents multiple workers from advancing the same job and allows recovery when ownership becomes stale.
- **Run**: The user-visible execution associated with a thread and message history; it remains the primary object shown in the timeline.
- **Run Event**: A persisted explanation of queue, worker, recovery, cancellation, model, error, and final-message transitions.
- **Pipeline Step**: A named stage within a run's background execution, limited in this slice to the stages needed to demonstrate recoverable queued work.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95% of run-start actions acknowledge queued or running status within 2 seconds during local validation.
- **SC-002**: A run continues to a terminal state after the initiating browser request is closed in 100% of smoke-test attempts.
- **SC-003**: In a two-worker local validation, no run produces more than one terminal event or more than one final assistant message across 50 consecutive job-claim attempts.
- **SC-004**: After a simulated worker interruption, recoverable work is either resumed or marked failed with a visible explanation within 30 seconds of the configured ownership window expiring.
- **SC-005**: 100% of cancellation requests for queued work prevent normal execution from starting.
- **SC-006**: 100% of cancellation requests for running work produce either a stopped terminal state or an already-terminal response with no conflicting final state.
- **SC-007**: A reconnecting user can see all persisted background execution events for a run in chronological order in 100% of smoke-test attempts.
- **SC-008**: Local diagnostics distinguish ready, paused, unhealthy, and degraded worker or queue states without exposing secrets in validation output.

## Assumptions

- M6 follows the current staged roadmap where M5 owns the LLM gateway and M6 owns worker, job queue, and pipeline execution.
- This feature is independent from the current `007-settings-placeholder` spec and uses its own feature directory, plan, tasks, and validation lifecycle.
- The first M6 slice targets local development and demonstration, not hosted multi-tenant production operation.
- Existing run, event, thread, message, stop, and history-first stream behavior remain product contracts to preserve.
- Provider-backed model execution from M5 can be invoked by background work, but adding new provider configuration UX is outside this feature.
- Real tool execution, desktop runtime permissions, activity recording, sandboxing, channels, multi-agent orchestration, and long-term memory are later M7+ capabilities unless explicitly represented as non-executed observable boundaries.
