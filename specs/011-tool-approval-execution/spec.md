# Feature Specification: M7 Tool Approval Execution Closure

**Feature Branch**: `011-tool-approval-execution`

**Created**: 2026-05-25

**Status**: Draft

**Input**: Complete the minimal safe approval execution loop on top of the merged M7 foundation: model requests `runtime.get_current_time`, run blocks on approval, user approves or denies, approved calls execute through the worker, terminal tool events replay through SSE, and the frontend shows the current lifecycle.

## User Scenarios & Testing

### User Story 1 - Approve or deny a pending tool call (Priority: P1)

A user sees an approval-required `runtime.get_current_time` tool card and chooses approve or deny. Repeated clicks or network retries are safe.

**Why this priority**: Approval is the explicit safety boundary. Without reliable idempotent decisions, execution cannot safely resume.

**Independent Test**: Create an approval-required tool call, call approve twice and deny twice on separate runs, and verify one decision event per call, scoped thread/run validation, and no duplicate resume or execution.

**Acceptance Scenarios**:

1. **Given** a tool call has `approval_status=required`, **When** approve is posted, **Then** the call transitions to `approved`, one `tool.call.approved` run event is written, and the run becomes resumable.
2. **Given** the same approve request is retried, **When** the call is already approved and not terminal, **Then** the API returns the current approved projection without a duplicate event.
3. **Given** a tool call has `approval_status=required`, **When** deny is posted, **Then** the call transitions to `denied`, one `tool.call.denied` run event is written, no execution is scheduled, and the run finishes as stopped/failed with a user-safe denial summary.
4. **Given** a caller uses the wrong thread, run, user, unknown id, or terminal/incompatible state, **When** approve or deny is posted, **Then** Loomi rejects the action without leaking cross-scope details or reversing terminal state.

---

### User Story 2 - Execute approved `runtime.get_current_time` (Priority: P2)

After approval, the worker resumes the blocked run, executes only the allowlisted current-time tool, and records executing plus a terminal result or error event.

**Why this priority**: The runnable slice must prove that approval leads to controlled execution, not only state changes.

**Independent Test**: Approve a pending current-time call and verify `tool.call.executing` then `tool.call.succeeded` with a redacted result. Force executor failure and verify `tool.call.failed` with redacted error code/message.

**Acceptance Scenarios**:

1. **Given** an approved `runtime.get_current_time` call, **When** the worker resumes, **Then** it writes `tool.call.executing` before execution.
2. **Given** execution succeeds, **When** the result is persisted, **Then** Loomi writes `tool.call.succeeded` with a redacted `result_summary`.
3. **Given** validation or execution fails, **When** the failure is persisted, **Then** Loomi writes `tool.call.failed` with stable redacted `error_code` and `error_message`.
4. **Given** any non-allowlisted tool, timezone other than omitted or `UTC`, shell/filesystem/network/MCP/browser request, or multi-tool execution attempt, **When** execution is considered, **Then** no such capability runs.

---

### User Story 3 - Replay and render approval execution states (Priority: P3)

A reconnecting frontend receives history-first SSE for approval, execution, and terminal events and updates ToolCallCard, RunRail, and Timeline without manual refresh.

**Why this priority**: Tool execution is safety-significant; the UI must make the lifecycle visible and auditable.

**Independent Test**: Replay a mixed run containing approval-required, approved/denied, executing, succeeded/failed events and verify the adapter and UI render the same states for history and live SSE.

**Acceptance Scenarios**:

1. **Given** history-first SSE includes approval/execution/result events, **When** the frontend adapter replays them, **Then** ToolCallCard reaches approved, executing, succeeded, failed, or denied state correctly.
2. **Given** ToolCallCard shows approve/deny controls, **When** the user clicks either action, **Then** the card shows loading, disables duplicate input, handles API errors, and updates from returned state or SSE.
3. **Given** Timeline or RunRail shows a run with tool lifecycle events, **When** approval/execution/result events arrive, **Then** they are visibly distinct from model stream rows.

## Edge Cases

