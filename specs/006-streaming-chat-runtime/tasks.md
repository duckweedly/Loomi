# Tasks: Streaming Chat Runtime

**Input**: Design documents from `specs/006-streaming-chat-runtime/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Test tasks are included because the feature specification and quickstart define scripted replay, component, smoke, and build validation criteria.

**Coordination Guard**: This task list is only for `specs/006-streaming-chat-runtime/`. Do not create, edit, or overwrite `specs/005-llm-gateway/plan.md` or `specs/005-llm-gateway/tasks.md` while executing these tasks.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks.
- **[Story]**: Maps to user stories from `specs/006-streaming-chat-runtime/spec.md`.
- Every task includes exact file paths.

---

## Phase 1: Setup (Shared Context)

**Purpose**: Establish the 006 scope and avoid cross-feature contamination with the concurrently active 005 work.

- [x] T001 Confirm the implementation target is `specs/006-streaming-chat-runtime/plan.md` and leave `specs/005-llm-gateway/plan.md` untouched
- [x] T002 [P] Review streaming behavior requirements in `specs/006-streaming-chat-runtime/contracts/chat-canvas-streaming.md` before editing `web/src/components/ChatCanvas.tsx`
- [x] T003 [P] Review composer action requirements in `specs/006-streaming-chat-runtime/contracts/composer-actions.md` before editing `web/src/components/Composer.tsx`
- [x] T004 [P] Review timeline grouping requirements in `specs/006-streaming-chat-runtime/contracts/runtime-event-groups.md` before editing `web/src/components/RunRail.tsx`
- [x] T005 [P] Review capability status requirements in `specs/006-streaming-chat-runtime/contracts/backend-capability-status.md` before editing `web/src/runtime/backendCapabilityStatus.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend shared runtime domain and state primitives that later user stories depend on.

**CRITICAL**: No user story implementation should begin until this phase is complete.

- [x] T006 Extend run, run event, assistant draft, and message attempt domain types in `web/src/domain.ts`
- [x] T007 Update shared execution adapter semantics for draft finalization, stopped drafts, and regenerated attempts in `web/src/runtime/executionAdapter.ts`
- [x] T008 Update selected-run and assistant draft reducer helpers for terminal-state safety in `web/src/state.ts`
- [x] T009 [P] Add foundational runtime script fixtures for model delta, final, error, stopped, and replayed events in `web/src/runtime/runtimeScripts.ts`
- [x] T010 [P] Update baseline runtime adapter tests for new draft statuses in `web/src/runtime/mockExecutionAdapter.test.ts`

**Checkpoint**: Shared domain, adapter, mock script, and state primitives are ready for independent user story implementation.

---

## Phase 3: User Story 1 - Read streaming assistant output naturally (Priority: P1) MVP

**Goal**: A sent user message produces a pending assistant bubble that streams partial output and reaches completed, failed, stopped, or recovering states without duplicate final content.

**Independent Test**: Send a message in mock Chat mode, replay success/failure/stop/recovery signals, and verify Chat Canvas displays one stable assistant draft/final bubble for the selected thread only.

### Tests for User Story 1

- [x] T011 [P] [US1] Add assistant draft derivation tests for pending, streaming, completed, failed, stopped, and recovering states in `web/src/runtime/chatCanvasState.test.ts`
- [x] T012 [P] [US1] Add Chat Canvas rendering tests for pending assistant bubble, streaming growth, failed draft, stopped draft, and recovered draft in `web/src/components/ChatCanvas.states.test.ts`
- [x] T013 [P] [US1] Add stale and replayed draft event tests for selected-thread isolation in `web/src/state.staleRuntime.test.ts`
- [x] T014 [P] [US1] Add real event mapping tests for `model.delta`, `model.final`, and `model.error` in `web/src/realApiClient.test.ts`

### Implementation for User Story 1

- [x] T015 [US1] Update chat canvas state derivation for pending, streaming, stopped, and recovering assistant draft states in `web/src/runtime/chatCanvasState.ts`
- [x] T016 [US1] Update assistant draft append, finalization, failure, stop, and recovery handling in `web/src/state.ts`
- [x] T017 [US1] Update mock execution scripts to emit pending assistant bubble, model delta, final, failed, stopped, and recovery-like draft sequences in `web/src/runtime/mockExecutionAdapter.ts`
- [x] T018 [US1] Map real API model delta, final, and error events into assistant draft updates without duplicate final messages in `web/src/realApiClient.ts`
- [x] T019 [US1] Render the assistant draft bubble and terminal draft states in `web/src/components/ChatCanvas.tsx`
- [x] T020 [US1] Keep Agent state motion aligned with streaming, stopped, failed, and recovering run states in `web/src/components/AgentStateMotion.tsx`

**Checkpoint**: User Story 1 is fully functional and testable independently.

