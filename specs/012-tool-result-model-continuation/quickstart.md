# Quickstart: Tool Result Model Continuation

This is a planning quickstart for the later implementation. It assumes Window A has landed approve/deny and approved `runtime.get_current_time` execution.

## Preconditions

- Local API, web app, database, and worker run with the current M7 setup.
- Provider can be fake or controlled for deterministic tool-call then continuation behavior.
- `runtime.get_current_time` is the only executable tool.
- The tool result event contains redacted result fields.

## Success Smoke

1. Start a thread run where the provider first requests `runtime.get_current_time`.
2. Observe `tool_call_requested` and `tool_call_approval_required`.
3. Approve the tool through Window A's approve path.
4. Observe `tool_call_approved`, `tool_call_executing`, and `tool_call_succeeded`.
5. Verify the worker builds one continuation request containing:
   - prior conversation messages,
   - the original assistant tool-call metadata,
   - one synthetic redacted tool result for `tc_1`.
6. Verify continuation streams `model_delta` events after `tool_call_succeeded`.
7. Verify one final assistant message includes the time result.
8. Verify `run_completed` follows final assistant persistence.

## Denied Smoke

1. Start a run that requests the MVP tool.
2. Deny the tool.
3. Verify no `tool_call_executing`.
4. Verify no continuation provider request.
5. Verify Timeline shows denied terminal state and no assistant answer claims tool output.

## Tool Failure Smoke

1. Force approved tool execution to return a safe error.
2. Verify `tool_call_failed` contains only redacted error metadata.
3. Verify no continuation provider request.
4. Verify run ends failed exactly once.

## Continuation Failure Smoke

1. Let the tool succeed.
2. Force continuation provider failure after zero or more deltas.
3. Verify provider error is redacted.
4. Verify any partial second-phase draft remains visible as failed context.
5. Verify no final assistant message is created.

## Unsupported Second Tool Smoke

1. Let the continuation provider request another tool.
2. Verify Loomi records an unsupported-loop failure.
3. Verify no second approval card is opened and no second tool executes.

## SSE Replay Smoke

For success, denied, tool-failed, and continuation-failed paths:

1. Replay from `after_sequence=0`.
2. Reconnect after `tool_call_succeeded`.
3. Reconnect after second-phase `model_delta`.
4. Verify the rebuilt frontend state matches uninterrupted live streaming.

## Documentation Validation

Implementation must update and validate:

- `docs-site/src/content/docs/architecture/tool-result-continuation.md`
- `docs-site/src/content/docs/api/tool-call-approval.md`
- `docs-site/src/content/docs/runbooks/local-m7.md`
- `docs-site/src/content/docs/devlog/2026-05-25-tool-result-continuation.md`
- `docs-site/src/content/docs/roadmap/current-status.md`

Run `bun run build` from `docs-site/` when docs are updated.
