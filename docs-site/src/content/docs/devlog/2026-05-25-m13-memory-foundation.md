---
title: 2026-05-25 M13 Memory Foundation
description: Implementation notes and validation for the first PG-backed memory slice.
---

## Completed

- Added `memory_entries` and `memory_write_proposals` migrations.
- Added product data memory models, in-memory service support, and PostgreSQL repository support.
- Added safe text memory search for approved user/thread-scoped entries.
- Added `RunContext.MemorySnapshot` and `memory_snapshot_loaded` run events.
- Added approval-gated write proposals, approval, denial, idempotency keys, and tombstone delete.
- Added safe memory audit events for source-run-linked proposal/approval/denial/delete operations.
- Added `/v1/memory` list/create/search/read/delete and `/v1/memory/write-proposals` approve/deny APIs.
- Added a runtime `MemoryProvider` abstraction with the current product data provider.
- Added Settings > Memory list/search/delete UI in real API mode.

## Validation

Targeted and full validation for this implementation:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke:

- Opened local web app at `http://127.0.0.1:5181/`.
- Opened Settings, selected Memory, verified the Memory category is available and the empty state renders.

## Boundaries

M13 is not a RAG system. It does not add embeddings, vector DB search, OpenViking provider implementation, marketplace/plugin memory providers, browser/activity recorder ingestion, automatic memory distillation, multi-agent long-term memory automation, worker/job queue rewrite, or MCP rewrite.

Memory distillation remains a later design topic. The current slice is explicit rows, safe snapshots, approval-gated writes, and user deletion control.
