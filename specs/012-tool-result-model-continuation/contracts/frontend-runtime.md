# Contract: Frontend Runtime Continuation

## Timeline

Timeline groups events in this visible order:

1. Initial model phase.
2. Tool request and approval-required state.
3. Approval decision and tool execution.
4. Tool result.
5. Continuation model phase.
6. Final assistant answer or terminal error.

Rules:

- Tool events remain in the tool-call group, not model text.
- Continuation `model_delta` rows remain model events but show phase-aware placement after the tool result.
- Redacted tool result appears in ToolCallCard and Timeline details.

## Assistant Draft

State transitions:

```text
empty
-> streaming(initial)
-> paused_for_tool(initial)
-> streaming(continuation)
-> final
```

Rules:

- Pre-tool text must not become a final assistant message when a tool call interrupts the initial phase.
- Continuation deltas may replace the visible draft or append after a phase separator in internal state, but the UI must not show duplicate assistant messages.
- `model_final` in continuation finalizes exactly one assistant message.
- `run_failed` after continuation error marks the draft failed and preserves partial text if present.

## ToolCallCard

The card must support:

- `approval_required`: approve/deny controls from Window A.
- `approved`: accepted and waiting to execute.
- `executing`: tool in progress.
- `succeeded`: redacted result visible.
- `denied`: terminal denial, no continuation.
- `failed`: redacted failure, no continuation.

## Replay

History replay must reconstruct the same:

- final assistant message,
- tool-call state,
- Timeline grouping,
- assistantDraft terminal state,
- run status.
