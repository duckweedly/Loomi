# Contract: Tool Catalog

## Endpoint

```http
GET /v1/tools/catalog
```

## Response

```json
{
  "tools": [
    {
      "name": "workspace.exec_command",
      "label": "Exec command",
      "group": "workspace",
      "capability": "exec",
      "approval_policy": "always_required",
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

## Safety

The response is read-only metadata. It must not include secrets, raw schemas, file contents, command output, or provider payloads.
