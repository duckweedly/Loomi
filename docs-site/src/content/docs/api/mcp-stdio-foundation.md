---
title: M11 MCP Stdio API and Events
description: Safe MCP discovery metadata, ToolSpec candidate shape, and future execution boundary.
---

M11 does not add public MCP management endpoints. Server config is local explicit configuration, and only safe discovery/availability metadata can appear in run events and Timeline/debug.

## Local Config Shape

Internal local config uses this shape:

```json
{
  "slug": "local-search",
  "display_name": "Local Search",
  "enabled": true,
  "transport": "stdio",
  "timeout_ms": 5000
}
```

Raw command, args, and env references are intentionally omitted from API-facing examples.

## Discovery Result

Successful discovery produces safe candidate metadata:

```json
{
  "server_slug": "local-search",
  "status": "succeeded",
  "tool_count": 1,
  "candidate_names": ["mcp.local-search.search"],
  "candidate_schema_hashes": {
    "mcp.local-search.search": "sha256:..."
  },
  "candidates": [
    {
      "mcp_tool_name": "search",
      "tool_spec_name": "mcp.local-search.search",
      "input_schema_hash": "sha256:...",
      "execution_enabled": false,
      "approval_required_for_future_execution": true
    }
  ]
}
```

The M11.5 local stdio smoke verifies this shape with an enabled local fixture server and a discovered `mcp.local-smoke.echo` candidate. The fixture only answers discovery/list-tools and treats any `tools/call` request as a test failure.

Discovery uses MCP stdio `Content-Length` frames. M12 execution uses the same framing for `initialize`, `notifications/initialized`, and the single approved `tools/call`.

Failure output is redacted:

```json
{
  "server_slug": "local-search",
  "status": "failed",
  "error_code": "mcp_discovery_timeout",
  "message": "[redacted]",
  "retryable": true
}
```

## Run Event Metadata

MCP availability can appear on existing pipeline/debug events:

```json
{
  "type": "pipeline_step_completed",
  "category": "progress",
  "metadata": {
    "step": "prepare_context",
    "mcp_servers_configured": 1,
    "mcp_servers_enabled": 1,
    "mcp_servers_succeeded": 1,
    "mcp_server_summaries": [
      {
        "server_safe_id": "mcp:local-search",
        "server_slug": "local-search",
        "enabled": true,
        "discovery_status": "succeeded",
        "candidate_count": 1,
        "candidate_names": ["mcp.local-search.search"],
        "redacted_error_code": "",
        "last_discovered_at": "2026-05-25T10:00:00Z",
        "execution_enabled": false
      }
    ],
    "mcp_candidate_count": 1,
    "mcp_candidate_names": ["mcp.local-search.search"],
    "mcp_non_executable_candidate_names": ["mcp.local-search.search"],
    "mcp_execution_enabled": false,
    "mcp_error_codes": [],
    "mcp_last_discovered_at": "2026-05-25T10:00:00Z"
  }
}
```

Frontend mapping also recognizes explicit MCP event names:

| Backend type | Frontend type | Group |
| --- | --- | --- |
| `mcp_discovery_succeeded` | `mcp.discovery.succeeded` | worker-job |
| `mcp_discovery_failed` | `mcp.discovery.failed` | error |
| `mcp_discovery_rejected` | `mcp.discovery.rejected` | error |
| `mcp_tools_available` | `mcp.tools.available` | worker-job |
| `mcp_tools_non_executable` | `mcp.tools.non_executable` | worker-job |

## Redaction

The following must not appear in API responses, run events, Timeline/debug, or docs examples:

- raw command paths
- args
- env values
- raw stdout or stderr
- tokens, credentials, Authorization headers
- secret-looking paths
- raw MCP tool arguments or results

## Execution Boundary

M12 implements the first approval-gated execution bridge for already-discovered local stdio candidates. Execution still must go through M7 approval and audit with redacted arguments/results and scoped run/thread/user checks. Remote MCP, OAuth, marketplace installs, DB-managed server admin, sandboxing, automation tools, and multi-step loops remain out of scope.
