# Contract: MCP RunContext Observability

## Purpose

Expose MCP discovery and tool availability as safe RunContext and Timeline/debug metadata without enabling MCP execution.

## RunContext Summary Shape

```json
{
  "mcp": {
    "servers_configured": 1,
    "servers_enabled": 1,
    "servers_succeeded": 1,
    "servers_failed": 0,
    "server_summaries": [
      {
        "server_safe_id": "mcp:local-search",
        "server_slug": "local-search",
        "enabled": true,
        "discovery_status": "succeeded",
        "candidate_count": 1,
        "candidate_names": ["mcp.local-search.search"],
        "redacted_error_code": "",
        "last_discovered_at": "2026-05-25T00:00:00Z",
        "execution_enabled": false
      }
    ],
    "candidate_count": 1,
    "candidate_names": ["mcp.local-search.search"],
    "non_executable_candidate_names": ["mcp.local-search.search"],
    "execution_enabled": false,
    "error_codes": [],
    "last_discovered_at": "2026-05-25T00:00:00Z"
  }
}
```

## Timeline/debug Labels

Suggested user-facing labels:

- `mcp.discovery.succeeded`: "MCP discovery succeeded"
- `mcp.discovery.failed`: "MCP discovery failed"
- `mcp.discovery.rejected`: "MCP config rejected"
- `mcp.tools.available`: "MCP tools available"
- `mcp.tools.non_executable`: "MCP execution disabled"

## Replay Rules

- Live SSE and history replay must show the same safe discovery and availability labels when events are persisted.
- Missing MCP metadata must not crash Timeline/debug.
- Discovery failure does not fail unrelated runs unless a later spec introduces required MCP dependencies.
- Safety errors must be visible as redacted codes/messages.

## Forbidden Metadata

Timeline/debug, Background tasks, persisted run events, and safe summaries must not include:

- env values
- raw args
- raw command paths
- raw stderr or stdout
- tokens, credentials, Authorization headers
- private filesystem paths or secret-looking file names
- raw provider payloads
- raw MCP tool arguments/results
- shell output, file contents, browser or desktop captured state
