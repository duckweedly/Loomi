# Quickstart: M13 Memory Foundation

Status: design-only. These commands and smokes are for the later implementation session.

## Design validation for this session

```bash
bun run --cwd docs-site build
git diff --check
```

## Implementation validation targets

Backend:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi
```

Frontend, if the memory UI/event replay slice is implemented:

```bash
bun test --cwd web
bun run --cwd web build
```

Docs:

```bash
bun run --cwd docs-site build
```

## Manual smoke expectations

### US1 RunContext snapshot

1. Seed approved, tombstoned, pending, denied, unsafe, and out-of-scope memory rows.
2. Start a run for the visible user/thread scope.
3. Confirm RunContext/debug includes only approved, non-deleted, scoped, redacted entries.
4. Confirm empty/unavailable states produce safe events and the run does not leak PG/query internals.

### US2 Approval-gated writes

1. Trigger an agent memory write proposal fixture.
2. Confirm the proposal is pending and not searchable.
3. Approve it and confirm exactly one approved Memory Entry appears.
4. Deny a second proposal and confirm it never appears in search/RunContext.
5. Retry approval/denial with the same idempotency key and confirm no duplicates.

### US3 User control

1. Open the minimal memory management UI or call list/search APIs.
2. Confirm only safe summaries and metadata are visible.
3. Delete one memory.
4. Confirm it disappears from list/search/RunContext immediately.
5. Confirm a safe tombstone/audit event remains.

## Non-goal checks

Implementation review should confirm there are no first-slice tasks or code for:

- vector DB, embeddings, or RAG orchestration;
- OpenViking provider;
- marketplace/plugin installation;
- sandbox/browser/activity recorder;
- multi-agent long-term memory automation;
- worker/job queue rewrite;
- MCP rewrite;
- automated memory distillation.
