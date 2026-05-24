# Contract: Composer Actions

## Purpose

Define send, stop, retry, regenerate, and continue behavior for the selected thread.

## Actions

| Action | Enabled when | Result |
|--------|--------------|--------|
| Send | Composer text is non-empty and selected thread has no pending/active run | User message appears and a new run starts or waits for runtime creation |
| Stop | Selected thread has an active stoppable run | Run transitions toward stopped; partial draft remains visible |
| Retry | Latest relevant run failed and original input context is recoverable | A new run attempt starts from the failed input context |
| Regenerate | A completed assistant response exists and no selected-thread run is active | Previous assistant response remains visible and a new assistant attempt is added |
| Continue | Thread is selected and text is non-empty with no active run | New user message continues the same thread |

## Keyboard rules

- Enter submits when Send or Continue is enabled.
- Shift+Enter inserts a newline.
- Empty or whitespace-only input never creates a message or run.

## Blocking rules

- Pending or active selected-thread run blocks Send and Continue.
- Pending or active selected-thread run blocks Retry and Regenerate.
- Actions in one thread must not affect another selected thread after switching.

## Failure preservation rules

- Failed sends preserve recoverable user input.
- Failed run context remains visible until retry, continue, regenerate, or thread selection changes.
- Retry does not erase the visible failure context until a new attempt is visible.

## Acceptance checks

- Empty submissions are rejected every time.
- Duplicate active-run submissions are blocked every time.
- Stop makes partial assistant content visible as stopped.
- Retry starts a new attempt without losing original user input context.
- Regenerate preserves the prior assistant response and adds a new assistant attempt.
