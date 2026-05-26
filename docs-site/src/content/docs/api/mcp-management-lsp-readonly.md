---
title: MCP Management + LSP Read-only API
description: Safe MCP server status API and LSP tool contract.
---

## GET /v1/mcp/servers

Returns local MCP server status projections.

```json
{
  "servers": [
    {
      "server_safe_id": "mcp_local-smoke",
      "server_slug": "local-smoke",
      "display_name": "Local Smoke",
      "transport": "stdio",
      "enabled": true,
      "config_source": "env",
      "discovery_status": "succeeded",
      "candidate_count": 1,
      "candidate_names": ["mcp.local-smoke.echo"],
      "execution_mode": "approval_gated",
      "redacted_error_code": "",
      "last_discovered_at": "2026-05-26T00:00:00Z"
    }
  ],
  "request_id": "req_..."
}
```

The response must not include raw `command`, `args`, env values, secrets, or host paths.

## LSP Tools

Catalog names:

- `lsp.diagnostics`
- `lsp.symbols`
- `lsp.references`

Common arguments:

- `path`: required workspace-relative file path
- `limit`: optional bounded result count
- `language`: optional metadata hint

`lsp.symbols` also accepts `query`.

`lsp.references` requires `line` and `column`.

Result summaries use:

```json
{
  "tool": "lsp.symbols",
  "scope": "lsp",
  "operation": "symbols",
  "path": "src/main.go",
  "count": 1,
  "truncated": false,
  "redaction_applied": false
}
```

All LSP tools are builtin, low risk, Work mode only, and approval-required.
