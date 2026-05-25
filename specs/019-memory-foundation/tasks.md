# Tasks: M13 Memory Foundation

**Input**: Design documents from `specs/019-memory-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Required by this feature because privacy, deletion, audit, approval, and RunContext behavior are safety-critical.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or only adds tests/docs
- **[Story]**: User story label for story phases only
- Every completed task records the actual file path used in this implementation slice. Some tests are consolidated into contract-style files instead of one file per planned topic.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm existing boundaries and create fixture coverage without changing behavior yet.

- [x] T001 Read `specs/019-memory-foundation/spec.md`, `specs/019-memory-foundation/plan.md`, `specs/019-memory-foundation/data-model.md`, and `specs/019-memory-foundation/contracts/` before editing code
- [x] T002 [P] Add memory fixture helpers for approved/pending/denied/tombstoned/out-of-scope entries in `internal/productdata/memory_service_test.go`
- [x] T003 [P] Add RunContext memory snapshot fixture setup in `internal/productdata/memory_service_test.go`
- [x] T004 [P] Add API contract fixture cases for memory list/search/delete/write approval in `internal/httpapi/memory_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared schema, models, redaction, audit, and provider boundary required before user stories.

**CRITICAL**: No user story work should begin until memory schema/status/redaction contracts are represented in tests.

- [x] T005 [P] Add migration coverage for `memory_entries` and memory write proposal persistence through `migrations/000009_m13_memory_foundation.up.sql` and backend compile/validation
- [x] T006 [P] Add redaction coverage for memory content/proposals/search/snapshots in `internal/productdata/memory_service_test.go` and `web/src/memory.test.ts`
- [x] T007 [P] Add authorization/scope coverage for memory reads and deletes in `internal/productdata/memory_service_test.go` and `internal/httpapi/memory_test.go`
- [x] T008 Define Memory Entry, Write Proposal, Search Result, Snapshot, Tombstone, and Audit models in `internal/productdata/models.go`
- [x] T009 Add PG migration for `memory_entries` and write proposal state in `migrations/000009_m13_memory_foundation.up.sql` and `migrations/000009_m13_memory_foundation.down.sql`
- [x] T010 Implement memory redaction helpers in `internal/runtime/memory_redaction.go`
- [x] T011 Define `MemoryProvider` interface and PG adapter boundary in `internal/runtime/memory.go`
- [x] T012 Wire safe memory audit event constants/helpers in `internal/productdata/models.go`

**Checkpoint**: Memory storage/status/redaction/provider contracts are test-covered and ready for user-story implementation.

---

## Phase 3: User Story 1 - Retrieve safe historical memory in RunContext (Priority: P1) MVP

**Goal**: A run receives a bounded safe memory snapshot from approved PG memory.

**Independent Test**: Seed scoped approved and excluded memories, prepare RunContext, and verify only safe entries appear with safe events/debug metadata.

### Tests for User Story 1

- [x] T013 [P] [US1] Add search tests for approved scoped entries and exclusion of pending/denied/tombstoned/unsafe/out-of-scope entries in `internal/productdata/memory_service_test.go`
- [x] T014 [P] [US1] Add RunContext snapshot tests for safe bounded entries, provenance, and summary metadata in `internal/productdata/memory_service_test.go`
- [x] T015 [P] [US1] Add event/audit tests for `memory_snapshot_loaded` safe metadata in `internal/productdata/memory_service_test.go`

### Implementation for User Story 1

- [x] T016 [US1] Implement scoped PG memory search in `internal/productdata/repository.go`
- [x] T017 [US1] Implement PG `MemoryProvider.SearchMemory` and `BuildSnapshot` in `internal/runtime/memory.go`
- [x] T018 [US1] Attach memory snapshot to RunContext preparation in `internal/productdata/service.go` and `internal/productdata/repository.go`
- [x] T019 [US1] Emit safe memory snapshot events/debug summaries through existing run event boundaries in `internal/productdata/service.go` and `internal/productdata/repository.go`

**Checkpoint**: US1 delivers Memory MVP: approved PG memories can influence a run only through safe RunContext snapshot.

---

## Phase 4: User Story 2 - Approval-gate agent memory writes (Priority: P2)

**Goal**: Agent-proposed memory writes remain pending until user approval and become searchable only after approval.

**Independent Test**: Propose, approve, deny, and retry memory writes; verify searchable state and audit events.

### Tests for User Story 2

- [x] T020 [P] [US2] Add repository/service tests for pending write proposal creation and duplicate idempotency in `internal/productdata/memory_service_test.go`
- [x] T021 [P] [US2] Add approve/deny tests proving only approved proposals create searchable entries in `internal/productdata/memory_service_test.go`
- [x] T022 [P] [US2] Add API tests for propose/approve/deny scoped authorization and idempotency in `internal/httpapi/memory_test.go`
- [x] T023 [P] [US2] Add event/audit tests for proposed/approved/denied states with safe metadata in `internal/productdata/memory_service_test.go`

### Implementation for User Story 2

- [x] T024 [US2] Implement memory write proposal persistence and idempotency in `internal/productdata/service.go`
- [x] T025 [US2] Implement approve/deny transitions that create or reject approved Memory Entries in `internal/productdata/service.go`
- [x] T026 [US2] Implement PG provider write proposal and decision adapter methods in `internal/runtime/memory.go`
- [x] T027 [US2] Add minimal propose/approve/deny HTTP handlers in `internal/httpapi/memory.go`
- [x] T028 [US2] Emit safe write proposal/decision events in `internal/productdata/service.go` and `internal/productdata/repository.go`

