# Implementation Plan: M27 Browser Automation Foundation

**Branch**: `[035-browser-automation-foundation]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/035-browser-automation-foundation/spec.md`

## Summary

M27 adds a stateful browser automation foundation through three builtin tools: `browser.open`, `browser.snapshot`, and `browser.click_link`. The slice reuses the existing ToolCatalog, RunContext, ToolBroker, approval, worker continuation, run-event, Settings, and RunRail boundaries. It intentionally avoids a real Chrome profile, cookies, JavaScript rendering, screenshots, form submission, downloads, authenticated browsing, artifact runtime, and multi-agent orchestration.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata/runtime/httpapi services, ToolBroker, worker approval resume path, React Settings/RunRail components, Go stdlib `net/http`

**Storage**: Existing run events and tool_calls projections; browser session state is run-scoped in process for the first foundation slice

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for Settings Tools and RunRail browser lifecycle

**Target Platform**: Local macOS/Darwin development first with deterministic local tests

**Project Type**: Go API/worker plus web/desktop-feeling shell

**Performance Goals**: One bounded navigation per approved open/click; snapshot reads current session state without network

**Constraints**: Work-mode only, approval required, public HTTP(S) only, no credentials, no private/local hosts, redirect/link validation, bounded response bytes, bounded links/text, no cookies/profile/JS/screenshots/forms/downloads

**Scale/Scope**: One local user, run-scoped page sessions, explicit approved navigation only

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature uses Loomi-owned browser runtime vocabulary and avoids copying another product's expression layer.
- **Runnable Vertical Slices**: PASS. The slice has executable open/snapshot/click backend evidence and visible Settings/RunRail states.
- **Core Flow Before Platform Complexity**: PASS. It follows workspace, sandbox, MCP, LSP, and web fetch foundations and defers full browser/profile/artifact runtimes.
- **Observable Agent Execution**: PASS. Browser requests/results are persisted through existing tool lifecycle events and visible in RunRail.
- **Safety, Permissions, and Data Boundaries**: PASS. Browser navigation is approval-gated, bounded, Work-mode only, and rejects private/local/credentialed targets.

## Project Structure

### Documentation (this feature)

```text
specs/035-browser-automation-foundation/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── browser-automation.md
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
├── browser_tools.go
├── browser_tools_test.go
├── tool_broker.go
├── tool_broker_test.go
├── tools.go
├── queued_runner.go
└── worker_test.go

internal/httpapi/
└── browser_automation_smoke_test.go

web/src/
├── components/SettingsView.tools.test.tsx
├── components/RunRail.runtime.test.ts
├── components/RunRail.tsx
├── mockApiClient.ts
└── mockData.ts

docs-site/src/content/docs/
├── architecture/browser-automation-foundation.md
├── api/browser-automation-foundation.md
├── runbooks/local-m27-browser-automation.md
├── devlog/2026-05-26-m27-browser-automation.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend the existing tool runtime rather than introducing a browser service or external browser dependency in the first slice.

## Complexity Tracking

No constitution violations.
