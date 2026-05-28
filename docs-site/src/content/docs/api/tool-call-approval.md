---
title: M7 Tool Call Approval API
description: Tool-call projection, event payloads, diagnostics fields, and Phase 2 API-facing contracts.
---

M7 supports the minimal approval execution loop for `runtime.get_current_time`: provider-requested calls are recorded, blocked for approval, approved or denied idempotently, executed by the worker after approval, and replayed through history-first SSE. M12 reuses the same approval projection for already-discovered local stdio MCP tools.

Local desktop Settings can also save one OpenAI-compatible `custom` model provider into the running local API process. That endpoint only returns redacted capability data and never echoes the API key.

## Local provider settings

`POST /v1/model-providers` saves the local OpenAI-compatible provider used by model gateway runs:

```json
{
  "base_url": "https://gateway.example.test/v1",
  "model": "gpt-5.5",
  "api_key": "sk-..."
}
```

Successful responses return only a redacted provider capability:

```json
{
  "provider": {
    "id": "custom",
    "family": "openai_compatible",
    "base_url": "https://gateway.example.test/v1",
    "model": "gpt-5.5",
    "status": "available"
  },
  "request_id": "req_..."
}
```

The saved provider updates both `GET /v1/model-providers` and the in-process model gateway provider list. The current slice does not write this provider to disk; restart the API to reset it unless environment variables also configure the provider.

## Tool-call projection

`tool_calls` stores the current redacted state for one model-requested tool call:

```json
{
  "id": "tool_...",
  "thread_id": "thread_...",
  "run_id": "run_...",
  "tool_call_id": "tc_1",
  "tool_name": "runtime.get_current_time",
  "arguments_summary": { "timezone": "UTC" },
  "approval_status": "required",
  "execution_status": "blocked",
  "requested_at": "2026-05-24T10:00:00Z",
  "updated_at": "2026-05-24T10:00:00Z"
}
```

A tool call is scoped by `thread_id`, `run_id`, and `tool_call_id`. The same `(run_id, tool_call_id)` request is idempotent and returns the existing projection without duplicating events. M7 allowed only one tool call per run; M22 starts a bounded Work-mode continuation path where a run can record a later workspace read tool call after the previous call reaches a terminal execution state.

## Run status

M7 extends active run states with:

| Status | Meaning |
| --- | --- |
| `blocked_on_tool_approval` | The run has a tool call waiting for user approval and must not execute that tool. |

This status participates in active-run uniqueness and readiness checks. Schema readiness now requires migration version `6`.

## Tool event metadata

Tool events are persisted as run events with redacted metadata:

```json
{
  "type": "tool_call_approval_required",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "runtime.get_current_time",
    "arguments_summary": { "timezone": "UTC" },
    "approval_status": "required",
    "execution_status": "blocked"
  }
}
```

Frontend API mapping converts these backend types to dotted runtime types such as `tool.call.approval_required` and keeps safe metadata available for replaying a stable `ToolCall` view model.

Known run/model/tool lifecycle events now include a Loomi run step summary in the same redacted metadata:

```json
{
  "run_step_kind": "approval",
  "run_step_status": "required",
  "run_step_summary": "Tool approval required"
}
```

`run_step_kind` can be `model_request`, `tool_requested`, `approval`, `tool_execution`, `continuation`, or `terminal`. `run_step_status` is a safe state such as `pending`, `required`, `approved`, `denied`, `running`, `succeeded`, `failed`, `completed`, or `stopped`. These fields are projection hints for timeline/debug and resume logic; clients must still treat `run_events.sequence` as the ordering source and `tool_calls` as the current tool-call projection.

For MCP calls, the same event names include safe source metadata:

```json
{
  "type": "tool_call_approval_required",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_mcp_1",
    "tool_name": "mcp.local-search.search",
    "tool_source": "mcp",
    "server_slug": "local-search",
    "arguments_summary": { "query": "status" },
    "approval_status": "required",
    "execution_status": "blocked"
  }
}
```

MCP approval is offered only when a prior discovery event lists the namespaced candidate and the selected persona allowed-tools snapshot resolves that same tool.

`tool_call_succeeded` may include a redacted result for model continuation:

```json
{
  "type": "tool_call_succeeded",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "runtime.get_current_time",
    "result_summary": {
      "iso_time": "2026-05-25T10:00:00Z",
      "timezone": "UTC",
      "source": "runtime"
    }
  }
}
```

If `result_for_model_redacted` is present, continuation uses that field. Otherwise it uses the safe `result_summary`. Raw executor output is never eligible for provider continuation.

Before provider serialization, oversized string fields inside that redacted result are compacted. The compacted string keeps early context, path/status/error-like lines, tail context, and a `[tool output compacted]` marker; small results are unchanged.

M92 preserves readable benign summaries during compaction. Lines containing credentials, tokens, secret markers, private host paths, or env-file references are redacted, but terminal labels such as `stdout` do not cause the whole result to collapse to `[redacted]`.

## Failure metadata

Terminal failures must carry a user-actionable summary in the error event and the final `run_failed` event:

