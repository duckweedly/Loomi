# Tasks: M8 Worker Job Closeout

**Input**: Design documents from `specs/013-m8-worker-job-closeout/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Included because this closeout patches retry/backoff behavior and must preserve worker ownership/recovery contracts.

**Organization**: Tasks are grouped by user story so each story can be audited, patched, and documented independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the closeout feature artifacts and point Spec Kit at this feature.

- [X] T001 Create closeout spec artifacts in `specs/013-m8-worker-job-closeout/`
- [X] T002 Update active Spec Kit feature pointer in `.specify/feature.json`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Audit existing M8 coverage before any implementation work.

- [X] T003 [P] [US1] Record M8 audit matrix in `specs/013-m8-worker-job-closeout/contracts/m8-audit.md`
- [X] T004 [US1] Verify existing M6 evidence across `specs/008-worker-job-pipeline/`, `internal/productdata/`, `internal/runtime/`, `internal/httpapi/`, and `docs-site/src/content/docs/`

**Checkpoint**: Existing coverage is known before patching.

---

## Phase 3: User Story 1 - Audit M8 Coverage (Priority: P1)

**Goal**: Decide which original M8 items are covered, patched, or out of scope.

**Independent Test**: Read the audit matrix and confirm all 10 original M8 acceptance targets have a result.

- [X] T005 [US1] Map each original M8 item to current evidence in `specs/013-m8-worker-job-closeout/contracts/m8-audit.md`
- [X] T006 [US1] Identify retry/backoff as the only behavior-level gap in `specs/013-m8-worker-job-closeout/research.md`

**Checkpoint**: Audit complete; no duplicate queue implementation needed.

---

## Phase 4: User Story 2 - Close the Smallest Gap (Priority: P2)

**Goal**: Add actual retry backoff during stale lease recovery.

**Independent Test**: Recover an expired leased job, verify it is scheduled in the future, then claim after that scheduled time.

### Tests for User Story 2

- [X] T007 [P] [US2] Add retry/backoff assertions in `internal/productdata/service_test.go`
- [X] T008 [P] [US2] Add contract retry/backoff assertions in `internal/productdata/repository_test.go`

### Implementation for User Story 2

- [X] T009 [US2] Add bounded retry backoff scheduling to in-memory recovery in `internal/productdata/service.go`
- [X] T010 [US2] Add bounded retry backoff scheduling to PostgreSQL recovery in `internal/productdata/repository.go`

**Checkpoint**: Retry/backoff behavior exists without changing worker topology.

---

## Phase 5: User Story 3 - Record M8 Closeout (Priority: P3)

**Goal**: Document that original M8 is closed by M6 plus the closeout patch.

**Independent Test**: Read roadmap and devlog for closeout result, evidence, validation, and non-goals.

- [X] T011 [US3] Update M8 status in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T012 [US3] Add closeout devlog in `docs-site/src/content/docs/devlog/2026-05-25-m8-worker-job-closeout.md`

**Checkpoint**: Documentation reflects closeout status.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Run validation and record any blockers.

- [X] T013 Run targeted validation `go test ./internal/productdata ./internal/runtime`
- [X] T014 Run requested backend validation `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...`
- [X] T015 Run docs validation `bun run --cwd docs-site build`

---

## Dependencies & Execution Order

Setup -> Audit -> Retry/backoff patch -> Documentation -> Validation.

## Parallel Execution Examples

```text
Task: T007 Add retry/backoff assertions in internal/productdata/service_test.go
Task: T008 Add contract retry/backoff assertions in internal/productdata/repository_test.go
```

## Implementation Strategy

1. Finish the audit first.
2. Patch only retry/backoff scheduling.
3. Update closeout docs.
4. Validate with targeted Go packages, requested Go packages, and docs build.
