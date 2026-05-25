# Tasks: RunContext Pipeline Foundation

**Input**: Design documents from `specs/014-run-context-pipeline-foundation/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Included because this feature changes worker context recovery, runtime stage orchestration, persisted events, replay behavior, and docs-site validation.

**Organization**: Tasks are grouped by user story so the durable-context MVP can be implemented and validated before pipeline extensibility polish.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Planning and Contracts)

**Purpose**: Confirm current M8/M7 baselines and establish the 014 planning artifacts before implementation.

- [X] T001 Verify M8 queue closeout and M7 continuation baseline against `specs/008-worker-job-pipeline/`, `specs/012-tool-result-model-continuation/`, `specs/013-m8-worker-job-closeout/`, `docs-site/src/content/docs/roadmap/current-status.md`, and `docs-site/src/content/docs/architecture/worker-job-pipeline.md`
- [X] T002 Confirm active Spec Kit feature pointer remains `specs/014-run-context-pipeline-foundation/plan.md` in `AGENTS.md` and `.specify/feature.json`
- [X] T003 [P] Record RunContext loader contract decisions in `specs/014-run-context-pipeline-foundation/contracts/run-context-loader.md`
- [X] T004 [P] Record pipeline stage event contract decisions in `specs/014-run-context-pipeline-foundation/contracts/pipeline-stage-events.md`
- [X] T005 [P] Record frontend Timeline/debug trace contract decisions in `specs/014-run-context-pipeline-foundation/contracts/frontend-debug-trace.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared failing tests and minimal stage/context interfaces that all user stories depend on.

- [X] T006 [P] Add RunContext loader test fixtures for run/thread/message/job/provider/tool data in `internal/productdata/service_test.go`
- [X] T007 [P] Add PostgreSQL RunContext loader contract tests in `internal/productdata/repository_test.go`
- [X] T008 [P] Add pipeline stage ordering and inserted test-stage coverage in `internal/runtime/pipeline_test.go`
- [X] T009 [P] Add stage event mapping tests in `web/src/realApiClient.test.ts`
- [X] T010 [P] Add frontend runtime grouping tests for pipeline stage trace in `web/src/runtime/runtimeEventGroups.test.ts`
- [X] T011 Define minimal RunContext data boundary types in `internal/productdata/models.go`
- [X] T012 Define minimal stage result types in `internal/runtime/pipeline.go`

**Checkpoint**: The implementation can target stable context and stage contracts.

---

## Phase 3: User Story 1 - Restore run context from durable state (Priority: P1) MVP

**Goal**: Worker execution prepares critical run context from durable data before runtime invocation.

**Independent Test**: Create a queued run and verify the worker records context prepared before provider/runtime invocation without request-local context.

### Tests for User Story 1

- [X] T013 [P] [US1] Add successful RunContext loader service test in `internal/productdata/service_test.go`
- [X] T014 [P] [US1] Add missing run/thread/message/provider route failure tests in `internal/productdata/service_test.go`
- [X] T015 [P] [US1] Add worker test proving `prepare_context` completes before runtime invocation in `internal/runtime/worker_test.go`
- [X] T016 [P] [US1] Add queued runner test preserving tool-result continuation context through RunContext in `internal/runtime/worker_test.go`

### Implementation for User Story 1

- [X] T017 [US1] Implement durable RunContext loading use case in `internal/productdata/service.go`
- [X] T018 [US1] Implement PostgreSQL reads for RunContext sources in `internal/productdata/repository.go`
- [X] T019 [US1] Add in-memory service support for RunContext tests in `internal/productdata/service.go`
- [X] T020 [US1] Wire `prepare_context` stage into queued worker execution in `internal/runtime/queued_runner.go`
- [X] T021 [US1] Ensure missing required context fails before runtime invocation in `internal/runtime/queued_runner.go`

**Checkpoint**: US1 works independently; worker can prepare context from durable state and fail safely before runtime invocation.

---

## Phase 4: User Story 2 - Observe a linear execution pipeline (Priority: P2)

**Goal**: Persist and display safe stage trace for `prepare_context`, `resolve_tools`, `invoke_runtime`, and `finalize`.

**Independent Test**: Start a run and verify live SSE/history replay show the same ordered stage trace in Timeline/debug views.

### Tests for User Story 2

- [X] T022 [P] [US2] Add runtime pipeline event ordering tests in `internal/runtime/pipeline_test.go`
- [X] T023 [P] [US2] Add HTTP/SSE history replay coverage for stage events in `internal/httpapi/runtime_test.go`
- [X] T024 [P] [US2] Add real API client mapping coverage for stage started/completed/failed events in `web/src/realApiClient.test.ts`
- [X] T025 [P] [US2] Add Timeline stage trace rendering test in `web/src/components/RunTimeline.runtime.test.ts`
- [X] T026 [P] [US2] Add RunRail/debug trace visibility test in `web/src/components/RunRail.runtime.test.ts`

### Implementation for User Story 2

- [X] T027 [US2] Implement stage event recording helpers in `internal/runtime/pipeline.go`
- [X] T028 [US2] Implement `resolve_tools`, `invoke_runtime`, and `finalize` stage wiring in `internal/runtime/queued_runner.go`
- [X] T029 [US2] Keep model invocation behavior delegated to existing code in `internal/runtime/runner.go`
- [X] T030 [US2] Keep tool continuation behavior delegated through queued runner in `internal/runtime/queued_runner.go`
- [X] T031 [US2] Map stage backend events to frontend runtime events in `web/src/realApiClient.ts`
- [X] T032 [US2] Group stage trace events for Timeline/debug display in `web/src/runtime/runtimeEventGroups.ts`
- [X] T033 [US2] Render safe stage summaries in `web/src/components/RunTimeline.tsx`
- [X] T034 [US2] Render safe stage summaries in debug/rail surfaces in `web/src/components/RunRail.tsx`

