---
title: MCP Management + LSP Read-only Foundation
description: Settings MCP config management and read-only LSP tool architecture.
---

This slice adds two bounded tool surfaces without changing the existing approval model:

- `/v1/mcp/servers` projects local stdio MCP config plus discovery events into safe status rows, and accepts saved local stdio config writes.
- `lsp.diagnostics`, `lsp.symbols`, `lsp.references`, `lsp.definition`, and `lsp.hover` are builtin low-risk tools for Work mode only.

## Boundaries

MCP management supports local stdio config save, delete, and discovery. It does not support remote MCP, OAuth, marketplace install, or background daemon management.

The API never returns command, args, env, token values, absolute host paths, or raw config. It only exposes server slug/display name, transport, enabled flag, config source, latest discovery status, candidate count/names, execution mode, redacted error code, and discovery timestamp.

Saved config lives in `mcp_server_configs` keyed by local user and slug. Env config from `LOOMI_MCP_SERVERS_JSON` remains supported; saved config with the same slug overrides the env projection. The worker MCP executor uses a config loader so saved updates are visible to later approved MCP executions, not just to the Settings page.

LSP tools reuse the existing `ToolCatalog -> RunContext -> approval -> ToolBroker -> worker continuation` path. They are always approval-required, workspace-scoped, and read-only. Chat mode filters them out with workspace and sandbox tools.

## Execution

`LSPToolExecutor` uses the same workspace scope guard as workspace tools:

- rejects absolute paths, traversal, sensitive paths, and symlink escape
- limits file reads and result counts
- returns workspace-relative paths only
- records safe summaries under `scope = lsp`

The first implementation is deterministic and local. It does not start or manage real language servers, does not install dependencies, and does not run package managers. Symbols, references, definition, hover, and diagnostics are intentionally lightweight until a later LSP daemon slice.

## Visibility

Settings > MCP displays a flat local stdio config form, save/delete actions, connection testing, and safe discovered tool rows. Settings > Tools shows LSP catalog entries as builtin `lsp` scope. RunRail labels LSP lifecycle rows as low-risk, read-only, and workspace-scoped.
