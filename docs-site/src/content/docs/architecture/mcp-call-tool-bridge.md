---
title: MCP Call Tool Bridge
description: M13 approval-gated minimal MCP-style tool bridge.
---

M13 adds `mcp.call_tool`, a minimal bridge for invoking one allowlisted local MCP-style tool through the existing tool-call lifecycle.

## Boundary

The bridge only supports:

- `server`: `local`
- `tool`: `echo`
- `arguments.message`: trimmed string, 1-500 characters

It does not start external MCP servers, connect sockets, discover installed tools, pass arbitrary JSON-RPC payloads, or call workspace/process APIs.

## Flow

1. Provider emits `mcp.call_tool`.
2. Product data validates the fixed server/tool allowlist, required approval state, message bounds, and secret-looking content.
3. Loomi records `tool.call.requested` and `tool.call.approval_required`.
4. User approves or denies through the existing approval UI.
5. Worker resolves the tool definition and normalizes arguments again.
6. Runtime executes `local.echo`.
7. Runtime stores a bounded result summary with `server`, `tool`, `message`, and `side_effect`.

## Safety

`mcp.call_tool` remains approval-required and is cataloged as:

- group: `mcp`
- capability: `call_tool`
- safety class: `mcp_bridge`
- risk: `medium`
- side effect: `mcp`

Both product data and runtime reject unknown fields, unknown server/tool values, empty messages, messages over 500 characters, and secret-looking strings such as tokens or `sk-` credentials.

## Non-Goals

M13 does not add external MCP server configuration, stdio transport, tool discovery, browser automation, spawn-agent, LSP, RAG, memory, auto-approval, or multi-agent delegation.
