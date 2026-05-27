# Implementation Plan: M23 Workspace Mutation Tools

**Branch**: `[031-workspace-mutation-tools]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/031-workspace-mutation-tools/spec.md`

## Summary

M23 adds the first approval-gated workspace mutation tools after M21 read tools and M22 bounded loops. The slice introduces `workspace.write_file` for new bounded text files and `workspace.edit` for exact bounded replacement, reusing the existing ToolCatalog, RunContext enabled snapshot, ToolBroker, tool-call approval, worker resume, and RunRail/Settings visibility. The implementation keeps mutations Work-mode-only, path-scoped, sensitive-path-denied, bounded, audited, and text-only.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing Go stdlib filesystem APIs, existing productdata/runtime/httpapi services, existing React Settings/RunRail components

**Storage**: Existing run events and tool_calls projections; host workspace filesystem under configured root

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for visible catalog/timeline

**Target Platform**: Local macOS/Darwin development first with localhost-compatible API/web commands

**Project Type**: Go API/worker plus web/desktop-feeling shell

**Performance Goals**: Bounded text writes/edits complete during normal approved-tool worker execution; large content is rejected before mutation

**Constraints**: Work-mode only; approval required; single workspace root; UTF-8 text only; no shell; no browser; no web fetch/search; no overwrite-by-default; no host absolute root in persisted metadata

**Scale/Scope**: First mutation slice for small source/config/document files, not bulk refactors or artifact runtime

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature implements Loomi-owned workspace mutation vocabulary and does not copy Arkloop expression.
- **Runnable Vertical Slices**: PASS. US1 and US2 each define independent approval-gated write/edit smoke tests.
- **Core Flow Before Platform Complexity**: PASS. This builds on existing tool approval, worker, RunContext, catalog, and bounded loop foundations; it explicitly excludes shell/browser/multi-agent runtime.
- **Observable Agent Execution**: PASS. All mutations go through persisted tool lifecycle events and visible Settings/RunRail surfaces.
- **Safety, Permissions, and Data Boundaries**: PASS. Mutation is approval-gated, bounded, scoped, redacted, and denied for sensitive paths.

## Project Structure

### Documentation (this feature)

```text
specs/031-workspace-mutation-tools/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── workspace-mutation-tools.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/productdata/
├── models.go
├── builtin_personas.go
├── tool_catalog.go
├── tool_catalog_test.go
└── service_test.go

internal/runtime/
├── tools.go
├── tools_test.go
├── tool_broker.go
├── tool_broker_test.go
├── workspace_tools.go
├── workspace_tools_test.go
├── gateway.go
├── gateway_test.go
└── worker_test.go

internal/httpapi/
└── workspace_mutation_tools_smoke_test.go

web/src/
├── components/SettingsView.tools.test.tsx
├── components/RunRail.runtime.test.ts
├── domain.ts
├── mockData.ts
└── realApiClient.ts

docs-site/src/content/docs/
├── api/workspace-mutation-tools.md
├── architecture/workspace-mutation-tools.md
├── runbooks/local-m23-workspace-mutation-tools.md
├── devlog/2026-05-26-m23-workspace-mutation-tools.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend the existing M18/M21 tool runtime boundaries rather than adding a separate mutation subsystem.

## Complexity Tracking

No constitution violations.
