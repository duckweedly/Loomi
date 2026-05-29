---
title: M31 Child Agent Run Handoff
description: Approval-gated child thread and child run creation for existing agent tasks.
---

M31 closes the next Arkloop/Craft parity gap after M29/M30 task lifecycle work: Loomi can now turn a durable agent task into a real child run without introducing an external worker pool or sandbox service.

Implemented:

- Added `agent.delegate` as a Work-mode, approval-required builtin agent tool.
- Added `child_thread_id`, `child_run_id`, `parent_tool_call_id`, and `delegated_at` to `agent_tasks`.
- `DelegateAgentTask` creates a new Work child thread, writes one bounded child user message from the task goal, starts a normal queued model-gateway child run through the existing background job pipeline, and links the child ids plus parent tool-call id back to the parent task.
- Approved `agent.delegate` leaves the parent tool call in `executing`, so the parent model does not continue until the child run is terminal.
- Parent runs now record an immediate `agent_child_run_started` event with safe child ids and parent tool-call id, so the timeline shows the handoff while the parent tool remains `executing`.
- Stopping a parent run now stops any non-terminal delegated child run, cancels its child job, and marks the delegated task failed instead of leaving an orphan child run queued.
- Retrying `agent.delegate` with the same `parent_tool_call_id` returns the existing child ids, so worker lease recovery does not fail an already-created child handoff.
- Worker reconciliation completes the parent delegate tool call after the child run finishes, stores a bounded child result summary, updates the task to `completed` or `failed`, and queues the parent continuation job.
- `agent.delegate` exposes only safe task status plus child ids, parent tool-call id, child terminal status, and bounded result summary; it does not return the child prompt, provider payload, credentials, paths, tool logs, stdout/stderr, or process handles.
- HTTP/CLI task projections include optional child ids and parent tool-call id.

Boundaries:

- Chat mode still filters agent tools out.
- `agent.delegate` must go through ToolBroker approval.
- Delegation is limited to the current thread task id.
- Terminal tasks and already delegated tasks cannot be delegated again.
- Child runs use their own RunContext, persona/tool allowlist, approval boundary, and worker job.
- Parent continuation receives only the reconciled safe child result after terminal child status.
- This is not a swarm scheduler, remote guest agent, external process, Docker/Firecracker sandbox, Redis queue rewrite, or autonomous worker pool.

Validation:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi ./internal/cli -run 'Test.*Agent|TestToolCatalogIncludesAgentRuntimeTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext|TestValidateAgentToolCallArguments|TestRendererPrintAgentTasks' -count=1
go test ./internal/runtime -run 'TestWorkerWaitsForDelegatedChildRunBeforeParentContinuation|TestWorkerExecutesApprovedAgentSpawnAndContinuesModel|TestWorkerDoesNotSpawnAgentTaskAfterStopOrDenied' -count=1
go test ./internal/productdata -run 'TestMemoryServiceDelegateAgentTaskCreatesChildRun|TestMemoryServiceReconcilesDelegatedAgentTaskAfterChildRunCompletes' -count=1
go test ./internal/productdata -run 'TestMemoryServiceDelegateAgentTaskCreatesChildRun|TestRepositoryContractDelegateAgentTaskRetryIsIdempotent|TestPostgresDelegateAgentTaskRetryIsIdempotent' -count=1 -v
```
