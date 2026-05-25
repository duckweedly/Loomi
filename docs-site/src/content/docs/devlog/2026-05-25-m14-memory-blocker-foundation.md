---
title: M14 Memory Blocker Foundation
description: Scope authorization, durable audit, redaction hardening, and grounded search filter shape before full Memory UX.
---

## Completed

- Hardened memory detail/delete authorization so thread-scoped entries require matching thread/source context. Wrong-thread and unscoped access returns generic `memory_not_found`.
- Added durable `memory_audit_events` as the user-readable audit source. Run timeline events remain an associated view when a source run can accept them.
- Preserved audit after terminal source runs for proposal, approval, denial, and delete operations.
- Expanded redaction for `/home/...`, Windows user paths, stdout/stderr, tool output, provider traces, key/env markers, and existing token/secret/Authorization patterns.
- Unified implemented search/list filters across backend and frontend client: `query`/`q`, `limit`, `scope_type`, `scope_id`, `source_thread_id`, `source_run_id`, `source_type`, and `include_tombstoned`.

## Tests Added

- In-memory service coverage for same-user wrong-thread read/delete denial and terminal-run audit durability.
- HTTP handler coverage for scoped thread detail/delete, `source_thread_id` filter, terminal-run audit, no duplicate deny audit, and unsafe audit response redaction.
- Postgres repository smoke coverage for scoped read/delete and terminal-run durable audit, gated by `LOOMI_TEST_DATABASE_URL`.
- Frontend real API client coverage for grounded snake_case memory filter fields.

## Deferred

Full M14 UX is still next: Settings > Memory filter controls, detail drawer/modal, delete confirmation, audit/history panel, and seeded browser smoke. This foundation does not add vector DB, embeddings/RAG, OpenViking, automatic distillation, activity recorder ingestion, MCP, worker queue, sandbox, or multi-agent rewrites.
