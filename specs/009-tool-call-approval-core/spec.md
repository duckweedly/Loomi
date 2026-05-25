# Feature Specification: M7 Tool Call Approval Core

**Feature Branch**: `009-tool-call-approval-core`

**Created**: 2026-05-24

**Status**: Draft

**Input**: User description: "M7: build the minimal safe Tool Call / Approval Core so Loomi can move from real model chat to model-requested tools that pass through an approval, safety, validation, audit, worker, and history-first SSE boundary. Planning only; do not implement code."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Observe tool requests safely (Priority: P1)

A user starts a model-backed run in a thread. If the model requests a supported tool, Loomi records the request as a first-class tool-call lifecycle event sequence, shows it in the timeline and ToolCallCard, validates and redacts the arguments, and does not execute anything until the required approval state is satisfied.

**Why this priority**: This is the core safety boundary for moving beyond M5's non-executed tool boundary. It delivers value even before rich approval UX or multi-step model continuation exists because users can see exactly what the model requested and why execution is blocked.

**Independent Test**: Use a fake provider or controlled model response that requests the MVP tool. Verify persisted history contains the tool call lifecycle up to `approval_required` or `approved`, the UI displays the tool name and redacted argument summary, and no shell, file, or arbitrary network action occurs.

**Acceptance Scenarios**:

1. **Given** a model-backed run emits a supported tool request, **When** Loomi receives the request, **Then** Loomi persists `tool_call_requested` and the next safe lifecycle event as ordered run events associated with the same thread and run.
2. **Given** the requested tool requires user approval, **When** the request is recorded, **Then** Loomi stops at `approval_required`, shows approve/deny controls, and does not execute the tool automatically.
3. **Given** model-provided arguments contain unexpected fields or sensitive-looking values, **When** Loomi records the request, **Then** only schema-valid and redacted argument summaries are visible in run events and UI.
4. **Given** a user reconnects to a run stream after the tool request was recorded, **When** history-first SSE replays the run, **Then** the tool events appear in the same order as originally persisted.

---

### User Story 2 - Approve or deny a pending tool call (Priority: P2)

A user reviews a pending tool request and chooses approve or deny. Loomi records the decision idempotently, resumes only the safe execution path after approval, and records denial as a terminal tool-call decision without executing the tool.

**Why this priority**: Approval is the product boundary that makes tool use user-controlled and auditable. Idempotent approve/deny prevents duplicate execution and makes UI retries safe.

**Independent Test**: Create a pending approval-required tool call. Send approve twice and verify a single `approved` decision and at most one execution attempt. Create another pending call, send deny twice, and verify a single `denied` decision and no execution attempt.

**Acceptance Scenarios**:

1. **Given** a tool call is waiting for approval, **When** the user approves it once or retries the approve action, **Then** Loomi records exactly one approved decision and proceeds toward execution once.
2. **Given** a tool call is waiting for approval, **When** the user denies it once or retries the deny action, **Then** Loomi records exactly one denied decision and never executes the tool.
3. **Given** a tool call is already approved, denied, executing, succeeded, failed, or cancelled, **When** the same decision request is repeated, **Then** Loomi returns the current state without creating duplicate lifecycle events.
4. **Given** a user tries to approve or deny a tool call from another thread or run, **When** Loomi validates the request, **Then** the action is rejected without exposing details outside the scoped thread/run boundary.

---

### User Story 3 - Execute the safest MVP internal tool (Priority: P3)

After approval or when a no-approval internal policy allows execution, Loomi executes exactly one allowlisted no-side-effect internal tool, records executing and terminal lifecycle events, and displays the result or redacted error in the ToolCallCard, RunRail, and Timeline.

**Why this priority**: A runnable vertical slice must demonstrate more than blocking. The safest useful tool proves the invocation path, validation, event recording, worker resume behavior, and UI result rendering without introducing shell, file, network, or secret risk.

**Independent Test**: Approve the MVP tool call and verify Loomi records `executing` then `succeeded` with a safe result. Force a validation or tool error and verify `failed` with a redacted error. Stop the run while waiting or executing and verify `cancelled` without later success.

