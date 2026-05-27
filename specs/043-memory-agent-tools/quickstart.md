# Quickstart: Memory Agent Tools

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesMemoryTools|TestMemoryToolCallValidation'
go test ./internal/runtime -run 'TestMemoryToolExecutor|TestWorkerContinuesAfterApprovedMemoryTool'
go test ./internal/httpapi -run TestToolsCatalog
bun test --cwd web src/components/SettingsView.tools.test.tsx
bun run --cwd docs-site build
```

Browser smoke:

1. Open Settings > Tools.
2. Verify memory tools appear as builtin approval-gated memory tools.
3. Start a run whose provider requests a memory tool.
4. Approve the tool call.
5. Verify RunRail shows request/approval/execution/success and no raw memory content.
