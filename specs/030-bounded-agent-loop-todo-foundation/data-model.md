# Data Model: M22 Bounded Agent Loop + Todo Foundation

## Agent Loop State

- `run_id`: Existing run id.
- `tool_call_count`: Number of tool calls accepted for this run.
- `max_tool_calls`: Configured hard limit for the run.
- `phase`: `initial_model`, `awaiting_tool_approval`, `executing_tool`, `continuing_model`, `completed`, `failed`, `stopped`, or `loop_limited`.
- `current_tool_call_id`: Optional current tool call id.

Validation:

- `tool_call_count` must not exceed `max_tool_calls`.
- Only one current tool call may be pending or executing.
- Terminal runs cannot accept new tool calls or continuation.

## Loop Tool Step

- `tool_call_id`: Provider tool call id.
- `tool_name`: Catalog tool name.
- `tool_index`: 1-based index within the run.
- `approval_status`: Existing tool approval status.
- `execution_status`: Existing tool execution status.
- `arguments_summary`: Existing redacted argument summary.
- `result_summary`: Existing redacted result summary.
- `error_code` / `error_message`: Safe failure summary.

Validation:

- Tool must be enabled in RunContext for this run.
- Tool request must pass existing argument validation.
- Approval is required before execution.

## Todo List Snapshot

- `items`: Ordered todo items.
- `updated_by`: `provider`, `runtime`, or `user`.
- `updated_at_event_id`: Run event id that produced the snapshot.
- `redaction_applied`: Boolean marker.

Validation:

- Snapshot must be bounded by item count and text length.
- Unsafe fields and secret-looking strings are redacted before persistence or UI replay.

## Todo Item

- `id`: Stable item id within the run.
- `title`: Short safe task title.
- `status`: `pending`, `running`, `completed`, `blocked`, or `failed`.
- `summary`: Optional safe summary.
- `redaction_applied`: Optional marker.

Validation:

- Title is required and length-bounded.
- Status must be one of the allowed values.
- Raw file contents, absolute private paths, shell commands, browser state, provider payloads, and secrets are not allowed.

## Loop Event

Existing run event with additional safe metadata:

- `loop_index`
- `loop_max`
- `model_phase`
- `tool_call_id`
- `tool_name`
- `todo_items`
- `redaction_applied`
- `loop_limit_reached`

Validation:

- Metadata must remain replay-safe and must not include raw provider/tool payloads.
