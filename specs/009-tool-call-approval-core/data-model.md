# Data Model: M7 Tool Call Approval Core

## Tool Call

A durable current-state projection for one model-requested internal tool invocation. Run events remain the audit and replay contract; Tool Call exists to make approve/deny idempotency, worker resume, and execution state safe under concurrency.

| Field | Type | Rules |
|-------|------|-------|
| `id` | string | Stable internal id, unique globally |
| `thread_id` | string | References one owned thread; must match the parent run |
| `run_id` | string | References one run |
| `tool_call_id` | string | Provider/model-facing tool call id when available, or Loomi-generated stable id; unique within a run |
| `tool_name` | string | Must match an allowlisted Tool Definition |
| `arguments_summary` | object | Schema-valid safe summary only; no raw provider payload; same name is used in API, events, and UI |
| `arguments_hash` | string | Stable hash of canonical validated arguments for idempotency/debug comparison without storing raw payload |
| `approval_status` | enum | `not_required`, `required`, `approved`, `denied`, `cancelled` |
| `execution_status` | enum | `not_started`, `blocked`, `executing`, `succeeded`, `failed`, `cancelled` |
| `result_redacted` | object/null | Safe result summary for successful execution |
| `error_code` | string/null | Stable user-safe error class |
| `error_message` | string/null | Redacted user-safe error summary |
| `requested_at` | timestamp | When Loomi recorded the request |
| `decided_at` | timestamp/null | When approval or denial was recorded |
| `started_at` | timestamp/null | When execution began |
| `completed_at` | timestamp/null | When execution reached terminal state |
| `updated_at` | timestamp | Latest state update |

Rules:

- A tool call belongs to exactly one thread and one run.
- `tool_call_id` must be unique within a run.
- Tool names must be allowlisted before approval or execution can proceed.
- Model-generated arguments are never trusted; only validated and redacted summaries may be persisted.
- `approval_status = required` pairs with `execution_status = blocked` until approve, deny, or cancel.
- `approval_status = denied` is terminal and must keep `execution_status = not_started` or `cancelled`; it must never execute.
- Terminal execution states are `succeeded`, `failed`, and `cancelled`.
- A terminal execution state must not transition to another terminal state.
- State updates must be idempotent for repeated approve/deny/resume requests.

Lifecycle mapping:

| Lifecycle | Approval Status | Execution Status | Required Event |
|-----------|-----------------|------------------|----------------|
| `requested` | `not_required` or `required` | `not_started` | `tool_call_requested` |
| `approval_required` | `required` | `blocked` | `tool_call_approval_required` |
| `approved` | `approved` | `not_started` | `tool_call_approved` |
| `denied` | `denied` | `not_started` | `tool_call_denied` |
| `executing` | `not_required` or `approved` | `executing` | `tool_call_executing` |
| `succeeded` | `not_required` or `approved` | `succeeded` | `tool_call_succeeded` |
| `failed` | any non-denied status | `failed` | `tool_call_failed` |
| `cancelled` | `cancelled` | `cancelled` | `tool_call_cancelled` |

State transitions:

```text
requested -> approval_required -> approved -> executing -> succeeded
requested -> approval_required -> approved -> executing -> failed
requested -> approval_required -> denied
requested -> approval_required -> cancelled
requested -> executing -> succeeded
requested -> executing -> failed
requested -> cancelled
executing -> cancelled
```

MVP constraint:

- M7 may execute at most one approved tool call per run. If a provider emits multiple tool calls, implementation tasks should either fail them safely or record unsupported/blocked events until a later milestone defines sequential execution.

## Tool Definition

An allowlisted internal tool contract.

| Field | Type | Rules |
|-------|------|-------|
| `name` | string | Stable name; MVP executable value is `runtime.get_current_time` |
| `description` | string | User-safe explanation shown in approval UI |
| `input_schema` | object | JSON-schema-like validation contract |
| `approval_policy` | enum | `always_required` or `not_required` |
| `safety_class` | enum | MVP class: `no_side_effect_internal` |
| `result_schema` | object | Safe result shape |
| `argument_redaction_policy` | enum/object | How validated arguments are summarized |
| `result_redaction_policy` | enum/object | How results/errors are summarized |
| `enabled` | boolean | Disabled tools must not execute |

MVP `runtime.get_current_time` definition:

| Attribute | Value |
|-----------|-------|
| `name` | `runtime.get_current_time` |
| `approval_policy` | `always_required` for M7 smoke, even though the tool is no-side-effect, to validate approval UX and idempotency |
| `safety_class` | `no_side_effect_internal` |
| `input_schema` | Object with optional `timezone` string; MVP allowlist is omitted or `UTC`, and omitted defaults to `UTC` |
| `result_schema` | Object containing `iso_time`, `timezone`, and `source` |
| Forbidden behavior | Shell, file reads/writes, arbitrary network, MCP, browser automation, secret access |

