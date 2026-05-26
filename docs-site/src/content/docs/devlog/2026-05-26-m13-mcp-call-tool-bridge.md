---
title: 2026-05-26 M13 MCP Call Tool Bridge
description: mcp.call_tool implementation notes.
---

## Completed

- Added Spec Kit artifacts under `specs/015-mcp-call-tool-bridge/`.
- Added allowlisted `mcp.call_tool`.
- Kept MCP calls approval-required through the existing tool lifecycle.
- Added product data validation for fixed `local.echo`, bounded message input, unknown field rejection, and secret-looking message rejection.
- Added runtime normalization and execution for `local.echo`.
- Added worker execution coverage for approved MCP call tools.
- Added catalog metadata with safety class `mcp_bridge`.
- Added ToolCallCard coverage for nested MCP arguments and result summaries.
- Added mock tool catalog entry for Settings > Tools visibility.
- Fixed seeded mock id allocation so browser smoke does not produce duplicate React message keys.

## Validation

Focused validation:

```bash
go test ./internal/productdata ./internal/runtime ./internal/httpapi
bun test --cwd web ./src/components/ToolCallCard.test.tsx
```

Full validation:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

Browser smoke:

- Open Settings > Tools.
- Confirm `mcp.call_tool`, `call_tool`, and `mcp_bridge` render.
- Checked browser console errors: `0`.
- Capture `m13-mcp-call-tool-smoke.png`.

## Known Limits

- The bridge only supports the built-in `local.echo` MCP-style tool.
- M13 does not add external MCP server processes, transports, tool discovery, browser automation, spawn-agent, LSP, RAG, memory, auto-approval, or multi-agent delegation.
