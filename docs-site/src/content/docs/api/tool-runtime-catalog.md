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

The discovery slice adds:

```json
{
  "name": "tool.load_tools",
  "display_name": "Load tools",
  "source": "builtin",
  "group": "discovery",
  "risk_level": "low",
  "approval_policy": "read_only",
  "execution_state": "executable",
  "safe_metadata": {
    "arguments": ["queries", "names", "limit"],
    "read_only": true,
    "scope": "runtime_catalog",
    "dynamic_schema_loader": false
  }
}
```

`skill.load_skill` uses the same read-only approval policy and returns installed skill summaries by `name`; it does not return full instruction bodies.

## Discovery Tool Results

`tool.load_tools` accepts optional `queries`, `names`, and `limit`, then returns safe descriptions for currently enabled tools only:

```json
{
  "operation": "load_tools",
  "scope": "runtime_catalog",
  "tools": [
    {
      "name": "workspace.read",
      "display_name": "Workspace read",
      "description": "Read a bounded UTF-8 text slice from one workspace file.",
      "group": "workspace",
      "risk_level": "low",
      "approval_policy": "always_required"
    }
  ],
  "dynamic_schema_loader": false
}
```

`skill.load_skill` accepts `name` and optional `limit`, then returns safe manifest fields such as `id`, `name`, `description`, `source`, `source_label`, `package`, and `installed`. It returns `instruction_loaded=false`.

## Event Metadata

Tool lifecycle events may include `tool_call_id`, `tool_name`, `tool_source`, `tool_group`, `candidate_schema_hash`, `arguments_summary`, `approval_status`, `execution_status`, `result_summary`, `error_code`, and `error_message`.

All metadata is redacted before persistence. API responses must not include secrets, raw tool args, raw tool results, MCP command/env/stderr, provider traces, or credential paths.

## Write Boundary

M18 does not add install, edit, enable, disable, approval policy override, dynamic provider schema injection, skill-body loading, or MCP management endpoints.
