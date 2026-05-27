---
title: Local M28 Artifact Runtime Validation
description: Commands for validating the M28 artifact runtime slice locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesArtifactRuntimeTools|TestWorkspaceLSPWebBrowserAndArtifactToolsOnlyEnabledForWorkModeRunContext|TestValidateArtifactToolCallArguments|TestMemoryService.*Artifact'
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run TestPostgresArtifactsAndAgentTasksUseThreadScope -count=1 -v
go test ./internal/runtime -run 'TestArtifact|TestToolBrokerExecutesArtifactCreateThroughOneEntrypoint|TestWorkerExecutesApprovedArtifactCreateAndContinuesModel|TestWorkerDoesNotCreateArtifactAfterStopOrDenied'
go test ./internal/httpapi -run 'TestM28Artifact'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts
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
2. Open Settings > Tools and confirm `artifact.create_text`, `artifact.read`, and `artifact.list` render as builtin, artifact-scoped, approval-required, medium risk, non-executable, and backed by real API storage when PostgreSQL is configured.
3. Open RunRail and confirm artifact lifecycle rows are visible.
4. Confirm the UI shows title/type/size/excerpt metadata without raw unbounded content, executable controls, credentials, or local paths.

M28 does not support binary artifacts, downloads, previews, iframe execution, filesystem export, browser integration, shell integration, artifact version graphs, marketplace packaging, or multi-agent orchestration.
