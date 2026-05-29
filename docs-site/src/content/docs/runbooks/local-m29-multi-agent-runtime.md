---
title: Local Multi-agent Runtime Validation
description: Commands for validating the coordination-only multi-agent runtime locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestValidateAgentToolCallArguments|TestMemoryServiceAgentTaskLifecycle|TestMemoryServiceDelegateAgentTaskCreatesChildRun|TestMemoryServiceReconcilesDelegatedAgentTaskAfterChildRunCompletes|TestRepositoryContractDelegateAgentTaskRetryIsIdempotent|TestToolCatalogIncludesAgentRuntimeTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext'
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run 'TestPostgresArtifactsAndAgentTasksUseThreadScope|TestPostgresReconcilesDelegatedAgentTaskAfterChildRunCompletes' -count=1 -v
go test ./internal/runtime -run 'TestAgent|TestToolBrokerExecutesAgentSpawnThroughOneEntrypoint|TestWorkerExecutesApprovedAgentSpawnAndContinuesModel|TestWorkerWaitsForDelegatedChildRunBeforeParentContinuation|TestWorkerDoesNotSpawnAgentTaskAfterStopOrDenied|TestGatewayRejectsAgentToolInChatMode'
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/runtime -run 'TestPostgresWorkerWaitsForDelegatedChildRunBeforeParentContinuation|TestPostgresAgentDelegateChildWorkerTerminalResumesParent' -count=1 -v
go test ./internal/httpapi -run 'TestM29Agent|TestM98CodeAgentParallelReadDelegateChildFinalSmoke'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts
```

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > Tools and confirm `agent.spawn`, `agent.list`, `agent.start`, `agent.complete`, and `agent.fail` render as builtin, agent-scoped, approval-required, medium risk, coordination-only, no autonomous execution, and backed by real API storage when PostgreSQL is configured. Confirm `agent.delegate` is approval-required and marked as autonomous child-run handoff.
3. Open RunRail and confirm agent lifecycle rows are visible.
4. Confirm task statuses can move `spawned -> in_progress -> completed` or `spawned/in_progress -> failed`, and terminal tasks cannot be restarted, delegated, or overwritten.
5. Confirm `agent.delegate` creates a child Work thread, a queued child model-gateway run, and records `child_thread_id`, `child_run_id`, and `parent_tool_call_id` on the parent task. Repeating the same parent tool-call id must return the existing child ids; a different parent tool-call id must not create a duplicate.
6. Confirm the parent run does not continue while the delegated child run is still active, then resumes only after the child run becomes terminal and reconciliation writes a bounded child result summary.
7. Confirm mixed code-agent runs can request multiple read-only workspace tools in one provider turn, receive all matching results together in continuation, then move through `agent.spawn -> agent.delegate -> child run -> parent final` without continuing the parent before the child is terminal.
8. Confirm the UI/CLI shows task id, role, goal, status, result summary, optional child ids, optional parent tool-call id, and safe timestamps without raw provider payloads, credentials, local paths, child execution logs, or external process ids.

The current multi-agent runtime supports approval-gated child run handoff, but still does not support external worker pools, OS process spawning, shell execution, Docker/Firecracker isolation, remote guest agents, long-term multi-agent memory, marketplace packaging, or background swarm orchestration.
