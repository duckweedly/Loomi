# Feature Specification: M13 MCP Call Tool Bridge

**Feature Branch**: `015-mcp-call-tool-bridge`

**Created**: 2026-05-26

**Status**: Draft

**Input**: Continue Arkloop code-agent coverage by adding the first approval-gated MCP call path after workspace tools and todo planning.

## User Scenarios & Testing

### User Story 1 - Execute a bounded MCP-style tool call (Priority: P1)

A model can request `mcp.call_tool` for an allowlisted local MCP-style server/tool pair. The request goes through the same approval flow as runtime and workspace tools, then the worker records a structured result.

**Why this priority**: Arkloop coverage needs MCP semantics, but Loomi should first prove the broker/lifecycle boundary before introducing external MCP processes.

**Independent Test**: Record, approve, and execute `mcp.call_tool` with `server: "local"` and `tool: "echo"`; verify the tool-call lifecycle succeeds and stores a redacted structured result.

**Acceptance Scenarios**:

1. **Given** a provider requests `mcp.call_tool` for `local.echo`, **When** the user approves it, **Then** the worker records a succeeded result with server, tool, and echoed message.
2. **Given** a request targets an unknown server or tool, **When** it is recorded or executed, **Then** Loomi rejects it as invalid/unsupported and does not execute arbitrary external code.
3. **Given** Settings > Tools is opened, **When** the catalog loads, **Then** `mcp.call_tool` appears as an approval-required MCP bridge with medium risk.

## Requirements

- **FR-001**: Loomi MUST support `mcp.call_tool` as an allowlisted tool.
- **FR-002**: The tool MUST require approval before execution.
- **FR-003**: The first supported pair MUST be `server: "local"` and `tool: "echo"`.
- **FR-004**: Arguments MUST include `server`, `tool`, and optional `arguments`.
- **FR-005**: `local.echo` MUST accept `arguments.message` as a string up to 500 characters and return a bounded result summary.
- **FR-006**: Unknown servers, unknown tools, invalid argument shapes, and secret-looking messages MUST be rejected or redacted before persistence.
- **FR-007**: The catalog MUST list `mcp.call_tool` with group `mcp`, capability `call_tool`, risk `medium`, and approval required.
- **FR-008**: The frontend MUST render MCP lifecycle/results through the existing ToolCallCard path and Settings catalog.
- **FR-009**: The feature MUST NOT start external MCP processes, open sockets, execute arbitrary tool names, install MCP servers, or add multi-agent delegation.

## Success Criteria

- **SC-001**: Backend tests prove request validation, catalog metadata, and approved worker execution for `mcp.call_tool`.
- **SC-002**: Frontend tests prove MCP result summaries render without raw object output.
- **SC-003**: Local validation passes `go test ./...`, `bun test --cwd web`, `bun run --cwd web build`, `bun run build` from `docs-site/`, and `git diff --check`.

## Assumptions

- M7-M12 approval, worker, catalog, and Settings surfaces exist.
- M13 is a lifecycle bridge, not full MCP host management.
