---
title: M7 Tool Call Approval API
description: Tool-call projection, event payloads, diagnostics fields, and Phase 2 API-facing contracts.
---

M7 adds the backend and frontend contracts needed for approval-gated internal tool calls. The US1 observable request slice records provider-requested `runtime.get_current_time` calls, replays approval-required events, and exposes scoped current-state reads. The US2 decision slice adds idempotent approve/deny HTTP actions and frontend client hooks. The US3 execution slice schedules the existing worker after approval, executes the approved `runtime.get_current_time` call, and records terminal result, failure, or cancellation events. US4 keeps mixed model/tool replay ordered and visibly grouped.

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

A tool call is scoped by `thread_id`, `run_id`, and `tool_call_id`. The same `(run_id, tool_call_id)` request is idempotent and returns the existing projection without duplicating events. M7 MVP allows only one tool call per run.

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

## Tool-call endpoints

Tool calls are scoped by thread, run, and provider `tool_call_id`:

- `GET /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/approve`
- `POST /v1/threads/{thread_id}/runs/{run_id}/tool-calls/{tool_call_id}/deny`

`approve` records one `tool_call_approved` event and moves the projection to:

```json
{
  "approval_status": "approved",
  "execution_status": "not_started"
}
```

Approval also queues one M6 worker job for the run. When the worker claims that job, the tool-call projection moves through:

```json
{
  "approval_status": "approved",
  "execution_status": "executing"
}
```

Successful execution records `tool_call_succeeded`, stores a redacted `result_summary`, emits `run_completed`, and leaves the projection at:

```json
{
  "approval_status": "approved",
  "execution_status": "succeeded",
  "result_summary": {
    "iso_time": "2026-05-24T10:00:00Z",
    "timezone": "UTC",
    "source": "runtime"
  }
}
```

Validation or executor failures record `tool_call_failed`, store stable redacted `error_code` and `error_message` fields, and emit `run_failed`.

`deny` records one `tool_call_denied` event, emits `run_stopped`, and moves the projection to:

```json
{
  "approval_status": "denied",
  "execution_status": "cancelled"
}
```

Repeated same decisions return HTTP 200 with the current projection and do not duplicate decision events. Conflicting reversals are rejected as HTTP 409 with error code `conflict`.

## Redaction and validation

Only schema-valid summaries may be persisted. `runtime.get_current_time` accepts no fields except optional `timezone`, and the only allowed value is `UTC`. Redaction runs before event/job/message metadata persistence and treats sensitive keys such as `api_key`, `authorization`, `password`, `secret`, `token`, and `credential` as always redacted.
