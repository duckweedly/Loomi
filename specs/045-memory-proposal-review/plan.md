# Implementation Plan: Memory Proposal Review

## Context

M44 creates pending post-run proposals, but users need a grounded Settings surface to review them. M45 adds the read and decision loop without changing the memory write model.

## Technical Approach

- Add `ListMemoryWriteProposals` to productdata service/repository.
- Add `GET /v1/memory/write-proposals` with safe projections.
- Add frontend `MemoryWriteProposal` domain type and real/mock API methods.
- Extend workspace state to load pending proposals and refresh after approve/deny.
- Render a pending proposal section in `MemoryPanel` with Save and Decline actions.

## Data Model

No migration. Reuses `memory_write_proposals`.

## Validation

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestListMemoryWriteProposals|TestMemoryHandlersListPendingWriteProposals' -count=1
bun test --cwd web src/memory.test.ts src/components/MemoryPanel.test.tsx
bun run --cwd web build
```

## Risks

- The list intentionally returns summaries, not raw content. Proposal editing is deferred.
- Terminal run audit remains the durable history source.
