# Tasks: M6 Worker Job Pipeline

**Input**: Design documents from `specs/008-worker-job-pipeline/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Included because the implementation plan and quickstart require backend, frontend, smoke, rollback, and docs validation for this worker/job/pipeline slice.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested as an independent increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks in the same phase
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish M6 schema/configuration scaffolding and shared validation hooks.

- [X] T001 Create M6 migration files in `migrations/000005_m6_worker_job_pipeline.up.sql` and `migrations/000005_m6_worker_job_pipeline.down.sql`
- [X] T002 [P] Add worker queue configuration fields and defaults in `internal/config/config.go`
- [X] T003 [P] Add worker queue configuration tests in `internal/config/config_test.go`
- [X] T004 [P] Add M6 backend domain constants for run statuses, job statuses, event types, and pipeline steps in `internal/productdata/models.go`
- [X] T005 [P] Add M6 frontend domain constants for queued, recovering, stopping, worker, and pipeline events in `web/src/domain.ts`
- [X] T006 Update schema readiness target and migration version expectations in `internal/db/readiness.go`
- [X] T007 [P] Add M6 readiness tests for schema version 5 in `internal/db/readiness_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Implement durable job persistence, shared service boundaries, and event/idempotency helpers required by every story.

**Critical**: No user story work should start until this phase is complete.

- [X] T008 Add migration SQL for `background_jobs`, lease fields, retry fields, and run status extensions in `migrations/000005_m6_worker_job_pipeline.up.sql`
- [X] T009 Add rollback SQL for M6 job tables and run status extensions in `migrations/000005_m6_worker_job_pipeline.down.sql`
- [X] T010 [P] Add repository tests for job creation, one-active-job-per-run, and one-active-run-per-thread in `internal/productdata/repository_test.go`
- [X] T011 Implement background job persistence methods in `internal/productdata/repository.go`
- [X] T012 [P] Add service tests for run/job atomic creation and redacted failure handling in `internal/productdata/service_test.go`
- [X] T013 Implement run/job service methods and terminal idempotency helpers in `internal/productdata/service.go`
- [X] T014 [P] Add runtime job coordinator tests for claim result mapping in `internal/runtime/jobs_test.go`
- [X] T015 Create runtime job coordinator for claim, completion, retry, and cancellation calls in `internal/runtime/jobs.go`
- [X] T016 [P] Add pipeline event recorder tests in `internal/runtime/pipeline_test.go`
- [X] T017 Create minimal pipeline step recorder in `internal/runtime/pipeline.go`
- [X] T018 [P] Add stream ordering regression tests for M6 worker events in `internal/runtime/stream_test.go`
- [X] T019 Extend run event stream handling for M6 queue and worker event types in `internal/runtime/stream.go`

**Checkpoint**: Durable job records, shared state transitions, event recording, and schema readiness are available for user story work.

---

## Phase 3: User Story 1 - Run Continues Outside the Request (Priority: P1)

**Goal**: A user starts a run, receives a queued/running acknowledgement quickly, and the run continues to a terminal outcome through background worker processing and persisted timeline events.

**Independent Test**: Start a run from an existing thread, disconnect the browser or event stream, and verify the run continues through persisted events to a single terminal outcome and final assistant message.

### Tests for User Story 1

- [X] T020 [P] [US1] Add HTTP test for queued run acknowledgement and active-run conflict in `internal/httpapi/runtime_test.go`
- [X] T021 [P] [US1] Add worker happy-path test for claim, run execution, finalization, and single terminal event in `internal/runtime/worker_test.go`
- [X] T022 [P] [US1] Add frontend runtime adapter tests for queued, running, and completed event mapping in `web/src/runtime/executionAdapter.test.ts`
- [X] T023 [P] [US1] Add frontend state tests for history replay after queued background events in `web/src/state.runtime.test.ts`

### Implementation for User Story 1

- [X] T024 [US1] Change run start service to create a queued run, background job, and `run_queued` event in `internal/productdata/service.go`
- [X] T025 [US1] Update run start handler to return `202` queued/running responses without synchronous execution in `internal/httpapi/runtime.go`
- [X] T026 [US1] Implement basic worker loop for claiming jobs and invoking run execution in `internal/runtime/worker.go`
- [X] T027 [US1] Route model-backed run execution through the worker boundary in `internal/runtime/runner.go`
- [X] T028 [US1] Record `job_claimed`, `pipeline_step_started`, `pipeline_step_completed`, and `run_completed` events during happy-path processing in `internal/runtime/pipeline.go`
- [X] T029 [US1] Wire worker startup and shutdown into the local API process in `cmd/loomi-api/main.go`
- [X] T030 [US1] Register background-capable runtime dependencies in `internal/httpapi/server.go`
- [X] T031 [US1] Extend real API client run types for queued and recovering statuses in `web/src/realApiClient.ts`
- [X] T032 [US1] Map queued, running, and completed worker events into frontend runtime state in `web/src/runtime/executionAdapter.ts`
- [X] T033 [US1] Render queued and background-running states in the existing chat canvas in `web/src/components/ChatCanvas.tsx`
- [X] T034 [US1] Render M6 queue and worker events in the existing run timeline in `web/src/components/RunTimeline.tsx`

