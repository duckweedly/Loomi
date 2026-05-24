# Feature Specification: Streaming Chat Runtime

**Feature Branch**: `[006-streaming-chat-runtime]`

**Created**: 2026-05-23

**Status**: Draft

**Input**: User description: "Improve the LLM streaming Chat Canvas experience, Run Timeline / Debug panel grouping, backend capability and mode status visibility, Composer interactions, and real thread/message data experience so the frontend naturally presents richer model/run events as backend capabilities arrive."

## Current State & Gap Analysis

- The product already has a shared runtime boundary, mock and real execution modes, run event streaming, assistant draft state, stop semantics, stale-event guards, chat canvas states, timeline shell, and thread/message history surfaces.
- The main visible gap is that partial assistant output is tracked as runtime state but is not yet presented as a streaming assistant bubble in the chat canvas.
- The timeline currently shows basic progress events, but richer model, worker, retry, cancellation, token, provider, and error events need clearer grouping before higher-volume events arrive.
- Backend mode and capability are present as basic indicators, but they do not yet make unavailable backend, missing model setup, provider issues, stream disconnects, or run recovery clear enough for development debugging.
- Composer interactions cover basic send validation and active-run blocking, but need a complete product loop for stop, retry, regenerate, continue conversation, failed-send recovery, and draft preservation.
- Thread and message surfaces exist, but need smoother empty, loading, error, retry, selected-thread, and history states so real model output feels stable once it is persisted through messages, runs, and events.

## Clarifications

### Session 2026-05-23

- Q: How should Regenerate handle the previous assistant response? → A: Preserve the previous response and add a new assistant attempt.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Read streaming assistant output naturally (Priority: P1)

A user sends a chat message and immediately sees a pending assistant bubble that streams partial model output, then transitions into a clear completed, failed, stopped, or recovered state without the conversation jumping or duplicating content.

**Why this priority**: This is the primary user-facing value of real model output. If streaming output feels unstable, later backend model integration will appear broken even when events are technically delivered.

**Independent Test**: Can be tested by sending a message into a thread and replaying partial output, completion, error, stop, and recovery signals while verifying the chat canvas remains readable and state transitions are explicit.

**Acceptance Scenarios**:

1. **Given** a selected thread with no active run, **When** the user sends a message and model output begins, **Then** the chat canvas shows the user message and a pending assistant bubble that grows as partial output arrives.
2. **Given** a pending assistant bubble with partial content, **When** the model completes successfully, **Then** the bubble transitions to a completed assistant message without losing or duplicating text.
3. **Given** a pending assistant bubble with partial content, **When** the model fails, **Then** the bubble displays a failed state, preserves any generated draft text, and offers a clear path to retry.
4. **Given** a pending assistant bubble with partial content, **When** the user stops the run, **Then** the bubble displays a stopped state and preserves the partial text as stopped draft content.
5. **Given** a run is restored after reconnecting or reselecting a thread, **When** a draft already exists for that run, **Then** the chat canvas restores the draft in the correct position with the correct non-terminal or terminal state.

---

### User Story 2 - Understand run execution through grouped timeline events (Priority: P2)

A developer or product tester reviews a run and sees execution events grouped by purpose: run lifecycle, model stream, worker or job activity, and errors.

**Why this priority**: Loomi’s agent execution must be observable. Grouping event categories before event volume grows prevents the debug surface from becoming noisy once model, provider, worker, retry, and cancellation signals are added.

**Independent Test**: Can be tested by replaying a run containing lifecycle, model progress, token usage, worker/job, retry, cancellation, and error events, then verifying each event appears in a predictable group with clear labels and severity.

**Acceptance Scenarios**:

1. **Given** a run contains lifecycle and model events, **When** the user opens the timeline, **Then** lifecycle events and model-stream events appear in separate groups.
2. **Given** a run contains queued, claimed, retrying, and cancellation events, **When** the timeline is displayed, **Then** worker/job state changes are grouped together and do not obscure model output.
3. **Given** a run contains provider or execution errors, **When** the user reviews the timeline, **Then** error events are grouped, visually distinct, and linked to the affected run state.
4. **Given** a run contains token usage or model metadata, **When** the timeline is displayed, **Then** the information is visible without overwhelming the primary conversation view.

---

### User Story 3 - See backend capability and mode status clearly (Priority: P3)

A developer can quickly tell whether the app is running against mock data, local simulated execution, real model execution, or an unavailable backend mode, and can understand stream disconnects or run recovery states without inspecting logs.

**Why this priority**: Backend capability visibility is essential during staged development. It reduces false bug reports by distinguishing frontend UX issues from backend availability, setup, or provider issues.

