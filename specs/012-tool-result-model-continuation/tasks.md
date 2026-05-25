# Tasks: Tool Result Model Continuation

**Input**: Design documents from `/specs/012-tool-result-model-continuation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Test tasks are included because the feature requires provider continuation prompt construction, SSE event ordering, final assistant content, and denied/failed path coverage.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Phase 1: Setup (Shared Planning and Dependency Check)

**Purpose**: Confirm Window A landed the execution result boundary before implementation starts.

- [x] T001 Verify Window A approve/deny and `runtime.get_current_time` execution contracts in `internal/productdata/models.go`, `internal/productdata/service.go`, `internal/runtime/worker.go`, and `docs-site/src/content/docs/api/tool-call-approval.md`
- [x] T002 Update this feature's assumptions if Window A result metadata names differ in `specs/012-tool-result-model-continuation/data-model.md`
- [x] T003 [P] Add implementation notes for the selected result projection fields in `specs/012-tool-result-model-continuation/contracts/provider-continuation.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared contracts needed before any user story work.

- [x] T004 Add provider continuation request/response test fixtures in `internal/runtime/gateway_test.go`
- [ ] T005 Add tool-result projection service tests in `internal/productdata/service_test.go`
- [x] T006 [P] Add frontend replay fixtures for two model phases in `web/src/runtime/realExecutionAdapter.test.ts`
- [x] T007 Define or reuse model phase metadata constants in `internal/productdata/models.go` and `web/src/domain.ts`

**Checkpoint**: Continuation tests can be written against stable contracts.

---

## Phase 3: User Story 1 - Continue after an approved tool succeeds (Priority: P1) MVP

**Goal**: Approved tool success feeds one redacted result into one continuation provider call and produces one final assistant message.

**Independent Test**: Fake provider requests the time tool, Window A execution succeeds, continuation provider receives synthetic tool result and returns final assistant text containing the time.

### Tests for User Story 1

- [x] T008 [P] [US1] Add provider continuation prompt construction test in `internal/runtime/providers_test.go`
- [x] T009 [P] [US1] Add worker success-path event ordering test in `internal/runtime/worker_test.go`
- [x] T010 [P] [US1] Add final assistant persistence test in `internal/productdata/service_test.go`
- [x] T011 [P] [US1] Add frontend success replay test for post-tool `model_delta` and one final message in `web/src/runtime/realExecutionAdapter.test.ts`

### Implementation for User Story 1

- [x] T012 [US1] Implement tool-result projection lookup for continuation in `internal/productdata/service.go`
- [x] T013 [US1] Add gateway-neutral synthetic tool result context in `internal/runtime/gateway.go`
- [x] T014 [US1] Serialize synthetic tool result for provider adapters in `internal/runtime/providers.go`
- [ ] T015 [US1] Resume provider continuation after `tool_call_succeeded` in `internal/runtime/queued_runner.go`
- [x] T016 [US1] Persist second-phase model deltas, final assistant message, and run completion in `internal/runtime/runner.go`
- [x] T017 [US1] Map continuation phase metadata through API/SSE if needed in `web/src/realApiClient.ts`
- [x] T018 [US1] Update assistantDraft continuation handling in `web/src/runtime/realExecutionAdapter.ts`

**Checkpoint**: Success path works end to end and creates exactly one final assistant message.

---

## Phase 4: User Story 2 - Keep denial terminal and understandable (Priority: P2)

**Goal**: Denied tool calls do not invoke tool execution or model continuation.

**Independent Test**: Deny a pending tool call and verify no continuation provider call and no assistant answer claiming tool output.

### Tests for User Story 2

- [ ] T019 [P] [US2] Add denied-path no-continuation test in `internal/runtime/worker_test.go`
- [ ] T020 [P] [US2] Add frontend denied replay test in `web/src/runtime/realExecutionAdapter.test.ts`

### Implementation for User Story 2

- [ ] T021 [US2] Ensure denied terminal handling skips continuation in `internal/runtime/queued_runner.go`
- [ ] T022 [US2] Ensure denied replay keeps ToolCallCard terminal and assistantDraft non-final in `web/src/runtime/realExecutionAdapter.ts`

**Checkpoint**: Denial is terminal and does not re-enter the model.

---

## Phase 5: User Story 3 - Fail safely on tool or continuation errors (Priority: P3)

**Goal**: Tool failure and continuation failure end the run with redacted, observable errors.

**Independent Test**: Force tool failure and provider failure separately; verify terminal events, no secret leakage, and no duplicate final messages.

