# Tasks: M4 Run, Event, and SSE

**Input**: Design documents from `/specs/003-m4-run-event-sse/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: TDD is required for this implementation. Test tasks appear before implementation tasks and must fail before code is written.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)
- Every task includes exact file paths

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare M4 schema, route registration points, and shared domain placeholders.

- [X] T001 Create M4 migration files in `migrations/000003_m4_run_event_sse.up.sql` and `migrations/000003_m4_run_event_sse.down.sql`
- [X] T002 [P] Add M4 run/event type constants and validation placeholders in `internal/productdata/models.go`
- [X] T003 [P] Create runtime package skeleton in `internal/runtime/simulator.go` and `internal/runtime/stream.go`
- [X] T004 [P] Create HTTP runtime handler skeleton in `internal/httpapi/runtime.go`
- [X] T005 Register M4 runtime routes in `internal/httpapi/server.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared persistence, readiness, errors, and simulator/stream primitives that all user stories need.

**Critical**: No user story work should begin until this phase is complete.

### Tests for Foundational Behavior

- [X] T006 [P] Add migration/readiness tests for schema version 3 in `internal/db/readiness_test.go`
- [X] T007 [P] Add run/event validation tests in `internal/productdata/service_test.go`
- [X] T008 [P] Add deterministic simulator tests in `internal/runtime/simulator_test.go`
- [X] T009 [P] Add history-then-live stream broadcaster tests in `internal/runtime/stream_test.go`
- [X] T010 [P] Add structured runtime error mapping tests in `internal/httpapi/errors_test.go`

### Implementation for Foundational Behavior

- [X] T011 Implement M4 migration up/down schema for `runs` and `run_events` in `migrations/000003_m4_run_event_sse.up.sql` and `migrations/000003_m4_run_event_sse.down.sql`
- [X] T012 Update schema readiness requirement from version 2 to version 3 in `internal/db/readiness.go`
- [X] T013 Implement Run, RunEvent, status/category constants, and validation helpers in `internal/productdata/models.go`
- [X] T014 Extend productdata service and repository interfaces for run/event operations in `internal/productdata/service.go`
- [X] T015 Implement in-memory run/event persistence helpers for tests in `internal/productdata/service.go`
- [X] T016 Implement pgx-backed run/event persistence and active-run constraints in `internal/productdata/repository.go`
- [X] T017 Implement deterministic local simulator step generation in `internal/runtime/simulator.go`
- [X] T018 Implement in-process history-then-live broadcaster in `internal/runtime/stream.go`
- [X] T019 Add runtime error codes `run_not_found` and `active_run_exists` in `internal/productdata/models.go` and `internal/httpapi/errors.go`

**Checkpoint**: M4 schema, domain, persistence, simulator, stream broadcaster, and errors are testable without HTTP or UI.

---

## Phase 3: User Story 1 - Start and observe a local run (Priority: P1) MVP

**Goal**: A user can start a deterministic local run for an existing active thread and see lifecycle status/events.

**Independent Test**: Create or seed a thread, start a run for it, verify a run id/status/source/thread id and initial lifecycle event exist; starting a second run for the same thread while active reports a conflict.

### Tests for User Story 1

- [X] T020 [P] [US1] Add service test for starting a run on an active thread in `internal/productdata/service_test.go`
- [X] T021 [P] [US1] Add service test rejecting a second active run for the same thread in `internal/productdata/service_test.go`
- [X] T022 [P] [US1] Add HTTP contract test for `POST /v1/threads/{thread_id}/runs` in `internal/httpapi/runtime_test.go`
- [X] T023 [P] [US1] Add frontend type/mapping test for M4 run status in `web/src/realApiClient.test.ts`

### Implementation for User Story 1

- [X] T024 [US1] Implement `StartRun` use case and active-run conflict handling in `internal/productdata/service.go`
- [X] T025 [US1] Implement repository start-run transaction and first lifecycle event persistence in `internal/productdata/repository.go`
- [X] T026 [US1] Implement `POST /v1/threads/{thread_id}/runs` handler in `internal/httpapi/runtime.go`
- [X] T027 [US1] Wire deterministic simulator start trigger from API startup/service wiring in `cmd/loomi-api/main.go`
- [X] T028 [US1] Add M4 run types to frontend domain in `web/src/domain.ts`
- [X] T029 [US1] Add `startRun` real API client method in `web/src/realApiClient.ts`

