# Quickstart: M25 MCP Management + LSP Read-only Foundation

## Focused Checks

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesLSPReadOnlyTools|TestLSPToolsOnlyEnabledForWorkModeRunContext'
go test ./internal/runtime -run 'TestLSP'
go test ./internal/httpapi -run 'TestM25'
bun test --cwd web ./src/components/SettingsView.mcp.test.tsx ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
```

Expected evidence:

1. `GET /v1/mcp/servers` returns safe local MCP server status without commands, args, env, secrets, or absolute private paths.
2. Settings > MCP renders configured, empty, failed, and succeeded states from real or mock status data.
3. Tool catalog includes `lsp.diagnostics`, `lsp.symbols`, and `lsp.references` as read-only, LSP-scoped, approval-required, executable builtin tools.
4. Work mode can enable LSP tools; Chat mode rejects them.
5. Approved LSP execution returns bounded, workspace-relative diagnostics/symbols/references and provider continuation.
6. RunRail shows LSP lifecycle rows without secret or host path leakage.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

Browser smoke must verify Settings > MCP, Settings > Tools LSP entries, RunRail LSP lifecycle visibility, and zero console errors.
