---
title: 2026-05-26 M25 MCP + LSP Read-only
description: Development log for M25 MCP status and LSP read-only tools.
---

## Completed

- Added `GET /v1/mcp/servers` safe read-only MCP server status.
- Added saved local stdio MCP config, delete, and connection-test discovery endpoints.
- Added Settings > MCP status rendering and Settings > Tools LSP catalog visibility.
- Added builtin `lsp.diagnostics`, `lsp.symbols`, `lsp.references`, `lsp.definition`, and `lsp.hover`.
- Routed LSP tools through RunContext, approval, ToolBroker, worker execution, and provider continuation.
- Added Chat-mode rejection, unsafe path rejection, invalid argument rejection, stopped/denied no-exec tests, and HTTP approve-execute-final smoke.
- Added RunRail LSP lifecycle labeling.

## Validation

Focused validation passed for productdata, runtime, httpapi, and RunRail/Settings tests during implementation.

Full closeout still requires the complete M25 validation set:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Limits

LSP behavior is deterministic and lightweight. It does not launch language servers, install dependencies, call package managers, or expose host paths. Definition and hover are best-effort scans over bounded workspace text files until a later real language-server slice. MCP management supports local stdio config only; remote MCP/OAuth and marketplace install remain out of scope.
