---
title: Local M28 Artifact Runtime Validation
description: Commands for validating the M28 artifact runtime slice locally.
---

## Focused Validation

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesArtifactRuntimeTools|TestWorkspaceLSPWebBrowserAndArtifactToolsOnlyEnabledForWorkModeRunContext|TestValidateArtifactToolCallArguments|TestMemoryService.*Artifact|TestWorkToolResolutionsFollowLatestUserIntent'
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run TestPostgresArtifactsAndAgentTasksUseThreadScope -count=1 -v
go test ./internal/runtime -run 'TestArtifact|TestToolBrokerExecutesArtifactCreateThroughOneEntrypoint|TestWorkerExecutesApprovedArtifactCreateAndContinuesModel|TestWorkerDoesNotCreateArtifactAfterStopOrDenied'
go test ./internal/httpapi -run 'TestM28Artifact'
bun test --cwd web SettingsView.tools.test.tsx RunRail.runtime.test.ts runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts runtime/artifactPreview.test.ts runtime/messageArtifactPreview.test.ts components/RightToolDrawer.preview.test.tsx
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
2. Open Settings > Tools and confirm `artifact.create_text`, `artifact.create_visual`, `artifact.read`, and `artifact.list` render as builtin, artifact-scoped, approval-required, medium risk, and backed by real API storage when PostgreSQL is configured.
3. Open RunRail and confirm artifact lifecycle rows are visible.
4. Ask Work mode to draw an SVG diagram. Confirm Loomi requests `artifact.create_visual`, the completed turn shows a visual artifact card, and the Preview drawer renders it inside the sandbox frame instead of dumping XML/CSS into chat text.
5. Drag the Preview drawer separator and confirm the right drawer width persists inside the desktop shell.
6. In the desktop shell, use the artifact titlebar action to open the bounded artifact body with the OS default local app.
7. Confirm the UI shows title/type/size/excerpt metadata without raw unbounded content, executable controls, credentials, or local paths.

M28 does not support binary artifacts, downloads, agent-side filesystem export, browser integration, shell execution, artifact version graphs, marketplace packaging, or multi-agent orchestration. Visual HTML/SVG preview is bounded to Loomi's sandboxed Preview frame; the desktop-only native-open action writes a bounded temporary copy for the user's local application.
