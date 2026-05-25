# Tasks: M7 Tool Call Approval Core

**Input**: Design documents from `specs/009-tool-call-approval-core/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [quickstart.md](./quickstart.md), [contracts/](./contracts/)

**Tests**: Included because the feature introduces a safety-critical approval boundary, idempotent APIs, worker block/resume behavior, redaction, and UI audit states that must be validated before completion.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested as an independent increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks in the same phase
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish M7 schema, constants, and shared tool contract scaffolding without implementing story behavior yet.

- [X] T001 Create M7 migration files in `migrations/000006_m7_tool_call_approval.up.sql` and `migrations/000006_m7_tool_call_approval.down.sql`
- [X] T002 [P] Add M7 tool-call lifecycle, approval status, execution status, and event type constants in `internal/productdata/models.go`
- [X] T003 [P] Add M7 frontend domain types for tool-call lifecycle, approval status, execution status, and tool events in `web/src/domain.ts`
- [X] T004 [P] Create allowlisted internal tool definition scaffold for `runtime.get_current_time` in `internal/runtime/tools.go`
- [X] T005 [P] Add tool argument/result redaction test skeletons in `internal/runtime/tools_test.go`
- [X] T006 Update schema readiness target and migration version expectations for M7 in `internal/db/readiness.go`
- [X] T007 [P] Add readiness tests for schema version 6 in `internal/db/readiness_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Implement durable tool-call projection, validation/redaction boundaries, scoped service APIs, and event helpers required by every user story.

**Critical**: No user story work should start until this phase is complete.

- [X] T008 Add migration SQL for the minimal `tool_calls` projection, unique `(run_id, tool_call_id)`, approval/execution status fields, redacted summaries, and rollback in `migrations/000006_m7_tool_call_approval.up.sql` and `migrations/000006_m7_tool_call_approval.down.sql`
- [X] T009 [P] Add repository tests for creating tool calls, scoped lookup by thread/run/tool_call_id, unique idempotency, and terminal-state guards in `internal/productdata/repository_test.go`
- [X] T010 Implement tool-call persistence and scoped lookup methods in `internal/productdata/repository.go`
- [X] T011 [P] Add service tests for schema validation, `timezone` omitted-or-UTC allowlist, argument redaction, unsupported tool rejection, duplicate `tool_call_id`, multiple tool request safe handling, and safe event metadata in `internal/productdata/service_test.go`
- [X] T012 Implement tool-call service methods for request recording, validation, redaction, and lifecycle event writing in `internal/productdata/service.go`
- [X] T013 [P] Add runtime tool registry tests for `runtime.get_current_time` schema validation, omitted-or-UTC timezone behavior, result shape, and forbidden capability invariants in `internal/runtime/tools_test.go`
- [X] T014 Implement the allowlisted internal tool registry and `runtime.get_current_time` executor in `internal/runtime/tools.go`
- [X] T015 [P] Add run event stream tests for M7 tool event ordering and history-first replay in `internal/runtime/stream_test.go`
- [X] T016 Extend run event stream handling for M7 tool lifecycle event categories/types in `internal/runtime/stream.go`
- [X] T017 [P] Add frontend adapter tests for replaying M7 tool events into a stable tool-call view model in `web/src/runtime/executionAdapter.test.ts`
- [X] T018 Add frontend state mapping for M7 tool-call view models in `web/src/runtime/executionAdapter.ts`

**Checkpoint**: Tool calls can be recorded safely, redacted, projected, and replayed without approval or execution UI yet.

---

## Phase 3: User Story 1 - Observe tool requests safely (Priority: P1) MVP

**Goal**: A model tool request becomes a persisted, redacted, approval-blocked tool-call lifecycle visible through history-first SSE and UI state, with no execution before approval.

**Independent Test**: Use a fake provider or controlled gateway event to request `runtime.get_current_time`; verify `tool_call_requested` and `tool_call_approval_required` are persisted, replayed, and rendered while no execution occurs.

### Tests for User Story 1