**Acceptance Scenarios**:

1. **Given** an approved MVP tool call, **When** the worker resumes the blocked run, **Then** Loomi records `tool_call_executing` and exactly one terminal tool event: `tool_call_succeeded`, `tool_call_failed`, or `tool_call_cancelled`.
2. **Given** the MVP tool produces a result, **When** Loomi persists the result, **Then** the result is redacted, associated with `thread_id`, `run_id`, `tool_call_id`, and visible through history-first SSE.
3. **Given** the run is stopped while a tool call is waiting or executing, **When** Loomi observes the cancellation, **Then** the tool call reaches `cancelled` and no later result is applied.
4. **Given** the model requests any non-allowlisted tool, **When** Loomi evaluates the request, **Then** the call fails or is denied at the safety boundary and no execution is attempted.

---

### User Story 4 - Keep tool events distinct from model streaming (Priority: P4)

A user or developer reviewing a run can distinguish model stream output from tool lifecycle, approval, execution, and result events in the RunRail and Timeline.

**Why this priority**: Tool calls are safety-significant state transitions. They must not be visually or contractually mixed into normal model output, or users cannot audit what the agent attempted.

**Independent Test**: Replay a run containing model deltas, a tool request, an approval decision, execution, and a result. Verify model events and tool events are grouped or labeled distinctly and remain ordered.

**Acceptance Scenarios**:

1. **Given** a run has both model output and tool events, **When** the user opens the timeline, **Then** tool events appear in a clear tool group or with distinct labels separate from model stream rows.
2. **Given** a tool event includes arguments or result metadata, **When** the UI renders it, **Then** it shows a safe summary and never shows raw provider payloads, API keys, or unredacted secrets.
3. **Given** a history replay includes completed tool calls, **When** the RunRail summarizes the run, **Then** tool lifecycle status contributes to a clear execution state without replacing model-stream state.

---

### Edge Cases

- A model emits malformed tool arguments, unknown fields, wrong types, unsupported tool names, duplicate tool call ids, or multiple tool calls in one provider response.
- A browser retries approve or deny after a network timeout.
- A user approves after another request already denied or cancelled the tool call.
- A worker claims a run that is blocked on approval, loses its lease, or resumes after approval.
- A run is stopped while a tool call is pending approval or executing.
- A tool execution fails internally or produces an unexpectedly large result.
- A provider emits text before and after a tool request, while MVP may not yet implement full multi-step tool-result continuation.
- History-first SSE reconnects after every lifecycle state.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Loomi MUST represent every supported tool call with this lifecycle: `requested`, `approval_required`, `approved`, `denied`, `executing`, `succeeded`, `failed`, `cancelled`.
- **FR-002**: Loomi MUST map every tool-call lifecycle transition to persisted run events that are available through history-first SSE replay.
- **FR-003**: Every persisted tool-call record or equivalent event projection MUST associate `thread_id`, `run_id`, `tool_call_id`, `tool_name`, validated/redacted argument summary, approval status, execution status, and result or error when available.
- **FR-004**: Loomi MUST choose `runtime.get_current_time` as the first MVP tool because it has no side effects, needs no user data access, needs no network access, needs no file access, can return a deterministic safe shape, and exercises the invocation path with the smallest privacy/security surface.
- **FR-005**: Loomi MUST reject or fail non-allowlisted tool names without executing them.
- **FR-006**: Loomi MUST NOT execute shell commands, read or write user files, perform arbitrary network requests, use MCP tools, automate browsers, run multi-agent workflows, use long-term memory/RAG, or bypass approval in M7.
- **FR-007**: Tool arguments generated by a model MUST be treated as untrusted data and MUST pass schema validation before any approval or execution state can proceed.
- **FR-008**: Run events, tool results, diagnostics, and UI summaries MUST redact secrets, credentials, raw provider payloads, and sensitive-looking argument values.
- **FR-009**: Tools marked as requiring approval MUST stop at `approval_required` and MUST NOT auto-execute.
- **FR-010**: Approve and deny actions MUST be idempotent: repeated same decisions MUST return the current state without duplicate approval, denial, execution, terminal, or result events; conflicting decisions that would reverse an incompatible terminal or already-progressed state MUST be rejected safely.
- **FR-011**: The M6 worker pipeline MUST treat approval-required tool calls as a blocked run state that releases or pauses active execution until a decision resumes or terminates the tool path.
- **FR-012**: Tool results and tool errors MUST be written back as run events associated with the originating tool call and visible in history replay.
- **FR-013**: M7 MUST define the boundary for feeding tool results back into the next model turn, but MVP MAY stop after recording the tool result and finalizing the run without implementing a full multi-step model loop.
- **FR-014**: The existing ToolCallCard MUST show tool name, argument summary, approval-required state, approve/deny controls, executing state, and result or error state.
- **FR-015**: RunRail and Timeline MUST group or label tool events distinctly from model stream events.
- **FR-016**: Cancellation MUST move pending or executing tool calls to `cancelled` and prevent later success or failure from overwriting cancellation.
- **FR-017**: The feature MUST preserve one active run per thread and existing run/event/message ownership boundaries.
- **FR-018**: Planning and implementation tasks MUST keep M7 scoped to minimal interfaces layered on M5 gateway and M6 worker/job pipeline, without broad worker rewrites unless a minimal resume/block interface is required.

