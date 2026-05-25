# Feature Specification: Tool Result Model Continuation

**Feature Branch**: `codex/012-tool-result-model-continuation`

**Created**: 2026-05-25

**Status**: Draft

**Input**: User description: "Design minimal tool result continuation after approval-blocked tool call execution. Planning only. No code implementation, no dangerous tools, no shell/filesystem/MCP, no multi-agent, no long-term memory/RAG."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Continue after an approved tool succeeds (Priority: P1)

A user asks a question that causes the model to request the approved MVP tool. After approval and successful `runtime.get_current_time` execution, Loomi feeds the redacted tool result back to the model and streams a final assistant answer into the same run.

**Why this priority**: This is the missing vertical slice after approval execution. It proves Loomi can move from "tool result is visible" to "the assistant uses the tool result to answer" without introducing a general multi-step agent loop.

**Independent Test**: Use a fake provider or controlled provider response that first requests `runtime.get_current_time`, then expects a tool result in the continuation context and returns a final assistant message that includes the tool result.

**Acceptance Scenarios**:

1. **Given** a run has `tool_call_succeeded` with a redacted result, **When** the worker resumes model continuation, **Then** the next provider request includes prior conversation context, the model's tool request, and a synthetic tool-result context item derived from persisted run events.
2. **Given** the continuation provider streams text, **When** Loomi receives the second model stream, **Then** it persists ordered `model_delta` events after `tool_call_succeeded` and before one final assistant message.
3. **Given** the continuation completes, **When** the user reads the chat, **Then** the final assistant message uses the tool result and the run reaches completed state.

---

### User Story 2 - Keep denial terminal and understandable (Priority: P2)

A user denies a tool request. Loomi records denial and finishes the run without invoking the tool or sending the denial back to the model for another answer.

**Why this priority**: Denial is a safety boundary. The minimal behavior should be deterministic, auditable, and not allow the model to negotiate around a human decision.

**Independent Test**: Create an approval-required tool call, deny it, and verify no tool execution, no continuation provider call, and a final visible denied state.

**Acceptance Scenarios**:

1. **Given** a pending tool call, **When** the user denies it, **Then** Loomi records `tool_call_denied` and a terminal run state without calling the model again.
2. **Given** the denied run is replayed over SSE, **When** the frontend rebuilds Timeline and Chat, **Then** the tool card shows denial and no synthetic assistant answer claims tool output.

---

### User Story 3 - Fail safely on tool or continuation errors (Priority: P3)

A tool execution or continuation provider request fails. Loomi records a redacted failure and ends the run predictably without exposing secrets or raw provider/tool payloads.

**Why this priority**: Tool-result continuation adds a second model call, so failures must remain visible and bounded before broader loop behavior exists.

**Independent Test**: Force the tool executor to fail, then force the continuation provider to fail, and verify each path produces redacted events, no secret leakage, and one terminal state.

**Acceptance Scenarios**:

1. **Given** an approved tool fails, **When** Loomi records `tool_call_failed`, **Then** it does not call the model again and the run ends failed with a redacted error.
2. **Given** a tool succeeds but continuation provider fails, **When** Loomi records the provider error, **Then** the run ends failed while preserving the tool result event and any partial continuation draft as terminal context.

---

### User Story 4 - Display two model-stream phases clearly (Priority: P4)

A user watches a run where the first model phase requests a tool and the second model phase answers after the tool result. Timeline, RunRail, and assistant draft behavior make the two stream phases understandable without mixing tool events into message text.

**Why this priority**: Tool continuation changes the visual run shape. Users need to see why the assistant paused, what the tool returned, and when the assistant resumed.

**Independent Test**: Replay the full success event sequence from history and live SSE, then compare chat draft, final message, tool card, and timeline ordering.

**Acceptance Scenarios**:

1. **Given** a first model phase emits text before requesting a tool, **When** the tool is requested, **Then** the pre-tool draft is preserved as partial context and is not finalized as the final assistant answer.
2. **Given** the second model phase streams after `tool_call_succeeded`, **When** the frontend receives `model_delta` events, **Then** assistantDraft resumes or replaces draft content according to a documented phase model and finalizes once.
3. **Given** a replay starts after completion, **When** the frontend rebuilds from events, **Then** Timeline shows model phase 1, approval/tool execution, model phase 2, and final assistant message in order.

### Edge Cases