**Independent Test**: Can be tested by switching between available execution modes and simulating backend unavailable, model setup missing, provider unavailable, stream disconnected, and run recovery states while verifying the UI communicates each condition distinctly.

**Acceptance Scenarios**:

1. **Given** the app is using mock or local simulated execution, **When** the user views the chat canvas or run surface, **Then** the mode is clearly labeled and not confused with real model execution.
2. **Given** the backend cannot be reached, **When** the user attempts a runtime action, **Then** the UI shows backend unavailable status and does not imply the model is thinking.
3. **Given** model setup or provider availability prevents real model execution, **When** the user starts a run, **Then** the UI explains the capability problem and preserves the user’s message or draft.
4. **Given** the event stream disconnects during an active run, **When** the UI detects the disconnect, **Then** the chat canvas and timeline show stream-disconnected status and distinguish it from model failure.
5. **Given** a run is being restored after reconnecting, **When** the user returns to the thread, **Then** the UI shows recovery status until the latest known run state is reconciled.

---

### User Story 4 - Compose, stop, retry, regenerate, and continue reliably (Priority: P4)

A user can send a message, avoid accidental duplicate sends, stop an active response, retry failed attempts, regenerate an assistant response, continue a conversation, and keep their input when a failure occurs.

**Why this priority**: These interactions determine whether early real-model usage feels usable, even before advanced backend capabilities are complete.

**Independent Test**: Can be tested by exercising composer input validation, keyboard send behavior, active-run blocking, stop, retry, regenerate, continue conversation, and failure recovery from a single thread.

**Acceptance Scenarios**:

1. **Given** the composer is empty or contains only whitespace, **When** the user attempts to send, **Then** no message is submitted and the user receives clear inline feedback.
2. **Given** a run is pending or active, **When** the user attempts to send another message, **Then** duplicate submission is prevented and the active run controls remain available.
3. **Given** a run is active, **When** the user selects stop, **Then** the run transitions to stopped and the composer becomes available for the next action.
4. **Given** a run failed after the user submitted input, **When** the failure is shown, **Then** the original user input remains recoverable and retry is available.
5. **Given** a completed assistant response exists, **When** the user selects regenerate, **Then** the previous assistant response remains visible and a new assistant attempt is added without losing conversation context.
6. **Given** a thread has prior messages, **When** the user submits another message, **Then** the conversation continues in the selected thread and the previous history remains visible.

---

### User Story 5 - Navigate real threads and messages smoothly (Priority: P5)

A user can browse threads, select a conversation, understand empty and loading states, recover from thread or message errors, and see persisted messages align with run and event state.

**Why this priority**: Model output eventually becomes message, run, and event data. If thread and message handling feels unstable, backend improvements will not translate into a coherent product experience.

**Independent Test**: Can be tested by loading the app with no thread, an empty thread, a thread with history, a loading thread, a failed message load, and a recovered thread while verifying each state is clear and actionable.

**Acceptance Scenarios**:

1. **Given** no thread is selected, **When** the chat canvas loads, **Then** it prompts the user to select or start a thread without showing stale messages.
2. **Given** a selected thread has no messages, **When** the thread opens, **Then** the empty state invites the user to start a conversation.
3. **Given** message history is loading, **When** the user selects a thread, **Then** a loading state appears instead of a blank or misleading conversation.
4. **Given** message history fails to load, **When** the error is shown, **Then** the user can retry loading without losing the selected thread context.
5. **Given** a persisted assistant message and its related run events exist, **When** the user selects the thread, **Then** the chat canvas and timeline agree on the latest run outcome.

---

### Edge Cases