**Checkpoint**: User Story 1 can be demonstrated through API start-run smoke and frontend type/client coverage.

---

## Phase 4: User Story 2 - Persist a timeline of run events (Priority: P1)

**Goal**: Run event history is durable, ordered, and visible after refresh.

**Independent Test**: Start a run, allow simulator events to persist, reload run events, and verify stable ordered lifecycle/progress/message/final or error categories.

### Tests for User Story 2

- [X] T030 [P] [US2] Add service test for ordered persisted events in `internal/productdata/service_test.go`
- [X] T031 [P] [US2] Add repository test for unique `(run_id, sequence)` event ordering in `internal/productdata/repository_test.go`
- [X] T032 [P] [US2] Add HTTP contract test for `GET /v1/runs/{run_id}/events` in `internal/httpapi/runtime_test.go`
- [X] T033 [P] [US2] Add frontend event mapping test for lifecycle/progress/message/error/final categories in `web/src/realApiClient.test.ts`

### Implementation for User Story 2

- [X] T034 [US2] Implement event append/list use cases with stable ordering in `internal/productdata/service.go`
- [X] T035 [US2] Implement pgx event insert/list queries in `internal/productdata/repository.go`
- [X] T036 [US2] Implement `GET /v1/runs/{run_id}` and `GET /v1/runs/{run_id}/events` handlers in `internal/httpapi/runtime.go`
- [X] T037 [US2] Connect simulator step persistence to run/event service in `internal/runtime/simulator.go`
- [X] T038 [US2] Add `getRun` and `getRunEvents` real API client methods in `web/src/realApiClient.ts`
- [X] T039 [US2] Render real M4 event categories in `web/src/components/RunTimeline.tsx`

**Checkpoint**: User Story 2 can be demonstrated by API history read and UI timeline reload after refresh.

---

## Phase 5: User Story 3 - Stream run updates to the web shell (Priority: P1)

**Goal**: The web shell receives persisted history first and live run events afterward without manual refresh.

**Independent Test**: Open the stream for a run, confirm existing events are delivered before live events, then confirm new events appear automatically and can be recovered after reconnect.

### Tests for User Story 3

- [X] T040 [P] [US3] Add SSE handler test for history-before-live delivery in `internal/httpapi/runtime_test.go`
- [X] T041 [P] [US3] Add SSE reconnect test with `after_sequence` in `internal/httpapi/runtime_test.go`
- [X] T042 [P] [US3] Add frontend stale stream guard test in `web/src/state.test.ts`
- [X] T043 [P] [US3] Add frontend stream state mapping test in `web/src/state.test.ts`

### Implementation for User Story 3

- [X] T044 [US3] Implement SSE `GET /v1/runs/{run_id}/events/stream` handler in `internal/httpapi/runtime.go`
- [X] T045 [US3] Publish persisted events to broadcaster after simulator append in `internal/runtime/stream.go`
- [X] T046 [US3] Add `subscribeRunEvents` real API client method using EventSource-compatible behavior in `web/src/realApiClient.ts`
- [X] T047 [US3] Update workspace state to select current run, subscribe to stream, dedupe events, and ignore stale stream updates in `web/src/state.ts`
- [X] T048 [US3] Update `web/src/components/RunRail.tsx` to bind agent state motion to real M4 run/event state
- [X] T049 [US3] Update `web/src/components/ChatCanvas.tsx` to show running/recoverable stream states from real run events

**Checkpoint**: User Story 3 can be demonstrated by live timeline updates and reconnect history recovery.

---

## Phase 6: User Story 4 - Stop a local run safely (Priority: P2)

**Goal**: A user can request best-effort cooperative stop and see stopped or already-terminal outcome.

**Independent Test**: Start a deterministic run, issue stop while active, verify stopped terminal state and timeline events; issue stop on a terminal run and verify already-terminal without changing final outcome.

### Tests for User Story 4

