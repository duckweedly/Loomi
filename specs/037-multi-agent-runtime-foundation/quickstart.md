# Quickstart: M29 Multi-agent Runtime Foundation

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesAgentRuntimeTools|TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext|TestValidateAgentToolCallArguments|TestMemoryService.*AgentTask'
go test ./internal/runtime -run 'TestAgent'
go test ./internal/httpapi -run 'TestM29Agent'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts
```

## Full Validation

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke

1. Start the web dev server.
2. Open Settings > Tools and verify agent tools show agent scope, medium risk, approval required, coordination-only, and autonomous execution disabled.
3. Open RunRail and verify agent spawn/list/complete lifecycle rows are visible without raw prompt or secret leakage.
