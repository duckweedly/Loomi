# Contract: MCP Server Config

## Purpose

Define the only accepted source of MCP server configuration for M11: explicit local stdio configuration.

## Config Shape

```json
{
  "slug": "local-search",
  "display_name": "Local Search MCP",
  "enabled": true,
  "transport": "stdio",
  "command_ref": "configured-local-command",
  "args_ref": "sensitive-args-reference",
  "env_ref": "sensitive-env-reference",
  "timeout_ms": 5000
}
```

## Validation Rules

- `slug`, `display_name`, `transport`, and `timeout_ms` are required.
- `transport` must be `stdio`.
- HTTP URLs, SSE endpoints, OAuth configuration, remote network endpoints, plugin marketplace references, and auto-install instructions are rejected.
- Commands, args, env values, tokens, credentials, and absolute private paths are sensitive.
- Safe summaries may include `slug`, `display_name`, enabled state, transport type, timeout bucket, and validation status.

## Rejected Config Result

```json
{
  "server_slug": "local-search",
  "status": "rejected",
  "error_code": "mcp_config_unsupported_transport",
  "message": "Only local stdio MCP servers are supported in this slice.",
  "retryable": false
}
```

The result must not include raw command, args, env, tokens, credentials, or secret-looking paths.
