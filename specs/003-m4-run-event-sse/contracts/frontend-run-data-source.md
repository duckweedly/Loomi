# Frontend Run Data Source Contract: M4 Real Run/Event Switching

M4 extends the existing real/mock data-source boundary with run/event behavior.

## Real API Mode

Trigger:

```text
VITE_LOOMI_API_BASE_URL is set
```

Behavior:

- Thread/message data continues to use the real M3 API.
- Starting a run uses `POST /v1/threads/{thread_id}/runs`.
- Current or latest run can be loaded from `GET /v1/threads/{thread_id}/runs/current`.
- Persisted run events can be loaded from `GET /v1/runs/{run_id}/events`.
- Live run events use `GET /v1/runs/{run_id}/events/stream`.
- Stop uses `POST /v1/runs/{run_id}/stop`.

## Mock Mode

Trigger:

```text
VITE_LOOMI_API_BASE_URL is absent or empty
```

Behavior:

- Existing mock thread/message/run demonstration remains usable.
- Mock run/event data must remain clearly mock-only.
- Mock mode may continue to drive frontend screenshots and local UI tests.

## Stream Loading Rules

When the web shell opens a run stream:

1. It must expect persisted history before live events.
2. It must dedupe by event id or sequence.
3. It must ignore stale stream events after the selected thread or run changes.
4. It must show a recoverable stream state when the stream disconnects unexpectedly.
5. It must reload persisted history after reconnect or refresh.

## UI Mapping Rules

- `lifecycle` events update run status and agent state motion.
- `progress` events add timeline progress rows.
- `message` events can append or display deterministic simulated assistant output.
- `error` events show failure state without leaking secrets.
- `final` events close the run lifecycle and stream state.

## Deferred Frontend Behavior

M4 does not add:

- LLM-generated model deltas.
- Real tool call cards backed by tool execution.
- Worker/job queue ownership UI.
- Desktop runtime permission prompts.
- Attachment, RAG, or catalog/plugin runtime behavior.