- [X] T019 [P] [US1] Add gateway tests for converting provider tool requests into `tool_call_requested` and `tool_call_approval_required` without execution in `internal/runtime/gateway_test.go`
- [X] T020 [P] [US1] Add HTTP/SSE test for history replay of approval-required tool events in `internal/httpapi/runtime_test.go`
- [X] T021 [P] [US1] Add ToolCallCard requested/approval-required rendering tests in `web/src/components/ToolCallCard.test.tsx`
- [X] T022 [P] [US1] Add Timeline grouping test that keeps tool events separate from model stream rows in `web/src/components/RunTimeline.test.tsx`

### Implementation for User Story 1

- [X] T023 [US1] Update provider/gateway conversion to pass allowlisted tool requests into the M7 service boundary in `internal/runtime/gateway.go`
- [X] T024 [US1] Record `tool_call_requested` and `tool_call_approval_required` events for approval-required `runtime.get_current_time` calls in `internal/productdata/service.go`
- [X] T025 [US1] Mark runs/jobs blocked on tool approval without treating the worker as failed or stale in `internal/runtime/worker.go`
- [X] T026 [US1] Expose current tool-call projection reads through scoped thread/run handler in `internal/httpapi/runtime.go`
- [X] T027 [US1] Register tool-call read route and dependencies in `internal/httpapi/server.go`
- [X] T028 [US1] Add real API client method for reading scoped tool-call state in `web/src/realApiClient.ts`
- [X] T029 [US1] Update ToolCallCard to show tool name, redacted argument summary, and approval-required controls disabled until handlers exist in `web/src/components/ToolCallCard.tsx`
- [X] T030 [US1] Update RunTimeline to render M7 tool request and approval-required events as a distinct tool group in `web/src/components/RunTimeline.tsx`
- [X] T031 [US1] Update RunRail to show waiting-for-tool-approval state without replacing model stream state in `web/src/components/RunRail.tsx`

**Checkpoint**: User Story 1 is independently functional as the safe observable tool-request MVP.

---

## Phase 4: User Story 2 - Approve or deny a pending tool call (Priority: P2)

**Goal**: Users can approve or deny a pending tool call through idempotent APIs and UI actions; denial never executes, and approval schedules exactly one resume.

**Independent Test**: Create pending tool calls; approve one 10 times and deny another 10 times; verify one decision event per call, no duplicate execution scheduling, and correct ToolCallCard terminal/approved states.

### Tests for User Story 2

- [ ] T032 [P] [US2] Add repository tests for idempotent approve, idempotent deny, conflicting decisions, and terminal-state conflicts in `internal/productdata/repository_test.go`
- [ ] T033 [P] [US2] Add service tests for `tool_call_approved` and `tool_call_denied` event idempotency in `internal/productdata/service_test.go`
- [ ] T034 [P] [US2] Add HTTP tests for approve/deny route scoping, retries, and conflict responses in `internal/httpapi/runtime_test.go`
- [ ] T035 [P] [US2] Add frontend API client tests for approve/deny parsing and safe retry behavior in `web/src/realApiClient.test.ts`
- [ ] T036 [P] [US2] Add ToolCallCard approve/deny interaction tests in `web/src/components/ToolCallCard.test.tsx`

### Implementation for User Story 2

- [ ] T037 [US2] Implement atomic approve and deny transitions in `internal/productdata/repository.go`
- [ ] T038 [US2] Implement approval decision service methods that record exactly one decision event, return 200/current state for repeated same decisions, and reserve conflicts for incompatible decision reversals in `internal/productdata/service.go`
- [ ] T039 [US2] Add scoped approve and deny HTTP handlers with same-decision idempotent 200 responses and conflicting-decision 409 responses in `internal/httpapi/runtime.go`
- [ ] T040 [US2] Register approve and deny routes in `internal/httpapi/server.go`
- [ ] T041 [US2] Wake or schedule the existing M6 worker/job pipeline exactly once after approval in `internal/runtime/jobs.go`
- [ ] T042 [US2] Ensure denied tool calls never reach execution and finalize the run through `run_stopped` in `internal/runtime/worker.go`
- [ ] T043 [US2] Add approve/deny client methods in `web/src/realApiClient.ts`
- [ ] T044 [US2] Wire ToolCallCard approve/deny controls to client actions and disable controls after decision in `web/src/components/ToolCallCard.tsx`
- [ ] T045 [US2] Map approved and denied events into frontend runtime state in `web/src/runtime/executionAdapter.ts`

