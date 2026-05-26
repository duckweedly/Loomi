---
title: Tool Runtime Catalog API
description: Read-only safe catalog API and tool event metadata for M18.
---

## `GET /v1/tools/catalog`

Returns the safe catalog available to the local Loomi runtime.

```json
{
  "tools": [
    {
      "name": "runtime.get_current_time",
      "display_name": "Current time",
      "description": "Returns the current UTC time.",
      "source": "builtin",
      "group": "runtime",
      "input_schema_hash": "sha256:...",
      "risk_level": "low",
      "approval_policy": "always_required",
      "enabled": true,
      "execution_state": "executable",
      "safe_metadata": {
        "arguments": ["timezone"]
      }
    }
  ],
  "request_id": "req_..."
}
```

MCP candidates use namespaced names such as `mcp.local-smoke.echo` and carry the discovery candidate schema hash when available.

The read-only catalog uses the latest successful discovery metadata for each MCP tool name. Because the API cannot prove that the current worker has the matching stdio executor configured, MCP catalog entries are shown as `non_executable` until a run-specific RunContext enables them for broker execution.

M21 also adds builtin workspace entries for `workspace.glob`, `workspace.grep`, and `workspace.read`. They appear under `group=workspace` with `safe_metadata.scope=workspace` and `safe_metadata.read_only=true`; the catalog does not expose the host absolute workspace root.

## Event Metadata

Tool lifecycle events may include `tool_call_id`, `tool_name`, `tool_source`, `tool_group`, `candidate_schema_hash`, `arguments_summary`, `approval_status`, `execution_status`, `result_summary`, `error_code`, and `error_message`.

All metadata is redacted before persistence. API responses must not include secrets, raw tool args, raw tool results, MCP command/env/stderr, provider traces, or credential paths.

## Write Boundary

M18 does not add install, edit, enable, disable, approval policy override, or MCP management endpoints.
