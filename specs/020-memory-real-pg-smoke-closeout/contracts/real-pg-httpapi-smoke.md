# Contract: Real PG HTTP API Memory Smoke

Status: current closeout evidence for `020-memory-real-pg-smoke-closeout`.

## Preconditions

- Local Postgres is reachable.
- `DATABASE_URL` points at the local Postgres database.
- `migrate -path migrations -database "$DATABASE_URL" up` has applied migrations through version `9`.
- `LOOMI_TEST_DATABASE_URL="$DATABASE_URL"` is set for the smoke test process.

## Required Covered Flow

1. `POST /v1/memory/write-proposals`
   - Creates a pending proposal in `memory_write_proposals`.
   - Response omits idempotency key and does not expose sensitive raw content.

2. `POST /v1/memory/write-proposals/{proposal_id}/approve`
   - Creates one approved `memory_entries` row.
   - Repeating approve returns the same entry and does not duplicate approval audit events.

3. `GET /v1/memory` and `POST /v1/memory/search`
   - Return the approved memory only while it is approved and in scope.
   - Exclude pending, denied, blocked/sensitive, tombstoned, and out-of-scope rows.

4. `PrepareRunContext` through the same real PG repository
   - Loads a safe memory snapshot after HTTP approval.
   - Excludes the entry immediately after HTTP deletion.
   - Emits safe `memory_snapshot_loaded` metadata.

5. `DELETE /v1/memory/{entry_id}`
   - Tombstones the entry and clears reusable content.
   - Repeating delete returns tombstoned status and does not duplicate delete audit events.

6. Out-of-scope and sensitive checks
   - Deleting another user's entry returns not found without confirming existence.
   - Sensitive content does not appear in API response bodies, RunContext safe summaries, or run-event metadata.

## Explicit Non-Contracts

- No new HTTP endpoint for RunContext inspection.
- No vector, embedding, RAG, OpenViking, distill, recorder, sandbox, MCP rewrite, or multi-agent memory behavior.
