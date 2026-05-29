---
title: Multi-agent Runtime Foundation
description: Approval-gated agent task runtime and child-run handoff architecture.
---

M29 added the first multi-agent runtime slice as three builtin tools: `agent.spawn`, `agent.list`, and `agent.complete`. M30 extended the coordination boundary with `agent.start` and `agent.fail`. M31 adds the first approval-gated child-run handoff with `agent.delegate`.

All agent tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. `agent.spawn`, `agent.list`, `agent.start`, `agent.complete`, and `agent.fail` only create or update bounded task records. `agent.delegate` creates a separate child Work thread and queued child model-gateway run, then keeps the parent tool call executing until the child run reaches a terminal state.

Work-mode intent detection keeps agent tools available when the latest user request asks for delegation, subagents, child agents, multi-agent work, or parallel child review. This lets a single code-agent run combine read-only workspace tools and approval-gated agent handoff in the same prepared tool snapshot.

## Boundaries

Agent tools are:

- builtin
- Work mode only
- approval required
- medium risk
- coordination-only for task record tools
- autonomous child-run handoff only for `agent.delegate`

Chat mode filters them out with workspace, sandbox, LSP, web, browser, and artifact tools. `agent.spawn` creates one task record with a supported role and bounded goal through `AgentTaskService`, backed by both in-memory service and PostgreSQL `agent_tasks` in the real API path. `agent.list` returns bounded summaries for the current thread. `agent.start` marks a spawned current-thread task `in_progress`. `agent.delegate` creates a new child thread, adds one child user message derived from the task goal, starts a queued child run via the normal `StartRun` background job path, and records `child_thread_id`, `child_run_id`, `parent_tool_call_id`, and `delegated_at` on the parent task. `agent.complete` and `agent.fail` close a spawned or in-progress task with a bounded result summary.

The durable lifecycle is intentionally small:

- `spawned`
- `in_progress`
- `completed`
- `failed`

Terminal states are not restartable, delegateable, or overwriteable. Already delegated tasks cannot be delegated a second time from a different parent tool call. Replaying the same `parent_tool_call_id` returns the existing child ids, which keeps worker lease recovery and approved-tool retry paths idempotent.

## Execution

`AgentToolExecutor` depends on `productdata.AgentTaskService`. Worker approved-tool resume injects the configured productdata service or repository. Most agent tools record tool success immediately and continue the provider with a safe result summary. `agent.delegate` is different: the approved tool starts the child run, records `agent_child_run_started` on the parent run with child ids and the parent tool-call id, leaves the parent tool call in `executing`, and returns control to the worker without continuing the parent provider.

On later worker ticks, `ReconcileAgentTaskChildRuns` checks delegated tasks with terminal child runs. Reconciliation completes the parent `agent.delegate` tool call with a bounded child result summary, updates the task to `completed` or `failed`, queues the parent run, and resumes parent continuation through the normal background job path.

Stopping a parent run also stops any non-terminal delegated child run created from that parent, cancels its queued/leased/retrying child job, cancels unresolved child tool calls, and marks the delegated task failed with a safe parent-stop summary. This prevents orphan child runs or hidden child tool approvals from continuing after the user has stopped the parent task.

Run events persist role, goal, status, task id, result summary, source thread, source run, and for `agent.delegate` only the child thread/run ids and parent tool-call id. The initial handoff event is visible immediately, but parent continuation still waits for terminal child reconciliation and never receives the child run message history, child tool timeline, raw provider payloads, credentials, local paths, stdout/stderr, or external process ids.

M31 still does not introduce Arkloop-style sandbox service, worker pool, remote guest agent, swarm scheduler, Redis queue rewrite, Docker/Firecracker isolation, or background OS process. A child run is a normal Loomi run in a separate thread with its own RunContext, tool catalog, persona allowlist, approval boundary, and worker job.

## Visibility

Settings > Tools shows agent task tools as agent-scoped, approval-required, medium risk, and coordination-only. `agent.delegate` is agent-scoped and approval-required but marked as autonomous child-run handoff. RunRail labels agent lifecycle rows separately from workspace, sandbox, LSP, web fetch, browser, artifact, MCP, and runtime tools.
