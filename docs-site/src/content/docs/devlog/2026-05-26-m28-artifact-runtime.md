---
title: 2026-05-26 M28 Artifact Runtime Foundation
description: Implementation notes and validation for M28 artifact tools.
---

## Completed

- Added Spec Kit feature `specs/036-artifact-runtime-foundation/`.
- Added builtin `artifact.create_text`, `artifact.create_visual`, `artifact.read`, and `artifact.list` catalog identity, default persona allowlist, Work-mode filtering, and safe tool-call metadata grouping.
- Added productdata text artifact records, in-memory service methods, PostgreSQL `artifacts` migration, and PostgresRepository create/read/list methods.
- Added `ArtifactToolExecutor` for non-executable text artifact creation, bounded reads, and safe lists.
- Routed artifact tools through ToolBroker, worker approved-tool resume, provider continuation, and HTTP smoke coverage.
- Added PG/in-memory alignment coverage for create/read/list and cross-thread no-leak behavior.
- Updated Settings > Tools, RunRail labels, mock catalog, seeded run data, and runtime scripts for visible artifact lifecycle metadata.
- Added bounded visual artifact support for SVG/HTML diagrams through `artifact.create_visual`; visual content is persisted as artifact type `visual`, returned as renderable `artifacts[]` content, and previewed only in the sandboxed UI frame.

## Validation

Focused validation during implementation:

```bash
go test ./internal/productdata -run 'TestToolCatalogIncludesArtifactRuntimeTools|TestWorkspaceLSPWebBrowserAndArtifactToolsOnlyEnabledForWorkModeRunContext|TestValidateArtifactToolCallArguments|TestMemoryService.*Artifact'
go test ./internal/runtime -run 'TestArtifact|TestToolBrokerExecutesArtifactCreateThroughOneEntrypoint|TestWorkerExecutesApprovedArtifactCreateAndContinuesModel|TestWorkerDoesNotCreateArtifactAfterStopOrDenied'
go test ./internal/httpapi -run 'TestM28Artifact'
bun test --cwd web runtimeScripts.test.ts mockExecutionAdapter.test.ts mockApiClient.test.ts SettingsView.tools.test.tsx RunRail.runtime.test.ts
```

Full closeout commands should also run before marking M28 complete:

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Non-goals

No binary artifacts, downloads, filesystem export, browser integration, shell integration, artifact version graph, marketplace packaging, or multi-agent orchestration were added. Visual SVG/HTML preview is limited to bounded artifact content in the sandboxed Preview frame.
