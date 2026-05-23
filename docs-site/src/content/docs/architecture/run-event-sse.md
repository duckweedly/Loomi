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

## Stream model

SSE streams use one recovery model for first load, refresh, and reconnect:

1. Validate the run belongs to the local identity.
2. Read persisted events with `sequence > after_sequence`.
3. Write those history events in ascending order.
4. Subscribe to live published events for the same run.
5. Close after a terminal event or terminal run snapshot.

The frontend dedupes by event id/sequence and ignores stale stream events if the selected thread or run changes.

## Stop model

Stop is cooperative and best-effort. M4 has no worker interrupt primitive, so the deterministic runner checks run status at step boundaries. Stopping an active run records stop lifecycle/final events and sets `stopped`; stopping a terminal run returns `already_terminal` without rewriting the terminal outcome.

## Safety notes

Event text is data, not instructions. Product data helpers redact obvious secret-bearing text such as database URLs, tokens, and secret strings before persistence/response. Later model/tool/worker event types are intentionally deferred.
