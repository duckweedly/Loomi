---
title: 2026-05-23 M4 Run/Event/SSE Devlog
description: Implementation notes, validation results, limitations, and deferred capabilities for M4.
---

## Completed work

M4 adds a local deterministic execution-observability slice:

- Migration `000003_m4_run_event_sse` creates `runs` and `run_events` with active-run and event-ordering constraints.
- Readiness now requires clean schema version `3` or later.
- Product data service and PostgreSQL repository support start, read current/latest run, append/list events, and cooperative stop.
- The deterministic local runner persists lifecycle/progress/message/final events and publishes only after persistence.
- HTTP endpoints implement start run, read run, read event history, history-first SSE stream, and stop run.
- Frontend real API mapping understands M4 run status, source, event categories, stale stream guards, event dedupe, and stopped/local simulated display language.

## Deferred capabilities

M4 intentionally does not add LLM calls, tool execution, worker queues, desktop runtime permissions, production auth, attachments, RAG, plugins, hosted multi-user behavior, or model/tool/worker-specific event types. All execution is labeled `local_simulated`.

## Safety notes

Event payload text is treated as data. Summary/content redaction catches obvious database URL, token, and secret-bearing strings before local persistence/response. Structured errors expose stable codes without leaking raw internal errors.

## Validation results

- `go test ./...` — PASS after review fixes for SSE ordering, stop publishing, terminal guards, and redaction.
- `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts` — PASS after installing `web` dependencies in the worktree and wiring real SSE subscription.
- `bun run --cwd web build` — PASS.
- M4 API/SSE smoke from quickstart — PASS on `127.0.0.1:18080` after repairing the existing local database migration state from version `2 dirty` to clean version `2` with `migrate force 2`, then applying M4 with `migrate up 1` to version `3`. Verified `/readyz`, `POST /v1/threads/{thread_id}/runs`, ordered persisted event history, SSE reconnect with `after_sequence=1` and `stream_closed`, active stop returning `stopped`, and terminal stop returning `already_terminal`.
- Browser smoke for real API mode — PASS with `VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5177`. Playwright loaded the page, observed real API network calls to `/v1/threads`, `/runs/current`, and `/events`, no console warnings/errors, and page text containing `Real API`, `Local simulated`, `closed`, and stopped timeline events.
- `bun run --cwd docs-site build` — PASS after adding the sidebar-referenced docs index/workflow/roadmap/ADR pages that were missing from the content collection.

## Known limitations

The SSE endpoint is in-process and tied to one local API process. If the process restarts, clients recover from persisted event history by reconnecting with `after_sequence`; live fan-out memory is not durable. Stop remains cooperative at deterministic runner step boundaries, not hard interruption.