## Approval Decision

A user decision for one pending tool call.

| Field | Type | Rules |
|-------|------|-------|
| `tool_call_id` | string | Scoped to the target thread and run |
| `decision` | enum | `approve` or `deny` |
| `idempotency_key` | string/null | Optional client-provided key; state machine must also be idempotent without it |
| `decided_by` | string | Local user identity |
| `reason` | string/null | Optional user-safe denial reason or note |
| `created_at` | timestamp | Decision time |

Rules:

- Approve and deny endpoints are idempotent.
- Repeating the same decision returns current tool call state and must not duplicate events.
- Conflicting decisions after a terminal decision must not execute or reverse the prior terminal state.
- Decisions must be scoped by `thread_id`, `run_id`, and `tool_call_id`.

## Tool Result

A redacted outcome generated by an allowlisted internal tool executor.

| Field | Type | Rules |
|-------|------|-------|
| `tool_call_id` | string | References one Tool Call |
| `status` | enum | `succeeded`, `failed`, or `cancelled` |
| `result_redacted` | object/null | Safe success payload |
| `error_code` | string/null | Stable safe failure class |
| `error_message` | string/null | Redacted failure text |
| `created_at` | timestamp | Result time |

Rules:

- Results must be persisted through run events.
- Results must never contain raw provider payloads, API keys, shell output, user files, arbitrary URL contents, or unredacted secrets.
- Large results must be summarized or rejected before persistence.

## Run Event Extensions

M7 reuses existing persisted run events and adds tool lifecycle events. These events are the history-first SSE and UI audit contract.

| Category | Event Type | Purpose |
|----------|------------|---------|
| `progress` | `tool_call_requested` | Model requested an allowlisted or rejected tool name |
| `progress` | `tool_call_approval_required` | Tool call is blocked awaiting user decision |
| `progress` | `tool_call_approved` | User approved execution |
| `progress` | `tool_call_denied` | User denied execution |
| `progress` | `tool_call_executing` | Worker/executor started the tool |
| `progress` | `tool_call_succeeded` | Tool completed with redacted result |
| `error` | `tool_call_failed` | Tool validation or execution failed safely |
| `progress` | `tool_call_cancelled` | Tool was cancelled before terminal success/failure |

Required metadata for each event:

| Field | Rules |
|-------|-------|
| `thread_id` | Present through base run event or metadata; must match run |
| `run_id` | Present through base run event |
| `tool_call_id` | Present in metadata |
| `tool_name` | Safe allowlisted name or safe rejected name summary |
| `arguments_summary` | Redacted, schema-valid summary when available |
| `approval_status` | Current approval state |
| `execution_status` | Current execution state |
| `result_summary` | Redacted success result when available |
| `error_code` / `error_message` | Redacted failure details when available |

Rules:

- Events are ordered by existing `(sequence, id)` semantics within a run.
- Tool events must be distinguishable from model stream events in event type and frontend grouping.
- No event may include raw provider payloads, credentials, Authorization headers, API keys, shell output, file contents, arbitrary URL contents, or unvalidated arguments.

## Worker Block/Resume State

The M6 worker-visible state for approval waits.

| State | Meaning |
|-------|---------|
| `running` | Worker is processing model/provider/tool-safe work |
| `blocked_on_tool_approval` | Run has a tool call waiting for user decision; worker must not keep executing that tool |
| `resumable_after_tool_approval` | Approval was recorded and existing job pipeline may resume execution |
| `stopped` | Stop request or terminal run cancellation prevents tool execution |

Rules:

- A worker must not hold an active lease indefinitely while waiting for human approval.
- Approval should schedule or wake the existing M6 job/run pipeline rather than create a parallel queue.
- Stop/cancel must win over later resume or success writes.
- Recovery must not duplicate tool terminal events.

## Tool Result Context Boundary

M7 defines but may not fully implement result-to-model continuation.

| Field | Rule |
|-------|------|
| `tool_call_id` | Identifies which model-requested tool result is being returned |
| `tool_name` | Must match executed Tool Definition |
| `result_for_model_redacted` | Safe summarized result only |
| `eligible_for_next_model_request` | Boolean boundary flag or equivalent design marker |

Rules:

- MVP may stop after recording tool result and finalizing the run.
- Future multi-step loops must consume only redacted tool results, never raw executor internals.
- Loop limits, repeated approval handling, and context budgeting are deferred.