---

## Phase 4: User Story 2 - Understand run execution through grouped timeline events (Priority: P2)

**Goal**: Timeline/debug surfaces group events into Run lifecycle, Model stream, Worker/job, and Error so richer backend events remain readable.

**Independent Test**: Replay a run with lifecycle, model, usage, worker, retry, cancellation, and error events, then verify each event appears in exactly one expected group with error precedence.

### Tests for User Story 2

- [X] T021 [P] [US2] Add event grouping tests for lifecycle, model stream, worker/job, error, unknown, and error-precedence cases in `web/src/runtime/runtimeEventGroups.test.ts`
- [X] T022 [P] [US2] Add Run Rail grouped timeline rendering tests in `web/src/components/RunRail.polish.test.ts`
- [X] T023 [P] [US2] Add Run Timeline mixed-event scenario tests for lifecycle/model/worker/error groups in `web/src/components/RunTimeline.runtime.test.ts`

### Implementation for User Story 2

- [X] T024 [US2] Create runtime event grouping utility with group, severity, label, detail, usage, and unknown-event mapping rules in `web/src/runtime/runtimeEventGroups.ts`
- [X] T025 [US2] Update Run Rail to render stable event groups and visually distinct error/cancellation states in `web/src/components/RunRail.tsx`
- [X] T026 [US2] Update Run Timeline composition to pass grouped selected-run events through the timeline surface in `web/src/components/RunTimeline.tsx`
- [X] T027 [US2] Preserve token usage and provider metadata as timeline/debug details instead of assistant message text in `web/src/realApiClient.ts`
- [X] T028 [US2] Add grouped model, worker, retry, cancelled, and provider-error mock events in `web/src/runtime/runtimeScripts.ts`

**Checkpoint**: User Story 2 is fully functional and testable independently after foundational tasks.

---

## Phase 5: User Story 3 - See backend capability and mode status clearly (Priority: P3)

**Goal**: The UI clearly distinguishes mock, local simulated, real model, backend unavailable, model setup missing, provider unavailable, stream disconnected, and run recovering states.

**Independent Test**: Switch or simulate each capability status and verify Chat Canvas, Timeline, and status chips communicate the condition without implying the model is thinking when it is not.

### Tests for User Story 3

- [X] T029 [P] [US3] Add capability precedence tests for recovering, stream disconnected, provider unavailable, setup missing, backend unavailable, real model, local simulated, and mock statuses in `web/src/runtime/backendCapabilityStatus.test.ts`
- [X] T030 [P] [US3] Add Chat Canvas status-chip rendering tests for capability states in `web/src/components/ChatCanvas.states.test.ts`
- [X] T031 [P] [US3] Add real adapter capability tests for unavailable/setup/provider/stream conditions in `web/src/runtime/realExecutionAdapter.test.ts`

### Implementation for User Story 3

- [X] T032 [US3] Create backend capability status derivation utility with precedence and display copy in `web/src/runtime/backendCapabilityStatus.ts`
- [X] T033 [US3] Track stream disconnected and run recovering presentation state in `web/src/state.ts`
- [X] T034 [US3] Surface backend unavailable, model setup missing, provider unavailable, and stream disconnected signals from real API runtime mapping in `web/src/realApiClient.ts`
- [X] T035 [US3] Render user-readable capability status chips and detail copy in `web/src/components/ChatCanvas.tsx`
- [X] T036 [US3] Show capability status consistently in timeline/run rail details in `web/src/components/RunRail.tsx`

**Checkpoint**: User Story 3 is fully functional and testable independently after foundational tasks.

---

## Phase 6: User Story 4 - Compose, stop, retry, regenerate, and continue reliably (Priority: P4)

**Goal**: Composer prevents invalid sends, supports stop, retry, regenerate, and continue actions, preserves recoverable input on failure, and keeps regenerated responses as new attempts.

**Independent Test**: In one selected thread, exercise empty submit, active-run blocking, stop, failure retry, regenerate after completion, and continuing conversation without refreshing the app.

### Tests for User Story 4

- [X] T037 [P] [US4] Add composer action availability tests for send, stop, retry, regenerate, and continue in `web/src/runtime/composerActions.test.ts`
- [X] T038 [P] [US4] Create Composer keyboard and disabled-state tests for empty input, Enter, Shift+Enter, and active-run blocking in `web/src/components/Composer.test.ts`
- [X] T039 [P] [US4] Add App-level retry and regenerate flow tests for selected thread context in `web/src/App.controls.test.ts`
- [X] T040 [P] [US4] Add state tests for failed-input preservation and regenerated assistant attempts in `web/src/state.runtime.test.ts`

### Implementation for User Story 4

