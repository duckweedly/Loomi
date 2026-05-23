# Contract: Execution Adapter

The execution adapter is the frontend runtime boundary shared by mock and future real runtime sources. It is a product contract, not a backend API contract for M3.5.

## Capability state

```text
runtimeCapability: available | unavailable
```

- Mock adapter reports `available`.
- Real adapter reports `unavailable` until backend run/event support exists.
- UI must display backend capability unavailable when a user attempts runtime execution while capability is unavailable.

## Operations

### sendMessage

Records or forwards the user's submitted message for a selected thread.

**Input**

- `threadId`
- `content`

**Output**

- User message visible in Chat Canvas.
- Initial runtime request state or backend capability unavailable state.

### createRun

Creates a runtime run for the submitted message.

**Input**

- `threadId`
- `messageId`
- `scriptId` for mock mode when selecting success/failure scripts.

**Output**

- Runtime run with status `pending` or `running`.
- First event `run.created`.

### subscribeRunEvents

Streams or emits ordered runtime events for a run.

**Input**

- `threadId`
- `runId`

**Output**

- Ordered Runtime Event values.
- Completion callback or terminal event.

### appendAssistantDelta

Applies assistant draft content to the active run.

**Input**

- `threadId`
- `runId`
- `delta`

**Output**

- Updated assistant draft for Chat Canvas.

### completeRun

Marks run successful and finalizes assistant content.

**Input**

- `threadId`
- `runId`
- final assistant content

**Output**

- Final assistant message.
- Terminal runtime status `completed`.
- Timeline event `run.completed`.

### failRun

Marks run failed.

**Input**

- `threadId`
- `runId`
- user-visible failure reason

**Output**

- Terminal runtime status `failed`.
- Timeline event `run.failed`.
- No successful assistant message is appended.

### stopRun

Stops a pending or running runtime run.

**Input**

- `threadId`
- `runId`

**Output**

- Terminal runtime status `stopped`.
- Timeline event `run.stopped`.
- Later script/runtime events are ignored.

## Contract rules

- All events include `threadId` and `runId` so stale-event guards can reject events for non-selected or superseded runs.
- Mock and real adapters expose the same visible event vocabulary even if transport differs later.
- Real adapter must not use mock execution as a hidden fallback when capability is unavailable.
