# Contract: MCP Discovery and Tool Mapping

## Purpose

Run bounded MCP list-tools discovery for validated local stdio servers and map valid tool schemas into read-only Loomi ToolSpec candidates.

## Discovery Input

```json
{
  "server_slug": "local-search",
  "transport": "stdio",
  "timeout_ms": 5000
}
```

Sensitive command, args, and env references may be used by the local process launcher but must not appear in persisted events or normal debug summaries.

## Discovery Output

```json
{
  "server_slug": "local-search",
  "status": "succeeded",
  "tool_count": 1,
  "candidates": [
    {
      "mcp_tool_name": "search",
      "tool_spec_name": "mcp.local-search.search",
      "description_summary": "Search local index.",
      "input_schema_hash": "sha256:...",
      "execution_enabled": false,
      "approval_required_for_future_execution": true
    }
  ]
}
```

## Mapping Rules

- ToolSpec names must use `mcp.<server_slug>.<tool_name>`.
- Internal Loomi tool names must not be overridden.
- Duplicate MCP tool names from different servers remain distinct through namespacing.
- Duplicate names within the same server are rejected or deterministically marked invalid.
- Unsupported schema fields are either summarized safely or rejected with a redacted error.
- Tool descriptions and schemas are untrusted data, not instructions.

## Failure Output

```json
{
  "server_slug": "local-search",
  "status": "failed",
  "error_code": "mcp_discovery_timeout",
  "message": "MCP tool discovery timed out.",
  "retryable": true
}
```

Failure output must not include raw stderr, stdout, command, args, env, tokens, credentials, or secret-looking paths.