- [X] T041 [US4] Create composer action availability helper for send, stop, retry, regenerate, and continue guards in `web/src/runtime/composerActions.ts`
- [X] T042 [US4] Update Composer props and rendering for stop, retry, regenerate, continue, disabled send, and inline empty-input feedback in `web/src/components/Composer.tsx`
- [X] T043 [US4] Wire stop, retry, regenerate, and continue handlers through Chat Canvas controls in `web/src/components/ChatCanvas.tsx`
- [X] T044 [US4] Implement retry and regenerate selected-thread state transitions while preserving previous assistant messages in `web/src/state.ts`
- [X] T045 [US4] Wire App-level handlers for retry, regenerate, stop, and continue without crossing thread mode boundaries in `web/src/App.tsx`
- [X] T046 [US4] Update mock runtime adapter behavior for retry and regenerate attempts in `web/src/runtime/mockExecutionAdapter.ts`

**Checkpoint**: User Story 4 is fully functional and testable independently after foundational tasks.

---

## Phase 7: User Story 5 - Navigate real threads and messages smoothly (Priority: P5)

**Goal**: Thread selection and message history states are clear for no thread, empty thread, loading, history, error, retry, and recovered latest run states.

**Independent Test**: Load no-thread, empty-thread, loading, error, retry, history, and recovered-run scenarios and verify Chat Canvas, Thread Sidebar, Timeline, and Composer stay synchronized.

### Tests for User Story 5

- [X] T047 [P] [US5] Add no-thread, empty-thread, loading, error, history, backend-unavailable, active-run, terminal-run, and recovering state tests in `web/src/runtime/chatCanvasState.test.ts`
- [X] T048 [P] [US5] Add Thread Sidebar selected-thread, loading, and error retry tests in `web/src/components/ThreadSidebar.actions.test.ts`
- [X] T049 [P] [US5] Add App-level selected-thread synchronization tests for Chat Canvas and Timeline latest-run agreement in `web/src/App.threadModes.test.ts`

### Implementation for User Story 5

- [X] T050 [US5] Update Chat Canvas state copy and retry affordances for no-thread, empty-thread, loading, history, error, backend-unavailable, active-run, terminal-run, and recovering states in `web/src/components/ChatCanvas.tsx`
- [X] T051 [US5] Update Thread Sidebar loading, selected, empty, and retry affordances without clearing selected context in `web/src/components/ThreadSidebar.tsx`
- [X] T052 [US5] Update selected-thread and message history retry state handling in `web/src/state.ts`
- [X] T053 [US5] Keep selected run, messages, composer guards, and timeline synchronized on thread changes in `web/src/App.tsx`

**Checkpoint**: User Story 5 is fully functional and testable independently after foundational tasks.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and cleanup that affect multiple stories.

- [X] T054 [P] Update frontend runtime architecture documentation for streaming draft bubbles, composer actions, capability status, and timeline grouping in `docs-site/src/content/docs/architecture/frontend-agent-runtime.md`
- [X] T055 [P] Update run event SSE architecture documentation for model delta/final/error interpretation and grouped timeline semantics in `docs-site/src/content/docs/architecture/run-event-sse.md`
- [X] T056 [P] Update run event SSE API documentation for event type/category examples used by streaming chat runtime in `docs-site/src/content/docs/api/run-event-sse.md`
- [X] T057 [P] Update frontend runtime smoke runbook with streaming, failure, stop, retry, regenerate, capability, and thread/message smoke scenarios in `docs-site/src/content/docs/runbooks/frontend-runtime-smoke.md`
- [X] T058 [P] Add devlog entry for completed 006 streaming chat runtime validation in `docs-site/src/content/docs/devlog/2026-05-23-streaming-chat-runtime.md`
- [X] T059 [P] Update Spec Kit workflow index with the 006 feature status and artifact links in `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T060 Run frontend unit and component validation for `web/package.json` with `bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"`
- [X] T061 Run production web build validation for `web/package.json` with `bun run --cwd web build`
- [X] T062 Run documentation build validation for `docs-site/package.json` with `bun run --cwd docs-site build`
- [X] T063 Perform browser smoke validation from `specs/006-streaming-chat-runtime/quickstart.md` against the running web renderer

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; can start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion; blocks all user story implementation.
- **User Story 1 (Phase 3)**: Depends on Foundational; recommended MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational; can run after US1 or in parallel with US1 if files are coordinated.
- **User Story 3 (Phase 5)**: Depends on Foundational; can run after US1 or in parallel with US2 if `web/src/realApiClient.ts` and `web/src/components/ChatCanvas.tsx` edits are coordinated.
- **User Story 4 (Phase 6)**: Depends on Foundational; safest after US1 because it relies on terminal draft and assistant attempt semantics.
- **User Story 5 (Phase 7)**: Depends on Foundational; safest after US1 because it shares Chat Canvas state derivation.
- **Polish (Final Phase)**: Depends on the desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: MVP; no user-story dependencies after Foundational.
- **US2 (P2)**: Independent timeline slice after Foundational; shares mock/real event metadata with US1.
- **US3 (P3)**: Independent status slice after Foundational; shares Chat Canvas header and real API mapping with US1.
- **US4 (P4)**: Depends conceptually on US1 draft/terminal attempt behavior for retry and regenerate.
- **US5 (P5)**: Depends conceptually on US1 selected-run presentation behavior for recovered and terminal states.

