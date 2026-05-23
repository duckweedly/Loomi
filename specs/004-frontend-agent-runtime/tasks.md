# Tasks: Frontend Agent Runtime Skeleton

**Input**: Design documents from `specs/004-frontend-agent-runtime/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Included because the feature specification defines independent tests, ordered runtime behavior, stale-event guarantees, and quickstart automated validation.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested as an independent vertical slice.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks.
- **[Story]**: User story label for story phases only.
- Each task includes exact repository-relative file paths.

## Phase 1: Setup (Shared Runtime Module Skeleton)

**Purpose**: Establish the frontend-only runtime boundary files without adding dependencies or backend behavior.

- [X] T001 Create the runtime module skeleton files in `web/src/runtime/executionAdapter.ts`, `web/src/runtime/chatCanvasState.ts`, `web/src/runtime/runtimeScripts.ts`, `web/src/runtime/mockExecutionAdapter.ts`, and `web/src/runtime/realExecutionAdapter.ts`
- [X] T002 [P] Add runtime module exports or import paths needed by the web app in `web/src/apiClient.ts` without changing thread/message API behavior

---

## Phase 2: Foundational (Blocking Runtime Contracts)

**Purpose**: Define shared types and state containers that every user story consumes. No user story work should start until this phase is complete.

- [X] T003 Add RuntimeStatus, RuntimeEvent, RuntimeRun, AssistantDraft, RuntimeScript, BackendCapabilityState, and ExecutionAdapter-facing types in `web/src/domain.ts`
- [X] T004 Implement the shared ExecutionAdapter interface and runtime capability contract in `web/src/runtime/executionAdapter.ts`
- [X] T005 Add selected-thread runtime state, active run lookup, script selection, and stale-event guard data structures in `web/src/state.ts`

**Checkpoint**: Runtime types and shared state shape exist; story work can begin.

---

## Phase 3: User Story 1 - 看到明确的 Chat 工作区状态 (Priority: P1)

**Goal**: Chat Canvas shows an explicit state for no thread, empty thread, loading, error, history, backend-unavailable, and runtime-derived states instead of unexplained blank space.

**Independent Test**: Use mock/local state inputs to switch through every Chat Canvas state without sending a message; each state renders concise Chinese text and the correct available action.

### Tests for User Story 1

- [X] T006 [P] [US1] Add failing pure-state tests for Chat Canvas state priority and every state value in `web/src/runtime/chatCanvasState.test.ts`
- [X] T007 [P] [US1] Add failing source/component tests for Chinese Chat Canvas empty/loading/error/backend-unavailable copy in `web/src/components/ChatCanvas.states.test.ts`

### Implementation for User Story 1

- [X] T008 [US1] Implement `deriveChatCanvasState` with the priority rules from `contracts/chat-canvas-states.md` in `web/src/runtime/chatCanvasState.ts`
- [X] T009 [US1] Thread loading, error, selected thread, messages, runtime, and backend capability inputs into Chat Canvas state selection in `web/src/state.ts`
- [X] T010 [US1] Render no-thread, empty-thread, loading, error, history, waiting-run, running, completed, failed, and backend-unavailable states in `web/src/components/ChatCanvas.tsx`
- [X] T011 [US1] Disable or label Composer input for no-thread, loading, error, backend-unavailable, and active-run blocked states in `web/src/components/Composer.tsx`
- [X] T012 [US1] Preserve existing message history rendering while routing explicit state branches through `web/src/components/ChatCanvas.tsx`

**Checkpoint**: US1 works independently: opening or forcing each state produces a visible Chat workspace state and no blank main area.

---

## Phase 4: User Story 2 - 用 mock 剧本体验一次完整 Agent 执行 (Priority: P1)

**Goal**: In mock mode, submitting a message immediately shows the user message, creates a run, plays deterministic success/failure/stopped scripts, and appends exactly one assistant reply only for success.

**Independent Test**: In mock mode, send one message with the success script and one with the failure script; verify ordered events, visible user message under 300 ms, correct terminal state, and no fake success reply on failure.

### Tests for User Story 2

- [X] T013 [P] [US2] Add failing tests for success, failure, and stopped script event order in `web/src/runtime/runtimeScripts.test.ts`
- [X] T014 [P] [US2] Add failing mock adapter tests for sendMessage, createRun, subscribeRunEvents, appendAssistantDelta, completeRun, failRun, and stopRun in `web/src/runtime/mockExecutionAdapter.test.ts`
- [X] T015 [P] [US2] Add failing runtime orchestration tests for immediate user message, repeated runs, blocked second send, and terminal-event handling in `web/src/state.test.ts`

### Implementation for User Story 2

- [X] T016 [US2] Implement deterministic success, failure, and stopped runtime scripts with stable event vocabulary in `web/src/runtime/runtimeScripts.ts`
- [X] T017 [US2] Implement the mock execution adapter using deterministic scripts and per-run identifiers in `web/src/runtime/mockExecutionAdapter.ts`
- [X] T018 [US2] Implement send-message-to-run orchestration, assistant draft accumulation, run completion, run failure, and run stopping in `web/src/state.ts`
- [X] T019 [US2] Hook Composer submit into the shared runtime flow and preserve immediate user-message display in `web/src/components/Composer.tsx`
- [X] T020 [US2] Render assistant draft, completed assistant message, failed state, and stopped state in `web/src/components/ChatCanvas.tsx`
- [X] T021 [US2] Add a visible stop-run action for the selected active run in `web/src/components/RunRail.tsx`

**Checkpoint**: US2 works independently: mock success, failure, and stopped runs can be executed from Chat mode and verified without a backend.

---

## Phase 5: User Story 3 - Chat、Timeline 和 Agent 状态徽章联动 (Priority: P2)

**Goal**: Chat Canvas, Run Timeline, and AgentStateMotion all reflect the same selected run state and event sequence.

**Independent Test**: Use one mock run event sequence to verify Chat Canvas, Timeline, and Agent badge transition together through waiting, running, completed, failed, and stopped semantics.

### Tests for User Story 3

- [X] T022 [P] [US3] Add failing tests for Timeline event rendering and terminal event ordering in `web/src/components/RunTimeline.runtime.test.ts`
- [X] T023 [P] [US3] Add failing tests for AgentStateMotion mapping from runtime pending, running, completed, failed, and stopped states in `web/src/components/AgentStateMotion.motion.test.ts`
- [X] T024 [P] [US3] Add failing stale-event and selected-thread switching tests in `web/src/useWorkspaceShellState.test.ts`

### Implementation for User Story 3

- [X] T025 [US3] Feed selected runtime events into RunRail and RunTimeline from the same active run source in `web/src/components/RunRail.tsx`
- [X] T026 [US3] Render ordered runtime milestones and terminal statuses in `web/src/components/RunTimeline.tsx`
- [X] T027 [US3] Drive AgentStateMotion from selected runtime status instead of independent decorative state in `web/src/components/AgentStateMotion.tsx`
- [X] T028 [US3] Apply stale-event guards when selected thread or active run changes in `web/src/useWorkspaceShellState.ts`
- [X] T029 [US3] Preserve Chat and Work mode-specific recent thread lists while attaching runtime state only to the selected Chat thread in `web/src/useWorkspaceShellState.ts`

**Checkpoint**: US3 works independently: all three surfaces show one coherent run state and old thread events do not affect the newly selected thread.

---

## Phase 6: User Story 4 - 为真实后端接入预留同一套状态机 (Priority: P2)

**Goal**: Mock and future real adapters share the same frontend state machine, while configured real API mode honestly reports runtime capability unavailable until M4/M5 exists.

**Independent Test**: Switch between mock mode and real API mode; mock mode executes deterministic scripts, while real mode enters backend-unavailable within one second of attempted execution without hidden mock fallback.

### Tests for User Story 4

- [X] T030 [P] [US4] Add failing real adapter tests for unavailable runtime capability and no mock fallback in `web/src/runtime/realExecutionAdapter.test.ts`
- [X] T031 [P] [US4] Add failing adapter selection tests for mock versus configured real API mode in `web/src/realApiClient.test.ts`
- [X] T032 [P] [US4] Add failing backend-unavailable UI state tests for attempted real-mode execution in `web/src/runtime/chatCanvasState.test.ts`

### Implementation for User Story 4

- [X] T033 [US4] Implement the real execution adapter that exposes unavailable runtime capability without executing mock scripts in `web/src/runtime/realExecutionAdapter.ts`
- [X] T034 [US4] Wire runtime adapter selection to mock or real data source mode in `web/src/apiClient.ts`
- [X] T035 [US4] Report real API runtime capability absence from `web/src/realApiClient.ts` while preserving durable thread/message behavior
- [X] T036 [US4] Keep mock durable thread/message behavior paired with mock runtime capability in `web/src/mockApiClient.ts`
- [X] T037 [US4] Surface backend-unavailable state on attempted real-mode execution through `web/src/state.ts` and `web/src/components/ChatCanvas.tsx`

**Checkpoint**: US4 works independently: real mode is honest about missing run/event backend support and does not fork the UI state model.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and cross-story consistency required by the Loomi constitution.

- [X] T038 [P] Document the frontend runtime state model, adapter boundary, stale-event guard, and deferred backend scope in `docs-site/src/content/docs/architecture/frontend-agent-runtime.md`
- [X] T039 [P] Document mock success, failure, stopped, stale-event, and real-mode backend-unavailable smoke checks in `docs-site/src/content/docs/runbooks/frontend-runtime-smoke.md`
- [X] T040 [P] Add a M3.5 implementation devlog with validation results and known limitations in `docs-site/src/content/docs/devlog/2026-05-23-m3-5-frontend-agent-runtime.md`
- [X] T041 [P] Link the M3.5 spec, plan, contracts, quickstart, and tasks from `docs-site/src/content/docs/spec-kit/workflow.md`
- [X] T042 Review sparse Chinese product microcopy across `web/src/components/ChatCanvas.tsx`, `web/src/components/Composer.tsx`, `web/src/components/RunTimeline.tsx`, and `docs-site/src/content/docs/architecture/frontend-agent-runtime.md`
- [X] T043 Run automated validation commands for `web/src/**/*.test.ts`, `web/vite.config.test.ts`, `web/`, and `docs-site/`: `bun test ./web/src/**/*.test.ts "web/vite.config.test.ts"`, `bun run --cwd web build`, and `bun run --cwd docs-site build`
- [X] T044 Run browser smoke for mock success, failure, stopped, stale-event switching, and real-mode backend-unavailable flows through `web/src/App.tsx`, then record results in `docs-site/src/content/docs/devlog/2026-05-23-m3-5-frontend-agent-runtime.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1; blocks all user stories.
- **US1 (Phase 3, P1)**: Depends on Phase 2; produces the MVP Chat Canvas state skeleton.
- **US2 (Phase 4, P1)**: Depends on Phase 2; can start in parallel with US1 after foundation, but full demo quality improves when US1 rendering exists.
- **US3 (Phase 5, P2)**: Depends on Phase 2 and benefits from US2 runtime events; implement after US2 for least churn.
- **US4 (Phase 6, P2)**: Depends on Phase 2 and US1 backend-unavailable rendering; can run in parallel with US3 after US1.
- **Polish (Phase 7)**: Depends on all selected user stories for final docs and validation.

