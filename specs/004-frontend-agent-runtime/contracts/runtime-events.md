# Contract: Runtime Events

Runtime events are the user-visible execution milestones that drive Run Timeline, Agent state motion, and Chat Canvas runtime states.

## Event vocabulary

| Event Type | Meaning | Timeline State | Agent Motion | Chat Canvas Effect |
|------------|---------|----------------|--------------|--------------------|
| `run.created` | A run was created for a user message | running | thinking | waiting-run |
| `context.loading` | Context is being collected or prepared | running | thinking | running |
| `assistant.thinking` | Assistant is planning the response | running | thinking | running |
| `assistant.drafting` | Assistant content is being drafted | running | speaking | running with draft |
| `assistant.message.completed` | Assistant response content is finalized | completed | speaking | final message ready |
| `run.completed` | Run finished successfully | completed | done | completed |
| `run.failed` | Run failed | failed | error | failed |
| `run.stopped` | Run was stopped by user or system | stopped | error | failed/stopped |

## Success script order

```text
run.created
context.loading
assistant.thinking
assistant.drafting
assistant.message.completed
run.completed
```

The success script appends exactly one final assistant message after `assistant.message.completed` and before or with `run.completed` becoming visible.

## Failure script order

```text
run.created
context.loading
assistant.thinking
run.failed
```

The failure script does not append a successful assistant message.

## Stopped script order

A stopped script can happen after `run.created` and before any terminal success/failure event.

```text
run.created
...
run.stopped
```

After `run.stopped`, no later event for that run is applied.

## Ordering rules

- Events must be applied in script/runtime order.
- Terminal events are final.
- Events must be ignored when their thread or run no longer matches the selected active runtime.
