# Feature Specification: RunContext Pipeline Foundation

**Feature Branch**: `[014-run-context-pipeline-foundation]`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "M9 RunContext + Pipeline foundation minimal slice: define RunContext loader from durable data, make worker execution independent from API process memory for critical context, evolve the recorder into observable linear stages, persist safe stage events visible in Timeline/debug panel, and make adding middleware/stages avoid large AgentLoop rewrites. Do not redo M8 worker/job queue; do not add Redis, external queues, multi-worker platform, Persona/Skill, MCP, Memory, Sandbox, Desktop Runtime, multi-agent, shell/filesystem/browser automation tools, or broad abstractions. docs-site must be updated during implementation."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Restore run context from durable state (Priority: P1)

As a Loomi developer validating background execution, I want a worker to prepare the critical run context from durable product data before runtime invocation, so that a queued run can execute without relying on request-scoped API memory.

**Why this priority**: This is the core M9 Step 71 foundation and the first blocker for reliable pipeline execution after the M8 worker queue closeout.

**Independent Test**: Create a run, clear any request-local runtime state by letting the worker pick it up asynchronously, and verify the worker records a context-prepared event whose safe summary includes the run/thread/message/job/provider route/tool availability facts needed for execution.

**Acceptance Scenarios**:

1. **Given** a queued run with persisted thread, messages, job metadata, and provider/model route, **When** a worker starts the run, **Then** Loomi prepares a RunContext from durable state and records `prepare_context` as completed before invoking the runtime.
2. **Given** the API request that created the run has already returned, **When** the worker executes the job, **Then** the run still has access to the current thread messages, job metadata, provider/model route, and enabled MVP tool list.
3. **Given** required context is missing or inconsistent, **When** the worker prepares context, **Then** the run fails safely with a redacted, persisted explanation instead of invoking the runtime with partial hidden state.

---

### User Story 2 - Observe a linear execution pipeline (Priority: P2)

As a user or developer reading a run Timeline/debug panel, I want the execution flow to show clear stage transitions, so that I can understand whether Loomi prepared context, resolved tools, invoked runtime, and finalized the run.

**Why this priority**: Observable stage events satisfy the constitution requirement for explainable agent execution and turn the existing recorder into the M9 Step 72 foundation.

**Independent Test**: Start a run and verify persisted events show the ordered stages `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize`, with safe summaries visible through history replay and live Timeline/debug views.

**Acceptance Scenarios**:

1. **Given** a normal run, **When** the worker executes it, **Then** the persisted event order includes started/completed records for `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize`.
2. **Given** the browser reconnects after completion, **When** the Timeline/debug panel rebuilds from history, **Then** it shows the same stage trace as the live stream.
3. **Given** a stage fails, **When** the run reaches a terminal state, **Then** the Timeline/debug panel shows the failed stage and redacted reason without leaking secrets, raw provider payloads, or unredacted tool results.

---

### User Story 3 - Add stages without rewriting the AgentLoop body (Priority: P3)

As a Loomi maintainer, I want the worker runtime to call named stages through a narrow pipeline boundary, so that the next middleware or stage can be inserted without editing a large AgentLoop control block.

**Why this priority**: M9 is foundation work. The value is a small extension seam that keeps future SafetyCheck, MemoryLoad, Persona, and richer tool stages deferred but easier to add later.

**Independent Test**: Add or reorder a harmless no-op stage in a targeted test fixture and verify the worker records it through the same stage contract without changing model invocation or finalization logic.

**Acceptance Scenarios**:

1. **Given** the four MVP stages are registered in order, **When** the worker executes a run, **Then** each stage receives the same prepared pipeline state and returns a bounded stage result.
2. **Given** a future stage is added in a test fixture, **When** the pipeline runs, **Then** the stage records started/completed events through the shared contract without modifying the main runtime invocation body.
3. **Given** a stage is not part of the MVP scope, **When** M9 foundation ships, **Then** it remains absent from runtime behavior rather than appearing as a placeholder or fake capability.

### Edge Cases

