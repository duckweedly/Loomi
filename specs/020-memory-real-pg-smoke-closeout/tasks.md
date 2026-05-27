# Tasks: M13.5 Memory Real PG Smoke Closeout

**Input**: Design documents from `specs/020-memory-real-pg-smoke-closeout/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required because this closeout exists to prove real PG/API behavior.

**Organization**: Tasks are grouped by user story to keep evidence and docs independently verifiable.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the closeout Spec Kit directory and context pointer.

- [x] T001 Create `specs/020-memory-real-pg-smoke-closeout/spec.md`
- [x] T002 Create `specs/020-memory-real-pg-smoke-closeout/plan.md`
- [x] T003 [P] Create `specs/020-memory-real-pg-smoke-closeout/research.md`
- [x] T004 [P] Create `specs/020-memory-real-pg-smoke-closeout/data-model.md`
- [x] T005 [P] Create `specs/020-memory-real-pg-smoke-closeout/contracts/real-pg-httpapi-smoke.md`
- [x] T006 [P] Create `specs/020-memory-real-pg-smoke-closeout/quickstart.md`
- [x] T007 [P] Create `specs/020-memory-real-pg-smoke-closeout/checklists/requirements.md`
- [x] T008 Update `.specify/feature.json` and `AGENTS.md` to point at `specs/020-memory-real-pg-smoke-closeout/plan.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Confirm current implementation and stale wording before adding evidence.

- [x] T009 Review existing M13 memory service/API/repository tests in `internal/productdata/` and `internal/httpapi/`
- [x] T010 Review `specs/019-memory-foundation/` status and contracts for stale Draft/planned language
- [x] T011 Review docs-site memory API/architecture/status pages for stale current-vs-deferred wording

**Checkpoint**: Closeout target is evidence/status cleanup only.

---

## Phase 3: User Story 1 - Prove PG-backed memory lifecycle (Priority: P1)

**Goal**: Add and run a real Postgres + HTTP API smoke for the M13 memory lifecycle.

**Independent Test**: `LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/httpapi -run TestM13MemoryRealPGHTTPAPISmoke -count=1 -v`

### Tests for User Story 1

- [x] T012 [US1] Add real PG/httpapi smoke test in `internal/httpapi/memory_real_pg_smoke_test.go`
- [x] T013 [US1] Ensure smoke verifies migrated `memory_entries` and `memory_write_proposals` tables in `internal/httpapi/memory_real_pg_smoke_test.go`
- [x] T014 [US1] Ensure smoke covers propose, approve, list/search, RunContext snapshot, delete/tombstone exclusion, duplicate decisions/deletes, out-of-scope non-leakage, and sensitive redaction in `internal/httpapi/memory_real_pg_smoke_test.go`

**Checkpoint**: Real PG/httpapi smoke passes when the test database is migrated through M13.

---

## Phase 4: User Story 2 - Close M13 status and documentation evidence (Priority: P2)

**Goal**: Mark implemented behavior as current and record M13.5 evidence without expanding feature scope.

**Independent Test**: Docs build plus manual review of M13 status/current/deferred wording.

- [x] T015 [US2] Update `specs/019-memory-foundation/spec.md` status to Implemented
- [x] T016 [US2] Update implemented/current wording in `specs/019-memory-foundation/contracts/memory-api.md`
- [x] T017 [US2] Update implemented/current wording in `specs/019-memory-foundation/contracts/memory-events.md`
- [x] T018 [US2] Update implemented/current wording in `specs/019-memory-foundation/contracts/memory-provider.md` while keeping distill/OpenViking deferred
- [x] T019 [US2] Add M13.5 devlog evidence in `docs-site/src/content/docs/devlog/2026-05-25-m13-5-memory-real-pg-smoke-closeout.md`
- [x] T020 [US2] Add local memory smoke runbook in `docs-site/src/content/docs/runbooks/local-m13-memory.md`
- [x] T021 [US2] Update memory API/architecture docs in `docs-site/src/content/docs/api/memory-foundation.md` and `docs-site/src/content/docs/architecture/memory-foundation.md`
- [x] T022 [US2] Update roadmap and Spec Kit workflow status in `docs-site/src/content/docs/roadmap/current-status.md` and `docs-site/src/content/docs/spec-kit/workflow.md`

---

## Phase 5: Validation & Closeout

**Purpose**: Run required commands and browser smoke when possible.

- [x] T023 Run real PG migration and `TestM13MemoryRealPGHTTPAPISmoke`
- [x] T024 Run `go test ./...`
- [x] T025 Run `bun test --cwd web`
- [x] T026 Run `bun run --cwd web build`
- [x] T027 Run `bun run --cwd docs-site build`
- [x] T028 Run `git diff --check`
- [x] T029 Run or document Settings > Memory browser smoke evidence in docs-site devlog/runbook

## Dependencies & Execution Order

- Setup precedes all work.
- Foundational review precedes code/doc edits.
- US1 and US2 can proceed in either order after foundational review.
- Final validation depends on all code and docs changes.

## Implementation Strategy

1. Create closeout Spec Kit artifacts.
2. Add the smallest real PG/httpapi smoke test.
3. Update M13 status and docs evidence.
4. Run real migration/smoke, full validation, and browser smoke or document blocker.
