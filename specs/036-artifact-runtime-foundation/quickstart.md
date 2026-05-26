# Quickstart: M28 Artifact Runtime Foundation

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesArtifactRuntimeTools|TestArtifactToolsOnlyEnabledForWorkModeRunContext|TestValidateArtifactToolCallArguments'
go test ./internal/runtime -run 'TestArtifact'
go test ./internal/httpapi -run 'TestM28Artifact'
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
2. Open Settings > Tools and verify artifact tools show artifact scope, medium risk, approval required, and non-executable.
3. Open RunRail and verify artifact create/read/list lifecycle rows are visible without raw unbounded content.