- [X] T050 [P] [US4] Add service test for cooperative stop on active run in `internal/productdata/service_test.go`
- [X] T051 [P] [US4] Add service test for stop on already-terminal run in `internal/productdata/service_test.go`
- [X] T052 [P] [US4] Add HTTP contract test for `POST /v1/runs/{run_id}/stop` in `internal/httpapi/runtime_test.go`
- [X] T053 [P] [US4] Add frontend stop-run state test in `web/src/state.test.ts`

### Implementation for User Story 4

- [X] T054 [US4] Implement stop-run use case and terminal-state guard in `internal/productdata/service.go`
- [X] T055 [US4] Implement repository stop transaction and stop/final events in `internal/productdata/repository.go`
- [X] T056 [US4] Make deterministic simulator check cooperative stop at step boundaries in `internal/runtime/simulator.go`
- [X] T057 [US4] Implement `POST /v1/runs/{run_id}/stop` handler in `internal/httpapi/runtime.go`
- [X] T058 [US4] Add `stopRun` real API client behavior in `web/src/realApiClient.ts`
- [X] T059 [US4] Update Chat Canvas, Run Timeline, and Run Rail stopped state rendering in `web/src/components/ChatCanvas.tsx`, `web/src/components/RunTimeline.tsx`, and `web/src/components/RunRail.tsx`

**Checkpoint**: User Story 4 can be demonstrated by cooperative stop API and consistent frontend stopped state.

---

## Phase 7: User Story 5 - Keep later platform capabilities deferred (Priority: P2)

**Goal**: M4 documents and enforces the execution boundary without pulling in LLM, tools, workers, desktop runtime, or production auth.

**Independent Test**: Review API responses, event payloads, and docs to confirm simulated execution is labeled, user-controlled event text is data, secrets are redacted, and deferred capabilities are explicit.

### Tests for User Story 5

- [X] T060 [P] [US5] Add test that simulator source is always labeled `local_simulated` in `internal/runtime/simulator_test.go`
- [X] T061 [P] [US5] Add event redaction/error test in `internal/productdata/service_test.go`
- [X] T062 [P] [US5] Add frontend test that real run labels do not claim LLM/tool/worker execution in `web/src/components/RunTimeline.tsx` tests or `web/src/realApiClient.test.ts`

### Implementation for User Story 5

