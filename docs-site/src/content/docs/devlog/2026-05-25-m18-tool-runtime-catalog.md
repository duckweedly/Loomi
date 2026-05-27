---
title: 2026-05-25 M18 Tool Runtime Catalog
description: Devlog for the M18 Tool Runtime + Tool Catalog foundation.
---

M18 added the first unified tool catalog and broker boundary. `runtime.get_current_time` and local stdio MCP execution now share a `ToolInvocation`/`ToolResult` path and broker policy checks before concrete execution.

## Completed

- Added safe tool catalog entries for builtin and discovered MCP tools.
- Added read-only `GET /v1/tools/catalog`.
- Added broker validation for approval, execution status, catalog membership, persona allowlist, enabled state, and MCP schema hash.
- Routed approved builtin and MCP worker resume through the broker.
- Added Settings > Tools read-only catalog rendering.
- Preserved existing M7/M12 event lifecycle and continuation behavior.
- Adjusted MCP catalog projection to keep the latest successful schema hash and avoid claiming executor availability in Settings/API when the worker executor cannot be proven.
- Adjusted broker execution to derive the execution catalog from the current prepared RunContext, preventing stale historical discovery hashes from blocking a valid current run.

## Later Discovery Update

- Added builtin `tool.load_tools` and `skill.load_skill` discovery helpers.
- Kept both helpers low-risk, read-only, and auto-approved.
- `tool.load_tools` returns safe descriptions for tools already enabled in the current run by exact name or keyword.
- `skill.load_skill` returns safe installed skill manifest summaries and explicitly does not return full skill instruction bodies.
- Provider schemas still come from the run enabled-tool snapshot; this update is not yet true dynamic schema injection.

## Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Focused discovery validation:

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateDiscovery|TestToolCatalogIncludesDiscovery|TestDiscovery|TestToolDefinitionsForPersona|TestGatewayExposesCodeAgentToolsToProvider' -count=1
go test ./internal/httpapi -run TestDiscoveryLoadToolsAutoApprovedSmoke -count=1
```

## Known Limits

The catalog is computed, not user-configurable. Settings > Tools has no write controls. Future workspace/artifact/sandbox/web/browser tools must add their own catalog entries and executor implementations through the broker.