- A background job references a missing run, deleted thread, or unavailable message history.
- A run has no provider/model route or points at a disabled/unavailable provider.
- The enabled MVP tool list is empty, or only `runtime.get_current_time` is enabled.
- A tool-call continuation run needs context reconstructed from persisted run events and messages.
- A user requests stop before or during one of the pipeline stages.
- A worker loses ownership after preparing context but before finalization.
- Stage event persistence fails after a run has already started execution.
- Stage metadata contains sensitive-looking provider, tool, or user-controlled fields.
- Existing mock/local runtime paths must remain understandable while real API mode uses durable context.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST define a RunContext loader that prepares context from durable product data for the current run before runtime invocation.
- **FR-002**: RunContext MUST include the run, thread, ordered conversation messages needed for provider calls, background job metadata, provider/model route, and enabled MVP tools.
- **FR-003**: RunContext MUST NOT include Persona/Skill, MCP, long-term Memory/RAG, Sandbox, Desktop Runtime, shell/filesystem/browser automation tools, or multi-agent state in this feature.
- **FR-004**: Worker execution MUST use the RunContext loader instead of relying on API request memory for critical execution context.
- **FR-005**: Missing or invalid required context MUST produce a redacted terminal failure and MUST NOT invoke provider/runtime work with partial hidden state.
- **FR-006**: The pipeline MUST expose the MVP stages `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize` in this order for normal execution.
- **FR-007**: Each MVP stage MUST record persisted run events for started, completed, and failed outcomes where applicable.
- **FR-008**: Stage event metadata MUST be a safe summary only and MUST NOT include provider credentials, authorization headers, raw provider requests/responses, raw tool results, file contents, shell output, or hidden local state.
- **FR-009**: Timeline and debug views MUST render the persisted stage trace from both live SSE and history replay.
- **FR-010**: The pipeline boundary MUST allow a new named stage or middleware to be added through a small registration/composition point rather than a large rewrite of the AgentLoop body.
- **FR-011**: The feature MUST reuse the existing M6/M8 worker/job queue, lease, retry, cancellation, and ownership guards without introducing a duplicate queue.
- **FR-012**: The feature MUST NOT introduce Redis, external queues, hosted multi-worker platform behavior, new tool families, or broad orchestration abstractions.
- **FR-013**: Existing M7 tool approval and tool-result continuation behavior MUST continue to work when the runtime is invoked from the new pipeline foundation.
- **FR-014**: Documentation MUST update docs-site architecture, API/event contract or runbook pages, roadmap/current-status, devlog, and Spec Kit references for the M9 foundation slice.

### Key Entities *(include if feature involves data)*

- **RunContext**: Prepared execution context for a run, assembled from durable run, thread, message, job, provider route, model route, and enabled MVP tool data.
- **Pipeline Stage**: A named linear execution step that receives pipeline state, records safe stage events, and returns success or a redacted failure.
- **Pipeline Trace Event**: A persisted run event describing a stage started, completed, or failed outcome with safe diagnostic metadata.
- **Tool Resolution Summary**: The safe result of deciding which MVP tools are enabled for the run, limited to allowed tool names and approval/execution availability.
- **Runtime Invocation Summary**: The safe record that the worker invoked the existing model/runtime boundary, without raw provider payloads or secrets.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of local worker run smoke attempts record `prepare_context` completed before any runtime/model invocation event.
- **SC-002**: A run created by the API can reach a terminal state after the request returns without any request-local context object in 100% of targeted worker tests.
- **SC-003**: Live SSE and history replay show the same ordered stage trace for `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize` in 100% of timeline/debug tests.
- **SC-004**: Missing required context produces one redacted terminal failure and zero provider/runtime invocations in automated tests.
- **SC-005**: A test-only stage can be inserted through the pipeline boundary without editing the main runtime invocation body.
- **SC-006**: The requested validation plan, including backend tests, related web runtime/timeline tests, web build, docs-site build, and browser smoke, is documented in quickstart/tasks.

## Assumptions

- Original M8 worker/job queue is closed by `specs/008-worker-job-pipeline` plus `specs/013-m8-worker-job-closeout`; this feature starts from that baseline.
- `specs/012-tool-result-model-continuation` has established the provider-neutral continuation context and single MVP tool continuation behavior to preserve.
- `runtime.get_current_time` remains the only executable MVP tool for this foundation slice.
- The first M9 slice is local-development scoped and linear; future Persona/Skill, MemoryLoad, SafetyCheck, MCP, Sandbox, Desktop Runtime, and multi-agent work require separate specs.
- Documentation changes are part of implementation done, but this planning pass stops before implementation until the user confirms.
