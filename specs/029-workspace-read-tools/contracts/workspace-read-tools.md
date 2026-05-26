# Contract: Workspace Read Tools

## Catalog

`GET /v1/tools/catalog` includes three executable read-only workspace entries:

```json
{
  "name": "workspace.read",
  "source": "builtin",
  "group": "workspace",
  "risk_level": "low",
  "approval_policy": "always_required",
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

## Approval Contract

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

Expected pre-approval result:

```text
tool_call_requested -> tool_call_approval_required
```

No filesystem content may be read before approval.

Expected post-approval result:

```text
tool_call_approved -> tool_call_executing -> tool_call_succeeded -> model continuation
```

## Failure Contract

Rejected paths return a failed tool execution after approval:

```json
{
  "error_code": "workspace_access_denied",
  "message": "Workspace path is outside the allowed scope."
}
```

Failure events include only safe path metadata and denial code. Sensitive file contents are never included.
