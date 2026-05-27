# Implementation Plan: M18 Tool Runtime + Tool Catalog Foundation

**Branch**: `025-tool-runtime-catalog-foundation` | **Date**: 2026-05-25 | **Spec**: `specs/025-tool-runtime-catalog-foundation/spec.md`  
**Status**: Complete candidate

## Summary

M18 turns Loomi tools into first-class runtime objects without expanding the executable tool surface. It adds a safe catalog, a broker/executor envelope, a read-only API/UI surface, and routes existing `runtime.get_current_time` plus local stdio MCP execution through one broker path. The implementation reuses M7 approval projection, M9 RunContext, M10 persona allowlists, M11/M12 MCP discovery/execution, M15 deterministic smoke patterns, and M17 documentation discipline.

## Technical Context

**Language/Version**: Go backend; React + TypeScript frontend; Starlight docs with Bun  
**Primary Dependencies**: stdlib Go HTTP/testing, existing productdata/runtime/httpapi packages, existing React test stack  
**Storage**: no new database table in M18; catalog is computed from builtin definitions and safe MCP discovery/run metadata  
**Testing**: Go unit/smoke tests, web component/API-client tests, web build, docs build, diff check  
**Constraints**: no workspace/shell/sandbox/browser/web/artifact execution; no remote MCP/OAuth/plugin marketplace; no worker queue rewrite; no multi-tool loop  
**Scale/Scope**: first runtime/catalog slice only, covering current builtin and local stdio MCP tools

## Constitution Check

- **Runnable Vertical Slice**: API, broker execution, RunContext, and Settings > Tools have automated tests and local validation.
- **Core Flow Before Platform Complexity**: M18 reuses current worker/job/run/event boundaries and defers future tool categories.
- **Observable Agent Execution**: tool lifecycle events remain the source of truth.
- **Safety, Permissions, and Data Boundaries**: catalog and broker expose only redacted safe metadata; no raw MCP process data or secrets.

## Project Structure

```text
internal/productdata/
  models.go                  # catalog model, ToolResolution metadata
  service.go                 # computed catalog and RunContext resolution
  repository.go              # PG parity for catalog/run context metadata if needed

internal/runtime/
  tools.go                   # builtin definition as catalog executor
  mcp_tools.go               # MCP candidate catalog mapping
  tool_broker.go             # ToolExecutor, ToolInvocation, ToolResult, broker
  queued_runner.go           # approved tool resume routes through broker

internal/httpapi/
  tools.go                   # GET /v1/tools/catalog
  runtime_test.go            # catalog API and redaction tests
  *_smoke_test.go            # builtin + MCP broker smoke

web/src/
  domain.ts                  # ToolCatalogItem type
  realApiClient.ts           # map Tools API
  apiClient.ts               # client contract
  components/SettingsView.tsx
  components/settingsCatalog.ts
  components/ToolsPanel.tsx

docs-site/src/content/docs/
  architecture/tool-runtime-catalog.md
  api/tool-runtime-catalog.md
  runbooks/local-m18-tool-runtime-catalog.md
  devlog/2026-05-25-m18-tool-runtime-catalog.md
```

## Data Model

See `data-model.md`.

## Contracts

See:

- `contracts/tools-catalog-api.md`
- `contracts/tool-broker.md`
- `contracts/tool-events-redaction.md`
- `contracts/settings-tools-ui.md`

## Implementation Phases

1. Create Spec Kit artifacts and run analysis.
2. Add failing productdata/runtime tests for catalog and broker policy.
3. Implement catalog model, builtin/MCP catalog mapping, and broker envelopes.
4. Route approved builtin and MCP execution through broker from worker resume.
5. Add Tools API and HTTP smoke coverage.
6. Add Settings > Tools read-only UI and web tests.
7. Update docs, roadmap, workflow, and mark spec/tasks complete candidate.
8. Run validation commands.

## Risk Notes

- M18 must preserve existing M7/M12 event order so current frontend projection keeps working.
- MCP catalog is computed from discovery metadata and current configs; it must not read real local credential files.
- Broker failure messages must be specific enough for tests and safe enough for API/event exposure.

## Post-Design Status

No known artifact conflicts after initial analysis. M18 is a foundation slice; future M19+ tools must add their own catalog entries and executor implementations through the broker.