### User Story Dependencies

- **US1 (P1)**: No dependency on other user stories; testable with local state inputs.
- **US2 (P1)**: No dependency on other user stories at adapter level; UI integration uses US1 state rendering if already complete.
- **US3 (P2)**: Depends on runtime events from US2 for realistic cross-surface linkage.
- **US4 (P2)**: Depends on US1 backend-unavailable state rendering; otherwise independent of mock scripts.

### Within Each User Story

- Tests come before implementation and should fail before the related implementation task is completed.
- Shared type changes come before adapter/state/UI changes.
- Runtime scripts come before mock adapter behavior.
- Adapter behavior comes before Composer and Chat Canvas integration.
- State orchestration comes before Timeline and AgentStateMotion linkage.
- Documentation updates should happen in the same work session as implementation changes.

### Parallel Opportunities

- T002 can run alongside T001 if the runtime import path is known.
- T006 and T007 can run in parallel for US1 tests.
- T013, T014, and T015 can run in parallel for US2 tests.
- T022, T023, and T024 can run in parallel for US3 tests.
- T030, T031, and T032 can run in parallel for US4 tests.
- T038, T039, T040, and T041 can run in parallel once implementation behavior is known.
- US1 and adapter-level parts of US2 can be developed in parallel after Phase 2 if implementers coordinate edits to `web/src/state.ts` and `web/src/components/ChatCanvas.tsx`.

