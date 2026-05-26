# Implementation Plan: M21 Workspace Read Tools

**Branch**: `017-mcp-approval-gated-execution` | **Date**: 2026-05-25 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/029-workspace-read-tools/spec.md`

## Summary

M21 adds the first bounded workspace read loop: `workspace.glob`, `workspace.grep`, and `workspace.read`. The implementation reuses the existing M18 catalog, ToolBroker, RunContext, provider tool-call request, approval, worker execution, result persistence, and continuation path. Workspace tools are executable only in Work mode persona scope, always approval-gated, rooted at `LOOMI_WORKSPACE_ROOT` or the Loomi repo root, deny traversal/symlink/sensitive paths, and return bounded safe text metadata.

## Technical Context

**Language/Version**: Go backend, TypeScript/React frontend, Starlight docs site.

**Primary Dependencies**: Existing `internal/productdata` tool catalog/run event/persona state, `internal/runtime` ToolBroker/Gateway/QueuedRunRouter, Go stdlib `filepath`, `fs`, `os`, `bufio`, `regexp`, and current Bun/Vite web stack.

**Storage**: Existing PostgreSQL and in-memory services for tool-call/run-event state; no new tables.

**Testing**: `go test ./...`, backend smoke/unit tests, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, `git diff --check`.

**Target Platform**: Local Loomi backend and web shell.

**Project Type**: Web application with Go API/worker backend and React frontend.

**Performance Goals**: Glob/grep/read complete on fixture roots within normal unit/smoke test time; grep/glob/read enforce bounded result sizes and do not scan unbounded output into memory.

**Constraints**: No shell/rg production implementation; no write/edit/shell/browser/web/artifact/sandbox tools; no host absolute path in UI; no sensitive content leakage in events or tool results.

**Scale/Scope**: Single configured workspace root per local backend process; one tool call per run remains the existing milestone boundary.

## Constitution Check

- **I. Mechanism Parity, Original Expression**: PASS. Arkloop is used only for mechanism study; Loomi keeps its own ToolCatalog, ToolBroker, RunContext, event, and UI language.
- **II. Runnable Vertical Slices**: PASS. Backend smoke covers successful glob/read/grep, denial until approval, boundary rejections, and continuation.
- **III. Core Flow Before Platform Complexity**: PASS. Adds only read-only workspace tools; explicitly defers shell, write, edit, sandbox, browser, web, artifact, and multi-tool loops.
- **IV. Observable Agent Execution**: PASS. Existing persisted tool-call events and timeline states remain the observable execution model.
- **V. Safety, Permissions, and Data Boundaries**: PASS. Workspace root, sensitive denylist, symlink boundary, approval gate, and redacted result metadata are required.

Post-design check: PASS. No constitution violations.

## Project Structure

### Documentation (this feature)

```text
specs/029-workspace-read-tools/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── workspace-read-tools.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/productdata/
├── models.go
├── tool_catalog.go
├── service.go
├── repository.go
├── builtin_personas.go
└── *_test.go

internal/runtime/
├── tools.go
├── workspace_tools.go
├── tool_broker.go
├── gateway.go
├── queued_runner.go
└── *_test.go

internal/httpapi/
└── *_smoke_test.go

web/src/
├── components/SettingsView.tsx
├── components/RunTimeline.tsx
├── domain.ts
└── *_test.tsx

docs-site/src/content/docs/
├── architecture/
├── api/
├── runbooks/
├── devlog/
├── roadmap/
└── spec-kit/
```

**Structure Decision**: Keep workspace tool execution in `internal/runtime` because it is a local runtime capability. Keep catalog metadata, persona allowlist, request validation, and event metadata in `internal/productdata` because those are durable product boundaries used by both in-memory and PostgreSQL services.

## Phase 0: Research

See [research.md](./research.md).

## Phase 1: Design & Contracts

See [data-model.md](./data-model.md), [contracts/workspace-read-tools.md](./contracts/workspace-read-tools.md), and [quickstart.md](./quickstart.md).

## Complexity Tracking

No constitution violations.
