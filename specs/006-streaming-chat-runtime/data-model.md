# Data Model: Streaming Chat Runtime

## Thread

Represents a conversation container selected from the sidebar.

### Fields

- `id`: Stable thread identifier.
- `mode`: Product area the thread belongs to, with Chat mode in scope for this feature.
- `title`: User-visible thread label.
- `messages`: Ordered message history for the selected thread.
- `latestRun`: Most recent run associated with the selected thread when available.
- `loadState`: `idle`, `loading`, `loaded`, or `error`.

### Relationships

- A Thread has many Messages.
- A Thread has zero or more Runs.
- The selected Thread drives visible Chat Canvas, Composer, Timeline, and capability status.

### Validation rules

- Thread selection must not show messages or run state from a previously selected thread.
- A selected thread may have at most one pending or active run that blocks new composer send.
- Thread load failure must preserve selected-thread context and offer retry.

## Message

Represents persisted user or assistant text in the visible conversation.

### Fields

- `id`: Stable message identifier.
- `threadId`: Parent thread identifier.
- `role`: `user` or `assistant`.
- `content`: User-visible text.
- `createdAt`: User-visible creation timestamp or label.
- `runId`: Optional run that produced the assistant message.
- `attemptOfMessageId`: Optional previous assistant message identifier when this message is a regenerated attempt.

### Relationships

- A Message belongs to one Thread.
- An assistant Message may be produced by one Run.
- A regenerated assistant Message preserves the prior assistant Message and links conceptually to the same conversation context.

### Validation rules

- Successful assistant draft finalization creates or reveals final assistant content exactly once.
- Regenerate must not delete, overwrite, or hide the previous assistant response for this feature.
- Message history errors must not clear composer input.

## Assistant Draft Bubble

Represents assistant output before or during finalization.

### Fields

- `runId`: Run that owns the draft.
- `threadId`: Thread the draft belongs to.
- `content`: Accumulated assistant text.
- `status`: `pending`, `streaming`, `completed`, `failed`, `stopped`, or `recovering`.
- `messageId`: Final assistant message identifier after successful completion when available.
- `lastEventId`: Latest applied event identifier when available.

### Relationships

- An Assistant Draft Bubble belongs to one Run.
- A Run has zero or one current Assistant Draft Bubble.
- A completed Assistant Draft Bubble maps to one assistant Message.

### State transitions

```text
pending -> streaming -> completed
pending -> streaming -> failed
pending -> streaming -> stopped
pending -> recovering -> streaming
pending -> recovering -> completed
streaming -> recovering -> streaming
streaming -> recovering -> failed
```

### Validation rules

- Empty draft content is valid while pending or recovering.
- Failed and stopped drafts preserve any accumulated content.
- Stale, duplicate, or out-of-order deltas must not corrupt visible content.
- A final completion after user stop must not convert a stopped draft into a completed message.

## Assistant Attempt

Represents one assistant response attempt in a thread, including regenerated attempts.

### Fields

- `id`: Stable attempt identifier or run-derived identifier.
- `threadId`: Parent thread identifier.
- `runId`: Run that generated the attempt.
- `sourceMessageId`: User message that prompted the attempt.
- `previousAssistantMessageId`: Previous assistant response when this is a regenerated attempt.
- `status`: `pending`, `streaming`, `completed`, `failed`, `stopped`, or `recovering`.

### Relationships

- An Assistant Attempt belongs to one Thread.
- An Assistant Attempt is produced by one Run.
- A regenerated Assistant Attempt references the previous assistant Message but does not replace it.

### Validation rules

- Retry of a failed run may reuse the original user input context.
- Regenerate after completion creates a new Assistant Attempt while preserving the completed assistant Message.
- Only the selected thread's active Assistant Attempt controls visible Composer blocking.

## Run

Represents one assistant execution attempt.

### Fields

- `id`: Stable run identifier.
- `threadId`: Parent thread identifier.
- `status`: `pending`, `running`, `completed`, `failed`, `stopped`, `cancelled`, `retrying`, or `recovering`.
- `events`: Ordered run events.
- `assistantDraft`: Current assistant draft state when available.
- `createdAt`: Creation timestamp or label.
- `completedAt`: Completion timestamp or label when terminal.
- `capabilityStatus`: Current backend/runtime capability status relevant to this run when available.