**Checkpoint**: User Story 2 can be validated independently with approve/deny API retries and UI controls.

---

## Phase 5: User Story 3 - Execute the safest MVP internal tool (Priority: P3)

**Goal**: Approved `runtime.get_current_time` calls execute through the existing worker boundary and record executing plus exactly one terminal result/error/cancel event.

**Independent Test**: Approve `runtime.get_current_time`, verify executing and succeeded events with a redacted timestamp result; force validation/executor failure; stop while pending/executing and verify cancellation wins.

### Tests for User Story 3

- [ ] T046 [P] [US3] Add worker tests for approval resume, single execution attempt, and two-worker race prevention in `internal/runtime/worker_test.go`
- [ ] T047 [P] [US3] Add runtime tool executor tests for successful `runtime.get_current_time` result and redacted failure in `internal/runtime/tools_test.go`
- [ ] T048 [P] [US3] Add service tests for `tool_call_executing`, `tool_call_succeeded`, `tool_call_failed`, and `tool_call_cancelled` terminal guards in `internal/productdata/service_test.go`
- [ ] T049 [P] [US3] Add stop/cancel HTTP or worker tests for pending and executing tool calls in `internal/httpapi/runtime_test.go`
- [ ] T050 [P] [US3] Add frontend tests for executing, succeeded, failed, and cancelled ToolCallCard states in `web/src/components/ToolCallCard.test.tsx`

### Implementation for User Story 3

- [ ] T051 [US3] Execute approved `runtime.get_current_time` calls through the M7 internal tool executor in `internal/runtime/worker.go`
- [ ] T052 [US3] Record `tool_call_executing` before executor invocation in `internal/productdata/service.go`
- [ ] T053 [US3] Record `tool_call_succeeded` with redacted result summary in `internal/productdata/service.go`
- [ ] T054 [US3] Record `tool_call_failed` with stable redacted error codes for validation/executor failures and finalize unsafe/failed tool runs through `run_failed` in `internal/productdata/service.go`
- [ ] T055 [US3] Implement cancellation precedence for pending, approved, and executing tool calls in `internal/productdata/repository.go`
- [ ] T056 [US3] Make worker stop/recovery logic avoid duplicate terminal tool events after cancellation or lease recovery in `internal/runtime/worker.go`
- [ ] T057 [US3] Define the MVP tool-result-to-model context boundary without full multi-step continuation in `internal/runtime/gateway.go`
- [ ] T058 [US3] Map executing, succeeded, failed, and cancelled events into frontend runtime state in `web/src/runtime/executionAdapter.ts`
- [ ] T059 [US3] Render executing, result, redacted error, and cancelled states in `web/src/components/ToolCallCard.tsx`

**Checkpoint**: User Story 3 completes the runnable M7 approval-gated internal tool execution slice.

---

## Phase 6: User Story 4 - Keep tool events distinct from model streaming (Priority: P4)

**Goal**: RunRail and Timeline make tool lifecycle, approval, execution, result, error, and cancellation events clearly distinct from model text streaming.

**Independent Test**: Replay mixed model/tool runs and verify grouping, labels, ordering, redaction, and equivalent state after reconnect.

### Tests for User Story 4

