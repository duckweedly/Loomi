# Contract: Run Event and SSE Ordering

## Success Path

```text
run_created
job_queued
job_claimed
model_started                  metadata.model_phase=initial
model_delta*                   metadata.model_phase=initial
tool_call_requested
tool_call_approval_required
# Window A: user approves
tool_call_approved
tool_call_executing
tool_call_succeeded            metadata.result_summary/result_for_model_redacted
model_started                  metadata.model_phase=continuation
model_delta*                   metadata.model_phase=continuation
model_final                    metadata.model_phase=continuation
assistant_message_created
run_completed
```

Rules:

- All events are persisted before SSE delivery.
- Existing history-first SSE remains the only stream contract.
- `tool_call_succeeded` must appear before continuation `model_started`.
- `assistant_message_created` may be represented by the existing final message event if that is the current contract, but replay must produce exactly one final assistant message.
- Event metadata should distinguish `initial` and `continuation` phases when the frontend needs draft reconstruction.

## Denied Path

```text
tool_call_requested
tool_call_approval_required
tool_call_denied
run_stopped or run_completed_with_denial
```

Rules:

- No continuation `model_started` after denial.
- No final assistant message that claims tool output.
- Timeline must display denial as the terminal explanation.

## Tool Failed Path

```text
tool_call_approved
tool_call_executing
tool_call_failed
run_failed
```

Rules:

- No continuation provider call after `tool_call_failed`.
- Failure metadata must be redacted.

## Continuation Failed Path

```text
tool_call_succeeded
model_started                  metadata.model_phase=continuation
model_delta*                   metadata.model_phase=continuation
model_error or provider_error
run_failed
```

Rules:

- Partial continuation draft may remain visible as failed terminal context.
- No duplicate final assistant message.

## SSE Replay Requirements

- A client reconnecting after any event sequence must rebuild the same Timeline and assistantDraft/final message state as a live stream.
- Dedupe continues to use event id and sequence.
- The second model phase uses the same `run_event` SSE envelope as all other persisted events.