- A stream sends duplicate or out-of-order partial output signals for the same run.
- A final completion signal arrives after the user has stopped the run.
- The user switches threads while an assistant draft is still streaming.
- The backend becomes unavailable after a user message is accepted but before assistant output begins.
- The event stream reconnects and replays events that were already displayed.
- A provider error contains no user-readable detail.
- Token usage or provider metadata arrives before or after final output.
- A retry or regenerate action is requested while another run is still active in the same thread.
- A thread loads successfully but its latest run is still being recovered.
- A failed send should not erase composer text that the user has not intentionally discarded.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The chat canvas MUST show a pending assistant bubble as soon as an assistant response is expected for a submitted user message.
- **FR-002**: The pending assistant bubble MUST append partial model output in order and keep the conversation layout stable while text grows.
- **FR-003**: The system MUST distinguish assistant draft states for pending, streaming, completed, failed, stopped, and recovering runs.
- **FR-004**: The system MUST preserve partial assistant draft content when a run fails, is stopped, or is recovered after reconnecting.
- **FR-005**: The system MUST convert successfully completed assistant output into the visible conversation history without duplicate assistant content.
- **FR-006**: The system MUST prevent stale, duplicate, or out-of-order run signals from corrupting the visible assistant draft or final message.
- **FR-007**: The timeline MUST group run events into at least Run lifecycle, Model stream, Worker/job, and Error groups.
- **FR-008**: The timeline MUST present run state changes, model progress, token usage, provider errors, queue/worker activity, retries, cancellation, and completion in the appropriate group when those signals are available.
- **FR-009**: The timeline MUST make error and cancellation events visually distinct from normal progress events.
- **FR-010**: The UI MUST expose the current execution mode in user-readable terms, including mock, local simulated, real model, backend unavailable, model setup missing, provider unavailable, stream disconnected, and run recovering states when applicable.
- **FR-011**: Runtime status indicators MUST clearly distinguish backend or stream availability problems from model generation failures.
- **FR-012**: The composer MUST reject empty or whitespace-only submissions without creating a message or run.
- **FR-013**: The composer MUST prevent duplicate message submission while a run for the selected thread is pending or active.
- **FR-014**: The user MUST be able to stop an active run and then continue the conversation from the same thread.
- **FR-015**: The user MUST be able to retry a failed run while preserving the original user input and visible failure context.
- **FR-016**: The user MUST be able to regenerate a completed assistant response by preserving the previous response and adding a new assistant attempt without losing earlier thread history.
- **FR-017**: The message surface MUST provide clear no-thread, empty-thread, loading, history, error, backend-unavailable, active-run, completed, failed, stopped, and recovering states.
- **FR-018**: Thread selection MUST keep the chat canvas, latest run status, and timeline/debug surface synchronized for the selected thread.
- **FR-019**: Message history errors MUST provide a retry path that does not clear the selected thread or composer content.
- **FR-020**: The feature MUST remain useful before full real-model and worker capabilities are available by displaying simulated, mock, unavailable, and deferred states honestly.

### Key Entities

- **Thread**: A conversation container with selected, empty, loading, error, and history states.
- **Message**: A user or assistant utterance that appears in the chat history and belongs to a thread.
- **Assistant Draft**: A temporary assistant response associated with a run before it becomes a completed message or terminal draft state.
- **Run**: A single assistant execution attempt with pending, active, completed, failed, stopped, cancelled, retrying, and recovering outcomes.
- **Run Event**: An observable execution signal that belongs to a run and appears in timeline/debug surfaces.
- **Backend Capability State**: The current execution capability visible to the user, such as mock, local simulated, real model, unavailable, setup missing, provider unavailable, disconnected, or recovering.
- **Composer Input**: The user’s pending text and submission controls for send, stop, retry, regenerate, and continue-conversation actions.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a streaming run, users see a pending assistant bubble within 500 ms of the first assistant-output signal in 95% of local development trials.
- **SC-002**: In scripted replay tests covering success, failure, stop, and reconnect recovery, 100% of assistant draft text remains visible in the correct state with no duplicated final message.
- **SC-003**: A developer can correctly identify the current execution mode or availability problem from the UI within 5 seconds in at least 90% of validation attempts.
- **SC-004**: Timeline validation scenarios containing lifecycle, model, worker/job, retry/cancel, token usage, and error signals place 100% of events into the expected event group.
- **SC-005**: Composer validation scenarios prevent 100% of empty submissions and duplicate active-run submissions while keeping recoverable user input after failure.
- **SC-006**: Thread/message validation scenarios for no thread, empty thread, loading, history, error, retry, and recovered latest run all display an explicit user-readable state.
- **SC-007**: During manual smoke testing, a tester can complete send, stop, retry, regenerate, and continue-conversation flows in one selected thread without refreshing the app, with regenerated responses preserving prior assistant attempts.

## Scope Boundaries & Non-Goals

- This feature is focused on frontend product behavior for chat, timeline/debug, capability status, composer interactions, and thread/message experience.
- This feature does not require introducing a new model provider, worker queue, authentication flow, desktop plugin, or external tool execution capability.
- This feature does not require changing the persisted data model unless planning discovers that current message, run, or event concepts cannot represent the required states.
- This feature should not hide mock, simulated, unavailable, or deferred backend states behind wording that implies real model execution.

## Assumptions

- Existing thread, message, run, event, and assistant draft concepts remain the core product language for this feature.
- The backend will increasingly provide model progress, completion, error, usage, worker, retry, cancellation, and recovery signals, but the frontend experience should remain useful before every signal exists.
- A selected thread can have at most one active or pending assistant run that blocks new message submission for that thread.
- Stopping, retrying, and regenerating are scoped to the currently selected thread and the relevant latest run or assistant response.
- Reconnect recovery means restoring the latest known run and draft state for the selected thread, not guaranteeing offline message composition or cross-device synchronization.
