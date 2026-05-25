# Tasks: Persona Skill Foundation

**Input**: Design documents from `specs/015-persona-skill-foundation/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Included because this feature changes durable data model, run creation/resolution, worker RunContext, runtime tool/model behavior, frontend Timeline/debug visibility, and docs-site validation.

**Organization**: Tasks are grouped by user story so the built-in Persona DB MVP can be implemented and validated before UI/debug polish.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Planning and Baseline)

**Purpose**: Confirm the M9 baseline and establish narrow persona contracts before implementation.

- [X] T001 Verify M9 RunContext/Pipeline baseline in `specs/014-run-context-pipeline-foundation/`, `internal/productdata/service.go`, `internal/runtime/queued_runner.go`, and `docs-site/src/content/docs/architecture/worker-job-pipeline.md`
- [X] T002 Confirm active Spec Kit feature pointer is `specs/015-persona-skill-foundation/plan.md` in `AGENTS.md` and `.specify/feature.json`
- [X] T003 [P] Review persona sync contract in `specs/015-persona-skill-foundation/contracts/persona-sync.md`
- [X] T004 [P] Review persona resolution contract in `specs/015-persona-skill-foundation/contracts/persona-resolution.md`
- [X] T005 [P] Review frontend persona summary contract in `specs/015-persona-skill-foundation/contracts/frontend-persona-summary.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared schema/types and failing tests that all user stories depend on.

- [X] T006 [P] Add Persona and PersonaVersion model tests in `internal/productdata/service_test.go`
- [X] T007 [P] Add PostgreSQL persona table migration contract tests in `internal/productdata/repository_test.go`
- [X] T008 [P] Add RunContext persona snapshot test fixture helpers in `internal/runtime/worker_test.go`
- [X] T009 [P] Add frontend persona summary mapping tests in `web/src/realApiClient.test.ts`
- [X] T010 Add durable Persona, PersonaVersion, PersonaSnapshot, and PersonaSummary types in `internal/productdata/models.go`
- [X] T011 Add persona-related repository interface methods in `internal/productdata/service.go`
- [X] T012 Add persona schema migrations in `migrations/000007_m10_persona_skill_foundation.up.sql` and `migrations/000007_m10_persona_skill_foundation.down.sql`

**Checkpoint**: The implementation can target stable persona model and persistence contracts.

---

## Phase 3: User Story 1 - Sync built-in personas into durable state (Priority: P1) MVP

**Goal**: Built-in persona definitions sync to durable DB with active version history.

**Independent Test**: Run built-in sync twice and verify durable persona/version records are created or updated idempotently without duplicate active versions.

### Tests for User Story 1

- [X] T013 [P] [US1] Add built-in persona sync create test in `internal/productdata/service_test.go`
- [X] T014 [P] [US1] Add built-in persona sync idempotency test in `internal/productdata/service_test.go`
- [X] T015 [P] [US1] Add built-in persona version-change and safe-field-change preservation test in `internal/productdata/service_test.go`
- [X] T016 [P] [US1] Add PostgreSQL persona sync persistence test in `internal/productdata/repository_test.go`
- [X] T017 [P] [US1] Add invalid built-in allowed-tool validation test in `internal/productdata/service_test.go`

### Implementation for User Story 1

- [X] T018 [US1] Add built-in persona config data in `internal/productdata/builtin_personas.go`
- [X] T019 [US1] Implement built-in persona validation in `internal/productdata/service.go`
- [X] T020 [US1] Implement in-memory built-in persona sync in `internal/productdata/service.go`
- [X] T021 [US1] Implement PostgreSQL persona upsert and version activation in `internal/productdata/repository.go`
- [X] T022 [US1] Wire built-in persona sync into API startup in `cmd/loomi-api/main.go`

**Checkpoint**: US1 works independently; persona DB has idempotent built-in sync and version history.

---

## Phase 4: User Story 2 - Run with selected or inherited persona (Priority: P2)

**Goal**: Thread/run/default persona resolution produces a run-scoped persona snapshot in RunContext before runtime invocation.

**Independent Test**: Create runs with explicit run persona, thread persona, and default fallback, then verify RunContext records the expected persona snapshot/version.

### Tests for User Story 2

