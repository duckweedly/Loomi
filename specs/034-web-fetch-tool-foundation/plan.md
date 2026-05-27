# Implementation Plan: M26 Web Fetch Tool Foundation

**Branch**: `[034-web-fetch-tool-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/034-web-fetch-tool-foundation/spec.md`

## Summary

M26 adds one read-only network tool, `web.fetch`, routed through the existing ToolCatalog, RunContext, ToolBroker, worker resume, run-event, Settings, and RunRail boundaries. It proves the web runtime boundary without browser automation, JavaScript rendering, crawler behavior, authenticated sessions, or artifact runtime.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata/runtime/httpapi services, existing ToolBroker and worker resume path, React Settings/RunRail components, Go stdlib `net/http`

**Storage**: Existing run events and tool_calls projections; no new database tables

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for Settings Tools and RunRail web lifecycle

**Target Platform**: Local macOS/Darwin development first with deterministic local tests

**Project Type**: Go API/worker plus web/desktop-feeling shell

**Performance Goals**: One bounded fetch per tool call; default timeout and byte limits keep requests short and stored excerpts small

**Constraints**: Chat/Work persona-gated, auto-approved public read, HTTP(S) only, no credentials, no private/local hosts, redirect validation, no cookies/auth/session/browser reuse, no raw response body persistence

**Scale/Scope**: One local user, single explicit URL fetch per call

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature uses Loomi-owned web tool vocabulary and does not copy another product's expression layer.
- **Runnable Vertical Slices**: PASS. The slice has executable backend request -> fetch -> continuation evidence and visible Settings/RunRail states.
- **Core Flow Before Platform Complexity**: PASS. This follows workspace/mutation/sandbox/MCP/LSP foundations and deliberately defers browser/search/artifact runtimes.
- **Observable Agent Execution**: PASS. Web fetch requests/results are persisted through existing tool lifecycle events and visible in RunRail.
- **Safety, Permissions, and Data Boundaries**: PASS. Network reads are auto-approved bounded public reads and reject private/local/credentialed URLs.

## Project Structure

### Documentation (this feature)

```text
specs/034-web-fetch-tool-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── web-fetch-tool.md
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
├── web_tools.go
├── web_tools_test.go
├── tool_broker.go
├── tool_broker_test.go
├── tools.go
├── queued_runner.go
└── worker_test.go

internal/httpapi/
└── web_fetch_tool_smoke_test.go

web/src/
├── components/SettingsView.tools.test.tsx
├── components/RunRail.runtime.test.ts
├── components/RunRail.tsx
├── domain.ts
├── mockApiClient.ts
└── mockData.ts

docs-site/src/content/docs/
├── architecture/web-fetch-tool.md
├── api/web-fetch-tool.md
├── runbooks/local-m26-web-fetch-tool.md
├── devlog/2026-05-26-m26-web-fetch-tool.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend the existing tool runtime rather than creating a separate web runtime service. Browser/search/artifact split stays as later features.

## Complexity Tracking

No constitution violations.