- [ ] T060 [P] [US4] Add RunTimeline mixed model/tool grouping and ordering tests in `web/src/components/RunTimeline.test.tsx`
- [ ] T061 [P] [US4] Add RunRail summary tests for waiting, executing, succeeded, failed, denied, and cancelled tool states in `web/src/components/RunRail.test.tsx`
- [ ] T062 [P] [US4] Add frontend runtime replay equivalence tests for history-first and live M7 tool events in `web/src/runtime/executionAdapter.test.ts`
- [ ] T063 [P] [US4] Add backend stream ordering regression tests for mixed model/tool/final events in `internal/runtime/stream_test.go`

### Implementation for User Story 4

- [ ] T064 [US4] Refine RunTimeline tool grouping labels and metadata summaries in `web/src/components/RunTimeline.tsx`
- [ ] T065 [US4] Refine RunRail tool lifecycle summaries without regressing M6 queued/running/recovering states in `web/src/components/RunRail.tsx`
- [ ] T066 [US4] Ensure frontend runtime replay produces the same tool view model for historical and live events in `web/src/runtime/executionAdapter.ts`
- [ ] T067 [US4] Ensure backend stream serialization preserves mixed model/tool/final event order in `internal/runtime/stream.go`

**Checkpoint**: User Story 4 makes M7 tool audit states clearly inspectable in UI and history replay.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, smoke coverage, security checks, and final validation spanning all user stories.

- [ ] T068 [P] Add M7 architecture documentation in `docs-site/src/content/docs/architecture/tool-call-approval.md`
- [ ] T069 [P] Add M7 API and event contract documentation in `docs-site/src/content/docs/api/tool-call-approval.md`
- [ ] T070 [P] Add local M7 validation and troubleshooting runbook in `docs-site/src/content/docs/runbooks/local-m7.md`
- [ ] T071 [P] Add M7 devlog with validation results, known limitations, and deferred multi-step loop notes in `docs-site/src/content/docs/devlog/2026-05-24-m7-tool-call-approval.md`
- [ ] T072 Update roadmap current status with M7 scope, blocked/resumable tool diagnostic visibility, and remaining desktop/MCP/memory/multi-agent boundaries in `docs-site/src/content/docs/roadmap/current-status.md`
- [ ] T073 Add local smoke coverage for fake/model tool request, approval wait, approve, deny, execution, validation failure, duplicate tool_call_id, multiple tool request safe handling, cancellation, and SSE reconnect in `specs/009-tool-call-approval-core/quickstart.md`
- [ ] T074 Run backend validation from `specs/009-tool-call-approval-core/quickstart.md` using `go test ./...`
- [ ] T075 Run frontend validation from `specs/009-tool-call-approval-core/quickstart.md` using `bun test ./web/src/*.test.ts ./web/src/*.test.tsx ./web/src/components/*.test.ts ./web/src/components/*.test.tsx ./web/src/runtime/*.test.ts`
- [ ] T076 Run web build validation from `specs/009-tool-call-approval-core/quickstart.md` using `bun run --cwd web build`
- [ ] T077 Run docs build validation from `specs/009-tool-call-approval-core/quickstart.md` using `bun run --cwd docs-site build`
- [ ] T078 Perform browser smoke for ToolCallCard approve/deny/result/error/cancel states in `web/`
- [ ] T079 Record final validation outcomes and exact skipped-command reasons in `docs-site/src/content/docs/devlog/2026-05-24-m7-tool-call-approval.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup** has no dependencies and can start immediately.
- **Phase 2: Foundational** depends on Phase 1 and blocks every user story.
- **Phase 3: US1** depends on Phase 2 and is the MVP observable safety boundary.
- **Phase 4: US2** depends on Phase 2 and benefits from US1 pending tool-call creation.
- **Phase 5: US3** depends on US2 approval semantics and completes execution.
- **Phase 6: US4** depends on Phase 2 and can start after US1 event mapping exists, but full coverage benefits from US2/US3 states.
- **Phase 7: Polish** depends on whichever user stories are included in the implementation slice.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2/US3 after Foundation; delivers non-executing approval-required visibility.
- **US2 (P2)**: Requires pending tool-call state from US1 or fixture creation; independently validates approve/deny idempotency.
- **US3 (P3)**: Requires approved state from US2 to execute the MVP tool safely.
- **US4 (P4)**: UI grouping can start with US1 events and expand as more terminal states exist.

### Story Completion Order

```text
Setup -> Foundation -> US1 observable request/approval wait -> US2 approve/deny -> US3 execute MVP tool -> US4 UI grouping polish -> Documentation/validation
```

US4 tests and UI grouping can proceed in parallel with US2/US3 if implementers coordinate shared files (`web/src/runtime/executionAdapter.ts`, `web/src/components/RunTimeline.tsx`, `web/src/components/RunRail.tsx`, `web/src/components/ToolCallCard.tsx`).

---

## Parallel Execution Examples

### User Story 1

```text
Task: T019 Add gateway tests for converting provider tool requests into M7 events in internal/runtime/gateway_test.go
Task: T020 Add HTTP/SSE history replay test in internal/httpapi/runtime_test.go
Task: T021 Add ToolCallCard requested/approval-required rendering tests in web/src/components/ToolCallCard.test.tsx
Task: T022 Add Timeline grouping test in web/src/components/RunTimeline.test.tsx
```

### User Story 2

```text
Task: T032 Add repository approve/deny idempotency tests in internal/productdata/repository_test.go
Task: T034 Add HTTP approve/deny retry tests in internal/httpapi/runtime_test.go
Task: T035 Add frontend API client approve/deny tests in web/src/realApiClient.test.ts
Task: T036 Add ToolCallCard approve/deny interaction tests in web/src/components/ToolCallCard.test.tsx
```

### User Story 3

```text
Task: T046 Add worker approval resume and race tests in internal/runtime/worker_test.go
Task: T047 Add runtime tool executor tests in internal/runtime/tools_test.go
Task: T048 Add terminal tool event guard tests in internal/productdata/service_test.go
Task: T050 Add ToolCallCard terminal state tests in web/src/components/ToolCallCard.test.tsx
```

### User Story 4

```text
Task: T060 Add RunTimeline mixed model/tool grouping tests in web/src/components/RunTimeline.test.tsx
Task: T061 Add RunRail tool summary tests in web/src/components/RunRail.test.tsx
Task: T062 Add runtime replay equivalence tests in web/src/runtime/executionAdapter.test.ts
Task: T063 Add backend mixed event ordering tests in internal/runtime/stream_test.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup.
2. Complete Phase 2 foundational persistence, validation, redaction, registry, and event replay.
3. Complete Phase 3 US1.
4. Stop and validate fake/model tool request -> `approval_required` -> history-first SSE -> ToolCallCard/Timeline visibility.
5. Demo the safety boundary before implementing approval decisions or execution.