- [X] T023 [P] [US2] Add run override persona resolution test in `internal/productdata/service_test.go`
- [X] T024 [P] [US2] Add thread inherited persona resolution test in `internal/productdata/service_test.go`
- [X] T025 [P] [US2] Add default built-in persona fallback test in `internal/productdata/service_test.go`
- [X] T026 [P] [US2] Add inactive or missing persona prepare-context failure test in `internal/productdata/service_test.go`
- [X] T027 [P] [US2] Add worker test proving persona snapshot and model route are prepared before runtime invocation in `internal/runtime/worker_test.go`
- [X] T028 [P] [US2] Add persona allowed tools intersection test in `internal/runtime/tools_test.go`
- [X] T029 [P] [US2] Add HTTP run creation persona override test in `internal/httpapi/runtime_test.go`
- [X] T030 [P] [US2] Add HTTP thread persona selection test in `internal/httpapi/product_test.go`

### Implementation for User Story 2

- [X] T031 [US2] Add thread and run persona reference fields in `internal/productdata/models.go`
- [X] T032 [US2] Implement persona resolution use case in `internal/productdata/service.go`
- [X] T033 [US2] Implement PostgreSQL persona resolution reads and run snapshot persistence in `internal/productdata/repository.go`
- [X] T034 [US2] Extend RunContext with persona snapshot and safe summary in `internal/productdata/models.go`
- [X] T035 [US2] Include resolved persona snapshot in `PrepareRunContext` in `internal/productdata/service.go`
- [X] T036 [US2] Include resolved persona snapshot in PostgreSQL `PrepareRunContext` in `internal/productdata/repository.go`
- [X] T037 [US2] Apply persona model route in queued runtime preparation in `internal/runtime/queued_runner.go`
- [X] T038 [US2] Intersect persona allowed tools with existing runtime allowlist in `internal/runtime/tools.go`
- [X] T039 [US2] Accept run persona override in `internal/httpapi/runtime.go`
- [X] T040 [US2] Accept thread persona selection in `internal/httpapi/product.go`

**Checkpoint**: US2 works independently; worker RunContext records the selected/inherited/default persona snapshot/version and applies model/tool behavior through existing boundaries.

---

## Phase 5: User Story 3 - Observe safe persona summary in Timeline/debug (Priority: P3)

**Goal**: Timeline/debug and minimal frontend surface show safe persona summary without exposing raw system prompt.

**Independent Test**: Create a run with a resolved persona and verify live SSE/history replay show the same safe persona summary while raw prompt text is absent.

### Tests for User Story 3

- [X] T041 [P] [US3] Add pipeline metadata redaction test for persona prompt absence in `internal/runtime/pipeline_test.go`
- [X] T042 [P] [US3] Add HTTP history/SSE persona summary test in `internal/httpapi/runtime_test.go`
- [X] T043 [P] [US3] Add real API client persona summary mapping test in `web/src/realApiClient.test.ts`
- [X] T044 [P] [US3] Add runtime event grouping test for persona summary in `web/src/runtime/runtimeEventGroups.test.ts`
- [X] T045 [P] [US3] Add Timeline persona summary rendering test in `web/src/components/RunTimeline.runtime.test.ts`
- [X] T046 [P] [US3] Add RunRail/debug persona summary rendering test in `web/src/components/RunRail.runtime.test.ts`
- [X] T047 [P] [US3] Add minimal persona selector test in `web/src/components/Composer.test.ts`

### Implementation for User Story 3

- [X] T048 [US3] Add safe persona summary to prepare-context pipeline metadata in `internal/runtime/pipeline.go`
- [X] T049 [US3] Expose safe persona fields in API run/event history responses in `internal/httpapi/runtime.go`
- [X] T050 [US3] Map safe persona summary fields in `web/src/realApiClient.ts`
- [X] T051 [US3] Replay persona summary into runtime state in `web/src/runtime/realExecutionAdapter.ts`
- [X] T052 [US3] Group persona summary for Timeline/debug in `web/src/runtime/runtimeEventGroups.ts`
- [X] T053 [US3] Render safe persona summary in `web/src/components/RunTimeline.tsx`
- [X] T054 [US3] Render safe persona summary in debug/rail surfaces in `web/src/components/RunRail.tsx`
- [X] T055 [US3] Implement minimal persona selector in `web/src/components/Composer.tsx`

**Checkpoint**: US3 works independently; persona is observable through safe summaries and prompt text is not exposed.

---

