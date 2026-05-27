# Implementation Plan: M28 Artifact Runtime Foundation

**Branch**: `[036-artifact-runtime-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M28 adds a non-executable artifact runtime foundation through three builtin tools: `artifact.create_text`, `artifact.read`, and `artifact.list`. The slice reuses ToolCatalog, RunContext, approval, ToolBroker, worker continuation, existing WorkPlan artifact projection, Settings, and RunRail. It intentionally avoids binary artifacts, previews, downloads, filesystem export, browser/runtime execution, and multi-agent orchestration.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata service interfaces, PostgresRepository, ToolBroker, worker approval resume path, WorkPlan artifact projection, React Settings/RunRail components

**Storage**: Productdata artifact records backed by both in-memory service and PostgreSQL `artifacts` table, with safe run-event summaries

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for Settings Tools and RunRail artifact lifecycle

**Constraints**: Work-mode only, approval required, bounded UTF-8 text, no execution/render/network/filesystem side effects, safe summaries only

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. Uses Loomi-owned artifact vocabulary.
- **Runnable Vertical Slices**: PASS. The slice has create/read/list backend proof and visible frontend state.
- **Core Flow Before Platform Complexity**: PASS. Adds storage-only artifact behavior before execution or rendering.
- **Observable Agent Execution**: PASS. Artifact requests/results go through run events and RunRail.
- **Safety, Permissions, and Data Boundaries**: PASS. Artifact tools are approval-gated, bounded, Work-mode only, and non-executable.

## Project Structure

```text
specs/036-artifact-runtime-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── artifact-runtime.md
└── tasks.md
```

Source changes target:

```text
internal/productdata/
internal/runtime/
internal/httpapi/
web/src/components/
web/src/runtime/
docs-site/src/content/docs/
```

## Complexity Tracking

No constitution violations.