**Checkpoint**: User Story 1 is independently functional as the MVP background execution slice.

---

## Phase 4: User Story 2 - Recover Work After Worker Interruption (Priority: P2)

**Goal**: Interrupted worker-owned jobs become visible, recoverable, and either resume safely or fail with a clear terminal explanation after retry exhaustion.

**Independent Test**: Start a run, interrupt the worker after claim, wait for lease expiry, resume processing, and verify recovery events plus no duplicate terminal events or final assistant messages.

### Tests for User Story 2

- [X] T035 [P] [US2] Add repository tests for lease expiry, stale claim recovery, and retry exhaustion in `internal/productdata/repository_test.go`
- [X] T036 [P] [US2] Add worker recovery tests for stale lease detection and retry scheduling in `internal/runtime/jobs_test.go`
- [X] T037 [P] [US2] Add idempotency tests for duplicate completion and duplicate final message prevention in `internal/productdata/service_test.go`
- [X] T038 [P] [US2] Add frontend runtime tests for recovering and failed worker states in `web/src/runtime/executionAdapter.test.ts`

### Implementation for User Story 2

- [X] T039 [US2] Implement lease acquisition, lease renewal, stale lease detection, and ownership version checks in `internal/productdata/repository.go`
- [X] T040 [US2] Implement recovery scheduling, attempt counting, retry exhaustion, and redacted terminal failure in `internal/productdata/service.go`
- [X] T041 [US2] Add lease renewal and stale-ownership stop behavior to the worker loop in `internal/runtime/worker.go`
- [X] T042 [US2] Add recovery coordinator behavior for expired leases and retryable jobs in `internal/runtime/jobs.go`
- [X] T043 [US2] Record `job_recovering`, `job_retry_scheduled`, `job_attempt_failed`, and `job_retry_exhausted` events in `internal/runtime/pipeline.go`
- [X] T044 [US2] Prevent stale workers from writing conflicting terminal outcomes in `internal/runtime/runner.go`
- [X] T045 [US2] Map recovering and retry-exhausted states into frontend state in `web/src/state.ts`
- [X] T046 [US2] Render recovering and retry-exhausted timeline states in `web/src/components/RunTimeline.tsx`

**Checkpoint**: User Story 2 can be validated independently with worker interruption and recovery smoke.

---

## Phase 5: User Story 3 - Cancel Background Execution (Priority: P3)

**Goal**: Users can request stop for queued or running background runs, and Loomi reaches a consistent stopped or already-terminal outcome without conflicting final output.

**Independent Test**: Request stop while a run is queued and while it is running, then verify normal execution is prevented or cooperatively stopped with no later conflicting final state.

### Tests for User Story 3

- [X] T047 [P] [US3] Add HTTP stop tests for queued, running, and already-terminal background runs in `internal/httpapi/runtime_test.go`
- [X] T048 [P] [US3] Add service tests for persisted stop requests and queued job cancellation in `internal/productdata/service_test.go`
- [X] T049 [P] [US3] Add worker cancellation tests for safe-boundary stop handling in `internal/runtime/worker_test.go`
- [X] T050 [P] [US3] Add frontend state tests for stopping and stopped event handling in `web/src/state.runtime.test.ts`

### Implementation for User Story 3

- [X] T051 [US3] Persist stop requests on active background runs in `internal/productdata/repository.go`
- [X] T052 [US3] Implement queued-job cancellation and already-terminal stop semantics in `internal/productdata/service.go`
- [X] T053 [US3] Update stop handler to return stop-requested or already-terminal responses for background runs in `internal/httpapi/runtime.go`
- [X] T054 [US3] Make worker processing observe stop requests at safe boundaries in `internal/runtime/worker.go`
- [X] T055 [US3] Prevent normal execution from starting for stopped queued jobs in `internal/runtime/jobs.go`
- [X] T056 [US3] Record `stop_requested` and `run_stopped` events for queued and running cancellation paths in `internal/runtime/pipeline.go`
- [X] T057 [US3] Map stopping and stopped worker states into the frontend runtime adapter in `web/src/runtime/executionAdapter.ts`
- [X] T058 [US3] Render stopping and stopped states in the chat canvas and run rail in `web/src/components/ChatCanvas.tsx` and `web/src/components/RunRail.tsx`