- Browser retries approve or deny after timeout.
- Approve is attempted after deny, after terminal execution, or for an unknown/wrong-scope tool call.
- Deny is attempted after approval but before execution starts.
- Worker resumes the same approved call more than once.
- Tool execution fails or produces unexpectedly verbose data.
- SSE reconnect happens after every lifecycle state.
- Model asks for a tool other than `runtime.get_current_time`, a non-UTC timezone, more than one tool, shell, filesystem, network, MCP, or browser automation.

## Requirements

### Functional Requirements

- **FR-001**: Loomi MUST expose `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve`.
- **FR-002**: Loomi MUST expose `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/deny`.
- **FR-003**: Approve and deny MUST be idempotent for repeated same-decision requests.
- **FR-004**: Approve MUST only transition `approval_status=required` tool calls to `approved`; repeated approved reads may return current state without new events.
- **FR-005**: Deny MUST only transition pre-execution safe states to `denied`; denied calls MUST NOT execute.
- **FR-006**: Deny MUST finalize the MVP run as stopped or failed and persist a visible safe event explaining that the user declined the tool.
- **FR-007**: Approve MUST move a run out of `blocked_on_tool_approval` into a worker-resumable state.
- **FR-008**: The worker MUST execute approved `runtime.get_current_time` calls and no other tool.
- **FR-009**: The worker MUST persist `tool.call.executing` before invoking the tool executor.
- **FR-010**: Successful execution MUST persist `tool.call.succeeded` with redacted `result_summary`.
- **FR-011**: Failed execution MUST persist `tool.call.failed` with redacted `error_code` and `error_message`.
- **FR-012**: The runtime tool schema MUST continue to allow timezone omitted or exactly `UTC` only.
- **FR-013**: The implementation MUST NOT add shell, filesystem, arbitrary network, MCP, browser automation, multi-tool concurrency, or multi-agent loop capabilities.
- **FR-014**: ToolCallCard approve/deny buttons MUST call the real API and expose loading, disabled, and error states.
- **FR-015**: History-first SSE MUST include approval, execution, result, failure, and denial events in persisted order.
- **FR-016**: The frontend runtime adapter MUST map succeeded, failed, and denied tool events into stable view-model states.
- **FR-017**: RunRail and Timeline MUST display approval, execution, result, failure, and denial events distinctly from model streaming.
- **FR-018**: Tests MUST cover backend approve idempotency, deny idempotency, wrong-scope/unknown/terminal rejection, worker execution, redaction, SSE replay, ToolCallCard actions, and frontend event mapping.
- **FR-019**: Documentation MUST update `architecture/tool-call-approval.md`, `api/tool-call-approval.md`, `runbooks/local-m7.md`, `devlog/2026-05-25-m7-approval-execution.md`, and `roadmap/current-status.md`.

### Key Entities

- **Tool Call**: Existing M7 projection for one model-requested tool with approval status, execution status, summaries, and timestamps.
- **Approval Decision**: Idempotent user action that approves or denies a required tool call.
- **Tool Execution**: Worker-owned execution attempt for an approved allowlisted tool.
- **Tool Result**: Redacted success or error summary persisted as run events and projection state.
- **Run Event**: Existing history-first SSE event stream extended by approval/execution terminal events.

## Success Criteria

- **SC-001**: Approve and deny retries create one decision event and do not duplicate execution.
- **SC-002**: Approved `runtime.get_current_time` completes from approval to visible succeeded/failed state in under 10 seconds locally.
- **SC-003**: Denied calls finish visibly without executor invocation.
- **SC-004**: SSE replay after reconnect reconstructs approval/execution/result states in ToolCallCard, RunRail, and Timeline.
- **SC-005**: Required backend, frontend, web build, docs build, and browser smoke validations run or report exact blockers.

## Assumptions

- M7 foundation from `009-tool-call-approval-core` is already merged into `origin/main`.
- Existing `tool_calls`, `blocked_on_tool_approval`, run events, placeholder ToolCallCard, allowlisted `runtime.get_current_time`, and strict timezone validation remain the base.
- MVP may stop after tool result or denial; full tool-result-to-model continuation remains out of scope.
