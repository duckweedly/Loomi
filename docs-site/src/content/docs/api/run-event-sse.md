---
title: M4 Run/Event/SSE API
description: Local M4 HTTP and SSE contracts for runs, events, and cooperative stop.
---

All M4 endpoints are local-development endpoints under the fixed local identity. Responses include a `request_id` for JSON endpoints. Run objects always use `source: "local_simulated"`.

## Start run

```http
POST /v1/threads/{thread_id}/runs
Content-Type: application/json
```

Request body is optional:

```json
{ "script_name": "m4_smoke" }
```

Success returns `201`:

```json
{
  "run": {
    "id": "run_...",
    "thread_id": "thr_...",
    "status": "running",
    "source": "local_simulated",
    "title": "Local simulated run",
    "created_at": "2026-05-23T00:00:00Z",
    "updated_at": "2026-05-23T00:00:00Z"
  },
  "request_id": "req_..."
}
```

A second active run for the same thread returns `409` with `active_run_exists`.

## Read current or specific run

```http
GET /v1/threads/{thread_id}/runs/current
GET /v1/runs/{run_id}
```

Both return a single run. Missing or inaccessible runs return `404` with `run_not_found`.

## List persisted events

```http
GET /v1/runs/{run_id}/events
GET /v1/runs/{run_id}/events?after_sequence=3
```

Success returns ordered events:

```json
{
  "events": [
    {
      "id": "evt_...",
      "run_id": "run_...",
      "thread_id": "thr_...",
      "sequence": 1,
      "category": "lifecycle",
      "type": "run_created",
      "summary": "Run created",
      "content": null,
      "metadata": { "script_name": "m4_smoke" },
      "created_at": "2026-05-23T00:00:00Z"
    }
  ],
  "request_id": "req_..."
}
```

Supported categories are `lifecycle`, `progress`, `message`, `error`, and `final`. Frontend runtime mapping also accepts model event types directly: `model.delta` carries assistant draft text in `content`, `model.final` carries final assistant text without creating duplicate visible chat content, and `model.error` maps to a failed assistant draft state. `run.recovering` keeps recovery visible until the latest run state is reconciled, while `run.stopped` and `run.failed` preserve any partial draft text as terminal context.

Frontend timeline grouping recognizes lifecycle events (`run.created`, `run.completed`, `run.stopped`, `run.cancelled`, retry/recovering states), model events (`model.started`, `model.delta`, `model.final`, `model.usage`, assistant draft/message events), worker/job events (`job.queued`, `job.claimed`, `worker.claimed`, `job.retrying`), and error events (`provider.error`, `stream.error`, `backend.unavailable`, `run.failed`). Usage metadata may include `input_tokens`, `output_tokens`, and `total_tokens`; provider metadata such as `provider` and `code` is displayed in timeline/debug details, not in message text. Event types containing `error`, `failed`, `unavailable`, or `timeout`, failed statuses, and error severities are grouped as Error even when another category is present.

## Stream events

```http
GET /v1/runs/{run_id}/events/stream?after_sequence=0
Accept: text/event-stream
```

The stream sends persisted history first, then live events after they are persisted:

```text
id: evt_...
event: run_event
data: {"event":{"id":"evt_...","sequence":1,"category":"lifecycle","type":"run_created"}}

```

Terminal streams may end with:

```text
event: stream_closed
data: {"run_id":"run_...","reason":"terminal"}

```

Clients should reconnect with the highest delivered sequence and dedupe by event id.

## Stop run

```http
POST /v1/runs/{run_id}/stop
```

Success returns:

```json
{
  "run": { "id": "run_...", "status": "stopped", "source": "local_simulated" },
  "result": "stopped",
  "request_id": "req_..."
}
```

Terminal runs return `result: "already_terminal"` and keep their existing status.

## Error codes

- `invalid_request` → malformed JSON or invalid query values.
- `thread_not_found` → missing, archived, or inaccessible thread.
- `run_not_found` → missing or inaccessible run.
- `active_run_exists` → thread already has an active run.
- `method_not_allowed` → unsupported method.
- `internal_error` → unexpected server failure with redacted message.
