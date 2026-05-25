# Research: M13.5 Memory Real PG Smoke Closeout

## Decision: Add one real PG/httpapi smoke instead of broad new tests

**Rationale**: M13 already has in-memory service, API handler, frontend, and repository tests. The missing evidence is the combined production path: migrated Postgres tables plus HTTP handlers plus RunContext readback.

**Alternatives considered**:

- Add many unit tests: rejected because the gap is integration evidence, not branch coverage.
- Add a new CLI smoke runner: rejected because Go test already fits existing validation and can be gated by `LOOMI_TEST_DATABASE_URL`.

## Decision: Keep migration execution explicit

**Rationale**: Loomi's database workflow intentionally keeps migrations outside API startup. The closeout documents and runs `migrate -path migrations -database "$DATABASE_URL" up` before the real PG smoke.

**Alternatives considered**:

- Auto-migrate inside tests or API startup: rejected because it would blur the explicit migration boundary.

## Decision: Treat RunContext validation as repository readback after HTTP writes

**Rationale**: The public API does not expose raw `PrepareRunContext`, and adding a debug endpoint would be new surface area. The smoke uses HTTP for user-visible memory lifecycle operations, then prepares RunContext through the same real PG repository backing the API.

**Alternatives considered**:

- Add a RunContext HTTP endpoint: rejected as new functionality.
- Use only repository calls: rejected because the closeout must prove the API chain too.

## Decision: Keep future memory directions deferred

**Rationale**: The requested closeout explicitly forbids vector DB, embeddings, RAG, OpenViking, automatic distill, activity recorder, sandbox, MCP rewrite, and multi-agent automatic memory.

**Alternatives considered**:

- Add richer memory management UX or distill design: rejected as M13 follow-up scope, not M13.5 closeout evidence.
