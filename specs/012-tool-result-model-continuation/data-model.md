# Data Model: Tool Result Model Continuation

## Continuation Context

Provider-ready context for the second model call after one successful tool execution.

| Field | Type | Rules |
|-------|------|-------|
| `thread_id` | string | Must match the parent run and conversation messages |
| `run_id` | string | Current run only |
| `model_phase` | enum | `continuation` for the second model call |
| `messages` | list | Existing conversation messages plus provider-facing synthetic items |
| `tool_call_id` | string | Must match the prior model-requested tool call |
| `tool_name` | string | Must match executed tool definition |
| `result_for_model_redacted` | object/string | Safe result payload only |
| `loop_count` | integer | MVP must be `1` when the continuation call starts |

Rules:

- Built from persisted messages plus current run events.
- Must include enough prior assistant tool-call metadata for provider adapters to associate the result with the request.
- Must not include raw provider payloads, raw tool arguments, raw executor internals, secrets, file contents, shell output, arbitrary network responses, or unredacted errors.
- Must not be persisted as a standalone chat message.

## Synthetic Tool Message

An in-memory gateway item representing a redacted tool result.

| Field | Type | Rules |
|-------|------|-------|
| `role` | enum | Gateway-neutral role such as `tool_result`, not necessarily persisted |
| `tool_call_id` | string | Links to the model's requested tool call id |
| `tool_name` | string | Safe allowlisted tool name |
| `content` | object/string | Redacted result content prepared for model context |
| `source_event_id` | string | Optional pointer to `tool_call_succeeded` event for diagnostics |

Rules:

- Exists only inside continuation prompt construction.
- May serialize as OpenAI-compatible `role: "tool"` inside a provider adapter.
- Must not require adding `role = tool` to durable Loomi messages for MVP.

## Tool Result Projection

Safe result state available after Window A's execution slice.

| Field | Type | Rules |
|-------|------|-------|
| `tool_call_id` | string | Unique within run |
| `tool_name` | string | MVP value `runtime.get_current_time` |
| `execution_status` | enum | Must be `succeeded` before continuation |
| `result_summary` | object/string | UI-safe result |
| `result_for_model_redacted` | object/string | Provider-safe result, may match result summary |
| `completed_at` | timestamp | Tool terminal time |

Rules:

- Derived from `tool_call_succeeded` event metadata and/or `tool_calls.result_redacted`.
- Result shape for `runtime.get_current_time` should contain only safe time fields such as `iso_time`, `timezone`, and `source`.

## Model Phase

A segment of provider streaming within one run.

| Phase | Meaning | Terminal Behavior |
|-------|---------|-------------------|
| `initial` | First model call before tool execution | May end with a tool request instead of final assistant answer |
| `continuation` | Second model call after tool success | Must end with final assistant answer or model/provider failure |

Rules:

- MVP permits at most `initial` and `continuation`.
- A tool request in `continuation` is an unsupported-loop failure.
- Event metadata should allow frontend grouping by phase without changing SSE transport.

## Assistant Draft

Frontend state for streaming assistant text.

| Field | Type | Rules |
|-------|------|-------|
| `run_id` | string | Current run |
| `phase` | enum | `initial` or `continuation` |
| `text` | string | Current draft text |
| `status` | enum | `streaming`, `paused_for_tool`, `final`, `failed` |

Rules:

- Pre-tool text must not create a final assistant message unless the provider explicitly finalizes without a tool.
- When a tool is requested, draft becomes `paused_for_tool`.
- During continuation, implementation may either append to the same draft with phase markers or replace visible draft with continuation text, but finalization must produce exactly one assistant message.
- History replay must reconstruct the same final message from persisted events and message state.

## State Transitions

Success path:

```text
initial model streaming
-> tool_call_requested
-> tool_call_approval_required
-> tool_call_approved
-> tool_call_executing
-> tool_call_succeeded
-> continuation_context_built
-> continuation model streaming
-> final assistant message
-> run_completed
```

Denied path:

```text
tool_call_approval_required
-> tool_call_denied
-> terminal denied/stopped run
```

Tool failure path:

```text
tool_call_approved
-> tool_call_executing
-> tool_call_failed
-> run_failed
```

Continuation failure path:

```text
tool_call_succeeded
-> continuation model streaming?
-> model/provider error
-> run_failed
```
