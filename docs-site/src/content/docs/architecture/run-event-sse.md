---
title: M4 Run/Event/SSE Architecture
description: Durable local simulated runs, ordered run events, history-first SSE, and cooperative stop boundaries.
---

M4 is Loomi's first execution-observability slice. It adds durable `runs` and `run_events` to the local Go API so the web shell can show an execution timeline that survives refresh and reconnect.

## Boundary

The M4 execution source is always `local_simulated`. It does not call an LLM, execute tools, start a worker queue, use a desktop runtime, or add production auth. This lets Loomi validate run lifecycle, event persistence, streaming, and stop semantics before later platform complexity exists.

## Data ownership

Runs and events are scoped to the fixed local identity from M3. A run belongs to one thread and one user. A thread may have only one active run (`pending` or `running`) at a time, while different threads can have active runs independently.

## Persistence first

PostgreSQL tables added by migration `000003_m4_run_event_sse` are the source of truth:

- `runs` stores lifecycle state, `source`, title, timestamps, and redacted error fields.
- `run_events` stores ordered timeline rows with unique `(run_id, sequence)`.
- Readiness requires clean schema version `3` or later before run/event endpoints are considered ready.

The simulator appends events through the product data service. Only after an event is persisted does the local runner publish it to the in-process broadcaster. This preserves the recovery rule that the database, not process memory, explains execution.

Postgres event sequence allocation is serialized per run before `max(sequence)+1` is calculated. Normal append paths already hold the run row; lower-level event helpers also take a transaction-scoped advisory lock keyed by run id so parallel tool execution, worker transitions, and memory-audit mirrors cannot race into duplicate `(run_id, sequence)` values. Memory audit mirrors write `memory_audit_events` and the related `run_events` row in one transaction when the source run still exists; failures are returned instead of being silently dropped.

## Stream model

SSE streams use one recovery model for first load, refresh, and reconnect:

1. Validate the run belongs to the local identity.
2. Subscribe to live events for non-terminal runs before reading history, so events committed during the replay window are not lost.
3. Read persisted events with `sequence > after_sequence`.
4. Write those history events in ascending order.
5. Continue live delivery only from the highest sequence already sent, with periodic persisted-event backfill from that cursor so a full in-process live buffer cannot permanently drop later events or terminal close markers.
6. Send periodic SSE comment heartbeats while an active stream is idle.
7. Close after a terminal event or terminal run snapshot.

`after_sequence` is an exclusive cursor. `GET /v1/runs/{run_id}/events?after_sequence=N` and the SSE stream both return only events with `sequence > N`, so clients can set the cursor to the last replayed event without receiving duplicates. The SSE handler treats the database as the authoritative catch-up source: if live delivery skips a sequence or no live event arrives after a burst, the handler re-reads `ListRunEvents` from the last sent sequence and sends any missed rows before closing on terminal state.

The heartbeat is a comment frame, so clients ignore it as data while intermediaries see active bytes. The frontend dedupes by event id and sequence and ignores stale stream events if the selected thread or run changes. If a replayed event and a live event share the same sequence, the replayed event wins and the live duplicate is ignored, preventing duplicate assistant deltas and duplicate tool lifecycle rows. Out-of-order lower-sequence deltas are retained in the timeline ordering but do not append stale text to the visible assistant draft. `model.delta` appends to the selected run assistant draft, `model.final` marks the draft completed without appending provider metadata to chat text, and `model.error` drives the failed draft state while preserving partial output. Once a run reaches `completed`, `failed`, `stopped`, or `cancelled`, later stream events cannot promote it into another terminal state in the visible run model.

Active-run reconciliation also uses the same cursor. When the stream is open or recovering, the web client asks for only events after the highest sequence already held locally and merges those with the current run snapshot instead of replaying the whole run history every poll. If the fetch-based SSE reader reaches EOF without a `stream_closed` event, the client treats it as a recoverable stream error and runs the same cursor-based reconciliation path, so transient network drops do not leave the UI pretending the stream is still live.

Timeline/debug rendering maps every event into one stable group: Run lifecycle, Model stream, Worker/job, or Error. Error-like types, failed statuses, and error severity override explicit group metadata. Token usage and provider metadata stay in event detail rows rather than assistant message text. This grouping is frontend-owned so M4 local simulated events and future M5 model/provider events share the same readability rules.

## Stop model

Stop is cooperative and best-effort. M4 local simulated work checks run status at step boundaries. Worker-owned runs also get an active stop watcher that cancels the runner context after `StopRun` writes the durable stop state, so provider streams and tool execution paths can halt before the next lease heartbeat. Stopping an active run records stop lifecycle/final events and sets `stopped`; stopping a terminal run returns `already_terminal` without rewriting the terminal outcome.

## Safety notes

Event text is data, not instructions. Product data helpers redact obvious secret-bearing text such as database URLs, tokens, and secret strings before persistence/response. Later model/tool/worker event types are intentionally deferred.
