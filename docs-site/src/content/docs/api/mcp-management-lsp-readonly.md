---
title: MCP Management + LSP Read-only API
description: Local MCP server config management API and LSP tool contract.
---

## GET /v1/mcp/servers

Returns local MCP server status projections from env config plus saved user config.

```json
{
  "servers": [
    {
      "server_safe_id": "mcp:local-smoke",
      "server_slug": "local-smoke",
      "display_name": "Local Smoke",
      "transport": "stdio",
      "enabled": true,
      "config_source": "local",
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

## POST /v1/mcp/servers

Saves or updates a local stdio MCP server config for the current local identity.

```json
{
  "slug": "local-search",
  "display_name": "Local Search",
  "enabled": true,
  "transport": "stdio",
  "command": "/path/to/mcp-server",
  "args": ["--profile", "default"],
  "env": { "MODE": "local" },
  "timeout_ms": 5000
}
```

The response returns a safe `server` projection only. Raw command, args, and env are stored for execution/discovery but are never echoed back.

Validation rejects missing slug/display name, invalid slugs, unsupported transports, enabled configs without a command, and remote `http`, `https`, `ws`, or `wss` endpoints.

## POST /v1/mcp/servers/:slug/discover

Runs a bounded local stdio `tools/list` discovery for the saved or env-backed config. The result is appended as safe discovery metadata and returned as a safe server status row.

## DELETE /v1/mcp/servers/:slug

Deletes the saved user config for `:slug` and returns the remaining safe server list. Env-backed config with the same slug is not removed.

## LSP Tools

Catalog names:

- `lsp.diagnostics`
- `lsp.symbols`
- `lsp.references`
- `lsp.definition`
- `lsp.hover`

Common arguments:

- `path`: required workspace-relative file path
- `limit`: optional bounded result count
- `language`: optional metadata hint

`lsp.symbols` also accepts `query`.

`lsp.references`, `lsp.definition`, and `lsp.hover` require `line` and `column`.

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

`lsp.definition` returns bounded best-effort definition items. `lsp.hover` returns a bounded hover object for the symbol at the requested position. All LSP tools are builtin, low risk, Work mode only, and approval-required.
