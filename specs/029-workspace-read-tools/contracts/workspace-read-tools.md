# Contract: Workspace Read Tools

## Catalog

`GET /v1/tools/catalog` includes three executable read-only workspace entries:

```json
{
  "name": "workspace.read",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "low",
  "approval_policy": "read_only",
  "enabled": true,
  "execution_state": "executable",
  "safe_metadata": {
    "read_only": true,
    "scope": "workspace",
    "arguments": ["path", "offset", "limit", "max_bytes"]
  }
}
```

The response must not expose the host absolute root.

## Read-Only Execution Contract

Provider tool call metadata:

```json
{
  "tool_call_id": "tc_workspace_read",
  "arguments_summary": {
    "path": "internal/runtime/tools.go",
    "offset": 0,
    "limit": 2048
  }
}
```

Expected read-only result:

```text
tool_call_requested -> tool_call_approved -> tool_call_executing -> tool_call_succeeded -> model continuation
```

`workspace.glob`, `workspace.grep`, and `workspace.read` are auto-approved bounded reads after the user has selected a workspace root. Workspace mutation tools remain approval-gated.

## Failure Contract

Rejected paths return a failed tool execution:

```json
{
  "error_code": "workspace_access_denied",
  "message": "Workspace path is outside the allowed scope."
}
```

Failure events include only safe path metadata and denial code. Sensitive file contents are never included.