**Checkpoint**: US2 makes agent memory writes possible only through explicit user approval.

---

## Phase 5: User Story 3 - User can view, search, and delete memory (Priority: P3)

**Goal**: The user can inspect safe memory metadata, search entries, and delete memories with immediate effect.

**Independent Test**: List/search visible entries, delete one, and verify it disappears from list/search/RunContext while retaining safe tombstone/audit metadata.

### Tests for User Story 3

- [x] T029 [P] [US3] Add API tests for list/search/read/delete authorization, safe search/delete audit metadata, no existence leak, and tombstone idempotency in `internal/httpapi/memory_test.go`
- [x] T030 [P] [US3] Add tests proving tombstoned entries are excluded immediately from search and snapshots and delete audit state is safe/idempotent in `internal/productdata/memory_service_test.go`
- [x] T031 [P] [US3] Add frontend API client tests for memory list/search/delete mapping in `web/src/memory.test.ts`
- [x] T032 [P] [US3] Add minimal MemoryPanel component tests for safe summary display and delete flow in `web/src/components/MemoryPanel.test.tsx`

### Implementation for User Story 3

- [x] T033 [US3] Implement list/search/read/delete HTTP handlers in `internal/httpapi/memory.go`
- [x] T034 [US3] Implement tombstone delete, safe search/delete audit metadata, and no-existence-leak errors in `internal/productdata/service.go`
- [x] T035 [US3] Add memory API client methods in `web/src/realApiClient.ts`
- [x] T036 [US3] Add minimal memory management UI in `web/src/components/MemoryPanel.tsx`
- [x] T037 [US3] Wire memory management entry point into the existing settings or shell surface in `web/src/App.tsx`

**Checkpoint**: US3 gives users visible control over approved memory.

---

## Phase 6: User Story 4 - Plan future distillation/provider boundaries without implementing them (Priority: P4)

**Goal**: Preserve future provider/distillation language in docs and contracts without adding first-slice implementation.

**Independent Test**: Review tasks and code diff to confirm PG provider is the only implementation and distillation/OpenViking remain planned-only.

### Implementation for User Story 4

- [x] T038 [P] [US4] Add planned-only architecture notes for MemoryProvider and distillation boundaries in `docs-site/src/content/docs/architecture/memory-foundation.md`
- [x] T039 [P] [US4] Add planned-only API/event notes in `docs-site/src/content/docs/api/memory-foundation.md`
- [x] T040 [US4] Confirm no implementation files introduce OpenViking, embeddings, vector DB, RAG orchestration, or automated distillation in the first slice

---

## Phase 7: Documentation and Validation

**Purpose**: Update docs-site and run required validation before reporting implementation completion.

- [x] T041 [P] Update Spec Kit workflow links/status in `docs-site/src/content/docs/spec-kit/workflow.md`
- [x] T042 [P] Update roadmap planned/completed boundaries in `docs-site/src/content/docs/roadmap/current-status.md`
- [x] T043 [P] Add implementation devlog only after code is implemented in `docs-site/src/content/docs/devlog/`
- [x] T044 Run backend validation command `go test ./...`
- [x] T045 Run frontend validation command `bun test --cwd web`
- [x] T046 Run frontend build command `bun run --cwd web build`
- [x] T047 Run docs validation command `bun run --cwd docs-site build`
- [x] T048 Run diff hygiene validation command `git diff --check`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **US1 (Phase 3)**: Depends on Foundational and is the MVP.
- **US2 (Phase 4)**: Depends on Foundational and may run after or beside US1 once shared provider/schema exists, but its approved entries are most valuable after US1 search/snapshot.
- **US3 (Phase 5)**: Depends on Foundational and should validate against US1 search/snapshot exclusion behavior.
- **US4 (Phase 6)**: Can run in parallel with docs work; must remain planned-only.
- **Documentation/Validation (Phase 7)**: Can begin during implementation but final validation depends on desired story scope.

### User Story Dependencies

- **User Story 1 (P1)**: Independent MVP after Foundational.
- **User Story 2 (P2)**: Independent write-safety slice after Foundational; full value combines with US1.
- **User Story 3 (P3)**: Independent user-control slice after Foundational; deletion must be validated against US1.
- **User Story 4 (P4)**: Documentation/planning only; no code dependency.

### Parallel Opportunities

- T002-T004 can run in parallel.
- T005-T007 can run in parallel.
- US1 tests T013-T015 can run in parallel.
- US2 tests T020-T023 can run in parallel.
- US3 tests T029-T032 can run in parallel.
- Documentation tasks T038-T042 can run in parallel once names/contracts are stable.

## Parallel Example: User Story 1

```text
Task: "T013 Add search tests for approved scoped entries and exclusion states in internal/productdata/memory_service_test.go"
Task: "T014 Add RunContext snapshot tests in internal/productdata/memory_service_test.go"
Task: "T015 Add event/audit tests for memory snapshot events in internal/productdata/memory_service_test.go"
```

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete US1 tests and implementation.
3. Validate that approved PG memories appear only through safe RunContext snapshots.
4. Stop before agent writes/UI if a minimal read-only memory MVP is desired.

### Full M13 First Slice

1. Complete US1 safe snapshot.
2. Complete US2 approval-gated writes.
3. Complete US3 user list/search/delete control.
4. Keep US4 provider/distill work as docs/design-only.
5. Run all validation commands in Phase 7.
