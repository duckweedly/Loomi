---
title: M11 MCP Stdio Foundation
description: Local stdio MCP discovery, read-only ToolSpec candidates, and safe availability observability.
---

M11 adds the first MCP foundation without opening tool execution. Loomi accepts explicit local stdio MCP server config, performs bounded discovery/list-tools, maps discovered schemas into namespaced read-only tool candidates, and exposes safe availability in RunContext and Timeline/debug.

This slice does not add MCP HTTP/SSE/OAuth, remote MCP, marketplace install, shell/filesystem/browser automation, sandboxing, or automatic MCP tool execution.

## Config Boundary

MCP server config is local and explicit for this slice. It may include a command, args, env values, and timeout for discovery, but normal product data and Timeline/debug must only show safe summaries:

- server slug and display name
- enabled/disabled state
- `stdio` transport
- timeout bucket
- whether args/env exist
- local config source
- execution mode: disabled

Command paths, raw args, env values, stderr, tokens, credentials, and secret-looking paths are sensitive.

## Discovery Boundary

Discovery uses the MCP list-tools shape and treats all server output as untrusted data. Tool names and schemas must pass validation before becoming candidates.

The current implementation supports:

- local stdio config validation
- discovery/list-tools response parsing
- disabled-config handling
- redacted failure summaries
- namespaced candidate mapping

The stdio runner is bounded by timeout and sends initialize plus tools/list messages. If discovery cannot complete safely, it records a redacted failure instead of leaking process output.

## Tool Candidate Mapping

Discovered MCP tools become read-only ToolSpec candidates:

```text
mcp.<server_slug>.<tool_name>
```

Examples:

- `mcp.local-search.search`
- `mcp.notes.lookup`

Candidates cannot override internal tools such as `runtime.get_current_time`. Duplicate names within one server are rejected. Duplicate names across servers remain distinct through namespacing.

Every MCP candidate is non-executable in M11:

- source: `mcp`
- execution state: disabled
- future approval policy: always required
- auto-execute: false

## Persona Integration

Persona allowed-tool names may reference discovered MCP candidates, but this only affects availability summaries. A persona entry such as `mcp.local-search.search` resolves as `discovered_non_executable`.

Persona config does not grant execution permission and does not replace user approval.

## RunContext and Timeline

RunContext can carry a safe MCP availability summary:

- configured/enabled/succeeded/failed server counts
- per-server safe id/slug, discovery status, candidate count, and timestamp
- safe namespaced candidate names
- non-executable candidate names
- execution enabled flag
- redacted error codes
- latest discovery timestamp

Timeline/debug maps MCP discovery and tool availability events into worker/job style rows:

- `mcp.discovery.succeeded`
- `mcp.discovery.failed`
- `mcp.discovery.rejected`
- `mcp.tools.available`
- `mcp.tools.non_executable`

Errors are grouped with error events. Missing MCP metadata must not crash replay.

## Future Execution Boundary

A later spec must define execution before any MCP tool invocation is allowed. That future path must reuse the M7 approval lifecycle:

- approval-required tool call projection
- scoped thread/run/user checks
- persisted audit events
- redacted argument summaries and hashes
- redacted result/error summaries
- worker ownership and cancellation guards
- no automatic execution from model output, persona config, or discovery metadata
