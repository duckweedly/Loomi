# Data Model: Frontend Agent Runtime Skeleton

## Chat Canvas State

Represents the visible state of the main chat work area.

### Values

- `no-thread`: No selected thread is available.
- `empty-thread`: A selected thread exists but has no messages and no active runtime.
- `loading`: Threads, messages, or run state are loading.
- `error`: Loading or user action failed unexpectedly.
- `history`: Selected thread has historical messages and no active runtime.
- `waiting-run`: User message is visible and the UI is waiting for runtime creation or first runtime event.
- `running`: Runtime events are actively progressing.
- `completed`: Runtime finished successfully and final assistant content is visible.
- `failed`: Runtime failed or was stopped.
- `backend-unavailable`: The selected data source cannot provide runtime/run-event behavior yet.

### Validation rules

- `loading` takes precedence while loading is true.
- `error` takes precedence when an unexpected recoverable error exists.
- `backend-unavailable` is used only when a real configured data source lacks runtime capability.
- `no-thread` is used when there is no selected thread.
- `empty-thread` requires selected thread, zero messages, and no active runtime.
- `history` requires at least one message and no active runtime state.
- Runtime states (`waiting-run`, `running`, `completed`, `failed`) are derived from the selected runtime state and must only apply to the selected thread.

## Runtime Run

Represents one frontend-observable Agent execution attempt for a thread.

### Fields

- `id`: Stable run identifier for the current execution attempt.
- `threadId`: Thread the run belongs to.
- `status`: `pending`, `running`, `completed`, `failed`, or `stopped`.
- `scriptId`: Mock script identifier when produced by mock runtime; absent for future real runtime.
- `events`: Ordered runtime events visible in Timeline.
- `assistantDraft`: Current assistant draft content and status.
- `createdAt`: User-visible creation time label or timestamp.
- `completedAt`: User-visible completion time label or timestamp when complete.

### State transitions

```text
pending -> running -> completed
pending -> running -> failed
pending -> running -> stopped
pending -> stopped
```

Completed, failed, and stopped are terminal for a single run. A new user message creates a new run rather than reusing a terminal run.

## Runtime Event

Represents one user-visible milestone in a runtime run.

### Event types

- `run.created`
- `context.loading`
- `assistant.thinking`
- `assistant.drafting`
- `assistant.message.completed`
- `run.completed`
- `run.failed`
- `run.stopped`

### Fields

- `id`: Stable event identifier within the run.
- `runId`: Parent runtime run.
- `threadId`: Thread the event belongs to.
- `type`: One of the event types above.
- `label`: Sparse user-facing label.
- `detail`: Concise event detail, Chinese in product UI where visible.
- `time`: User-visible time label.
- `status`: `pending`, `running`, `completed`, `failed`, or `stopped`.
- `assistantDelta`: Optional assistant text delta for drafting events.

### Validation rules

- Events must be ordered by script/runtime sequence.
- Events from a non-selected thread must not update visible selected-thread state.
- A terminal event prevents later script events from applying to the same run.

## Runtime Script

Represents a deterministic mock execution scenario.

### Fields

- `id`: `success` or `failure` for the first M3.5 slice.
- `name`: Chinese learning/debug label.
- `steps`: Ordered runtime event and optional assistant delta steps.
- `finalAssistantMessage`: Required for success script, absent for failure script.
- `terminalStatus`: `completed` or `failed`.

### Validation rules

- Success script must include at least six user-visible milestones and end with one assistant reply.
- Failure script must end in failed state and must not append final assistant success content.
- Scripts must be repeatable and generate independent run/event identifiers per execution.

## Assistant Draft

Represents assistant content while a run is drafting.

### Fields

- `content`: Current accumulated assistant text.
- `status`: `empty`, `drafting`, `completed`, `failed`, or `stopped`.
- `messageId`: Final message id once completed.

### Validation rules

- Draft content may be empty while waiting or thinking.
- Completed draft becomes a normal assistant message exactly once.
- Failed or stopped draft does not become a successful assistant message.

## Execution Adapter

Represents the shared runtime boundary consumed by frontend UI.

### Capabilities

- Send user message.
- Create runtime run.
- Subscribe to runtime events.
- Append assistant delta.
- Complete runtime run.
- Fail runtime run.
- Stop runtime run.
- Report backend runtime capability.

### Validation rules

- Mock and real adapters must expose the same UI-visible state semantics.
- Real adapter may report capability unavailable before M4/M5, but must not silently substitute mock runtime behavior.

## Backend Capability State

Represents whether the current data source can provide runtime behavior.

### Values

- `available`: Runtime behavior can be executed or subscribed to.
- `unavailable`: Runtime behavior is not yet supported by the configured backend.

### Validation rules

- Mock mode reports available for deterministic scripts.
- Real API mode reports unavailable until run/event backend support exists.

## Stale Event Guard

Represents the rule that prevents old async events from affecting the wrong selected thread.

### Fields

- `requestedThreadId`: Thread that initiated the runtime action.
- `currentSelectedThreadId`: Thread selected when the event attempts to apply.
- `runId`: Runtime run that emitted the event.

### Validation rules

- Apply event only when `requestedThreadId === currentSelectedThreadId` and run id is still current for that thread.
- Ignore events from stopped or superseded runs.
