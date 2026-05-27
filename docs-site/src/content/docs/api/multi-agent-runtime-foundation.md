---
title: Multi-agent Runtime Tool API
description: Catalog, arguments, and result contract for M29 agent coordination tools.
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

`agent.list` uses `read_only = true`. `agent.complete` uses `read_only = false`. All three keep `coordination_only = true` and `autonomous_execution = false`.

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

`agent.complete`:

```json
{
  "task_id": "agt_...",
  "result_summary": "No safety issue found"
}
```

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

`agent.list` returns `tasks` and `count`. `agent.complete` returns the task status and bounded result summary.

Events and continuation context persist safe summaries only. They do not include raw provider payloads, credentials, local paths, child model runs, external process handles, or cross-thread delegation data.

## Read-only HTTP Projection

The CLI can inspect coordination records through:

- `GET /v1/threads/:thread_id/agent-tasks?limit=20`

Responses include bounded safe task fields: task id, thread/run ids, role, goal, status, result summary, and timestamps. This endpoint is read-only. Creating or completing tasks remains an approval-gated Work-mode tool flow through `agent.spawn` and `agent.complete`; Loomi still does not launch child model runs or external workers in this slice.
