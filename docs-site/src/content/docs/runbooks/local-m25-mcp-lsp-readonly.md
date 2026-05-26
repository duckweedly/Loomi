---
title: Local MCP Management + LSP Read-only Validation
description: Commands for validating MCP config management and LSP read-only tools locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestValidateLSPToolCallArguments|TestToolCatalogIncludesLSPReadOnlyTools|TestWorkspaceAndLSPToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestLSPReadOnlyTools|TestGatewayRejectsLSPToolInChatMode|TestWorkerDoesNotExecuteLSPToolAfterStopOrDenied|TestWorkerExecutesApprovedLSPToolAndContinuesModel|TestToolBrokerExecutesLSPToolThroughOneEntrypoint'
go test ./internal/httpapi -run 'TestM25MCPServersHandlerReturnsSafeReadOnlyStatus|TestMCPServersHandlerSavesDiscoversAndDeletesConfig|TestM25LSPReadonlyApproveExecuteFinalSmoke'
bun test --cwd web ./src/components/SettingsView.mcp.test.tsx ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

## Full Closeout

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Manual Smoke

1. Start the API and web app.
2. Open Settings > MCP and confirm the flat local stdio form renders without a nested preview card.
3. Save a disabled MCP config and confirm it appears in the list without command, args, env, secrets, or host paths.
4. Click connection test and confirm the discovery status updates.
5. Delete the saved config and confirm it leaves the list.
6. Open Settings > Tools and confirm `lsp.diagnostics`, `lsp.symbols`, and `lsp.references` render as builtin LSP tools.
7. Run or seed a Work mode tool lifecycle and confirm RunRail labels LSP rows as low-risk, read-only, and workspace-scoped.

MCP management supports local stdio save, delete, and connection testing. It does not support remote MCP/OAuth, marketplace install, real language server process management, package-manager diagnostics, or shell-backed LSP discovery.
