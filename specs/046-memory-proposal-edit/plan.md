# Implementation Plan: Memory Proposal Edit

## Context

M45 made pending memory proposals visible and actionable, but users can only accept or reject the generated text. M46 adds a narrow review edit loop so approval can save a user-confirmed memory summary.

## Technical Approach

- Add `UpdateMemoryWriteProposal` to productdata service/repository.
- Reuse `memory_write_proposals` columns; no migration.
- Add `PATCH /v1/memory/write-proposals/{proposal_id}`.
- Add frontend API and state method for proposal edits.
- Extend `MemoryPanel` pending proposal cards with inline edit controls.

## Data Model

No migration. Updating a pending proposal changes `title`, `summary`, and `content` to the edited safe summary. Approval then creates the approved memory entry from those fields.

## Validation

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestUpdateMemoryWriteProposal|TestMemoryHandlersUpdateWriteProposal' -count=1
bun test --cwd web src/components/MemoryPanel.test.tsx src/memory.test.ts
bun run --cwd web build
```

## Risks

- Editing writes the safe summary back into `content` so approval persists the user-confirmed text. This intentionally discards the previous hidden raw proposal text.
- Post-run LLM distillation and semantic retrieval remain deferred.