### Tests for User Story 3

- [ ] T023 [P] [US3] Add tool-failed no-continuation test in `internal/runtime/worker_test.go`
- [ ] T024 [P] [US3] Add continuation provider failure test in `internal/runtime/runner_test.go`
- [x] T025 [P] [US3] Add unsupported second tool request test in `internal/runtime/providers_test.go`
- [ ] T026 [P] [US3] Add frontend failed continuation replay test in `web/src/runtime/realExecutionAdapter.test.ts`

### Implementation for User Story 3

- [ ] T027 [US3] Skip continuation after `tool_call_failed` in `internal/runtime/queued_runner.go`
- [ ] T028 [US3] Record redacted continuation provider failures in `internal/runtime/runner.go`
- [x] T029 [US3] Fail safely when continuation requests another tool in `internal/runtime/providers.go`
- [x] T030 [US3] Preserve partial failed continuation draft in `web/src/runtime/realExecutionAdapter.ts`

**Checkpoint**: Failure paths are visible, redacted, and terminal.

---

## Phase 6: User Story 4 - Display two model-stream phases clearly (Priority: P4)

**Goal**: Timeline, RunRail, ToolCallCard, and assistantDraft make the initial and continuation model phases understandable.

**Independent Test**: Replay a complete success run and verify phase ordering, one final chat message, and visible tool result.

### Tests for User Story 4

- [ ] T031 [P] [US4] Add RunTimeline two-phase grouping test in `web/src/components/RunTimeline.runtime.test.ts`
- [ ] T032 [P] [US4] Add RunRail continuation status test in `web/src/components/RunRail.runtime.test.ts`
- [ ] T033 [P] [US4] Add ToolCallCard succeeded result display test in `web/src/components/ToolCallCard.test.tsx`

### Implementation for User Story 4

- [ ] T034 [US4] Update Timeline phase rendering in `web/src/components/RunTimeline.tsx`
- [ ] T035 [US4] Update RunRail continuation/final state rendering in `web/src/components/RunRail.tsx`
- [ ] T036 [US4] Update ToolCallCard result display if Window A result fields changed in `web/src/components/ToolCallCard.tsx`

**Checkpoint**: Browser smoke can explain the full two-phase run.

---

## Phase 7: Documentation and Validation

**Purpose**: Update docs-site and validate the planning-to-implementation handoff.

- [x] T037 [P] Add architecture page in `docs-site/src/content/docs/architecture/tool-result-continuation.md`
- [x] T038 Update continuation event/result details in `docs-site/src/content/docs/api/tool-call-approval.md`
- [x] T039 Update local smoke steps in `docs-site/src/content/docs/runbooks/local-m7.md`
- [x] T040 [P] Add implementation devlog in `docs-site/src/content/docs/devlog/2026-05-25-tool-result-continuation.md`
- [x] T041 Update status and next steps in `docs-site/src/content/docs/roadmap/current-status.md`
- [x] T042 Run backend validation commands for runtime/productdata changes
- [ ] T043 Run web validation commands for runtime/UI changes
- [x] T044 Run `bun run build` from `docs-site/`
- [ ] T045 Perform browser smoke for success, denied, and failed paths

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1** depends on Window A landing approve/deny and execution.
- **Phase 2** depends on Phase 1.
- **US1** depends on Phase 2 and is the MVP.
- **US2** can proceed after Phase 2 but should verify Window A denial semantics first.
- **US3** can proceed after Phase 2 and should follow US1 gateway contracts.
- **US4** depends on event metadata and frontend state behavior from US1-US3.
- **Documentation and Validation** depends on selected implementation behavior.

### Parallel Opportunities

- T003, T006, T008-T011, T019-T020, T023-T026, T031-T033, T037, and T040 can run in parallel when they touch different files.

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete US1 success path.
3. Validate one approved tool result creates one final assistant answer.
4. Add denied, failed, and UI polish paths.

### Hard Stops

- Stop if Window A does not provide a redacted `tool_call_succeeded` result.
- Stop if implementing continuation requires adding shell/filesystem/network/MCP tools.
- Stop if a second tool-call loop appears necessary for MVP.

### Current Blockers

- T015, T019-T023, T027, and full browser smoke depend on Window A landing approve/deny endpoints and approved `runtime.get_current_time` execution.
- T043 is partially validated by `bun test src/runtime/realExecutionAdapter.test.ts`; broader web tests are blocked on an existing `web/src/i18n.ts` syntax error on `origin/main` and missing local `web/node_modules`.
