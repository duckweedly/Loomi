# Implementation Plan: M25 MCP Management + LSP Read-only Foundation

**Branch**: `[033-mcp-management-lsp-readonly]` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/033-mcp-management-lsp-readonly/spec.md`

## Summary

M25 productizes two adjacent Arkloop parity surfaces without broadening execution risk: a read-only Settings > MCP management surface over existing local stdio MCP config/discovery state, and first read-only LSP tools (`lsp.diagnostics`, `lsp.symbols`, `lsp.references`) routed through the same ToolCatalog, RunContext, ToolBroker, approval, worker, run-event, Settings, and RunRail boundaries used by M21-M24.

The slice intentionally avoids writable MCP configuration, remote MCP/OAuth, marketplace install, full language-server process lifecycle, shell/package-manager diagnostics, browser/web/artifact runtime, activity recording, and multi-agent orchestration.

## Technical Context

**Language/Version**: Go backend; TypeScript/React frontend; Astro/Starlight docs

**Primary Dependencies**: Existing productdata/runtime/httpapi services, existing local MCP config/discovery code, existing workspace scope guard, existing React Settings/RunRail components

**Storage**: Existing run events and tool_calls projections; no new database tables for this slice

**Testing**: `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run --cwd docs-site build`, browser smoke for Settings MCP/Tools and RunRail visibility

**Target Platform**: Local macOS/Darwin development first with localhost-compatible API/web commands

**Project Type**: Go API/worker plus web/desktop-feeling shell

**Performance Goals**: MCP status lists are bounded by configured server count and existing discovery metadata; LSP scans are bounded by file size/result limits and must not traverse unbounded workspace content

**Constraints**: MCP management read-only; LSP Work-mode only; LSP approval required; workspace-relative paths; no shell/network/language-server process in this slice; safe metadata only

**Scale/Scope**: One local user, local stdio MCP visibility, bounded read-only code-intelligence queries

## Constitution Check

- **Mechanism Parity, Original Expression**: PASS. The feature uses Loomi-owned MCP/LSP vocabulary and avoids copying another product's expression layer.
- **Runnable Vertical Slices**: PASS. US1 has a visible Settings MCP surface; US2 has an approve -> execute -> continue LSP smoke; US3 covers audit UI.
- **Core Flow Before Platform Complexity**: PASS. This builds on existing catalog, approval, worker, RunContext, MCP, workspace, and sandbox foundations; remote MCP/browser/artifact/multi-agent remain excluded.
- **Observable Agent Execution**: PASS. LSP tool calls go through persisted tool lifecycle events and visible RunRail rows.
- **Safety, Permissions, and Data Boundaries**: PASS. MCP config is redacted/read-only; LSP is approval-gated, workspace-scoped, bounded, and read-only.

## Project Structure

### Documentation (this feature)

```text
specs/033-mcp-management-lsp-readonly/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── mcp-lsp-readonly.md
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
├── lsp_tools.go
├── lsp_tools_test.go
├── gateway.go
├── gateway_test.go
└── worker_test.go

internal/httpapi/
├── mcp_management_test.go
└── lsp_readonly_smoke_test.go

web/src/
├── components/SettingsView.mcp.test.tsx
├── components/SettingsView.tools.test.tsx
├── components/RunRail.runtime.test.ts
├── components/SettingsView.tsx
├── domain.ts
├── realApiClient.ts
├── mockApiClient.ts
└── mockData.ts

docs-site/src/content/docs/
├── architecture/mcp-management-lsp-readonly.md
├── api/mcp-management-lsp-readonly.md
├── runbooks/local-m25-mcp-lsp-readonly.md
├── devlog/2026-05-26-m25-mcp-lsp-readonly.md
├── roadmap/current-status.md
└── spec-kit/workflow.md
```

**Structure Decision**: Extend M18/M21-M24 tool runtime and M11/M12 MCP foundations instead of creating a separate MCP admin service or LSP daemon.

## Complexity Tracking

No constitution violations.
