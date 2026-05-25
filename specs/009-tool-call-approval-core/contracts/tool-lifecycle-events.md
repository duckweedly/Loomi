# Contract: M7 Tool Lifecycle Events

M7 tool events extend the existing run event stream. They are persisted first, then delivered by history-first SSE.

## Event Ordering

Tool events use the existing per-run sequence order. The expected happy path for the MVP approval-required tool is:

```text
model_request_started
model_output_delta?                  # Optional text before tool request
tool_call_requested
tool_call_approval_required          # Run becomes blocked_on_tool_approval
# user approves
tool_call_approved
tool_call_executing
tool_call_succeeded
run_completed                        # MVP may finalize without next model turn
```

Denial path:

```text
tool_call_requested
tool_call_approval_required
tool_call_denied
run_stopped                          # MVP treats user denial as a visible stopped run with no tool execution
```

Failure path:

```text
tool_call_requested
tool_call_failed                     # validation, unsupported tool, or executor failure
run_failed                           # MVP treats unsafe/failed tool execution as a visible failed run
```

Cancellation path:

```text
tool_call_requested
tool_call_approval_required or tool_call_executing
stop_requested
tool_call_cancelled
run_stopped
```

## Required Metadata

Every tool lifecycle event metadata object must contain:

| Field | Requirement |
|-------|-------------|
| `tool_call_id` | Stable id scoped to one run |
| `tool_name` | Safe name, preferably allowlisted exact name |
| `arguments_summary` | Redacted schema-valid summary when validation succeeds; omitted or safe error summary when validation fails |
| `approval_status` | Current approval status after the event |
| `execution_status` | Current execution status after the event |
| `result_summary` | Present only on `tool_call_succeeded`, redacted |
| `error_code` | Present on failed/cancelled states when useful |
| `error_message` | Redacted user-safe failure or cancellation summary |

Metadata must not contain raw provider payloads, API keys, Authorization headers, shell output, file contents, arbitrary URL contents, or unvalidated argument objects.

## Lifecycle Events

### `tool_call_requested`

Emitted after Loomi detects a provider/model tool request and before approval or execution decisions.

Rules:

- Must be persisted before any `approval_required`, `executing`, `failed`, or `cancelled` state.
- Must include `tool_call_id` and `tool_name`.
- Must include redacted argument summary only after schema validation; malformed arguments may be represented by a safe validation error.

### `tool_call_approval_required`

Emitted when the tool definition requires approval.

Rules:

- Must stop execution until an approve, deny, or cancellation decision.
- Must make the run visible as blocked on approval.
- Must not be followed by `tool_call_executing` unless `tool_call_approved` is recorded first.

### `tool_call_approved`

Emitted when a user approves a pending tool call.

Rules:

- Idempotent repeats must not duplicate this event.
- Must schedule or wake M6 worker/resume behavior exactly once.

### `tool_call_denied`

Emitted when a user denies a pending tool call.

Rules:

- Must be terminal for the tool call.
- Must not be followed by `tool_call_executing`.
- Idempotent repeats must not duplicate this event.

### `tool_call_executing`

Emitted when the internal executor begins the allowlisted tool.

Rules:

- Must require either `approval_status = approved` or an explicit `not_required` policy.
- MVP implementation should still use approval for `runtime.get_current_time` to validate the approval loop.

### `tool_call_succeeded`

Emitted when the internal tool returns a safe result.

Rules:

- Must include redacted result summary.
- Must be the only success event for a tool call.

### `tool_call_failed`

Emitted when validation, policy, or execution fails.

Rules:

- Must include stable error code and redacted message.
- Must not expose raw argument or executor internals.

### `tool_call_cancelled`

Emitted when stop/cancel wins over pending or executing tool work.

Rules:

- Must prevent later succeeded/failed writes from replacing cancellation.
- Must be replayable through history-first SSE.

## Frontend Grouping Contract

Timeline and RunRail must treat these events as tool events, not model stream events. `ToolCallCard` derives its display from the latest event or current tool-call projection:

- Requested: tool name and argument summary.
- Approval required: approve/deny controls.
- Approved: decision accepted, waiting to execute or resuming.
- Denied: terminal non-executing state.
- Executing: progress state.
- Succeeded: redacted result.
- Failed: redacted error.
- Cancelled: cancelled state.