| Code | Meaning |
| --- | --- |
| `provider_error`, `provider_timeout`, `provider_misconfigured`, `provider_rate_limited` | Upstream provider request failed or is not ready. |
| `tool_validation_failed` | Provider requested a tool with invalid arguments. The message includes the validation reason, such as a missing path. |
| `permission_not_authorized` | Tool or command is not authorized for the current run, mode, or approval boundary. |
| `workspace_unbound` | No usable workspace folder is bound for a workspace tool run. |
| `bounded_limit_reached` | A process/tool hit timeout or bounded output limits. |

Clients should display `error_message` or the event summary before falling back to a generic failed state. The generic "run failed" label is only a heading; it is not enough diagnostic content.

## Tool result continuation

After an approved tool succeeds, runtime can build one continuation request from:

- persisted thread messages through the triggering user message
- the matching `tool_call_requested` event
- the matching `tool_call_succeeded` event

The provider-neutral continuation context uses in-memory roles:

| Role | Purpose |
| --- | --- |
| `assistant_tool_call` | Replays the model's prior tool request to the provider adapter. |
| `tool_result` | Supplies the redacted tool result for the same `tool_call_id`. |

OpenAI-compatible providers serialize these as an assistant `tool_calls` message followed by a matching `tool` message. Loomi does not persist a durable `messages.role = tool` row for this MVP.

The second model stream reuses existing run events with `metadata.model_phase = "continuation"`. For M22, continuation can request another enabled workspace read tool and runtime records it as a fresh approval-required tool call:

```json
{
  "type": "tool_call_approval_required",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_read_2",
    "tool_name": "workspace.read",
    "approval_status": "required",
    "execution_status": "blocked"
  }
}
```

Only one non-terminal tool call may exist in a run. Workspace continuation is capped at three accepted tool calls; exceeding the cap records `tool_loop_limit_reached` and fails the run without recording the extra call. Continuation requests for non-workspace tools, tools outside the run's enabled tool snapshot, or Chat-mode-only tools still record `unsupported_tool_loop` and fail without execution. Repeating an already-requested `tool_call_id` during continuation records `duplicate_tool_call_id` and does not duplicate approval-required events.

M95 adds a runtime guard before recording Work-mode workspace tool requests. Directory inventory/classification intent must start with `workspace.tree_summary` or `workspace.list_directory`; first-call `workspace.grep`, broad inventory `workspace.glob`, and repeated same-argument `workspace.read`, `workspace.list_directory`, or `workspace.grep` requests fail with `tool_planner_guardrail`. These rejected requests do not create `tool_calls` rows; the terminal `run_failed` event carries safe metadata including `guardrail`, `tool_name`, `arguments_summary`, and, when applicable, `recommended_tool`.

M80 clarifies the durable resume contract without adding fields. If a worker retry sees an approved tool call that already reached `tool_call_succeeded`, it may resume the missing continuation only when the event suffix has no continuation start, later tool request, or final run event. Pending `tool_call_requested` events without a matching `tool_call_succeeded` are not serialized into provider continuation input.

M97 keeps that contract but moves the retry decision onto `RebuildRunStepState`: pending tools remain separate from completed results, a later continuation step suppresses duplicate continuation, and terminal steps stop late model/tool writes.

## Diagnostics

`GET /v1/diagnostics/worker-queue` now includes M7 counters:

```json
{
  "diagnostics": {
    "queue_status": "ready",
    "worker_status": "ready",
    "queued_count": 0,
    "leased_count": 0,
    "stale_count": 0,
    "retrying_count": 0,
    "blocked_tool_approval_count": 1,
    "resumable_tool_call_count": 0,
    "dead_count": 0,
    "updated_at": "2026-05-24T10:00:00Z"
  }
}
```

`blocked_tool_approval_count` counts tool calls with `approval_status = required` and `execution_status = blocked`. `resumable_tool_call_count` counts calls approved but not started.

## Approve and deny endpoints

The scoped tool-call read and decision paths are:

- `GET /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/deny`

Approve is valid from `approval_status = required` and `execution_status = blocked`. It records `tool_call_approved`, changes execution to `not_started`, and queues the existing M6 worker path for resume. Repeated approve returns the current approved projection without duplicating events or jobs.

Deny is valid before execution starts. It records `tool_call_denied`, cancels pending run jobs, marks the tool execution cancelled, and writes `run_stopped`. Repeated deny returns the current denied projection without duplicate events.

Wrong thread/run/user scope returns not found. Incompatible states such as terminal execution or reversing a denied call return a safe invalid request error.

## Execution events

Approved worker execution writes:

```json
{
  "type": "tool_call_succeeded",
  "category": "progress",
  "metadata": {
    "tool_call_id": "tc_1",
    "tool_name": "runtime.get_current_time",
    "approval_status": "approved",
    "execution_status": "succeeded",
    "result_summary": { "timezone": "UTC", "source": "runtime" }
  }
}
```

Failures use `tool_call_failed` with redacted `error_code` and `error_message`, then `run_failed`. MCP failures use safe codes such as `mcp_config_unavailable`, `mcp_stdio_timeout`, `mcp_stdio_exit`, `mcp_stdio_invalid_response`, or `mcp_tool_execution_failed`.

## Redaction and validation

Only schema-valid summaries may be persisted. `runtime.get_current_time` accepts no fields except optional `timezone`, and the only allowed value is `UTC`. Redaction runs before event/job/message metadata persistence and treats sensitive keys such as `api_key`, `authorization`, `password`, `secret`, `token`, and `credential` as always redacted.
