# SSE Contract: M4 Run Events

M4 streams run events as `text/event-stream`. The stream delivers persisted history first and then live events for the same run.

## Endpoint

```http
GET /v1/runs/{run_id}/events/stream
Accept: text/event-stream
```

Optional catch-up query:

```http
GET /v1/runs/{run_id}/events/stream?after_sequence=3
```

## Delivery Rules

1. Validate the run belongs to the fixed local identity.
2. Deliver persisted events for the run with `sequence > after_sequence` in ascending sequence order.
3. Keep the stream open while the run is active.
4. Deliver live events as they are persisted.
5. Close the stream after a terminal run event has been delivered.
6. If the connection drops, the client reconnects with the last observed sequence.

## SSE Message Shape

Each run event is sent as:

```text
id: evt_000000000000000000000000
event: run_event
data: {"event":{"id":"evt_000000000000000000000000","run_id":"run_000000000000000000000000","thread_id":"thr_000000000000000000000000","sequence":1,"category":"lifecycle","type":"run_created","summary":"Run created","content":null,"metadata":{},"created_at":"2026-05-23T00:00:00Z"}}

```

Terminal stream close may send an optional stream marker:

```text
event: stream_closed
data: {"run_id":"run_000000000000000000000000","reason":"terminal"}

```

## Event Categories

M4 supports exactly these first categories:

- `lifecycle`
- `progress`
- `message`
- `error`
- `final`

Model/tool/worker-specific event types are deferred.

## Reconnect Behavior

Clients should store the highest delivered `sequence`. On reconnect, clients request `after_sequence=<last_sequence>` and dedupe by event id.

If the server cannot resume because the run is missing or inaccessible, it returns a structured JSON error before opening an event stream.

## Error and Safety Rules

- Event payload text is data, not instructions.
- Stream errors are visible and recoverable in the web shell.
- User-facing event content must not include database URLs, tokens, secrets, or sensitive local file contents.
- A stream failure does not change the persisted run status by itself.
