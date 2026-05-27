# Tasks: MCP Approval-Gated Execution

**Input**: Design documents from `specs/017-mcp-approval-gated-execution/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Required by the feature request: backend tests, worker tests, frontend replay tests, docs-site updates, and validation commands.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or only adds tests/docs
- **[Story]**: User story label for story phases only
- Every task includes exact file paths for the next implementation session

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm M7/M11 baselines and create the smallest shared test fixtures.

- [X] T001 Read `specs/017-mcp-approval-gated-execution/spec.md`, `specs/017-mcp-approval-gated-execution/plan.md`, and `specs/017-mcp-approval-gated-execution/contracts/` before editing code
- [X] T002 [P] Add MCP approval fixture data for discovered candidate plus persona allowed-tools in `internal/runtime/mcp_execution_test.go`
- [X] T003 [P] Add MCP approval fixture data for tool-call projection and run events in `internal/productdata/repository_test.go`
- [X] T004 [P] Add frontend MCP event replay fixtures in `web/src/runtime/realExecutionAdapter.test.ts` and `web/src/runtime/runtimeEventGroups.test.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared safe types, redaction, and status mapping required before any story executes.

**CRITICAL**: No user story work should begin until foundational redaction/status contracts are represented in tests.

- [X] T005 [P] Add backend redaction tests for MCP arguments/results/errors in `internal/runtime/mcp_stdio_test.go` and `internal/runtime/mcp_worker_execution_test.go`
- [X] T006 [P] Add productdata projection status tests for MCP source/status transitions in `internal/productdata/repository_test.go`
- [X] T007 Define MCP-safe projection and event metadata fields in `internal/productdata/models.go`
- [X] T008 Define MCP redaction helper boundaries in `internal/runtime/mcp_redaction.go`
- [X] T009 Wire MCP-safe event metadata constants or mapping helpers in `internal/productdata/models.go`

**Checkpoint**: Shared metadata and redaction boundaries are test-covered and ready for user-story implementation.

---

## Phase 3: User Story 1 - Gate discovered MCP tool execution through M7 approval (Priority: P1) MVP

**Goal**: Provider-requested MCP calls create approval-blocked M7 projections and execute nothing until approved.

**Independent Test**: A discovered persona-allowed MCP candidate request becomes one approval-required projection; disallowed requests are rejected; denial never starts the MCP process.

### Tests for User Story 1

- [X] T010 [P] [US1] Add backend test for discovered persona-allowed MCP request creating approval-blocked projection in `internal/runtime/mcp_execution_test.go`
- [X] T011 [P] [US1] Add backend test for undiscovered, disabled, unnamespaced, conflicting, and persona-disallowed MCP request rejection in `internal/runtime/mcp_execution_test.go`
- [X] T012 [P] [US1] Add idempotency coverage for repeated `(run_id, provider_tool_call_id)` projection recording in `internal/productdata/repository_test.go` and `internal/runtime/mcp_execution_test.go`
- [X] T013 [P] [US1] Add approve/deny API metadata regression tests for MCP tool calls in `internal/httpapi/runtime_test.go`

### Implementation for User Story 1

- [X] T014 [US1] Implement MCP candidate plus persona allowed-tools resolution before projection in `internal/runtime/gateway.go`
- [X] T015 [US1] Implement approval-blocked MCP projection creation/reuse in `internal/productdata/service.go`
- [X] T016 [US1] Emit redacted MCP approval-required and rejection events in `internal/runtime/gateway.go`
- [X] T017 [US1] Preserve approve/deny idempotency and scoped errors for MCP tool calls in `internal/httpapi/runtime.go`

**Checkpoint**: US1 blocks every valid MCP request behind M7 approval and rejects invalid requests without process startup.

---

## Phase 4: User Story 2 - Execute one approved local stdio MCP tool safely in the worker (Priority: P2)

**Goal**: The worker executes exactly one approved local stdio MCP tool call under ownership, timeout, cancellation, redaction, and no-duplicate guards.

**Independent Test**: Approve one MCP tool call, run worker execution, and verify one redacted success/failure with no duplicate process startup after retry/recovery.

### Tests for User Story 2

- [X] T018 [P] [US2] Add worker test for approved MCP execution requiring current lease ownership in `internal/runtime/mcp_worker_execution_test.go`
- [X] T019 [P] [US2] Add worker stopped-run and stale-execution tests preventing process startup in `internal/runtime/mcp_worker_execution_test.go`
- [X] T020 [P] [US2] Add retry/recovery tests proving started/succeeded/failed/cancelled states do not re-execute in `internal/runtime/mcp_worker_execution_test.go`
- [X] T021 [P] [US2] Add stdio lifecycle tests for success, timeout, early exit, stderr, invalid JSON, unsafe result, and cleanup in `internal/runtime/mcp_stdio_test.go`
- [X] T022 [P] [US2] Add redacted result/error event persistence tests in `internal/productdata/repository_test.go` and `internal/runtime/mcp_worker_execution_test.go`

### Implementation for User Story 2

- [X] T023 [US2] Implement worker readiness guard for approved MCP projections in `internal/runtime/queued_runner.go`
- [X] T024 [US2] Implement mark-started-before-invocation transition in `internal/productdata/service.go`
- [X] T025 [US2] Implement bounded stdio MCP invocation facade in `internal/runtime/mcp_stdio.go`
- [X] T026 [US2] Implement MCP execution orchestration and redacted success/failure events in `internal/runtime/queued_runner.go`
- [X] T027 [US2] Implement recovery handling for stale started MCP execution without re-execution in `internal/runtime/queued_runner.go`