### Key Entities *(include if feature involves data)*

- **Tool Call**: A model-requested internal tool invocation associated with one thread and one run. It has a stable `tool_call_id`, tool name, validated/redacted arguments, approval state, execution state, result or error summary, and timestamps.
- **Tool Definition**: An allowlisted internal tool contract with name, description, JSON schema, approval policy, execution safety class, and result redaction policy.
- **Approval Decision**: A user action that approves or denies a pending tool call and can be replayed safely without duplicate effects.
- **Tool Result**: A redacted success or failure payload generated by the internal tool executor and persisted through run events.
- **Run Event**: The existing persisted event contract extended with tool lifecycle events and safe metadata.
- **Worker Block/Resume State**: The M6 worker-visible state that distinguishes a run waiting for approval from queued/running/recovering/stopped terminal states.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of tool-call lifecycle transitions in the M7 smoke path are visible through persisted history and history-first SSE replay in order.
- **SC-002**: 100% of approval-required MVP tool calls remain unexecuted until an approve action is recorded.
- **SC-003**: Repeating approve or deny requests 10 times for the same tool call creates one decision event and at most one execution attempt.
- **SC-004**: The MVP tool-call smoke path completes from model tool request to visible result or denial in under 10 seconds in local validation after the user decision is made.
- **SC-005**: Non-allowlisted tool names, malformed arguments, and sensitive-looking argument values produce safe failure or redacted display states with no shell, file, arbitrary network, MCP, browser automation, or secret exposure.
- **SC-006**: A user can identify tool lifecycle events separately from model stream events in the Timeline and RunRail during browser smoke without inspecting developer logs.
- **SC-007**: Stopping a run while a tool call is pending or executing produces a visible cancelled tool state and no later succeeded event in all local smoke attempts.

## Assumptions

- M6 worker/job pipeline is present on `origin/main` and provides durable background jobs, lease/recovery/cancel behavior, diagnostics, and pipeline events.
- M5 LLM gateway already normalizes provider tool/function-call output into non-executed boundary events; M7 upgrades the allowlisted subset from blocked-only to approval-gated internal execution.
- The first M7 slice uses a fixed local development identity and existing thread/run ownership model.
- `runtime.get_current_time` is the only executable MVP tool for M7; `thread.list_recent_messages` and `diagnostics.get_worker_queue` remain design alternatives because they expose user or operational data and require stricter disclosure/redaction policy.
- MVP supports one executable tool call per run; duplicate `tool_call_id` or multiple simultaneous tool requests must fail safely with redacted events unless a later plan adds sequential handling.
- Tool-result continuation into another provider request is a designed boundary only for M7; full multi-step agent loops are out of scope.
- Documentation site updates are part of done during implementation, but this planning request only creates the specs and docs-site update plan.
