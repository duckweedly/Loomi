---
title: M46 Memory Proposal Edit
description: Pending memory proposals can now be edited before approval.
---

M46 closes the review gap after M45 proposal visibility.

Implemented:

- Added productdata and Postgres pending proposal update support.
- Added `PATCH /v1/memory/write-proposals/{proposal_id}`.
- Proposal edits accept only title and summary and keep raw content/idempotency keys hidden from the frontend.
- Approval now persists the edited title and summary after the user saves changes.
- Settings > Memory pending cards expose inline edit/save/cancel controls before Save or Decline.

Validation:

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestUpdateMemoryWriteProposal|TestMemoryHandlersUpdateWriteProposal' -count=1
bun test --cwd web src/components/MemoryPanel.test.tsx src/memory.test.ts
```

Deferred: bulk edit, editing approved entries, LLM distillation, embeddings/vector retrieval, and external semantic provider execution.
