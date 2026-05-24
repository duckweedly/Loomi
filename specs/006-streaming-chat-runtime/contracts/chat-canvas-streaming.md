# Contract: Chat Canvas Streaming

## Purpose

Define how the Chat Canvas presents user messages, assistant draft bubbles, terminal draft states, and finalized assistant messages for the selected thread.

## Inputs

- Selected thread, or no selected thread.
- Ordered messages for the selected thread.
- Latest selected-thread run when available.
- Assistant draft associated with the latest selected-thread run when available.
- Stream state and backend capability status.

## Visible states

| Condition | Chat Canvas output |
|-----------|--------------------|
| No selected thread | No-thread empty prompt; no stale messages |
| Selected thread has no messages and no run | Empty-thread prompt |
| Message history loading | Loading state |
| Message history failed | Error state with retry path |
| User message accepted and run pending | User message plus pending assistant bubble |
| Assistant deltas arriving | Assistant bubble grows in place |
| Run completed | Assistant bubble becomes completed assistant content exactly once |
| Run failed | Failed assistant bubble preserves partial content and offers retry affordance |
| Run stopped | Stopped assistant bubble preserves partial content |
| Run recovering | Recovering assistant bubble/status until latest run state is reconciled |
| Backend unavailable | Backend status is explicit and not described as model thinking |

## Draft update rules

- Append only deltas that belong to the selected thread and current run.
- Ignore duplicate or stale events that have already been applied.
- Ignore events from a previous selected thread after the user switches threads.
- Preserve partial draft text on failure, stop, and recovery.
- Do not append token usage or provider metadata into the assistant message body.

## Finalization rules

- A successful final signal creates or reveals one completed assistant message.
- The visible draft must not duplicate the final assistant message.
- A stop, failure, or cancellation terminal state prevents later final signals from converting that same run into completed content.
- Regenerate preserves the previous assistant response and adds a new assistant attempt.

## Acceptance checks

- Streaming success shows one user message, one growing assistant bubble, and one final assistant response.
- Streaming failure shows preserved partial content and a failed state.
- Stop shows preserved partial content and a stopped state.
- Recovery restores the latest known draft content and status for the selected thread.
- Thread switching never applies old run output to the newly selected thread.
