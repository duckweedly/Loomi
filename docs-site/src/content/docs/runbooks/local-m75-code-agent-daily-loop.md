---
title: Local M75 Code-Agent Daily Loop Validation
description: Local reproduction steps for the Work-mode code task loop.
---

## Focused Smoke

```bash
go test ./internal/httpapi ./internal/runtime -run 'Test.*Workspace|Test.*Sandbox|Test.*Bounded|Test.*CodeAgent' -count=1
bun test --cwd web ./src/components/RunRail.runtime.test.ts ./src/components/ToolCallCard.test.tsx
```

Expected evidence:

1. The HTTP smoke provider requests `workspace.grep`, `workspace.read`, `workspace.patch_preview`, `workspace.patch_apply`, and `sandbox.exec_command`.
2. `workspace.grep` and `workspace.read` execute without user approval.
3. `workspace.patch_preview`, `workspace.patch_apply`, and `sandbox.exec_command` block on approval before executing.
4. The patch preview does not mutate the file.
5. Patch apply mutates only after approval and only while the preview is fresh.
6. Sandbox exec runs only after approval and returns bounded result metadata.
7. Terminal runs reject pending mutation approval and do not resume execution.
8. The final timeline contains continuation events and `run_completed`, followed by the assistant summary.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
```

M75 is complete only when the focused smoke, full Go test suite, web tests, web build, and docs build pass locally.

## M76 Continuation Reliability Smoke

Focused command:

```bash
go test ./internal/runtime ./internal/httpapi -run 'Test.*Continuation|Test.*Bounded|Test.*CodeAgent|Test.*Tool' -count=1
```

Expected evidence:

1. The M76 smoke provider requests at least six tools in one run: `workspace.grep`, `workspace.read`, `workspace.patch_preview`, `workspace.patch_apply`, `sandbox.exec_command`, `workspace.read`, then a final assistant message.
2. Each `tool_call_id` has exactly one terminal tool event.
3. Only the current blocked tool remains pending for approval; older approved/executing/succeeded tools do not stay pending.
4. Provider continuation requests carry tool result messages in durable event order, growing from the first result to all six results.
5. The assistant message for the run is persisted once.
6. After `run_completed`, late model output and late tool continuation requests are rejected.

## M98 Parallel Tool And Retry Smoke

Focused command:

```bash
go test ./internal/runtime -run 'TestHTTPProviderFlushesMultipleOpenAIToolCalls|TestGatewayRecordsMultipleProviderToolCallsInOneTurn|TestQueuedRunRouterExecutesParallelAutoApprovedToolsBeforeContinuation|TestGatewayRetriesTransientProviderFailureBeforeOutput|TestGatewayDoesNotRetryProviderFailureAfterVisibleOutput' -count=1
```

Expected evidence:

1. OpenAI-compatible streamed tool calls with multiple indexes emit every tool call, not only the first one.
2. Gateway records multiple auto-approved read-only tool calls from one provider turn.
3. Queued runner executes the ready auto-approved batch before starting continuation.
4. Continuation receives all completed batch tool-call/result pairs in durable event order.
5. Retryable provider failure before visible output schedules a retry and can recover.
6. Provider failure after visible output is not retried, preventing duplicate output or tool state.