### Within Each User Story

- Write tests before implementation tasks in the same story.
- Domain/state changes before component rendering changes.
- Runtime adapter and API mapping changes before App-level wiring.
- Story checkpoint validation before starting the next priority story when working sequentially.

---

## Parallel Opportunities

- Setup review tasks T002-T005 can run in parallel.
- Foundational T009 and T010 can run in parallel after T006-T008 are assigned carefully.
- US1 tests T011-T014 can run in parallel.
- US2 tests T021-T023 can run in parallel.
- US3 tests T029-T031 can run in parallel.
- US4 tests T037-T040 can run in parallel.
- US5 tests T047-T049 can run in parallel.
- Documentation tasks T054-T059 can run in parallel after implementation behavior is known.

---

## Parallel Example: User Story 1

```bash
Task: "Add assistant draft derivation tests in web/src/runtime/chatCanvasState.test.ts"
Task: "Add Chat Canvas assistant draft rendering tests in web/src/components/ChatCanvas.states.test.ts"
Task: "Add stale/replayed draft event tests in web/src/state.staleRuntime.test.ts"
Task: "Add real model event mapping tests in web/src/realApiClient.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "Add event grouping tests in web/src/runtime/runtimeEventGroups.test.ts"
Task: "Add Run Rail grouped rendering tests in web/src/components/RunRail.polish.test.ts"
Task: "Add Run Timeline mixed-event tests in web/src/components/RunTimeline.runtime.test.ts"
```

## Parallel Example: User Story 3

```bash
Task: "Add capability precedence tests in web/src/runtime/backendCapabilityStatus.test.ts"
Task: "Add Chat Canvas status chip tests in web/src/components/ChatCanvas.states.test.ts"
Task: "Add real adapter capability tests in web/src/runtime/realExecutionAdapter.test.ts"
```

## Parallel Example: User Story 4

```bash
Task: "Add composer action availability tests in web/src/runtime/composerActions.test.ts"
Task: "Add Composer keyboard/disabled tests in web/src/components/Composer.tsx"
Task: "Add App retry/regenerate tests in web/src/App.controls.test.ts"
Task: "Add state tests for failure preservation in web/src/state.runtime.test.ts"
```

## Parallel Example: User Story 5

```bash
Task: "Add Chat Canvas state matrix tests in web/src/runtime/chatCanvasState.test.ts"
Task: "Add Thread Sidebar retry tests in web/src/components/ThreadSidebar.actions.test.ts"
Task: "Add App selected-thread synchronization tests in web/src/App.threadModes.test.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 setup guard.
2. Complete Phase 2 foundational domain/state/runtime prerequisites.
3. Complete Phase 3 User Story 1.
4. Validate US1 with targeted tests and browser smoke from `specs/006-streaming-chat-runtime/quickstart.md`.
5. Stop and demo streaming assistant bubble before adding timeline grouping, capability status, composer actions, or thread/message polish.

### Incremental Delivery

1. Setup + Foundational → shared runtime state ready.
2. US1 → streaming assistant bubble MVP.
3. US2 → grouped timeline/debug readability.
4. US3 → backend capability and mode honesty.
5. US4 → composer stop/retry/regenerate/continue loop.
6. US5 → thread/message state polish.
7. Polish → docs, validation, and smoke coverage.

### Parallel Team Strategy

1. Keep 005 work isolated from `specs/006-streaming-chat-runtime/`.
2. Finish Phase 2 before splitting implementation.
3. Assign US1 and US2 to separate workers only if edits to `web/src/runtime/runtimeScripts.ts` and `web/src/realApiClient.ts` are coordinated.
4. Assign US3 after capability status utility ownership is clear.
5. Assign US4 and US5 after US1 state semantics are stable.

---

## Notes

- [P] tasks are parallelizable only when workers coordinate shared files named in the task list.
- Story labels map to `specs/006-streaming-chat-runtime/spec.md` user stories.
- Do not modify `specs/005-llm-gateway/*` from this task list.
- Do not update `.specify/feature.json` while 005 and 006 workflows are both active; use explicit `SPECIFY_FEATURE_DIRECTORY=specs/006-streaming-chat-runtime` for 006 scripts.
- No commits are required by this task list unless the user explicitly asks for commits.
