---
title: M14 Memory Management Audit UX
description: Spec Kit and API/UX contract prep for Settings > Memory management and scoped memory audit history.
---

M14 is scoped as a Memory management and audit UX slice on top of M13. It is not a distillation, embedding/RAG, OpenViking, activity recorder, MCP, worker queue, sandbox, or multi-agent rewrite milestone.

## Completed In This Prep

- Created `specs/021-memory-management-audit-ux/` with spec, plan, research, data model, API contract, quickstart, tasks, and requirements checklist.
- Defined Settings > Memory done standard: list/search/filter, detail drawer/modal, delete confirmation, loading/empty/error/tombstoned states, scoped audit/history, safe metadata only, and seeded browser smoke.
- Captured blocker fixes as part of the M14 contract: thread-scoped memory read/delete authorization, terminal-run audit retention, expanded local/provider-output redaction, and unified list/search filter shape.
- Added low-risk MemoryPanel error rendering so failed memory loads do not silently masquerade as successful stale data.
- Implemented the blocker foundation: `memory_audit_events` migration, generic not found for out-of-scope thread memory detail/delete, terminal-run durable audit, `source_thread_id` filter shape, and redaction for `/home`, Windows paths, stdout/stderr, tool output, provider traces, and key/env markers.

## Contract Notes

The intended API shape is:

- `GET /v1/memory` and `POST /v1/memory/search` with shared implemented filters: `query/q`, `limit`, `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, and `include_tombstoned`.
- `GET /v1/memory/entries/{entry_id}` for safe detail with scope/source context for thread-scoped entries.
- `DELETE /v1/memory/entries/{entry_id}` after UI confirmation, with the same scope/source context.
- `GET /v1/memory/audit` for real safe history backed by durable `memory_audit_events`.

Audit items expose event type, ids, scope/source/status, redaction marker, and time only. Forbidden data includes raw content, secrets, provider traces, tool output, local paths, env values, Authorization headers, and credentials.

## Validation

Required validation for this prep slice:

- `go test ./...` when backend blocker fixes are present.
- `bun test --cwd web` and `bun run --cwd web build` when frontend error-state/client contract prep is present.
- `bun run --cwd docs-site build`.
- `git diff --check`.

Full M14 implementation still requires browser seeded entry smoke for list/search/filter/detail/delete confirmation/audit history.