- [X] T063 [US5] Ensure run/event responses label `source: local_simulated` and avoid model/tool/worker terminology in `internal/productdata/models.go` and `internal/httpapi/runtime.go`
- [X] T064 [US5] Ensure event summaries/content are redacted before persistence or response in `internal/productdata/service.go`
- [X] T065 [US5] Document M4 architecture boundaries in `docs-site/src/content/docs/architecture/run-event-sse.md`
- [X] T066 [US5] Document M4 endpoint and SSE contracts in `docs-site/src/content/docs/api/run-event-sse.md`
- [X] T067 [US5] Document local M4 setup, validation, stream troubleshooting, and rollback in `docs-site/src/content/docs/runbooks/local-m4.md`
- [X] T068 [US5] Update Spec Kit workflow status for M4 in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T069 [US5] Add M4 devlog with validation results and deferred capability notes in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`

**Checkpoint**: User Story 5 can be verified by docs review, redaction tests, and absence of LLM/tool/worker claims in M4 runtime outputs.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Validate the complete M4 slice and keep documentation, contracts, and code aligned.

- [X] T070 [P] Update `specs/003-m4-run-event-sse/quickstart.md` if implementation commands differ from planned smoke commands
- [X] T071 [P] Verify `specs/003-m4-run-event-sse/contracts/http-m4.openapi.yaml`, `specs/003-m4-run-event-sse/contracts/sse-run-events.md`, and `docs-site/src/content/docs/api/run-event-sse.md` agree on endpoint names, event categories, and error codes
- [X] T072 Run `go test ./...` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`
- [X] T073 Run `bun test ./web/src/*.test.ts ./web/src/components/*.test.ts` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`
- [X] T074 Run `bun run --cwd web build` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`
- [X] T075 Run M4 API/SSE smoke from `specs/003-m4-run-event-sse/quickstart.md` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`
- [X] T076 Run browser smoke for real API run/event mode, or document exact blocker and API/SSE fallback evidence in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`
- [X] T077 Run `bun run --cwd docs-site build` and record the result in `docs-site/src/content/docs/devlog/2026-05-23-m4-run-event-sse.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **US1 Start and observe run (Phase 3)**: Depends on Foundational; MVP slice.
- **US2 Persist timeline (Phase 4)**: Depends on US1 because events require a run.
- **US3 Stream updates (Phase 5)**: Depends on US1 and US2 because stream delivery needs persisted runs/events.
- **US4 Stop run (Phase 6)**: Depends on US1 and uses US2 events; can begin after basic event persistence exists.
- **US5 Deferred boundaries/docs (Phase 7)**: Depends on US1-US4 behavior enough to document real outputs; redaction tests can start after Foundational.
- **Polish (Phase 8)**: Depends on all selected user stories.

### User Story Dependencies

- **US1 (P1)**: First MVP after Foundational; no dependency on other stories.
- **US2 (P1)**: Depends on US1 run creation.
- **US3 (P1)**: Depends on US1 run creation and US2 event history.
- **US4 (P2)**: Depends on US1 run lifecycle and US2 event persistence.
- **US5 (P2)**: Depends on all implemented behavior for final docs, but redaction/source-label tests can be prepared in parallel after Foundational.

### Within Each User Story

- Tests must be written and observed failing before implementation.
- Models and validation before service behavior.
- Service/repository behavior before HTTP handlers.
- HTTP contracts before frontend integration.
- Frontend domain/client mapping before component rendering.
- Story checkpoint must pass before moving to the next priority story.

### Parallel Opportunities

- T002, T003, and T004 can run in parallel after T001.
- T006 through T010 can run in parallel after setup skeletons exist.
- Within US1, T020 through T023 can be written in parallel before implementation.
- Within US2, T030 through T033 can be written in parallel before implementation.
- Within US3, T040 through T043 can be written in parallel before implementation.
- Within US4, T050 through T053 can be written in parallel before implementation.
- Within US5, T060 through T062 can be written in parallel; docs tasks T065 through T069 can be drafted in parallel once behavior names settle.
- Polish docs/contract consistency tasks T070 and T071 can run before final command validation.

---

## Parallel Example: User Story 1

```bash
# Tests that can be created together before implementation:
Task: "Add service test for starting a run on an active thread in internal/productdata/service_test.go"
Task: "Add service test rejecting a second active run for the same thread in internal/productdata/service_test.go"
Task: "Add HTTP contract test for POST /v1/threads/{thread_id}/runs in internal/httpapi/runtime_test.go"
Task: "Add frontend type/mapping test for M4 run status in web/src/realApiClient.test.ts"
```

## Parallel Example: User Story 3

```bash
# Tests that can be created together before implementation:
Task: "Add SSE handler test for history-before-live delivery in internal/httpapi/runtime_test.go"
Task: "Add SSE reconnect test with after_sequence in internal/httpapi/runtime_test.go"
Task: "Add frontend stale stream guard test in web/src/state.test.ts"
Task: "Add frontend stream state mapping test in web/src/state.test.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup.
2. Complete Phase 2 foundational schema/domain/persistence/simulator/stream primitives.
3. Complete Phase 3 User Story 1.
4. Validate independent start-run API smoke and tests.
5. Stop for review before implementing streaming or stop behavior.

### Incremental Delivery

1. Setup + Foundational → run/event data layer ready.
2. US1 → start and observe local deterministic run.
3. US2 → durable timeline history after refresh.
4. US3 → live history-then-live event stream.
5. US4 → cooperative stop.
6. US5 → safety/deferred-boundary documentation.
7. Polish → full quickstart validation and docs build.

### Parallel Team Strategy

With multiple agents after Foundational:

1. Agent A: US1 service/API start-run tests and implementation.
2. Agent B: US2 event history tests and repository/service work after US1 interfaces stabilize.
3. Agent C: US3 stream broadcaster/SSE tests after event persistence contract stabilizes.
4. Agent D: US5 documentation and redaction tests once runtime output names settle.

---

## Notes

- [P] tasks use different files or can be written as tests before shared implementation.
- Story labels map tasks to user stories for traceability.
- Keep M4 deterministic and local; do not introduce LLM, tools, worker queues, desktop runtime, production auth, attachments, RAG, or plugin runtime.
- Keep event payload content as data, not instructions.
- Commit after each completed story or logical checkpoint, not after every tiny test edit if that would create churn.
