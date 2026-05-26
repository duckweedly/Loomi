---
title: M25 MCP Management + LSP Read-only Foundation
description: Settings MCP status and read-only LSP tool architecture.
---

M25 adds two bounded tool surfaces without changing the existing approval model:

- `GET /v1/mcp/servers` projects local stdio MCP config plus discovery events into safe read-only status rows.
- `lsp.diagnostics`, `lsp.symbols`, and `lsp.references` are builtin low-risk tools for Work mode only.

## Boundaries

MCP management is read-only. The API never returns command, args, env, token values, absolute host paths, or raw config. It only exposes server slug/display name, transport, enabled flag, config source, latest discovery status, candidate count/names, execution mode, redacted error code, and discovery timestamp.

LSP tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. They are always approval-required, workspace-scoped, and read-only. Chat mode filters them out with workspace and sandbox tools.

## Execution

`LSPToolExecutor` uses the same workspace scope guard as workspace tools:

- rejects absolute paths, traversal, sensitive paths, and symlink escape
- limits file reads and result counts
- returns workspace-relative paths only
- records safe summaries under `scope = lsp`

The first implementation is deterministic and local. It does not start or manage real language servers, does not install dependencies, and does not run package managers. Symbols/references/diagnostics are intentionally lightweight until a later LSP daemon slice.

## Visibility

Settings > MCP displays server status. Settings > Tools shows LSP catalog entries as builtin `lsp` scope. RunRail labels LSP lifecycle rows as low-risk, read-only, and workspace-scoped.
