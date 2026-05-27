# Tasks: M22 Bounded Agent Loop + Todo Foundation

**Input**: Design documents from `/specs/030-bounded-agent-loop-todo-foundation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Required by user request and specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Phase 1: Setup

- [X] T001 Update AGENTS.md current Spec Kit feature pointer to `specs/030-bounded-agent-loop-todo-foundation/plan.md`
- [X] T002 [P] Add Spec Kit workflow/current-status links for M22 in `docs-site/src/content/docs/spec-kit/workflow.md` and `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T003 [P] Review M21 `unsupported_tool_loop` and tool continuation boundaries in `internal/runtime/gateway.go` and `internal/runtime/queued_runner.go`

---

## Phase 2: Foundational

- [X] T004 Add bounded loop state/event metadata constants and validation helpers in `internal/productdata/models.go`
- [X] T005 Add safe todo metadata validation/redaction helpers in `internal/productdata/models.go`
- [X] T006 Extend in-memory and PostgreSQL service paths to persist/replay loop/todo metadata safely in `internal/productdata/service.go` and `internal/productdata/repository.go`
- [X] T007 Update runtime continuation routing to permit sequential bounded tool calls while preserving one pending/executing tool invariant in `internal/runtime/gateway.go` and `internal/runtime/queued_runner.go`

---

## Phase 3: User Story 1 - Bounded Multi-Step Tool Continuation (Priority: P1)

**Goal**: A Work mode run can complete after at least two sequential approved tool calls and a final provider response.

**Independent Test**: Backend smoke runs `workspace.glob -> approve -> result -> workspace.read -> approve -> result -> final`.

- [X] T008 [P] [US1] Add backend smoke for two approved sequential workspace tools in `internal/httpapi/bounded_agent_loop_smoke_test.go`
- [X] T009 [P] [US1] Add runtime unit tests for loop limit, repeated tool call id, disallowed later tool, and one pending tool invariant in `internal/runtime/gateway_test.go`
- [X] T010 [US1] Implement bounded continuation loop state until T008-T009 pass in `internal/runtime/gateway.go`
- [X] T011 [US1] Wire approved later tool execution and continuation resume until T008 passes in `internal/runtime/queued_runner.go`

---

## Phase 4: User Story 2 - Observable Todo State (Priority: P2)

**Goal**: Work mode can replay a safe todo list from run events into Work Plan View without exposing unsafe values.

**Independent Test**: Backend/frontend fixtures produce todo updates and UI replay shows current todo statuses after refresh.

- [X] T012 [P] [US2] Add productdata tests for safe todo metadata redaction and bounds in `internal/productdata/service_test.go`
- [X] T013 [P] [US2] Add frontend projection tests for todo snapshots and Chat/Work isolation in `web/src/workModeProjection.test.ts`
- [X] T014 [US2] Implement todo metadata normalization and event replay projection in `internal/productdata/models.go` and `web/src/workModeProjection.ts`
- [X] T015 [US2] Render todo statuses in Work Plan View without executable controls in `web/src/components/WorkPlanView.tsx`

---

## Phase 5: User Story 3 - Operator Control and Failure Visibility (Priority: P3)

**Goal**: Timeline and run detail surfaces explain loop phases, approval waits, stops, failures, and loop-limit state.

**Independent Test**: Web runtime tests render approval-required, continuation, loop-limit, stopped, failed, and completed loop states.

- [X] T016 [P] [US3] Add RunRail/RunTimeline rendering tests for loop index, continuation, loop limit, and stopped state in `web/src/components/RunRail.runtime.test.ts`
- [X] T017 [US3] Update runtime event mapping and timeline copy for loop/todo states in `web/src/components/RunRail.tsx` and `web/src/realApiClient.ts`
- [X] T018 [US3] Add backend stop/cancel-between-tools test in `internal/runtime/worker_test.go`
- [X] T019 [US3] Preserve terminal stop/cancel semantics across blocked-on-approval and continuation states in `internal/runtime/queued_runner.go`

---

## Phase 6: Documentation & Validation

- [X] T020 [P] Update docs-site architecture/api/runbook/devlog pages for M22 loop/todo behavior
- [X] T021 Run `go test ./...`
- [X] T022 Run `bun test --cwd web`
- [X] T023 Run `bun run --cwd web build`
- [X] T024 Run `bun run --cwd docs-site build`
- [X] T025 Run `git diff --check`
- [X] T026 Run browser smoke for Work Plan todo replay and timeline loop states

---

## Dependencies & Execution Order

- Setup and Foundational tasks block all user stories.
- US1 is the MVP and must pass before todo/timeline polish is called complete.
- US2 depends on safe loop/todo metadata shape from Foundational and can proceed after US1 backend metadata names settle.
- US3 depends on loop events from US1 and todo metadata from US2.
- Documentation and validation run after code behavior is complete.

## Parallel Opportunities

- T002 and T003 can run in parallel.
- T008 and T009 target different test scopes.
- T012 and T013 target backend/frontend projections separately.
- T016 and T018 target frontend/backend control visibility separately.
- T020 docs can run after event names settle.

## Implementation Strategy

1. Prove US1 with backend smoke before touching write/edit/shell features.
2. Add todo replay only as safe metadata; do not create a separate task system.
3. Make loop phases visible in timeline and keep stop/cancel semantics strict.
4. Validate with backend tests, frontend tests/build, docs build, diff check, and browser smoke.
