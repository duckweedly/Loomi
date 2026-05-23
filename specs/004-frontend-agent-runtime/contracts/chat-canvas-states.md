# Contract: Chat Canvas States

Chat Canvas states define what the main work area shows before, during, and after runtime execution.

## State priority

1. `loading`
2. `error`
3. `backend-unavailable`
4. `no-thread`
5. runtime terminal/active states
6. `empty-thread`
7. `history`

## States

### no-thread

Shown when no selected thread exists.

Visible outcome: concise Chinese empty state and a create-chat action.

### empty-thread

Shown when selected thread has no messages and no active runtime.

Visible outcome: concise Chinese prompt to send the first message.

### loading

Shown while selected thread data is loading.

Visible outcome: loading indicator; Composer disabled.

### error

Shown when thread/message/runtime loading or action fails unexpectedly.

Visible outcome: concise Chinese error summary and retry action when available.

### history

Shown when selected thread has messages and no active runtime state.

Visible outcome: normal message history and enabled Composer.

### waiting-run

Shown after a user message is visible but before first meaningful runtime progress.

Visible outcome: user message remains visible; runtime area indicates waiting.

### running

Shown while runtime events are active.

Visible outcome: message history plus optional assistant draft; Timeline and Agent motion update.

### completed

Shown after successful runtime completion.

Visible outcome: final assistant reply appears exactly once; Timeline completed; Agent motion done.

### failed

Shown after runtime failure or stop.

Visible outcome: failure state is visible; no fake success assistant reply appears.

### backend-unavailable

Shown when configured real data source cannot provide runtime behavior yet.

Visible outcome: concise Chinese explanation that backend run/event capability is not connected; no hidden mock execution.

## Cross-surface consistency

- Chat Canvas, Run Timeline, and Agent state motion must derive from the same selected runtime state.
- A surface must not show success while another shows failure for the same selected run.
- Stale events from old selections must not alter the current visible state.