**Checkpoint**: US2 executes one approved local stdio MCP call safely and never duplicates execution during retry/recovery.

---

## Phase 5: User Story 3 - Continue the provider once with the redacted MCP tool result (Priority: P3)

**Goal**: A successful MCP tool result is passed to the provider once as a redacted continuation, then the run finalizes or fails safely on additional tool requests.

**Independent Test**: A successful MCP execution produces one continuation with matching redacted result and rejects any second tool request.

### Tests for User Story 3

- [X] T028 [P] [US3] Add continuation context test using matching redacted MCP result in `internal/runtime/gateway_test.go`
- [X] T029 [P] [US3] Add denied/failed/cancelled/no-continuation tests in `internal/runtime/gateway_test.go`
- [X] T030 [P] [US3] Add unsupported additional tool request test for MCP continuation in `internal/runtime/gateway_test.go`
- [X] T031 [P] [US3] Add frontend replay tests for MCP approval/execution/continuation event types in `web/src/runtime/runtimeEventGroups.test.ts`
- [X] T032 [P] [US3] Add frontend adapter tests for persisted MCP event metadata in `web/src/runtime/realExecutionAdapter.test.ts`

### Implementation for User Story 3

- [X] T033 [US3] Extend provider-neutral continuation context for MCP tool source in `internal/runtime/gateway.go`
- [X] T034 [US3] Emit MCP continuation started/succeeded/failed and loop-unsupported events through existing gateway continuation paths in `internal/runtime/gateway.go` and `internal/runtime/queued_runner.go`
- [X] T035 [US3] Map MCP event types to frontend runtime groups in `web/src/runtime/runtimeEventGroups.ts`
- [X] T036 [US3] Preserve MCP safe metadata in real API replay mapping in `web/src/runtime/realExecutionAdapter.ts`

**Checkpoint**: US3 completes the one-result continuation loop and frontend replay remains stable.

---

## Phase 6: Documentation and Validation

**Purpose**: Update docs-site and run required validation before reporting implementation completion.

- [X] T037 [P] Update approval API documentation with MCP execution states in `docs-site/src/content/docs/api/tool-call-approval.md`
- [X] T038 [P] Add MCP approval-gated execution API/events documentation in `docs-site/src/content/docs/api/mcp-approval-gated-execution.md`
- [X] T039 [P] Add architecture documentation in `docs-site/src/content/docs/architecture/mcp-approval-gated-execution.md`
- [X] T040 [P] Add local validation runbook in `docs-site/src/content/docs/runbooks/local-m12-mcp-approval-execution.md`
- [X] T041 [P] Update roadmap status for M12 boundaries in `docs-site/src/content/docs/roadmap/current-status.md`
- [X] T042 [P] Add devlog with validation evidence in `docs-site/src/content/docs/devlog/2026-05-25-m12-mcp-approval-gated-execution.md`
- [X] T043 [P] Update Spec Kit workflow/status links in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T044 Run backend validation command `go test ./...`
- [X] T045 Run frontend validation command `bun test --cwd web`
- [X] T046 Run frontend build command `bun run --cwd web build`
- [X] T047 Run docs validation command `bun run --cwd docs-site build`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **US1 (Phase 3)**: Depends on Foundational and is the MVP.
- **US2 (Phase 4)**: Depends on US1 approval projection and decision flow.
- **US3 (Phase 5)**: Depends on US2 successful redacted execution result.
- **Documentation/Validation (Phase 6)**: Can begin during implementation but final validation depends on desired story scope.

### User Story Dependencies

- **US1 (P1)**: Independent MVP after Foundational.
- **US2 (P2)**: Requires US1 because execution cannot exist before approval projection.
- **US3 (P3)**: Requires US2 because continuation needs a successful redacted MCP result.

### Parallel Opportunities

- T002-T004 can run in parallel.
- T005-T006 can run in parallel.
- US1 tests T010-T013 can run in parallel before implementation.
- US2 tests T018-T022 can run in parallel after foundational metadata is clear.
- US3 tests T028-T032 can run in parallel after continuation/event contracts are understood.
- Documentation tasks T037-T043 can run in parallel with implementation once event names stabilize.

## Parallel Example: User Story 1

```text
Task: "T010 Add backend test for discovered persona-allowed MCP request creating approval-blocked projection in internal/runtime/mcp_execution_test.go"
Task: "T011 Add backend test for undiscovered, disabled, unnamespaced, conflicting, and persona-disallowed MCP request rejection in internal/runtime/mcp_execution_test.go"
Task: "T012 Add idempotency test for repeated projection recording in internal/productdata/mcp_tool_call_test.go"
Task: "T013 Add approve/deny API metadata regression tests for MCP tool calls in internal/httpapi/runtime_test.go"
```

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete US1 tests and implementation.
3. Validate that MCP tool requests are approval-blocked and invalid requests execute nothing.
4. Stop if only the approval gate is desired.

### Full M12 Slice

1. Complete US1.
2. Complete US2 to execute one approved local stdio MCP call safely.
3. Complete US3 to continue once with redacted result.
4. Update docs-site.
5. Run all validation commands in Phase 6.
