# Data Model: M7 Tool Approval Execution Closure

## Tool Call

Existing projection for one tool request.

Required behavior:

- Scoped by `thread_id`, `run_id`, and `tool_call_id`.
- `approval_status` transitions from `required` to `approved` or `denied`.
- `execution_status` transitions from blocked/not-started to `executing`, then `succeeded` or `failed`.
- Terminal statuses cannot be overwritten by conflicting decisions.
- Result and error fields store redacted summaries only.

## Approval Decision

User action submitted through approve/deny endpoints.

Rules:

- Approve is valid only from `required`.
- Repeated approve returns current approved state without duplicate events.
- Deny is valid from pre-execution safe states only.
- Repeated deny returns current denied state without duplicate events.
- Conflicting reversal of denied, executing, succeeded, failed, cancelled, or incompatible approved state is rejected.

## Tool Execution

Worker action for an approved call.

Rules:

- Execute only `runtime.get_current_time`.
- Validate timezone omitted or `UTC` immediately before execution.
- Persist `tool.call.executing` before invoking the executor.
- Persist exactly one terminal event: `tool.call.succeeded` or `tool.call.failed`.

## Run Event

Persisted audit event and SSE payload.

Required event names:

- `tool.call.approved`
- `tool.call.denied`
- `tool.call.executing`
- `tool.call.succeeded`
- `tool.call.failed`

Existing foundation event names may be mapped if already standardized, but docs and frontend must use one stable contract.

## Frontend Tool View Model

Adapter-owned state derived from history-first and live SSE.

Fields:

- `toolCallId`
- `toolName`
- `approvalStatus`
- `executionStatus`
- `argumentsSummary`
- `resultSummary`
- `errorCode`
- `errorMessage`
- `actionState`: `idle`, `approving`, `denying`, or `error`
