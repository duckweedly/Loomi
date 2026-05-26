# Implementation Plan: M13 MCP Call Tool Bridge

**Branch**: `015-mcp-call-tool-bridge` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M13 adds `mcp.call_tool`, an approval-gated MCP-style bridge with a single built-in local server/tool pair: `local.echo`. It proves Loomi can route MCP-shaped calls through tool approval, worker execution, result summaries, catalog visibility, and timeline rendering without external process management.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React/Vite frontend; Bun frontend/docs.

**Primary Dependencies**: Existing runtime tool registry, productdata validation, worker execution, ToolCallCard, Settings catalog. No new dependency.

**Storage**: Existing `tool_calls.arguments_summary` and `tool_calls.result_summary`. No migration.

**Testing**: TDD required. Backend validation/runtime/worker tests before implementation; frontend ToolCallCard/settings tests before UI/catalog updates.

## Constitution Check

- **Runnable Vertical Slice**: PASS. Provider request -> approval -> worker execution -> timeline result.
- **Core Flow Before Platform Complexity**: PASS. Internal MCP bridge before external MCP server lifecycle.
- **Observable Agent Execution**: PASS. MCP calls are persisted as tool lifecycle events.
- **Safety/Data Boundaries**: PASS. Approval-required, allowlisted, no sockets/processes/secrets.
- **Documentation Definition of Done**: PASS.

## File Targets

- Backend: `internal/productdata/models.go`, `internal/productdata/service_test.go`, `internal/runtime/tools.go`, `internal/runtime/tools_test.go`, `internal/runtime/worker.go`, `internal/runtime/worker_test.go`, `internal/httpapi/tools_test.go`
- Frontend: `web/src/mockApiClient.ts`, `web/src/components/ToolCallCard.test.tsx`
- Docs: `docs-site/src/content/docs/architecture/mcp-call-tool-bridge.md`, `docs-site/src/content/docs/api/mcp-call-tool-bridge.md`, `docs-site/src/content/docs/runbooks/local-m13.md`, `docs-site/src/content/docs/devlog/2026-05-26-m13-mcp-call-tool-bridge.md`, roadmap/workflow updates