**Checkpoint**: User Story 3 can be validated independently through queued and running cancellation smoke.

---

## Phase 6: User Story 4 - Inspect Queue and Worker Health (Priority: P4)

**Goal**: Developers can distinguish ready, paused, unhealthy, degraded, stale, retrying, stopped, completed, and failed background execution states through diagnostics and timeline output.

**Independent Test**: Create queued, running, recovered, failed, stopped, stale, retrying, and dead jobs, then verify diagnostics and run timelines explain each state without exposing secrets.

### Tests for User Story 4

- [X] T059 [P] [US4] Add queue diagnostics repository tests for queued, leased, stale, retrying, dead, and redacted states in `internal/productdata/repository_test.go`
- [X] T060 [P] [US4] Add HTTP diagnostics endpoint tests for ready, paused, unhealthy, degraded, and stopped states in `internal/httpapi/health_test.go`
- [X] T061 [P] [US4] Add real API client tests for worker queue diagnostics parsing in `web/src/realApiClient.test.ts`
- [X] T062 [P] [US4] Add timeline rendering tests for worker diagnostic event summaries in `web/src/components/RunTimeline.test.tsx`

### Implementation for User Story 4

- [X] T063 [US4] Implement queue diagnostic aggregation in `internal/productdata/repository.go`
- [X] T064 [US4] Implement redacted worker queue diagnostic service output in `internal/productdata/service.go`
- [X] T065 [US4] Add worker heartbeat, paused, unhealthy, degraded, and stopped status reporting in `internal/runtime/worker.go`
- [X] T066 [US4] Add worker queue diagnostics handler in `internal/httpapi/diagnostics.go`
- [X] T067 [US4] Register the worker queue diagnostics route in `internal/httpapi/server.go`
- [X] T068 [US4] Add worker queue diagnostics client method in `web/src/realApiClient.ts`
- [X] T069 [US4] Render safe worker and queue event metadata in the run timeline in `web/src/components/RunTimeline.tsx`

**Checkpoint**: User Story 4 can be validated independently through diagnostics and timeline inspection.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, smoke coverage, final validation, and cleanup that spans multiple user stories.