### State transitions

```text
pending -> running -> completed
pending -> running -> failed
pending -> running -> stopped
pending -> running -> cancelled
pending -> running -> retrying -> running
pending -> recovering -> running
pending -> recovering -> completed
running -> recovering -> running
```

### Validation rules

- Completed, failed, stopped, and cancelled are terminal for a single run attempt.
- Retry or regenerate creates a new run attempt rather than mutating a terminal run into a new execution.
- Events from a non-selected or superseded run must not update the selected Chat Canvas.

## Run Event

Represents one observable execution signal.

### Fields

- `id`: Stable event identifier within or across a run.
- `runId`: Parent run identifier.
- `threadId`: Parent thread identifier.
- `type`: Canonical event type such as `run.created`, `model.delta`, `provider.error`, `worker.claimed`, or `run.cancelled`.
- `group`: `run-lifecycle`, `model-stream`, `worker-job`, or `error`.
- `severity`: `info`, `progress`, `warning`, or `error`.
- `label`: Concise user-visible label.
- `detail`: Optional user-visible detail.
- `time`: User-visible time label.
- `assistantDelta`: Optional text delta for model stream events.
- `usage`: Optional token or usage summary when available.

### Validation rules

- Error severity takes visual precedence when an event could fit multiple groups.
- Model delta events may update assistant draft content and timeline detail.
- Token usage events must be visible in timeline/debug without polluting the primary chat message text.
- Replayed events must be deduplicated before changing visible state.

## Timeline Event Group

Represents a semantic grouping for Run Events.

### Values

- `run-lifecycle`: Creation, start, active, completion, stop, cancellation, retry, recovery.
- `model-stream`: Model start, partial output, final output, token usage, model metadata.
- `worker-job`: Queue, worker claim, lease/ownership, retry scheduling, job state.
- `error`: Provider, backend, stream, model, validation, and execution errors.

### Validation rules

- Every displayed run event maps to exactly one primary group.
- Error group wins when an event has both lifecycle and error semantics.
- Groups may be empty, but visible group labels must stay stable across runs.

## Backend Capability Status

Represents user-readable execution capability and availability.

### Values

- `mock`: Deterministic local mock execution.
- `local-simulated`: Real API path with simulated local model/run behavior.
- `real-model`: Real provider/model execution is available.
- `backend-unavailable`: Backend cannot be reached or cannot provide runtime behavior.
- `model-setup-missing`: Required model setup is absent.
- `provider-unavailable`: Provider is unavailable or rejected execution.
- `stream-disconnected`: Event stream disconnected before terminal reconciliation.
- `run-recovering`: The UI is reconciling latest known run/draft state.

### Validation rules

- Capability status must not imply real model execution in mock or local simulated mode.
- Backend or stream availability problems must be visually distinct from model generation failure.
- Recovery status remains visible until the latest known run state is reconciled.

## Composer Input and Action State

Represents the user's pending input and available actions.

### Fields

- `text`: Current composer text.
- `canSend`: True when text is non-empty and no selected-thread run blocks send.
- `canStop`: True when the selected-thread run is active and stoppable.
- `canRetry`: True when the latest relevant run failed and original input context is recoverable.
- `canRegenerate`: True when a completed assistant response exists and no selected-thread run is active.
- `canContinue`: True when a thread is selected and composer input can be submitted as the next user turn.

### Validation rules

- Empty or whitespace-only text must not create a message or run.
- Active or pending selected-thread run blocks new send.
- Failed sends preserve recoverable user input.
- Stop makes composer available for the next valid action after the run reaches stopped state.

## Chat Canvas Presentation State

Represents the main chat area's visible state.

### Values

- `no-thread`
- `empty-thread`
- `loading`
- `history`
- `error`
- `backend-unavailable`
- `waiting-run`
- `streaming`
- `completed`
- `failed`
- `stopped`
- `recovering`

### Validation rules

- Loading and error states take precedence over history display when selected-thread data is not ready.
- Assistant draft presentation is scoped to the selected thread and latest relevant run.
- Terminal run states remain visible until the user continues, retries, regenerates, or selects another thread.
