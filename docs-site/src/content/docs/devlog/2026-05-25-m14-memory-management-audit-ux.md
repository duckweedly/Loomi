---
title: M14 Memory Management Audit UX
description: Spec Kit and API/UX contract prep for Settings > Memory management and scoped memory audit history.
---

M14 is scoped as a Memory management and audit UX slice on top of M13. It is not a distillation, embedding/RAG, OpenViking, activity recorder, MCP, worker queue, sandbox, or multi-agent rewrite milestone.

This entry records the M14 Spec Kit, UX/API contract, blocker foundation, US1 Settings > Memory management UX, US2 safe audit history, and seeded browser smoke. M14 can now be treated as a full UX complete candidate.

## Completed In This Candidate

- Created `specs/021-memory-management-audit-ux/` with spec, plan, research, data model, API contract, quickstart, tasks, and requirements checklist.
- Defined Settings > Memory done standard: list/search/filter, detail drawer/modal, delete confirmation, loading/empty/error/tombstoned states, scoped audit/history, safe metadata only, and seeded browser smoke.
- Captured blocker fixes as part of the M14 contract: thread-scoped memory read/delete authorization, terminal-run audit retention, expanded local/provider-output redaction, and unified list/search filter shape.
- Added low-risk MemoryPanel error rendering so failed memory loads do not silently masquerade as successful stale data.
- Implemented the blocker foundation: `memory_audit_events` migration, generic not found for out-of-scope thread memory detail/delete, explicit invalid request for thread list/search without `scope_id`, terminal-run durable audit, `source_thread_id` filter shape, and redaction for `/home`, Windows paths, stdout/stderr, tool output, provider traces, and key/env markers.

## US1 Management UX

- Settings > Memory now renders real approved/tombstoned memory entries through the existing API client state instead of fake rows.
- Added grounded search/filter controls for `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, `include_tombstoned`, and `limit`.
- Added a safe detail panel that shows title, summary, scope/source metadata, status, timestamps, and redaction markers only.
- Added delete confirmation so the delete API is not called from a single accidental click.
- Thread-scoped detail/delete calls reuse the entry scope/source context; missing context becomes a safe UI error instead of leaking the target id.
- List/detail/delete errors clear or avoid stale success-looking data, so API failures are visible.

## US2 Audit History UX

- Settings > Memory now loads real history through `GET /v1/memory/audit`.
- The history renders only safe metadata: event type, memory/proposal/thread/run ids, scope/source/status, redaction marker, and time.
- Covered audit events are `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, `memory_deleted`, and `memory_snapshot_loaded`.
- Backend unavailable, empty, and error states do not fabricate history rows; the UI shows the actual unavailable/empty/error state.
- Fixed the frontend delete request body so UI-only list fields such as `limit` are not sent to the backend delete endpoint.

## Contract Notes

The intended API shape is:

- `GET /v1/memory` and `POST /v1/memory/search` with shared implemented filters: `query/q`, `limit`, `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, and `include_tombstoned`.
- `GET /v1/memory/entries/{entry_id}` for safe detail with scope/source context for thread-scoped entries.
- `DELETE /v1/memory/entries/{entry_id}` after UI confirmation, with the same scope/source context.
- `GET /v1/memory/audit` for real safe history backed by durable `memory_audit_events`.

Audit items expose event type, ids, scope/source/status, redaction marker, and time only. Forbidden data includes raw content, secrets, provider traces, tool output, local paths, env values, Authorization headers, and credentials.

## Validation

Required validation for this candidate:

- `go test ./...` when backend blocker fixes are present.
- `bun test --cwd web` and `bun run --cwd web build` when frontend error-state/client contract prep is present.
- `bun run --cwd docs-site build`.
- `git diff --check`.

## Seeded Browser Smoke

Date: 2026-05-25.

Ports and commands:

- API: `APP_ENV=local HTTP_ADDR=127.0.0.1:18080 DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable go run ./cmd/loomi-api`
- Web: `VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080 bun run --cwd web dev --host 127.0.0.1 --port 5173`
- Database: local Postgres on `127.0.0.1:55433`; migration advanced from version `9` to clean version `10`.

Seed:

- Thread: `thr_1779701971994970000_b2ab6b56d4c5`
- Run: `run_1779701972024596000_bad6f6217636`
- Kept approved memory: `mem_1779701972093540000_89e109af7e29`
- Deleted smoke memory after confirmation: `mem_1779702112854813000_f094a9187f91`

Result:

- Settings > Memory opened in real API mode.
- Search `m14-smoke-173931`, `scope_type=thread`, `scope_id=thr_1779701971994970000_b2ab6b56d4c5`, and `source_run_id=run_1779701972024596000_bad6f6217636` returned the seeded approved memories.
- Detail panel opened for a seeded memory and showed safe summary, scope/source/status/redaction/timestamps only.
- Delete confirmation appeared before the delete request; confirming tombstoned the selected memory and refreshed the list.
- Audit history showed real `memory_deleted`, `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, and `memory_snapshot_loaded` events.
- Browser console errors: none after the delete request body fix.

M14 full UX done gate is satisfied for this candidate.

## M14.1 Review Fix

This review fix keeps M14 scoped to memory management/audit polish and does not start M15.

- `/v1/memory/audit` now accepts the same thread/source filter shape as list/search for Settings > Memory: `scope_type=thread&scope_id`, `source_thread_id`, `source_run_id`, and `limit`. Thread filters return only matching thread history.
- Settings memory list, audit history, and detail loading use latest-request guards so older responses from quick filter/search/detail changes cannot overwrite newer UI state.
- Delete confirmation now shows the selected title plus scope/source metadata. Delete failure closes the confirmation and keeps the existing list instead of replacing it with an empty-looking error state.

Validation for this fix should include `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.