- The provider emits more than one tool call in a run.
- The continuation provider requests another tool after receiving the first tool result.
- The model emits pre-tool assistant text before the tool call.
- The tool result contains sensitive-looking keys, large payloads, or executor internals.
- The user denies the tool request.
- The approved tool fails.
- The continuation provider fails after some second-phase deltas.
- SSE reconnects between tool success and the second model phase.
- Window A lands approve/deny and execution contracts with slightly different endpoint names or event metadata.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST build tool-result continuation context from persisted conversation messages plus ordered run-event projection for the current run.
- **FR-002**: Loomi MUST represent the tool result in provider context as a synthetic provider-level tool result item that is derived from `tool_call_succeeded`, includes `tool_call_id`, `tool_name`, and redacted result content, and is not stored as a normal chat message.
- **FR-003**: Loomi MUST NOT persist a durable `messages.role = tool` row for MVP unless implementation discovers that the existing message schema cannot replay completed runs without it; the preferred source of truth is run events plus tool-call projection.
- **FR-004**: Loomi MUST keep the persistent message roles limited to user/assistant/system for MVP unless a later schema review proves a role extension is necessary.
- **FR-005**: Provider gateway interfaces MUST expose a continuation request shape that can include prior assistant tool-call metadata and one synthetic tool result while keeping provider-specific formatting inside provider adapters.
- **FR-006**: MVP MUST allow at most one tool call per run. If the first provider response contains multiple tool calls, Loomi MUST fail safely before execution. If the continuation provider requests another tool, Loomi MUST record a redacted unsupported-loop failure and end the run failed.
- **FR-007**: On the success path, the persisted event order MUST be: first model phase, tool request, approval required, approved, executing, succeeded, continuation model phase, final assistant message, run completed.
- **FR-008**: On the denied path, Loomi MUST NOT call the model again. It MUST record denial and complete or stop the run with a visible terminal denial state.
- **FR-009**: On tool execution failure, Loomi MUST NOT call the model again. It MUST record a redacted `tool_call_failed` event and fail the run.
- **FR-010**: On continuation provider failure, Loomi MUST record a redacted model/provider error, preserve any second-phase draft as terminal context, and fail the run exactly once.
- **FR-011**: SSE MUST deliver the second model phase as normal ordered model events after `tool_call_succeeded`, and replay MUST reconstruct the same ordering without a separate stream endpoint.
- **FR-012**: Frontend runtime state MUST distinguish pre-tool assistant draft text from post-tool continuation text and MUST create exactly one final assistant message for the run.
- **FR-013**: Redaction MUST happen before any tool result is written to run events, passed to the provider, exposed over SSE, shown in UI, or persisted in final assistant context metadata.
- **FR-014**: Tool result continuation MUST NOT introduce shell tools, filesystem tools, arbitrary network tools, MCP, browser automation, multi-agent behavior, long-term memory, or RAG.
- **FR-015**: The plan MUST identify Window A dependencies and must not require changes to approve/deny API paths unless a specific interface gap is found during implementation.
- **FR-016**: Documentation updates for implementation MUST cover architecture/tool-result-continuation, api/tool-call-approval, runbooks/local-m7, devlog, and roadmap/current-status.

### Key Entities *(include if feature involves data)*

- **Continuation Context**: Provider-ready conversation context assembled for the second model call after a successful tool result.
- **Synthetic Tool Message**: In-memory provider context item that represents a redacted tool result and links to the model's prior tool call id.
- **Tool Result Projection**: Safe result data derived from `tool_call_succeeded` events and/or tool-call current-state projection.
- **Model Phase**: A segment of model streaming within one run, such as `initial` before tool execution and `continuation` after tool result.
- **Assistant Draft**: Frontend-visible transient assistant text that may span or reset across model phases but finalizes once.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The approved success smoke path produces exactly one tool execution, exactly one continuation provider call, exactly one final assistant message, and one completed run in 100% of local test attempts.
- **SC-002**: Provider continuation tests prove the second request contains the redacted `runtime.get_current_time` result and does not contain raw executor internals or sensitive-looking keys.
- **SC-003**: SSE replay and live streaming produce identical event ordering for the full success path, including the second `model_delta` phase.
- **SC-004**: Denied and tool-failed paths make zero continuation provider calls in automated tests.
- **SC-005**: If the continuation provider requests another tool, Loomi records one safe unsupported-loop failure and does not execute a second tool.
- **SC-006**: Browser smoke shows the tool result in Timeline/ToolCallCard and the final assistant answer in Chat without duplicate final messages.

## Assumptions

- `origin/main` contains M7 approval-blocked tool-call foundation and the redacted tool-call projection.
- Window A is responsible for approve/deny endpoints, enabled approval UI actions, and approved `runtime.get_current_time` execution.
- This feature starts after Window A exposes a successful `tool_call_succeeded` event with redacted result metadata.
- `runtime.get_current_time` remains the only executable tool for this slice.
- Current persistent chat messages do not need `role = tool`; provider adapters can build provider-specific tool result items from run events.
- One active run per thread remains enforced.
- Documentation changes are implementation tasks, not part of this planning-only request.
