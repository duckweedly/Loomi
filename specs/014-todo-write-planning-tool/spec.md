# Feature Specification: M12 Todo Write Planning Tool

**Feature Branch**: `014-todo-write-planning-tool`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue Arkloop code-agent coverage by adding a planning/todo tool after the workspace read/write/exec loop.

## User Scenarios & Testing

### User Story 1 - Publish an auditable plan (Priority: P1)

A model can request `runtime.todo_write` with a bounded list of plan items. The request is visible in the existing tool-call approval flow, and after approval the worker records a structured result that appears in the run timeline.

**Why this priority**: Arkloop-style code agents need a planning surface before higher-level delegation. The plan must be observable and auditable through Loomi's existing run/event model.

**Independent Test**: Record, approve, and execute `runtime.todo_write`; verify the tool-call lifecycle reaches succeeded and stores counts plus sanitized item summaries.

**Acceptance Scenarios**:

1. **Given** a provider requests `runtime.todo_write` with valid todo items, **When** the user approves it, **Then** the worker records a succeeded tool-call result with total/pending/in-progress/completed counts.
2. **Given** a todo item is missing a title, has an invalid status, or the list exceeds the bound, **When** the request is recorded, **Then** Loomi rejects it as invalid before approval.
3. **Given** Settings > Tools is opened, **When** the catalog loads, **Then** `runtime.todo_write` appears as a low-risk runtime planning tool.

## Requirements

- **FR-001**: Loomi MUST support `runtime.todo_write` as an allowlisted tool.
- **FR-002**: The tool MUST require the existing approval flow before execution.
- **FR-003**: Tool arguments MUST be `items`, an array of 1-20 objects with `title` and optional `status`.
- **FR-004**: `status` MUST be one of `pending`, `in_progress`, or `completed`; missing status defaults to `pending`.
- **FR-005**: Item titles MUST be trimmed, non-empty, and no longer than 160 characters.
- **FR-006**: Execution MUST return structured counts and sanitized item summaries through `result_summary`.
- **FR-007**: The catalog MUST list `runtime.todo_write` with capability `plan`, risk `low`, side effect `none`, and no secrets.
- **FR-008**: The frontend MUST render todo_write lifecycle/results through the existing timeline/ToolCallCard path.
- **FR-009**: The feature MUST NOT add editable todo state, a separate todo database table, multi-agent delegation, MCP, LSP, or auto-approval.

## Success Criteria

- **SC-001**: Backend tests prove request validation, catalog metadata, and approved worker execution for `runtime.todo_write`.
- **SC-002**: Frontend tests prove todo_write result summaries render without custom execution controls.
- **SC-003**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run build` from `docs-site/`, and `git diff --check`.

## Assumptions

- M7-M11 approval, worker, catalog, and Settings surfaces exist.
- Todo planning is a runtime visibility tool, not a durable task manager.