- [X] T070 [P] Add M6 architecture documentation in `docs-site/src/content/docs/architecture/worker-job-pipeline.md`
- [X] T071 [P] Add M6 API and event contract documentation in `docs-site/src/content/docs/api/worker-job-pipeline.md`
- [X] T072 [P] Add local M6 validation and troubleshooting runbook in `docs-site/src/content/docs/runbooks/local-m6.md`
- [X] T073 [P] Update Spec Kit workflow status for M6 in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T074 [P] Update roadmap current status with M6 scope and remaining M7+ boundaries in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T075 [P] Add M6 devlog with validation results and known limitations in `docs-site/src/content/docs/devlog/2026-05-24-m6-worker-job-pipeline.md`
- [X] T076 Add local smoke coverage for queued acknowledgement, reconnect, recovery, cancellation, and diagnostics in `specs/008-worker-job-pipeline/quickstart.md`
- [X] T077 Run backend validation from `specs/008-worker-job-pipeline/quickstart.md` using `go test ./...`
- [X] T078 Run frontend validation from `specs/008-worker-job-pipeline/quickstart.md` using `bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts`
- [X] T079 Run web build validation from `specs/008-worker-job-pipeline/quickstart.md` using `bun run --cwd web build`
- [X] T080 Run docs build validation from `specs/008-worker-job-pipeline/quickstart.md` using `bun run --cwd docs-site build`
- [X] T081 Record final validation outcomes and any exact skipped-command reasons in `docs-site/src/content/docs/devlog/2026-05-24-m6-worker-job-pipeline.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** has no dependencies and can start immediately.
- **Phase 2: Foundational** depends on Phase 1 and blocks every user story.
- **Phase 3: US1** depends on Phase 2 and is the MVP slice.
- **Phase 4: US2** depends on Phase 2; it can start after foundational work, but it is easiest to validate after US1 worker happy path exists.
- **Phase 5: US3** depends on Phase 2; it can start after foundational work, but running-stop validation benefits from US1 worker processing.
- **Phase 6: US4** depends on Phase 2; diagnostics can be built in parallel with US2/US3 once job states exist.
- **Phase 7: Polish** depends on whichever user stories are included in the implementation slice.

### User Story Dependencies

- **US1 (P1)**: No dependency on other stories after Phase 2; recommended MVP.
- **US2 (P2)**: Uses the same job/worker foundation as US1; independent validation is worker interruption and recovery.
- **US3 (P3)**: Uses the same job/worker foundation as US1; independent validation is queued and running cancellation.
- **US4 (P4)**: Uses job/worker state from prior stories; independent validation is diagnostics and timeline visibility for constructed states.

### Story Completion Order

```text
Setup -> Foundation -> US1 MVP -> US2 recovery -> US3 cancellation -> US4 diagnostics -> Polish
```

US2, US3, and US4 may proceed in parallel after Phase 2 if implementers coordinate shared files (`internal/productdata/service.go`, `internal/productdata/repository.go`, `internal/runtime/worker.go`, `internal/runtime/jobs.go`, `internal/runtime/pipeline.go`, `web/src/state.ts`, `web/src/components/RunTimeline.tsx`).

---

## Parallel Execution Examples

### User Story 1

```text
Task: T020 Add HTTP test for queued run acknowledgement and active-run conflict in internal/httpapi/runtime_test.go
Task: T021 Add worker happy-path test for claim, run execution, finalization, and single terminal event in internal/runtime/worker_test.go
Task: T022 Add frontend runtime adapter tests for queued, running, and completed event mapping in web/src/runtime/executionAdapter.test.ts
Task: T023 Add frontend state tests for history replay after queued background events in web/src/state.runtime.test.ts
```

### User Story 2

```text
Task: T035 Add repository tests for lease expiry, stale claim recovery, and retry exhaustion in internal/productdata/repository_test.go
Task: T036 Add worker recovery tests for stale lease detection and retry scheduling in internal/runtime/jobs_test.go
Task: T037 Add idempotency tests for duplicate completion and duplicate final message prevention in internal/productdata/service_test.go
Task: T038 Add frontend runtime tests for recovering and failed worker states in web/src/runtime/executionAdapter.test.ts
```

### User Story 3

```text
Task: T047 Add HTTP stop tests for queued, running, and already-terminal background runs in internal/httpapi/runtime_test.go
Task: T048 Add service tests for persisted stop requests and queued job cancellation in internal/productdata/service_test.go
Task: T049 Add worker cancellation tests for safe-boundary stop handling in internal/runtime/worker_test.go
Task: T050 Add frontend state tests for stopping and stopped event handling in web/src/state.runtime.test.ts
```

### User Story 4

```text
Task: T059 Add queue diagnostics repository tests for queued, leased, stale, retrying, dead, and redacted states in internal/productdata/repository_test.go
Task: T060 Add HTTP diagnostics endpoint tests for ready, paused, unhealthy, degraded, and stopped states in internal/httpapi/health_test.go
Task: T061 Add real API client tests for worker queue diagnostics parsing in web/src/realApiClient.test.ts
Task: T062 Add timeline rendering tests for worker diagnostic event summaries in web/src/components/RunTimeline.test.tsx
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup.
2. Complete Phase 2 foundational job, event, and schema work.
3. Complete Phase 3 US1 tasks.
4. Stop and validate queued acknowledgement, background continuation, history replay, and single terminal finalization.
5. Demo the M6 background execution slice before adding recovery, cancellation, and diagnostics breadth.

### Incremental Delivery

1. Setup + Foundation: durable queue, shared transitions, event types, and worker scaffolding.
2. US1: background execution MVP.
3. US2: recovery and retry safety.
4. US3: cancellation semantics.
5. US4: diagnostics and health visibility.
6. Polish: documentation, smoke validation, and final validation recording.

### Validation Gates

- After Phase 3, run US1 backend/frontend tests and a manual queued-run smoke from `specs/008-worker-job-pipeline/quickstart.md`.
- After Phase 4, run recovery and idempotency tests plus worker interruption smoke from `specs/008-worker-job-pipeline/quickstart.md`.
- After Phase 5, run queued and running cancellation smoke from `specs/008-worker-job-pipeline/quickstart.md`.
- After Phase 6, run diagnostics smoke and verify redaction from `specs/008-worker-job-pipeline/quickstart.md`.
- Before completion, run all validation commands listed in Phase 7.

## Task Summary

- **Total tasks**: 81
- **Setup**: 7 tasks
- **Foundational**: 12 tasks
- **US1**: 15 tasks
- **US2**: 12 tasks
- **US3**: 12 tasks
- **US4**: 11 tasks
- **Polish**: 12 tasks
- **Suggested MVP**: Complete Phase 1, Phase 2, and Phase 3 only.
