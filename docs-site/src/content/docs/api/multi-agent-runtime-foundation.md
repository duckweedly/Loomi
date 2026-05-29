---
title: Multi-agent Runtime Tool API
description: Catalog, arguments, lifecycle, and result contract for agent task and child-run handoff tools.
---

## Catalog Entries

`GET /v1/tools/catalog` includes:

```json
{
  "name": "agent.spawn",
  "source": "builtin",
  "group": "agent",
  "risk_level": "medium",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "agent",
    "read_only": false,
    "coordination_only": true,
    "autonomous_execution": false,
    "arguments": ["role", "goal"]
  }
}
```

`agent.list` uses `read_only = true`. `agent.start`, `agent.complete`, and `agent.fail` use `read_only = false` with `coordination_only = true` and `autonomous_execution = false`.

M31 adds `agent.delegate` as the first child-run handoff tool. It is still Work-mode only and approval required, but its safe metadata is different:

```json
{
  "name": "agent.delegate",
  "safe_metadata": {
    "scope": "agent",
    "read_only": false,
    "coordination_only": false,
    "autonomous_execution": true,
    "arguments": ["task_id"]
  }
}
```

## Arguments

`agent.spawn`:

```json
{
  "role": "reviewer",
  "goal": "Review implementation"
}
```

Supported roles are `researcher`, `implementer`, and `reviewer`.

`agent.list`:

```json
{
  "limit": 20
}
```

`agent.start`:

```json
{
  "task_id": "agt_..."
}
```

`agent.delegate`:

```json
{
  "task_id": "agt_..."
}
```

`agent.complete`:

```json
{
  "task_id": "agt_...",
  "result_summary": "No safety issue found"
}
```

`agent.fail`:

```json
{
  "task_id": "agt_...",
  "result_summary": "Blocked by missing context"
}
```

## Lifecycle

Agent tasks now use a bounded durable lifecycle:

| Status | Meaning |
| --- | --- |
| `spawned` | Task record exists and has not started. |
| `in_progress` | Approved `agent.start` marked the task active. |
| `completed` | Approved `agent.complete` closed the task successfully. |
| `failed` | Approved `agent.fail` closed the task with a safe failure summary. |

`agent.start` only moves `spawned -> in_progress`. `agent.delegate` creates a new Work-mode child thread, writes one bounded child user message from the task goal, starts a queued child model-gateway run through the existing background job path, marks the parent task `in_progress`, and records the parent tool-call id. Replaying `agent.delegate` with the same `parent_tool_call_id` returns the existing child ids instead of creating a duplicate child run. The parent `agent.delegate` tool call stays `executing` until the child run reaches `completed`, `failed`, `cancelled`, or `stopped`. `agent.complete` and `agent.fail` only close `spawned` or `in_progress` tasks. Terminal tasks cannot be restarted, delegated, or overwritten.

## Result Summary

```json
{
  "tool": "agent.spawn",
  "scope": "agent",
  "operation": "spawn",
  "task_id": "agt_...",
  "role": "reviewer",
  "goal": "Review implementation",
  "status": "spawned",
  "autonomous_execution": false,
  "redaction_applied": false
}
```

`agent.list` returns `tasks` and `count`. `agent.start`, `agent.complete`, and `agent.fail` return the task status and bounded result summary.

`agent.delegate` first returns the same safe task summary plus `child_thread_id`, `child_run_id`, `parent_tool_call_id`, and `delegated_at` to persisted task state. The parent run also records an `agent_child_run_started` progress event with the child ids, task id, and parent tool-call id so the timeline can show the handoff before the child finishes. That first result is not passed to parent model continuation. After child terminal reconciliation, the parent tool result includes `child_status`, child ids, `parent_tool_call_id`, and a bounded assistant-message summary from the child run. The parent continuation job preserves the parent run's workspace-root snapshot from the original run job, so delegated child completion does not silently unbind later workspace tools. It does not include child prompt text, raw provider payloads, credentials, local paths, child tool logs, stdout/stderr, or external process handles.

Events and continuation context persist safe summaries only. Parent run continuation can see the child ids, parent tool-call id, child terminal status, task status, and bounded child result summary, but not the child run message history or tool timeline.

If the parent run is stopped while a delegated child run is still non-terminal, the child run is stopped, its active queued/leased/retrying job is cancelled, unresolved child tool calls are marked `cancelled`, and the delegated task moves to `failed` with a safe parent-stop result summary. No child prompt, child history, or tool logs are copied into the parent response.

## Read-only HTTP Projection

The CLI can inspect coordination records through:

- `GET /v1/threads/:thread_id/agent-tasks?limit=20`

Responses include bounded safe task fields: task id, parent thread/run ids, role, goal, status, result summary, optional `child_thread_id`, optional `child_run_id`, optional `parent_tool_call_id`, optional `delegated_at`, and timestamps. This endpoint is read-only. Creating, starting, delegating, completing, and failing tasks remains an approval-gated Work-mode tool flow.

M31 child-run handoff is not an external worker pool, swarm runtime, OS process, shell, Docker/Firecracker sandbox, or remote agent service. Child runs use the existing Loomi thread/run/job/approval boundaries.
