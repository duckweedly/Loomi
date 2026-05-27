---
title: M45 Memory Proposal Review
description: Pending memory proposals are now visible and actionable in Settings > Memory.
---

M45 closes the user-control loop after M44 post-run proposal creation.

Implemented:

- Added productdata and Postgres safe pending proposal list support.
- Added `GET /v1/memory/write-proposals`, defaulting to pending proposals.
- Proposal list responses omit raw content and idempotency keys.
- Added frontend `MemoryWriteProposal` mapping plus list/approve/deny API methods.
- Settings > Memory now shows pending proposals above saved memories with Save and Decline actions.
- Approval refreshes proposals, saved entries, and audit history; denial refreshes proposals and audit history.

Validation:

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestListMemoryWriteProposals|TestMemoryHandlersListPendingWriteProposals' -count=1
bun test --cwd web src/memory.test.ts src/components/MemoryPanel.test.tsx
bun run --cwd web build
```

Deferred: proposal editing, bulk approve/deny, LLM distillation, embeddings/vector retrieval, and external semantic provider execution.
