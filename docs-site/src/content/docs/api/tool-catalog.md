---
title: Tool Catalog
description: M11 read-only tool catalog API contract.
---

## Endpoint

```http
GET /v1/tools/catalog
```

Returns deterministic metadata for current allowlisted tools.

## Response

```json
{
  "tools": [
    {
      "name": "workspace.exec_command",
      "label": "Exec command",
      "group": "workspace",
      "capability": "exec",
      "approval_policy": "required",
      "safety_class": "workspace_exec",
      "risk_level": "high",
      "side_effect": "process",
      "enabled": true,
      "description": "Run one bounded argv command inside the workspace."
    }
  ],
  "updated_at": "2026-05-26T00:00:00Z"
}
```

## Included Tools

| Tool | Capability | Risk | Side Effect |
| --- | --- | --- | --- |
| `runtime.get_current_time` | `time` | `low` | `none` |
| `workspace.glob` | `read` | `medium` | `read` |
| `workspace.grep` | `read` | `medium` | `read` |
| `workspace.read_file` | `read` | `medium` | `read` |
| `workspace.write_file` | `write` | `high` | `write` |
| `workspace.edit` | `write` | `high` | `write` |
| `workspace.exec_command` | `exec` | `high` | `process` |

All entries currently use `approval_policy: "required"`.

## Redaction Contract

The endpoint is read-only and must not return secrets, provider credentials, raw provider payloads, file content, command output, or executable command examples.
