# Implementation Plan: M12 Todo Write Planning Tool

**Branch**: `014-todo-write-planning-tool` | **Date**: 2026-05-26 | **Spec**: [spec.md](./spec.md)

## Summary

M12 adds `runtime.todo_write`, an approval-gated planning tool that lets model runs publish a bounded structured todo list through existing tool-call events and result summaries.

## Technical Context

**Language/Version**: Go 1.23 backend; TypeScript/React/Vite frontend; Bun frontend/docs.

**Primary Dependencies**: Existing runtime tool registry, productdata validation, worker approved execution, frontend ToolCallCard, Settings tool catalog. No new dependency.

**Storage**: Existing `tool_calls.arguments_summary` and `tool_calls.result_summary`. No migration.

**Testing**: TDD required. Backend unit/service/worker tests before implementation; frontend ToolCallCard/settings catalog tests before UI/catalog updates.

## Constitution Check

- **Runnable Vertical Slice**: PASS. Provider request -> approval -> worker execution -> timeline result.
- **Core Flow Before Platform Complexity**: PASS. Planning tool before MCP/spawn/LSP.
- **Observable Agent Execution**: PASS. Todo state is visible through persisted tool lifecycle events.
- **Safety/Data Boundaries**: PASS. Bounded, approval-required, no secrets, no file/process side effects.
- **Documentation Definition of Done**: PASS.

## File Targets

- Backend: `internal/productdata/models.go`, `internal/productdata/service_test.go`, `internal/runtime/tools.go`, `internal/runtime/tools_test.go`, `internal/runtime/worker.go`, `internal/runtime/worker_test.go`
- Frontend: `web/src/mockApiClient.ts`, `web/src/components/ToolCallCard.test.tsx`
- Docs: `docs-site/src/content/docs/architecture/todo-write-planning-tool.md`, `docs-site/src/content/docs/api/todo-write-planning-tool.md`, `docs-site/src/content/docs/runbooks/local-m12.md`, `docs-site/src/content/docs/devlog/2026-05-26-m12-todo-write-planning-tool.md`, roadmap/workflow updates