**Checkpoint**: US2 works independently; stage trace is persisted and replayable.

---

## Phase 5: User Story 3 - Add stages without rewriting the AgentLoop body (Priority: P3)

**Goal**: Provide a narrow pipeline composition point so future stages can be inserted without rewriting the main runtime invocation body.

**Independent Test**: Insert a test-only no-op stage and verify it records through the shared stage contract.

### Tests for User Story 3

- [X] T035 [P] [US3] Add test-only inserted stage coverage in `internal/runtime/pipeline_test.go`
- [X] T036 [P] [US3] Add failure short-circuit coverage for a stage fixture in `internal/runtime/pipeline_test.go`

### Implementation for User Story 3

- [X] T037 [US3] Refactor pipeline execution into ordered stage composition in `internal/runtime/pipeline.go`
- [X] T038 [US3] Keep queued runner orchestration limited to stage registration and execution in `internal/runtime/queued_runner.go`
- [X] T039 [US3] Ensure absent future stages are not represented as placeholders in `internal/runtime/pipeline.go`

**Checkpoint**: Future stage insertion has a small tested seam without adding a broad workflow engine.

---

## Phase 6: Documentation and Validation

**Purpose**: Update docs-site and run the exact validation plan.

- [X] T040 [P] Update RunContext/pipeline architecture in `docs-site/src/content/docs/architecture/worker-job-pipeline.md`
- [X] T041 [P] Update stage event API contract in `docs-site/src/content/docs/api/worker-job-pipeline.md`
- [X] T042 [P] Add local M9 validation runbook in `docs-site/src/content/docs/runbooks/local-m9.md`
- [X] T043 [P] Add M9 implementation devlog in `docs-site/src/content/docs/devlog/2026-05-25-m9-run-context-pipeline-foundation.md`
- [X] T044 Update roadmap status and next steps in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T045 Update Spec Kit workflow/status reference in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T046 Run backend validation `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T047 Run related web runtime test `web/src/runtime/realExecutionAdapter.test.ts` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T048 Run related web runtime grouping test `web/src/runtime/runtimeEventGroups.test.ts` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T049 Run related Timeline test `web/src/components/RunTimeline.runtime.test.ts` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T050 Run related RunRail test `web/src/components/RunRail.runtime.test.ts` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T051 Run web build `bun run --cwd web build` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T052 Run docs build `bun run --cwd docs-site build` and record result in `specs/014-run-context-pipeline-foundation/quickstart.md`
- [X] T053 Perform browser smoke showing context prepared, tools resolved, runtime invoked, and finalized, then record result in `specs/014-run-context-pipeline-foundation/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No implementation dependency; confirms inputs and contracts.
- **Phase 2**: Depends on Phase 1; blocks all user stories.
- **US1**: Depends on Phase 2 and is the MVP.
- **US2**: Depends on US1 because stage trace needs prepared context and runtime invocation boundaries.
- **US3**: Depends on US2 because the composition seam should wrap real stage behavior.
- **Documentation and Validation**: Depends on implemented behavior and selected event names.

### User Story Dependencies

- **User Story 1 (P1)**: MVP; no dependency on later stories.
- **User Story 2 (P2)**: Builds on US1; independently testable through persisted stage trace.
- **User Story 3 (P3)**: Builds on US2; independently testable through inserted no-op/failure stage fixtures.

### Parallel Opportunities

- T003-T005 can run in parallel.
- T006-T010 can run in parallel.
- T013-T016 can run in parallel after T011-T012.
- T022-T026 can run in parallel after US1 contracts are stable.
- T035-T036 can run in parallel after pipeline composition exists.
- T040-T043 can run in parallel after implementation behavior is known.

## Parallel Execution Examples

```text
Task: T006 Add RunContext loader test fixtures in internal/productdata/service_test.go
Task: T008 Add pipeline stage ordering and inserted test-stage coverage in internal/runtime/pipeline_test.go
Task: T010 Add frontend runtime grouping tests in web/src/runtime/runtimeEventGroups.test.ts
```

```text
Task: T024 Add real API client mapping coverage in web/src/realApiClient.test.ts
Task: T025 Add Timeline stage trace rendering test in web/src/components/RunTimeline.runtime.test.ts
Task: T026 Add RunRail/debug trace visibility test in web/src/components/RunRail.runtime.test.ts
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Implement US1 only.
3. Validate that worker execution prepares RunContext from durable state and fails safely on missing context.
4. Add US2 stage trace visibility.
5. Add US3 composition seam.
6. Update docs and run the full validation plan.

### Hard Stops

- Stop if implementation requires a duplicate worker queue or queue schema redesign.
- Stop if RunContext requires Persona/Skill, MCP, Memory/RAG, Sandbox, Desktop Runtime, multi-agent, or broad tool execution to satisfy MVP.
- Stop if stage trace needs raw provider payloads, credentials, raw tool results, file contents, shell output, or hidden local state.
- Stop if M7 tool-result continuation regresses while routing through the pipeline.
