# Feature Specification: M13.5 Memory Real PG Smoke Closeout

**Feature Branch**: `[020-memory-real-pg-smoke-closeout]`

**Created**: 2026-05-25

**Status**: Implemented

**Input**: User description: "Create and complete M13.5 / 020-memory-real-pg-smoke-closeout as a closeout/evidence slice for the completed M13 Memory Foundation. Add real Postgres migration and HTTP/API smoke evidence for memory_entries and memory_write_proposals, close stale Draft/planned language, update docs-site, and avoid new memory platform features."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Prove PG-backed memory lifecycle (Priority: P1)

As a Loomi maintainer, I want a real Postgres and HTTP API smoke for the M13 memory lifecycle, so that M13 can be closed with evidence from the production repository path instead of in-memory-only tests.

**Why this priority**: M13 added safety-critical memory persistence and user-control behavior. The closeout is not complete until the migration-backed path proves the same behavior.

**Independent Test**: Apply migrations through M13 to a local Postgres database, run the real PG/httpapi smoke test, and verify proposal, approval, list/search, RunContext, delete, idempotency, authorization, and redaction behavior.

**Acceptance Scenarios**:

1. **Given** migrations through `000009_m13_memory_foundation` are applied, **When** a memory write is proposed through HTTP and approved, **Then** exactly one approved memory entry is visible through list/search and eligible for RunContext.
2. **Given** the approved memory is deleted through HTTP, **When** list/search and RunContext are reloaded, **Then** the tombstoned entry is immediately excluded.
3. **Given** approval, denial, and deletion requests are retried, **When** the same real Postgres rows are inspected through API and run events, **Then** entries and audit events are not duplicated.
4. **Given** out-of-scope or sensitive memory content exists, **When** API responses, RunContext safe summaries, and memory run events are inspected, **Then** they do not leak existence or sensitive raw content.

---

### User Story 2 - Close M13 status and documentation evidence (Priority: P2)

As a Loomi developer, I want M13's Spec Kit and docs-site pages to distinguish implemented behavior from deferred future work, so that future agents do not treat completed memory foundations as Draft or confuse deferred distill/OpenViking work with current scope.

**Why this priority**: Incorrect status language causes duplicate implementation attempts and scope creep.

**Independent Test**: Review `specs/019-memory-foundation/`, `docs-site/src/content/docs/`, and this closeout directory to confirm implemented/current language matches shipped behavior and deferred language remains only for future distill/OpenViking/RAG-style work.

**Acceptance Scenarios**:

1. **Given** M13 implementation is complete, **When** `specs/019-memory-foundation/spec.md` is reviewed, **Then** its status is `Implemented`.
2. **Given** contracts are reviewed, **When** current API/event/provider behavior is described, **Then** implemented PG/API/RunContext behavior is labeled current and future distill/OpenViking remains deferred.
3. **Given** docs-site is built, **When** roadmap, runbook, API, architecture, devlog, and Spec Kit workflow pages are reviewed, **Then** they include M13.5 evidence and avoid adding new feature scope.

### Edge Cases

- The local API or web dev server cannot start because a port is already occupied.
- `LOOMI_TEST_DATABASE_URL` is unset, so the real PG smoke is skipped during normal `go test ./...`.
- The Postgres database has not been migrated through M13.
- Sensitive content contains token, secret, authorization, private path, or connection-string markers.
- Retried approve, deny, or delete requests happen after the original state transition already succeeded.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The closeout MUST add a `specs/020-memory-real-pg-smoke-closeout/` Spec Kit directory and identify it as a closeout/evidence slice, not a new feature expansion.
- **FR-002**: The closeout MUST validate the real Postgres migration path creates and uses `memory_entries` and `memory_write_proposals`.
- **FR-003**: The closeout MUST include a real PG/httpapi smoke test file that exists in the repository.
- **FR-004**: The smoke MUST cover create/propose memory write, approve proposal, list/search visibility, RunContext safe snapshot loading, tombstone delete exclusion, duplicate approve/deny/delete idempotency, out-of-scope non-leakage, and sensitive response/event exclusion.
- **FR-005**: The closeout MUST update M13 status and stale planned/design-only wording for implemented API/event/provider behavior while keeping distill, OpenViking, vector DB, embedding, RAG, activity recorder, sandbox, MCP rewrite, and multi-agent memory automation deferred.
- **FR-006**: The closeout MUST update docs-site devlog, runbook, roadmap/current-status, spec-kit workflow, and any API/architecture pages whose current behavior language is stale.
- **FR-007**: The closeout MUST record browser Settings > Memory smoke evidence when local web/API startup is available, or document the blocker and equivalent backend smoke evidence when it is not.
- **FR-008**: The closeout MUST NOT add vector DB, embedding, RAG, OpenViking, automatic distillation, activity recorder, sandbox, MCP rewrite, or multi-agent automatic memory behavior.

### Key Entities

- **Closeout Evidence Slice**: Spec Kit artifact set and docs/test evidence proving completed M13 behavior.
- **Real PG Smoke**: An integration test using a real Postgres repository and HTTP handlers.
- **M13 Status Update**: Spec and docs language that marks implemented behavior as current and keeps future work deferred.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A named real PG/httpapi smoke test passes when `LOOMI_TEST_DATABASE_URL` points at a database migrated through M13.
- **SC-002**: Required validation commands complete: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, and `git diff --check`.
- **SC-003**: `specs/019-memory-foundation/spec.md` reads `Status: Implemented`.
- **SC-004**: Docs-site includes M13.5 evidence and local memory smoke commands.
- **SC-005**: The resulting diff contains no new implementation of deferred vector/RAG/OpenViking/distill/recorder/sandbox/MCP rewrite/multi-agent memory scope.

## Assumptions

- The existing local Postgres compose service and explicit `migrate` workflow remain the canonical local database path.
- The real PG smoke may be skipped by default unless `LOOMI_TEST_DATABASE_URL` is set; the closeout run must set it and execute the test once.
- RunContext validation can use the product repository directly after HTTP creates the approved memory, because both share the same real Postgres store.