---

## Parallel Example: User Story 1

```text
Task: "Add failing pure-state tests for Chat Canvas state priority and every state value in web/src/runtime/chatCanvasState.test.ts"
Task: "Add failing source/component tests for Chinese Chat Canvas empty/loading/error/backend-unavailable copy in web/src/components/ChatCanvas.states.test.ts"
```

## Parallel Example: User Story 2

```text
Task: "Add failing tests for success, failure, and stopped script event order in web/src/runtime/runtimeScripts.test.ts"
Task: "Add failing mock adapter tests for runtime adapter operations in web/src/runtime/mockExecutionAdapter.test.ts"
Task: "Add failing runtime orchestration tests in web/src/state.test.ts"
```

## Parallel Example: User Story 3

```text
Task: "Add failing tests for Timeline event rendering in web/src/components/RunTimeline.runtime.test.ts"
Task: "Add failing tests for AgentStateMotion runtime status mapping in web/src/components/AgentStateMotion.motion.test.ts"
Task: "Add failing stale-event and selected-thread switching tests in web/src/useWorkspaceShellState.test.ts"
```

## Parallel Example: User Story 4

```text
Task: "Add failing real adapter tests in web/src/runtime/realExecutionAdapter.test.ts"
Task: "Add failing adapter selection tests in web/src/realApiClient.test.ts"
Task: "Add failing backend-unavailable UI state tests in web/src/runtime/chatCanvasState.test.ts"
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 (US1) to make Chat Canvas states explicit and independently testable.
3. Stop and validate US1 with state tests and a browser-visible state smoke.

### Demo Slice

1. Complete MVP First.
2. Complete Phase 4 (US2) to demonstrate one full mock success/failure/stopped Agent execution.
3. Validate success/failure/stopped scripts before moving to linkage or real-mode work.

### Incremental Delivery

1. Add US1 for explicit Chat workspace states.
2. Add US2 for deterministic mock Agent execution.
3. Add US3 for Chat/Timeline/Agent badge linkage.
4. Add US4 for future real adapter honesty and shared state-machine boundary.
5. Complete Phase 7 docs, automated validation, and browser smoke.

### Parallel Team Strategy

1. Finish Setup and Foundational tasks together.
2. Assign US1 and adapter-level US2 tests in parallel.
3. After US2 runtime events exist, assign US3 linkage and US4 real-mode boundary in parallel.
4. Keep edits to `web/src/state.ts`, `web/src/components/ChatCanvas.tsx`, and `web/src/useWorkspaceShellState.ts` coordinated because those files span multiple stories.

---

## Notes

- M3.5 remains frontend-only: no real run persistence, SSE, LLM gateway, worker queue, tool execution, desktop runtime, new database table, or new frontend framework.
- Mock and real adapters must share visible state semantics; real mode must not silently fallback to mock runtime behavior.
- Product UI microcopy should stay sparse; learning documentation under `docs-site/src/content/docs/` should explain the state model and adapter boundary in Chinese.
- Each completed story must be independently testable before continuing to the next priority.