## Phase 6: Documentation and Validation

**Purpose**: Update docs-site and run the exact validation plan.

- [X] T056 [P] Add Persona/Skill foundation architecture page in `docs-site/src/content/docs/architecture/persona-skill-foundation.md`
- [X] T057 [P] Add Persona/Skill foundation API/event contract page in `docs-site/src/content/docs/api/persona-skill-foundation.md`
- [X] T058 [P] Add local M10 persona validation runbook in `docs-site/src/content/docs/runbooks/local-m10-persona.md`
- [X] T059 [P] Add M10 implementation devlog in `docs-site/src/content/docs/devlog/2026-05-25-m10-persona-skill-foundation.md`
- [X] T060 Update roadmap status and next steps in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T061 Update Spec Kit workflow/status reference in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T062 Run backend validation `go test ./internal/productdata ./internal/runtime ./internal/httpapi ./cmd/...` and record result in `specs/015-persona-skill-foundation/quickstart.md`
- [X] T063 Run related web runtime/UI tests for persona selector/display, Timeline, and RunRail, then record result in `specs/015-persona-skill-foundation/quickstart.md`
- [X] T064 Run web build `bun run --cwd web build` and record result in `specs/015-persona-skill-foundation/quickstart.md`
- [X] T065 Run docs build `bun run --cwd docs-site build` and record result in `specs/015-persona-skill-foundation/quickstart.md`
- [X] T066 Perform browser smoke selecting persona or using default persona, confirming Timeline/debug safe summary and RunContext persona version, then record result in `specs/015-persona-skill-foundation/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1**: No implementation dependency; confirms inputs and contracts.
- **Phase 2**: Depends on Phase 1; blocks all user stories.
- **US1**: Depends on Phase 2 and is the MVP.
- **US2**: Depends on US1 because run resolution needs durable built-in persona records.
- **US3**: Depends on US2 because Timeline/debug summary needs resolved RunContext persona snapshot.
- **Documentation and Validation**: Depends on implemented behavior and selected UI path.

### User Story Dependencies

- **User Story 1 (P1)**: MVP; no dependency on later stories.
- **User Story 2 (P2)**: Builds on US1; independently testable through RunContext persona snapshot/version.
- **User Story 3 (P3)**: Builds on US2; independently testable through live/replayed Timeline/debug persona summary.

### Parallel Opportunities

- T003-T005 can run in parallel.
- T006-T009 can run in parallel.
- T013-T017 can run in parallel after T010-T012.
- T023-T030 can run in parallel after US1 persistence is stable.
- T041-T047 can run in parallel after safe persona summary fields are defined.
- T056-T059 can run in parallel after implementation behavior is known.

## Parallel Execution Examples

```text
Task: T013 Add built-in persona sync create test in internal/productdata/service_test.go
Task: T016 Add PostgreSQL persona sync persistence test in internal/productdata/repository_test.go
Task: T017 Add invalid built-in allowed-tool validation test in internal/productdata/service_test.go
```

```text
Task: T027 Add worker test proving persona snapshot is prepared before runtime invocation in internal/runtime/worker_test.go
Task: T029 Add HTTP run creation persona override test in internal/httpapi/runtime_test.go
Task: T030 Add HTTP thread persona selection test in internal/httpapi/product_test.go
```

```text
Task: T043 Add real API client persona summary mapping test in web/src/realApiClient.test.ts
Task: T045 Add Timeline persona summary rendering test in web/src/components/RunTimeline.runtime.test.ts
Task: T046 Add RunRail/debug persona summary rendering test in web/src/components/RunRail.runtime.test.ts
```

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Implement US1 only.
3. Validate built-in persona sync and version history.
4. Add US2 run/thread/default resolution and RunContext snapshot.
5. Add US3 safe Timeline/debug frontend visibility.
6. Update docs and run the full validation plan.

### Hard Stops

- Stop if implementation requires a full Skill marketplace, plugin install flow, MCP, Memory/RAG, Sandbox/Desktop Runtime, or multi-agent behavior.
- Stop if persona tool names require creating new executable tool families instead of narrowing the existing allowlist.
- Stop if normal Timeline/debug needs raw system prompt text, provider credentials, raw provider payloads, raw tool results, file contents, shell output, or hidden local state.
- Stop if M9 RunContext/Pipeline behavior regresses or worker/job queue needs to be redesigned.