### Incremental Delivery

1. Setup + Foundation: schema, constants, projection, validation, redaction, event contracts.
2. US1: observable approval-required tool requests, no execution.
3. US2: idempotent approve/deny APIs and UI actions.
4. US3: approved `runtime.get_current_time` execution and terminal result/error/cancel states.
5. US4: polished grouping and replay equivalence across ToolCallCard/RunRail/Timeline.
6. Polish: docs-site updates, smoke validation, final validation recording.

### Validation Gates

- After Phase 3, run US1 backend/frontend tests and fake-provider approval-required smoke from `specs/009-tool-call-approval-core/quickstart.md`.
- After Phase 4, run approve/deny idempotency tests and smoke.
- After Phase 5, run worker resume, execution, failure, cancellation, and duplicate-prevention tests plus browser smoke.
- After Phase 6, run mixed event replay and UI grouping tests.
- Before completion, run all validation commands listed in Phase 7.

## Task Summary

- **Total tasks**: 79
- **Setup**: 7 tasks
- **Foundational**: 11 tasks
- **US1**: 13 tasks
- **US2**: 14 tasks
- **US3**: 14 tasks
- **US4**: 8 tasks
- **Polish**: 12 tasks
- **Suggested MVP**: Complete Phase 1, Phase 2, and Phase 3 only.
