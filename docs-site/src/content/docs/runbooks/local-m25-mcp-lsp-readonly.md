---
title: Local M25 MCP + LSP Read-only Validation
description: Commands for validating M25 MCP status and LSP read-only tools locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestValidateLSPToolCallArguments|TestToolCatalogIncludesLSPReadOnlyTools|TestWorkspaceAndLSPToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestLSPReadOnlyTools|TestGatewayRejectsLSPToolInChatMode|TestWorkerDoesNotExecuteLSPToolAfterStopOrDenied|TestWorkerExecutesApprovedLSPToolAndContinuesModel|TestToolBrokerExecutesLSPToolThroughOneEntrypoint'
go test ./internal/httpapi -run 'TestM25MCPServersHandlerReturnsSafeReadOnlyStatus|TestM25LSPReadonlyApproveExecuteFinalSmoke'
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
2. Open Settings > MCP and confirm configured local stdio servers render without command, args, env, secrets, or host paths.
3. Open Settings > Tools and confirm `lsp.diagnostics`, `lsp.symbols`, and `lsp.references` render as builtin LSP tools.
4. Run or seed a Work mode tool lifecycle and confirm RunRail labels LSP rows as low-risk, read-only, and workspace-scoped.

M25 does not support editing MCP config, remote MCP/OAuth, marketplace install, real language server process management, package-manager diagnostics, or shell-backed LSP discovery.
