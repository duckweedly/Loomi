# Contract: M29 Multi-agent Runtime Foundation

## Catalog

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
    "coordination_only": true,
    "autonomous_execution": false,
    "arguments": ["role", "goal"]
  }
}
```

`agent.list` and `agent.complete` use the same group and approval policy. `agent.list` is read-only; `agent.spawn` and `agent.complete` mutate coordination records only.

## Spawn Arguments

```json
{
  "role": "reviewer",
  "goal": "Review the artifact runtime implementation for safety gaps."
}
```

## Complete Arguments

```json
{
  "task_id": "agt_...",
  "result_summary": "No raw content leakage found."
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
  "goal": "Review the artifact runtime implementation for safety gaps.",
  "status": "spawned",
  "redaction_applied": false
}
```

## Rejections

- Chat mode RunContext filtering
- missing role/goal/task_id/result_summary
- unsupported role
- oversized role/goal/result
- unknown or out-of-thread task
- denied/stopped/terminal calls
- unsupported arguments
