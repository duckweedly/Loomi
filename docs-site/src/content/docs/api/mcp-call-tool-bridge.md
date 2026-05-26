---
title: MCP Call Tool Bridge
description: M13 mcp.call_tool contract.
---

## Tool Name

| Tool | Approval | Capability | Risk | Side Effect | Safety |
| --- | --- | --- | --- | --- | --- |
| `mcp.call_tool` | Required | `call_tool` | `medium` | `mcp` | `mcp_bridge` |

## Arguments

```json
{
  "server": "local",
  "tool": "echo",
  "arguments": {
    "message": "hello mcp"
  }
}
```

Rules:

- `server` is required and must be `local`.
- `tool` is required and must be `echo`.
- `arguments` is required.
- `arguments.message` is required, trimmed, and capped at 500 characters.
- secret-looking strings are rejected before approval.
- unknown fields are rejected.

## Result

```json
{
  "server": "local",
  "tool": "echo",
  "message": "hello mcp",
  "side_effect": "none"
}
```

The result is stored in the existing tool-call `result_summary` projection and rendered by ToolCallCard.

## Catalog

`GET /v1/tools/catalog` includes `mcp.call_tool` after runtime tools and before workspace tools.
