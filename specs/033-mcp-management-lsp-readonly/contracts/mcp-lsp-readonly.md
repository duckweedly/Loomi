# Contract: MCP Management + LSP Read-only Foundation

## MCP Status API

```http
GET /v1/mcp/servers
```

Response:

```json
{
  "servers": [
    {
      "server_safe_id": "mcp:local-search",
      "server_slug": "local-search",
      "display_name": "Local Search",
      "transport": "stdio",
      "enabled": true,
      "config_source": "local",
      "discovery_status": "succeeded",
      "candidate_count": 1,
      "candidate_names": ["mcp.local-search.search"],
      "execution_mode": "approval_gated",
      "redacted_error_code": "",
      "last_discovered_at": "2026-05-26T12:00:00Z"
    }
  ]
}
```

Rules:

- The response is read-only status; no write endpoint is added in M25.
- Do not expose command, args, env, raw payloads, absolute host roots, or secrets.
- Empty config returns `servers: []`.

## LSP Tool Catalog Entries

```json
{
  "name": "lsp.symbols",
  "source": "builtin",
  "group": "lsp",
  "risk_level": "low",
  "approval_policy": "always_required",
  "execution_state": "executable",
  "safe_metadata": {
    "scope": "lsp",
    "read_only": true,
    "arguments": ["path", "query", "line", "column", "include_declaration", "language", "limit"]
  }
}
```

The same contract shape applies to `lsp.diagnostics` and `lsp.references`.

## LSP Tool Requests

`lsp.symbols`:

```json
{
  "path": "internal/runtime/tools.go",
  "query": "Tool",
  "limit": 20
}
```

`lsp.references`:

```json
{
  "path": "internal/runtime/tools.go",
  "line": 12,
  "column": 4,
  "include_declaration": false,
  "limit": 20
}
```

`lsp.diagnostics`:

```json
{
  "path": "internal/runtime/tools.go",
  "limit": 20
}
```

Rules:

- `path` is required, relative, and workspace-scoped.
- `query` is optional for symbols, ignored for diagnostics and references.
- `line` and `column` are required for references and must be positive integers.
- `include_declaration` is optional for references.
- `limit` is optional and executor-bounded.
- LSP tools reject traversal, absolute, sensitive, symlink-escape, and unsupported paths before reading.

## LSP Result Summary

```json
{
  "tool": "lsp.symbols",
  "scope": "lsp",
  "operation": "symbols",
  "path": "internal/runtime/tools.go",
  "items": [
    {
      "name": "ToolBroker",
      "kind": "type",
      "path": "internal/runtime/tools.go",
      "line": 12,
      "preview": "type ToolBroker struct"
    }
  ],
  "count": 1,
  "truncated": false,
  "redaction_applied": false
}
```

`lsp.references` results use the same envelope with `operation: "references"` and a `position` object:

```json
{
  "tool": "lsp.references",
  "scope": "lsp",
  "operation": "references",
  "path": "internal/runtime/tools.go",
  "position": { "line": 12, "column": 4 },
  "items": [
    {
      "path": "internal/runtime/tool_broker.go",
      "line": 22,
      "column": 8,
      "preview": "ToolBroker"
    }
  ],
  "count": 1,
  "truncated": false,
  "redaction_applied": false
}
```

Failures use the existing `tool_call_failed` event path. Validation failures must happen before file reads outside the workspace boundary.
